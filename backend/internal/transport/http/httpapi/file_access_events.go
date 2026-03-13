package httpapi

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListFileAccessEvents(
	writer http.ResponseWriter,
	request *http.Request,
	params ListFileAccessEventsParams,
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

	var executableID *uuid.UUID
	if params.ExecutableId != nil {
		parsed := *params.ExecutableId
		executableID = &parsed
	}

	items, total, err := handler.admin.ListFileAccessEvents(request.Context(), domain.FileAccessEventListOptions{
		ListOptions:  listOptions,
		MachineID:    machineID,
		ExecutableID: executableID,
	})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	mapped := make([]FileAccessEventSummary, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, mapFileAccessEventSummary(item))
	}

	writeJSON(writer, http.StatusOK, FileAccessEventListResponse{
		Rows:  mapped,
		Total: total,
	})
}

func (handler *Server) GetFileAccessEvent(writer http.ResponseWriter, request *http.Request, id Id) {
	event, err := handler.admin.GetFileAccessEvent(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "file access event not found"})
		return
	}

	writeJSON(writer, http.StatusOK, mapFileAccessEvent(event))
}

func (handler *Server) DeleteFileAccessEvent(writer http.ResponseWriter, request *http.Request, id Id) {
	if err := handler.admin.DeleteFileAccessEvent(request.Context(), id); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "file access event not found"})
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
