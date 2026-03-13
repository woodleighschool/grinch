package httpapi

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListExecutionEvents(
	writer http.ResponseWriter,
	request *http.Request,
	params ListExecutionEventsParams,
) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	var machineID *uuid.UUID
	if params.MachineId != nil {
		parsed := *params.MachineId
		machineID = &parsed
	}

	var userID *uuid.UUID
	if params.UserId != nil {
		parsed := *params.UserId
		userID = &parsed
	}

	var executableID *uuid.UUID
	if params.ExecutableId != nil {
		parsed := *params.ExecutableId
		executableID = &parsed
	}

	items, total, err := handler.admin.ListExecutionEvents(request.Context(), domain.ExecutionEventListOptions{
		ListOptions:  listOptions,
		MachineID:    machineID,
		UserID:       userID,
		ExecutableID: executableID,
	})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	mapped := make([]ExecutionEventSummary, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, mapExecutionEventSummary(item))
	}

	writeJSON(writer, http.StatusOK, ExecutionEventListResponse{
		Rows:  mapped,
		Total: total,
	})
}

func (handler *Server) GetExecutionEvent(writer http.ResponseWriter, request *http.Request, id Id) {
	event, err := handler.admin.GetExecutionEvent(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "execution event not found"})
		return
	}

	writeJSON(writer, http.StatusOK, mapExecutionEvent(event))
}

func (handler *Server) DeleteExecutionEvent(writer http.ResponseWriter, request *http.Request, id Id) {
	if err := handler.admin.DeleteExecutionEvent(request.Context(), id); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "execution event not found"})
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
