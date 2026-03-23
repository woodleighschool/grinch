package apihttp

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListFileAccessEvents(
	writer http.ResponseWriter,
	request *http.Request,
	params ListFileAccessEventsParams,
) {
	listOptions, err := parseListOptions(
		params.Limit,
		params.Offset,
		params.Search,
		params.Sort,
		params.Order,
		params.Ids,
	)
	if err != nil {
		writeError(writer, err)
		return
	}

	decisions, err := parseOptionalValues(params.Decision, domain.ParseFileAccessDecision)
	if err != nil {
		writeError(writer, err)
		return
	}

	items, total, err := handler.fileAccessEvents.ListFileAccessEvents(
		request.Context(),
		domain.FileAccessEventListOptions{
			ListOptions: listOptions,
			MachineID:   params.MachineId,
			Decisions:   decisions,
		},
	)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, FileAccessEventListResponse{
		Rows:  items,
		Total: total,
	})
}

func (handler *Server) GetFileAccessEvent(writer http.ResponseWriter, request *http.Request, id Id) {
	event, err := handler.fileAccessEvents.GetFileAccessEvent(request.Context(), id)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, event)
}

func (handler *Server) DeleteFileAccessEvent(writer http.ResponseWriter, request *http.Request, id Id) {
	if err := handler.fileAccessEvents.DeleteFileAccessEvent(request.Context(), id); err != nil {
		writeError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
