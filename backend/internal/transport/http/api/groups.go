package apihttp

import (
	"net/http"

	appgroups "github.com/woodleighschool/grinch/internal/app/groups"
)

type groupWriteRequestBody struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func (handler *Server) ListGroups(writer http.ResponseWriter, request *http.Request, params ListGroupsParams) {
	listOptions, err := parseListOptions(
		params.Limit,
		params.Offset,
		params.Search,
		params.Sort,
		params.Order,
		params.Ids,
	)
	if err != nil {
		writeError(writer, err)
		return
	}

	items, total, err := handler.groups.ListGroups(request.Context(), listOptions)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, GroupListResponse{
		Rows:  items,
		Total: total,
	})
}

func (handler *Server) CreateGroup(writer http.ResponseWriter, request *http.Request) {
	var body groupWriteRequestBody
	if err := decodeJSONBody(request, &body); err != nil {
		writeError(writer, err)
		return
	}

	group, err := handler.groups.CreateGroup(request.Context(), appgroups.WriteInput{
		Name:        body.Name,
		Description: optionalString(body.Description),
	})
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusCreated, group)
}

func (handler *Server) GetGroup(writer http.ResponseWriter, request *http.Request, id Id) {
	group, err := handler.groups.GetGroup(request.Context(), id)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, group)
}

func (handler *Server) UpdateGroup(writer http.ResponseWriter, request *http.Request, id Id) {
	var body groupWriteRequestBody
	if err := decodeJSONBody(request, &body); err != nil {
		writeError(writer, err)
		return
	}

	updated, err := handler.groups.UpdateGroup(
		request.Context(),
		id,
		appgroups.WriteInput{
			Name:        body.Name,
			Description: optionalString(body.Description),
		},
	)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, updated)
}

func (handler *Server) DeleteGroup(writer http.ResponseWriter, request *http.Request, id Id) {
	if err := handler.groups.DeleteGroup(request.Context(), id); err != nil {
		writeError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
