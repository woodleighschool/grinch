package httpapi

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListMachines(writer http.ResponseWriter, request *http.Request, params ListMachinesParams) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	items, total, err := handler.admin.ListMachines(request.Context(), domain.MachineListOptions{
		ListOptions: listOptions,
		UserID:      params.UserId,
	})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, MachineListResponse{
		Rows:  mapSliceValue(items, mapMachineSummary),
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
