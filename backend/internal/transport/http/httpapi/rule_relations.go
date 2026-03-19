package httpapi

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListMachineRules(
	writer http.ResponseWriter,
	request *http.Request,
	params ListMachineRulesParams,
) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order, nil)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	items, total, err := handler.admin.ListMachineRules(request.Context(), domain.MachineRuleListOptions{
		ListOptions: listOptions,
		MachineID:   params.MachineId,
	})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, MachineRuleListResponse{
		Rows:  items,
		Total: total,
	})
}

func (handler *Server) ListRuleMachines(
	writer http.ResponseWriter,
	request *http.Request,
	params ListRuleMachinesParams,
) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order, nil)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	items, total, err := handler.admin.ListRuleMachines(request.Context(), domain.RuleMachineListOptions{
		ListOptions: listOptions,
		RuleID:      params.RuleId,
	})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, RuleMachineListResponse{
		Rows:  items,
		Total: total,
	})
}
