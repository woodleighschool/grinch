package apihttp

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListExecutionEvents(
	writer http.ResponseWriter,
	request *http.Request,
	params ListExecutionEventsParams,
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

	decisions, err := parseOptionalValues(params.Decision, domain.ParseExecutionDecision)
	if err != nil {
		writeError(writer, err)
		return
	}

	items, total, err := handler.store.ListExecutionEvents(request.Context(), domain.ExecutionEventListOptions{
		ListOptions:  listOptions,
		MachineID:    params.MachineId,
		UserID:       params.UserId,
		ExecutableID: params.ExecutableId,
		Decisions:    decisions,
	})
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, ExecutionEventListResponse{
		Rows:  items,
		Total: total,
	})
}

func (handler *Server) GetExecutionEvent(writer http.ResponseWriter, request *http.Request, id Id) {
	event, err := handler.store.GetExecutionEvent(request.Context(), id)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, event)
}

func (handler *Server) DeleteExecutionEvent(writer http.ResponseWriter, request *http.Request, id Id) {
	if err := handler.store.DeleteExecutionEvent(request.Context(), id); err != nil {
		writeError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
