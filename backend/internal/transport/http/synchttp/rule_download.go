package synchttp

import (
	"net/http"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/go-chi/chi/v5"
)

func (handler *Handler) ruleDownload(writer http.ResponseWriter, request *http.Request) {
	message := &syncv1.RuleDownloadRequest{}
	if err := handler.decodeRequest(request, message); err != nil {
		handler.fail(writer, err)
		return
	}

	machineID, err := parseMachineID(chi.URLParam(request, "machine_id"))
	if err != nil {
		handler.fail(writer, err)
		return
	}

	response, err := handler.service.HandleRuleDownload(request.Context(), machineID, message)
	if err != nil {
		handler.fail(writer, err)
		return
	}
	handler.writeResponse(writer, response)
}
