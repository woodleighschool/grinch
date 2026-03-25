package apihttp

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (s *Server) ListFileAccessEvents(
	w http.ResponseWriter,
	r *http.Request,
	params ListFileAccessEventsParams,
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

	decisions, err := parseOptionalValues(params.Decision, domain.ParseFileAccessDecision)
	if err != nil {
		writeError(w, err)
		return
	}

	items, total, err := s.store.ListFileAccessEvents(
		r.Context(),
		domain.FileAccessEventListOptions{
			ListOptions: listOptions,
			MachineID:   params.MachineId,
			Decisions:   decisions,
		},
	)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, FileAccessEventListResponse{
		Rows:  items,
		Total: total,
	})
}

func (s *Server) GetFileAccessEvent(w http.ResponseWriter, r *http.Request, id Id) {
	event, err := s.store.GetFileAccessEvent(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, event)
}

func (s *Server) DeleteFileAccessEvent(w http.ResponseWriter, r *http.Request, id Id) {
	if err := s.store.DeleteFileAccessEvent(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	writeNoContent(w)
}
