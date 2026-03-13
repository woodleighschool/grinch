package httpapi

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListMachines(writer http.ResponseWriter, request *http.Request, params ListMachinesParams) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	var userID *uuid.UUID
	if params.UserId != nil {
		value := *params.UserId
		userID = &value
	}

	items, total, err := handler.admin.ListMachines(request.Context(), domain.MachineListOptions{
		ListOptions: listOptions,
		UserID:      userID,
	})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	mapped := make([]MachineSummary, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, mapMachineSummary(item))
	}

	writeJSON(writer, http.StatusOK, MachineListResponse{
		Rows:  mapped,
		Total: total,
	})
}

func (handler *Server) GetMachine(writer http.ResponseWriter, request *http.Request, id Id) {
	machine, err := handler.admin.GetMachine(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "machine not found"})
		return
	}

	writeJSON(writer, http.StatusOK, mapMachine(machine))
}

func (handler *Server) DeleteMachine(writer http.ResponseWriter, request *http.Request, id Id) {
	if err := handler.admin.DeleteMachine(request.Context(), id); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "machine not found"})
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
