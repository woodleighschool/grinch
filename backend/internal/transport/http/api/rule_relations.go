package apihttp

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (s *Server) ListMachineRules(
	w http.ResponseWriter,
	r *http.Request,
	params ListMachineRulesParams,
) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order, nil)
	if err != nil {
		writeError(w, err)
		return
	}

	items, total, err := s.store.ListMachineRules(r.Context(), domain.MachineRuleListOptions{
		ListOptions: listOptions,
		MachineID:   &params.MachineId,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, MachineRuleListResponse{
		Rows:  items,
		Total: total,
	})
}

func (s *Server) ListRuleMachines(
	w http.ResponseWriter,
	r *http.Request,
	params ListRuleMachinesParams,
) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order, nil)
	if err != nil {
		writeError(w, err)
		return
	}

	items, total, err := s.store.ListRuleMachines(r.Context(), domain.RuleMachineListOptions{
		ListOptions: listOptions,
		RuleID:      &params.RuleId,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, RuleMachineListResponse{
		Rows:  items,
		Total: total,
	})
}
