package apihttp //nolint:dupl // structurally similar to users.go by design

import (
	"net/http"
)

func (s *Server) ListExecutables(
	w http.ResponseWriter,
	r *http.Request,
	params ListExecutablesParams,
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

	items, total, err := s.store.ListExecutables(r.Context(), listOptions)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ExecutableListResponse{
		Rows:  items,
		Total: total,
	})
}

func (s *Server) GetExecutable(w http.ResponseWriter, r *http.Request, id Id) {
	executable, err := s.store.GetExecutable(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, executable)
}
