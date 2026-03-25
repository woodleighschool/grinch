package apihttp

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (s *Server) ListExecutionEvents(
	w http.ResponseWriter,
	r *http.Request,
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
		writeError(w, err)
		return
	}

	decisions, err := parseOptionalValues(params.Decision, domain.ParseExecutionDecision)
	if err != nil {
		writeError(w, err)
		return
	}

	items, total, err := s.store.ListExecutionEvents(r.Context(), domain.ExecutionEventListOptions{
		ListOptions:  listOptions,
		MachineID:    params.MachineId,
		UserID:       params.UserId,
		ExecutableID: params.ExecutableId,
		Decisions:    decisions,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ExecutionEventListResponse{
		Rows:  items,
		Total: total,
	})
}

func (s *Server) GetExecutionEvent(w http.ResponseWriter, r *http.Request, id Id) {
	event, err := s.store.GetExecutionEvent(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, event)
}

func (s *Server) DeleteExecutionEvent(w http.ResponseWriter, r *http.Request, id Id) {
	if err := s.store.DeleteExecutionEvent(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	writeNoContent(w)
}
