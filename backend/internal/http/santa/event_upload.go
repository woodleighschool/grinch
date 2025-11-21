package santa

import (
	"errors"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/store"
	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

type eventUploadHandler struct {
	store  *store.Store
	logger *slog.Logger
}

func (h *eventUploadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	machineIdentifier := strings.TrimSpace(chi.URLParam(r, "machineID"))
	if machineIdentifier == "" {
		respondError(w, http.StatusBadRequest, "machine id required")
		return
	}
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
	if err := unmarshalProtoJSON(bodyBytes, &req); err != nil {
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
		payload, err := marshalProtoJSON(evt)
		if err != nil {
			h.logger.Warn("skip event marshal", "machine", machineIdentifier, "index", idx, "err", err)
			continue
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
			Payload:    payload,
			OccurredAt: occurred,
		}); err != nil {
			h.logger.Error("insert event", "err", err)
			respondError(w, http.StatusInternalServerError, "persist failed")
			return
		}
	}
	if _, err := h.store.TouchMachine(ctx, machine.ID, machine.ClientVersion.String, machine.SyncCursor.String, machine.RuleCursor.String); err != nil {
		h.logger.Warn("touch machine", "err", err)
	}
	respondProtoJSON(w, http.StatusOK, &syncv1.EventUploadResponse{EventUploadBundleBinaries: []string{}})
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
