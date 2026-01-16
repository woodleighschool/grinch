//nolint:dupl // entry is similar to preflight but handles a different request type.
package handlers

import (
	"net/http"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"

	httpstatus "github.com/woodleighschool/grinch/internal/transport/http/status"
	"github.com/woodleighschool/grinch/internal/transport/http/sync/helpers"
)

// Postflight handles POST /postflight/{machine_id}.
func (h *Handler) Postflight(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger(ctx).With("op", "postflight")

	machineID, err := helpers.ParseMachineID(r)
	if err != nil {
		log.WarnContext(ctx, "invalid machine id", "error", err)
		w.WriteHeader(httpstatus.Status(err))
		return
	}

	req := &syncv1.PostflightRequest{}
	decodeErr := helpers.DecodeRequest(r, req)
	if decodeErr != nil {
		log.WarnContext(ctx, "invalid postflight payload", "error", decodeErr)
		w.WriteHeader(httpstatus.Status(decodeErr))
		return
	}

	resp, err := h.sync.Postflight(ctx, machineID, req)
	if err != nil {
		log.ErrorContext(ctx, "postflight failed", "error", err)
		w.WriteHeader(httpstatus.Status(err))
		return
	}

	if encodeErr := helpers.EncodeResponse(w, r, resp); encodeErr != nil {
		log.ErrorContext(ctx, "encode response", "error", encodeErr)
	}
}
