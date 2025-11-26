package santa

import (
	"errors"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"encoding/json"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/store"
	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

// eventUploadHandler ingests Santa events + file metadata.
type eventUploadHandler struct {
	store  *store.Store
	logger *slog.Logger
}

// Handle ingests an EventUploadRequest and persists compacted event rows.
func (h *eventUploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	machineIdentifier := strings.TrimSpace(chi.URLParam(r, "machineID"))
	if machineIdentifier == "" {
		respondError(w, http.StatusBadRequest, "machine id required")
		return
	}
	format := requestWireFormat(r)
	body, err := decodeBody(r)
	if err != nil {
		h.logger.Error("decode event body", "err", err)
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	defer func() {
		if closeErr := body.Close(); closeErr != nil {
			h.logger.Warn("failed to close request body", "err", closeErr)
		}
	}()
	bodyBytes, err := io.ReadAll(io.LimitReader(body, 8<<20))
	if err != nil {
		h.logger.Error("read event payload", "err", err)
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	var req syncv1.EventUploadRequest
	if err := decodeWireMessage(format, bodyBytes, &req); err != nil {
		h.logger.Error("decode event payload", "err", err)
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	ctx := r.Context()
	machine, err := h.store.GetMachineByIdentifier(ctx, machineIdentifier)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "machine not registered")
			return
		}
		h.logger.Error("fetch machine", "err", err)
		respondError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	for idx, evt := range req.GetEvents() {
		if evt == nil {
			continue
		}
		fullPayload, err := marshalProtoJSON(evt)
		if err != nil {
			h.logger.Warn("skip event marshal", "machine", machineIdentifier, "index", idx, "err", err)
			continue
		}

		fileMeta, strippedPayload, err := splitEventPayload(evt, fullPayload)
		if err != nil {
			h.logger.Warn("skip event payload", "machine", machineIdentifier, "index", idx, "err", err)
			continue
		}

		if fileMeta.SHA256 != "" {
			if err := h.store.UpsertFile(ctx, sqlc.UpsertFileParams{
				Sha256:       fileMeta.SHA256,
				Name:         fileMeta.Name,
				SigningID:    pgtype.Text{String: fileMeta.SigningID, Valid: fileMeta.SigningID != ""},
				Cdhash:       pgtype.Text{String: fileMeta.CDHash, Valid: fileMeta.CDHash != ""},
				SigningChain: fileMeta.SigningChain,
				Entitlements: fileMeta.Entitlements,
			}); err != nil {
				h.logger.Error("upsert file", "sha", fileMeta.SHA256, "err", err)
			}
		}
		if len(strippedPayload) == 0 {
			strippedPayload = []byte("{}")
		}

		userID := resolveUserID(ctx, h.store, h.logger, evt.GetExecutingUser())
		var occurred pgtype.Timestamptz
		if ts := timestampFromFloat(evt.GetExecutionTime()); !ts.IsZero() {
			occurred = pgtype.Timestamptz{Time: ts, Valid: true}
		}
		if _, err := h.store.InsertEvent(ctx, sqlc.InsertEventParams{
			MachineID:  machine.ID,
			UserID:     uuidPtrToPgtype(userID),
			Kind:       evt.GetDecision().String(),
			Payload:    strippedPayload,
			OccurredAt: occurred,
			FileSha256: pgtype.Text{String: fileMeta.SHA256, Valid: fileMeta.SHA256 != ""},
		}); err != nil {
			h.logger.Error("insert event", "err", err)
			respondError(w, http.StatusInternalServerError, "persist failed")
			return
		}
	}
	if _, err := h.store.TouchMachine(ctx, machine.ID, machine.ClientVersion.String, machine.SyncCursor.String, machine.RuleCursor.String); err != nil {
		h.logger.Warn("touch machine", "err", err)
	}
	respondProto(w, r, http.StatusOK, &syncv1.EventUploadResponse{EventUploadBundleBinaries: []string{}})
}

func timestampFromFloat(value float64) time.Time {
	if value == 0 {
		return time.Time{}
	}
	integer, frac := math.Modf(value)
	sec := int64(integer)
	nsec := int64(frac * float64(time.Second))
	return time.Unix(sec, nsec).UTC()
}

type eventFileMetadata struct {
	SHA256       string
	Name         string
	SigningID    string
	CDHash       string
	SigningChain []byte
	Entitlements []byte
}

func splitEventPayload(evt *syncv1.Event, payload []byte) (eventFileMetadata, []byte, error) {
	meta := eventFileMetadata{
		SHA256:    evt.GetFileSha256(),
		Name:      evt.GetFileName(),
		SigningID: evt.GetSigningId(),
		CDHash:    evt.GetCdhash(),
	}
	if len(payload) == 0 {
		return meta, []byte("{}"), nil
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return eventFileMetadata{}, nil, err
	}
	if v, ok := raw["signing_chain"]; ok {
		meta.SigningChain = cloneRawJSON(v)
	}
	if v, ok := raw["entitlement_info"]; ok {
		meta.Entitlements = cloneRawJSON(v)
	}
	for _, key := range []string{"signing_chain", "entitlement_info"} {
		delete(raw, key)
	}
	var stripped []byte
	if len(raw) == 0 {
		stripped = []byte("{}")
	} else {
		var err error
		stripped, err = json.Marshal(raw)
		if err != nil {
			return eventFileMetadata{}, nil, err
		}
	}
	return meta, stripped, nil
}

func cloneRawJSON(msg json.RawMessage) []byte {
	if len(msg) == 0 {
		return nil
	}
	dup := make([]byte, len(msg))
	copy(dup, msg)
	return dup
}
