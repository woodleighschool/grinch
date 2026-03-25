package apihttp

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (s *Server) ListMachines(w http.ResponseWriter, r *http.Request, params ListMachinesParams) {
	listOptions, err := parseListOptions(
		params.Limit,
		params.Offset,
		params.Search,
		params.Sort,
		params.Order,
		params.Ids,
	)
	if err != nil {
		writeError(w, err)
		return
	}

	ruleSyncStatuses, err := parseOptionalValues(params.RuleSyncStatus, domain.ParseMachineRuleSyncStatus)
	if err != nil {
		writeError(w, err)
		return
	}

	clientModes, err := parseOptionalValues(params.ClientMode, domain.ParseMachineClientMode)
	if err != nil {
		writeError(w, err)
		return
	}

	items, total, err := s.store.ListMachines(r.Context(), domain.MachineListOptions{
		ListOptions:      listOptions,
		UserID:           params.UserId,
		RuleSyncStatuses: ruleSyncStatuses,
		ClientModes:      clientModes,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, MachineListResponse{
		Rows:  items,
		Total: total,
	})
}

func (s *Server) GetMachine(w http.ResponseWriter, r *http.Request, id Id) {
	machine, err := s.store.GetMachine(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, machine)
}

func (s *Server) DeleteMachine(w http.ResponseWriter, r *http.Request, id Id) {
	if err := s.store.DeleteMachine(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	writeNoContent(w)
}
