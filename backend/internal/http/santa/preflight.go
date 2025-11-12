package santa

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/proto"

	"github.com/woodleighschool/grinch/internal/store"
	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

const (
	defaultFullSyncInterval = 600
	defaultEventBatchSize   = 100
)

type preflightHandler struct {
	store  *store.Store
	logger *slog.Logger
}

func (h *preflightHandler) Handle(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			h.logger.Warn("failed to close request body", "err", closeErr)
		}
	}()
	machineIdentifier := strings.TrimSpace(chi.URLParam(r, "machineID"))
	if machineIdentifier == "" {
		respondError(w, http.StatusBadRequest, "machine id required")
		return
	}
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		h.logger.Error("read preflight payload", "err", err)
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	var req syncv1.PreflightRequest
	if err := unmarshalProtoJSON(bodyBytes, &req); err != nil {
		h.logger.Error("decode preflight payload", "err", err)
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	serialNumber := strings.TrimSpace(req.GetSerialNumber())
	if serialNumber == "" {
		respondError(w, http.StatusBadRequest, "serial_num required")
		return
	}
	primaryUser := strings.TrimSpace(req.GetPrimaryUser())
	if primaryUser == "" {
		respondError(w, http.StatusBadRequest, "primary_user required")
		return
	}
	clientMode := normalizeClientMode(req.GetClientMode())
	ctx := r.Context()
	userID := resolveUserID(ctx, h.store, h.logger, primaryUser)
	machine, err := h.store.UpsertMachine(ctx, sqlc.UpsertMachineParams{
		ID:                uuid.New(),
		MachineIdentifier: machineIdentifier,
		Serial:            serialNumber,
		Hostname:          coalesce(req.GetHostname(), "unknown"),
		UserID:            uuidPtrToPgtype(userID),
		ClientVersion:     textPtr(req.GetSantaVersion()),
	})
	if err != nil {
		h.logger.Error("upsert machine", "err", err)
		respondError(w, http.StatusInternalServerError, "upsert failed")
		return
	}
	cleanSyncRequested := machine.CleanSyncRequested || req.GetRequestCleanSync()
	machine, err = h.store.UpdateMachinePreflightState(ctx, sqlc.UpdateMachinePreflightStateParams{
		ID:                   machine.ID,
		PrimaryUser:          textPtr(primaryUser),
		ClientMode:           strings.ToLower(clientMode.String()),
		CleanSyncRequested:   cleanSyncRequested,
		LastPreflightPayload: bodyBytes,
	})
	if err != nil {
		h.logger.Error("update machine preflight", "err", err)
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}
	syncType := syncv1.SyncType_NORMAL
	if req.GetRequestCleanSync() || machine.CleanSyncRequested {
		syncType = syncv1.SyncType_CLEAN
	}
	overrideAction := syncv1.FileAccessAction_NONE
	resp := &syncv1.PreflightResponse{
		ClientMode:               clientMode,
		SyncType:                 syncType.Enum(),
		BatchSize:                defaultEventBatchSize,
		FullSyncIntervalSeconds:  defaultFullSyncInterval,
		OverrideFileAccessAction: overrideAction.Enum(),
		EnableBundles:            proto.Bool(true),
		EnableTransitiveRules:    proto.Bool(false),
		DeprecatedBundlesEnabled: proto.Bool(true),
		BlockUsbMount:            proto.Bool(false),
	}
	if _, err := h.store.TouchMachine(ctx, machine.ID, machine.ClientVersion.String, machine.SyncCursor.String, machine.RuleCursor.String); err != nil {
		h.logger.Warn("touch machine", "err", err)
	}
	respondProtoJSON(w, http.StatusOK, resp)
}

func normalizeClientMode(mode syncv1.ClientMode) syncv1.ClientMode {
	switch mode {
	case syncv1.ClientMode_MONITOR,
		syncv1.ClientMode_LOCKDOWN,
		syncv1.ClientMode_STANDALONE:
		return mode
	default:
		return syncv1.ClientMode_MONITOR
	}
}

func textPtr(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func coalesce(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func uuidPtrToPgtype(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}
