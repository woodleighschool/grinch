package handlers

import (
	"net/http"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"

	httpstatus "github.com/woodleighschool/grinch/internal/transport/http/status"
	"github.com/woodleighschool/grinch/internal/transport/http/sync/helpers"
)

// RuleDownload handles POST /ruledownload/{machine_id}.
func (h *Handler) RuleDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := h.logger(ctx).With("op", "ruledownload")

	machineID, err := helpers.ParseMachineID(r)
	if err != nil {
		log.WarnContext(ctx, "invalid machine id", "error", err)
		w.WriteHeader(httpstatus.Status(err))
		return
	}

	req := &syncv1.RuleDownloadRequest{}
	decodeErr := helpers.DecodeRequest(r, req)
	if decodeErr != nil {
		log.WarnContext(ctx, "invalid rule payload", "error", decodeErr)
		w.WriteHeader(httpstatus.Status(decodeErr))
		return
	}

	offset, err := helpers.ParseCursor(req.GetCursor())
	if err != nil {
		log.WarnContext(ctx, "invalid cursor", "error", err)
		w.WriteHeader(httpstatus.Status(err))
		return
	}

	resp, err := h.sync.RuleDownload(ctx, machineID, offset)
	if err != nil {
		log.ErrorContext(ctx, "ruledownload failed", "error", err)
		w.WriteHeader(httpstatus.Status(err))
		return
	}

	if encodeErr := helpers.EncodeResponse(w, r, resp); encodeErr != nil {
		log.ErrorContext(ctx, "encode response", "error", encodeErr)
	}
}
