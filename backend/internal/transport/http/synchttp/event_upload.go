package synchttp

import (
	"net/http"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/go-chi/chi/v5"
)

func (handler *Handler) eventUpload(writer http.ResponseWriter, request *http.Request) {
	message := &syncv1.EventUploadRequest{}
	if err := handler.decodeRequest(request, message); err != nil {
		handler.fail(writer, err, http.StatusBadRequest)
		return
	}

	machineID, err := parseMachineID(chi.URLParam(request, "machine_id"))
	if err != nil {
		handler.fail(writer, err, http.StatusBadRequest)
		return
	}

	response, err := handler.service.HandleEventUpload(request.Context(), machineID, message)
	if err != nil {
		handler.fail(writer, err, http.StatusInternalServerError)
		return
	}
	handler.writeResponse(writer, response)
}
