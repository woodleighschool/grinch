package handlers

import (
	"net/http"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"

	httpstatus "github.com/woodleighschool/grinch/internal/transport/http/status"
	"github.com/woodleighschool/grinch/internal/transport/http/sync/helpers"
)

// EventUpload handles POST /eventupload/{machine_id}.
func (h *Handler) EventUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger(ctx).With("op", "eventupload")

	machineID, err := helpers.ParseMachineID(r)
	if err != nil {
		log.WarnContext(ctx, "invalid machine id", "error", err)
		w.WriteHeader(httpstatus.Status(err))
		return
	}

	req := &syncv1.EventUploadRequest{}
	decodeErr := helpers.DecodeRequest(r, req)
	if decodeErr != nil {
		log.WarnContext(ctx, "invalid event payload", "error", decodeErr)
		w.WriteHeader(httpstatus.Status(decodeErr))
		return
	}

	events := helpers.ConvertEvents(machineID, req)
	serviceErr := h.sync.EventUpload(ctx, machineID, events)
	if serviceErr != nil {
		log.ErrorContext(ctx, "eventupload failed", "error", serviceErr)
		w.WriteHeader(httpstatus.Status(serviceErr))
		return
	}

	if encodeErr := helpers.EncodeResponse(w, r, &syncv1.EventUploadResponse{}); encodeErr != nil {
		log.ErrorContext(ctx, "encode response", "error", encodeErr)
	}
}
