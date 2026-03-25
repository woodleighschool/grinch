package synchttp

import (
	"context"
	"net/http"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/preflight/{machine_id}", h.preflight)
	r.Post("/eventupload/{machine_id}", h.eventUpload)
	r.Post("/ruledownload/{machine_id}", h.ruleDownload)
	r.Post("/postflight/{machine_id}", h.postflight)
}

func (h *Handler) preflight(w http.ResponseWriter, r *http.Request) {
	handleSyncRequest(h, w, r, &syncv1.PreflightRequest{}, h.service.HandlePreflight)
}

func (h *Handler) eventUpload(w http.ResponseWriter, r *http.Request) {
	handleSyncRequest(h, w, r, &syncv1.EventUploadRequest{}, h.service.HandleEventUpload)
}

func (h *Handler) ruleDownload(w http.ResponseWriter, r *http.Request) {
	handleSyncRequest(h, w, r, &syncv1.RuleDownloadRequest{}, h.service.HandleRuleDownload)
}

func (h *Handler) postflight(w http.ResponseWriter, r *http.Request) {
	handleSyncRequest(h, w, r, &syncv1.PostflightRequest{}, h.service.HandlePostflight)
}

func handleSyncRequest[Req proto.Message, Resp proto.Message](
	h *Handler,
	w http.ResponseWriter,
	r *http.Request,
	req Req,
	handle func(context.Context, uuid.UUID, Req) (Resp, error),
) {
	if err := h.decodeRequest(r, req); err != nil {
		h.writeError(w, r, err)
		return
	}

	machineID, err := parseMachineID(chi.URLParam(r, "machine_id"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}

	resp, err := handle(r.Context(), machineID, req)
	if err != nil {
		writeStatusOnly(w, statusCodeForError(err))
		return
	}

	h.writeProtoResponse(w, r, resp)
}
