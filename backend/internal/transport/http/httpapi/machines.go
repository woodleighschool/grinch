package httpapi

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListMachines(writer http.ResponseWriter, request *http.Request, params ListMachinesParams) {
	listOptions, err := parseListOptions(
		params.Limit,
		params.Offset,
		params.Search,
		params.Sort,
		params.Order,
		params.Ids,
	)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	ruleSyncStatuses, err := parseOptionalValues(params.RuleSyncStatus, domain.ParseMachineRuleSyncStatus)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	clientModes, err := parseOptionalValues(params.ClientMode, domain.ParseMachineClientMode)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	items, total, err := handler.admin.ListMachines(request.Context(), domain.MachineListOptions{
		ListOptions:      listOptions,
		UserID:           params.UserId,
		RuleSyncStatuses: ruleSyncStatuses,
		ClientModes:      clientModes,
	})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, MachineListResponse{
		Rows:  items,
		Total: total,
	})
}

func (handler *Server) GetMachine(writer http.ResponseWriter, request *http.Request, id Id) {
	machine, err := handler.admin.GetMachine(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "machine not found"})
		return
	}

	writeJSON(writer, http.StatusOK, machine)
}

func (handler *Server) DeleteMachine(writer http.ResponseWriter, request *http.Request, id Id) {
	if err := handler.admin.DeleteMachine(request.Context(), id); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "machine not found"})
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
