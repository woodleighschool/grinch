//nolint:dupl // entry is similar to postflight but handles a different request type.
package handlers

import (
	"net/http"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"

	httpstatus "github.com/woodleighschool/grinch/internal/transport/http/status"
	"github.com/woodleighschool/grinch/internal/transport/http/sync/helpers"
)

// Preflight handles POST /preflight/{machine_id}.
func (h *Handler) Preflight(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger(ctx).With("op", "preflight")

	machineID, err := helpers.ParseMachineID(r)
	if err != nil {
		log.WarnContext(ctx, "invalid machine id", "error", err)
		w.WriteHeader(httpstatus.Status(err))
		return
	}

	req := &syncv1.PreflightRequest{}
	decodeErr := helpers.DecodeRequest(r, req)
	if decodeErr != nil {
		log.WarnContext(ctx, "invalid preflight payload", "error", decodeErr)
		w.WriteHeader(httpstatus.Status(decodeErr))
		return
	}

	resp, err := h.sync.Preflight(ctx, machineID, req)
	if err != nil {
		log.ErrorContext(ctx, "preflight failed", "error", err)
		w.WriteHeader(httpstatus.Status(err))
		return
	}

	if encodeErr := helpers.EncodeResponse(w, r, resp); encodeErr != nil {
		log.ErrorContext(ctx, "encode response", "error", encodeErr)
	}
}
