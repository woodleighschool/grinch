package apihttp //nolint:dupl // structurally similar to executables.go by design

import (
	"net/http"
)

func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request, params ListUsersParams) {
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

	items, total, err := s.store.ListUsers(r.Context(), listOptions)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, UserListResponse{
		Rows:  items,
		Total: total,
	})
}

func (s *Server) GetUser(w http.ResponseWriter, r *http.Request, id Id) {
	user, err := s.store.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}
