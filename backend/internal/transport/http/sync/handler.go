// Package sync provides HTTP handlers for the Santa sync protocol.
package sync

import (
	"context"
	"net/http"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/domain/errx"
	"github.com/woodleighschool/grinch/internal/domain/santa"
	"github.com/woodleighschool/grinch/internal/logging"
)

// Handler serves Santa sync protocol endpoints.
type Handler struct {
	svc santa.SyncService
}

// NewHandler constructs a Handler.
func NewHandler(svc santa.SyncService) *Handler {
	return &Handler{svc: svc}
}

// Preflight handles POST /preflight/{machine_id}.
func (h *Handler) Preflight(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logging.FromContext(ctx).With("op", "preflight")

	machineID, err := parseMachineID(r)
	if err != nil {
		writeError(ctx, w, log, err, "invalid machine id")
		return
	}
	log = log.With("machine_id", machineID)

	req := &syncv1.PreflightRequest{}
	if err = decodeRequest(r, req); err != nil {
		writeError(ctx, w, log, err, "decode failed")
		return
	}

	resp, err := h.svc.Preflight(ctx, machineID, req)
	if err != nil {
		writeError(ctx, w, log, err, "preflight failed")
		return
	}

	if err = encodeResponse(w, r, resp); err != nil {
		log.Error("encode failed", "error", err)
	}
}

// EventUpload handles POST /eventupload/{machine_id}.
func (h *Handler) EventUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logging.FromContext(ctx).With("op", "eventupload")

	machineID, err := parseMachineID(r)
	if err != nil {
		writeError(ctx, w, log, err, "invalid machine id")
		return
	}
	log = log.With("machine_id", machineID)

	req := &syncv1.EventUploadRequest{}
	if err = decodeRequest(r, req); err != nil {
		writeError(ctx, w, log, err, "decode failed")
		return
	}

	resp, err := h.svc.EventUpload(ctx, machineID, req)
	if err != nil {
		writeError(ctx, w, log, err, "eventupload failed")
		return
	}

	if err = encodeResponse(w, r, resp); err != nil {
		log.Error("encode failed", "error", err)
	}
}

// RuleDownload handles POST /ruledownload/{machine_id}.
func (h *Handler) RuleDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logging.FromContext(ctx).With("op", "ruledownload")

	machineID, err := parseMachineID(r)
	if err != nil {
		writeError(ctx, w, log, err, "invalid machine id")
		return
	}
	log = log.With("machine_id", machineID)

	req := &syncv1.RuleDownloadRequest{}
	if err = decodeRequest(r, req); err != nil {
		writeError(ctx, w, log, err, "decode failed")
		return
	}

	resp, err := h.svc.RuleDownload(ctx, machineID, req)
	if err != nil {
		writeError(ctx, w, log, err, "ruledownload failed")
		return
	}

	if err = encodeResponse(w, r, resp); err != nil {
		log.Error("encode failed", "error", err)
	}
}

// Postflight handles POST /postflight/{machine_id}.
func (h *Handler) Postflight(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logging.FromContext(ctx).With("op", "postflight")

	machineID, err := parseMachineID(r)
	if err != nil {
		writeError(ctx, w, log, err, "invalid machine id")
		return
	}
	log = log.With("machine_id", machineID)

	req := &syncv1.PostflightRequest{}
	if err = decodeRequest(r, req); err != nil {
		writeError(ctx, w, log, err, "decode failed")
		return
	}

	resp, err := h.svc.Postflight(ctx, machineID, req)
	if err != nil {
		writeError(ctx, w, log, err, "postflight failed")
		return
	}

	if err = encodeResponse(w, r, resp); err != nil {
		log.Error("encode failed", "error", err)
	}
}

func parseMachineID(r *http.Request) (uuid.UUID, error) {
	raw := chi.URLParam(r, "machine_id")
	if raw == "" {
		return uuid.Nil, errx.NotFound("machine_id required")
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errx.NotFound("invalid machine_id")
	}

	return id, nil
}

func writeError(ctx context.Context, w http.ResponseWriter, log logging.Logger, err error, msg string) {
	status := errx.Status(err)

	if status >= http.StatusInternalServerError {
		log.ErrorContext(ctx, msg, "error", err)
	} else {
		log.WarnContext(ctx, msg, "error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{}`))
}
