package apihttp

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListMachineRules(
	writer http.ResponseWriter,
	request *http.Request,
	params ListMachineRulesParams,
) {
	if params.MachineId == nil {
		writeError(writer, badRequestError("machine_id is required"))
		return
	}

	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order, nil)
	if err != nil {
		writeError(writer, err)
		return
	}

	items, total, err := handler.store.ListMachineRules(request.Context(), domain.MachineRuleListOptions{
		ListOptions: listOptions,
		MachineID:   params.MachineId,
	})
	if err != nil {
		writeError(writer, err)
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
	if params.RuleId == nil {
		writeError(writer, badRequestError("rule_id is required"))
		return
	}

	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order, nil)
	if err != nil {
		writeError(writer, err)
		return
	}

	items, total, err := handler.store.ListRuleMachines(request.Context(), domain.RuleMachineListOptions{
		ListOptions: listOptions,
		RuleID:      params.RuleId,
	})
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, RuleMachineListResponse{
		Rows:  items,
		Total: total,
	})
}
