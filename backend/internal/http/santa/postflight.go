package santa

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/store"
	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

type postflightHandler struct {
	store  *store.Store
	logger *slog.Logger
}

func (h *postflightHandler) Handle(w http.ResponseWriter, r *http.Request) {
	machineIdentifier := strings.TrimSpace(chi.URLParam(r, "machineID"))
	if machineIdentifier == "" {
		respondError(w, http.StatusBadRequest, "machine id required")
		return
	}
	body, err := decodeBody(r)
	if err != nil {
		h.logger.Error("decode postflight body", "err", err)
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	defer func() {
		if closeErr := body.Close(); closeErr != nil {
			h.logger.Warn("failed to close request body", "err", closeErr)
		}
	}()
	bodyBytes, err := io.ReadAll(io.LimitReader(body, 1<<20))
	if err != nil {
		h.logger.Error("read postflight payload", "err", err)
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	var req syncv1.PostflightRequest
	if err := unmarshalProtoJSON(bodyBytes, &req); err != nil {
		h.logger.Error("decode postflight payload", "err", err)
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
	if _, err := h.store.UpdateMachinePostflightState(ctx, sqlc.UpdateMachinePostflightStateParams{
		ID:                 machine.ID,
		LastRulesReceived:  uint32ToInt4(req.GetRulesReceived()),
		LastRulesProcessed: uint32ToInt4(req.GetRulesProcessed()),
		CleanSyncRequested: false,
	}); err != nil {
		h.logger.Error("update postflight state", "err", err)
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}
	if _, err := h.store.TouchMachine(ctx, machine.ID, machine.ClientVersion.String, machine.SyncCursor.String, machine.RuleCursor.String); err != nil {
		h.logger.Warn("touch machine", "err", err)
	}
	respondProtoJSON(w, http.StatusOK, &syncv1.PostflightResponse{})
}

func uint32ToInt4(value uint32) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(value), Valid: true}
}
