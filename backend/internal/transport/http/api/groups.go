package apihttp

import (
	"net/http"

	appgroups "github.com/woodleighschool/grinch/internal/app/groups"
)

type groupWriteRequestBody struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func (s *Server) ListGroups(w http.ResponseWriter, r *http.Request, params ListGroupsParams) {
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

	items, total, err := s.groups.ListGroups(r.Context(), listOptions)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, GroupListResponse{
		Rows:  items,
		Total: total,
	})
}

func (s *Server) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var body groupWriteRequestBody
	if err := decodeJSONBody(r, &body); err != nil {
		writeError(w, err)
		return
	}

	group, err := s.groups.CreateGroup(r.Context(), appgroups.WriteInput{
		Name:        body.Name,
		Description: optionalString(body.Description),
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, group)
}

func (s *Server) GetGroup(w http.ResponseWriter, r *http.Request, id Id) {
	group, err := s.groups.GetGroup(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, group)
}

func (s *Server) UpdateGroup(w http.ResponseWriter, r *http.Request, id Id) {
	var body groupWriteRequestBody
	if err := decodeJSONBody(r, &body); err != nil {
		writeError(w, err)
		return
	}

	updated, err := s.groups.UpdateGroup(
		r.Context(),
		id,
		appgroups.WriteInput{
			Name:        body.Name,
			Description: optionalString(body.Description),
		},
	)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) DeleteGroup(w http.ResponseWriter, r *http.Request, id Id) {
	if err := s.groups.DeleteGroup(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	writeNoContent(w)
}
