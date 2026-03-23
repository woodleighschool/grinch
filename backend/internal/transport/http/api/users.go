package apihttp //nolint:dupl // structurally similar to executables.go by design

import (
	"net/http"
)

func (handler *Server) ListUsers(writer http.ResponseWriter, request *http.Request, params ListUsersParams) {
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

	items, total, err := handler.store.ListUsers(request.Context(), listOptions)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, UserListResponse{
		Rows:  items,
		Total: total,
	})
}

func (handler *Server) GetUser(writer http.ResponseWriter, request *http.Request, id Id) {
	user, err := handler.store.GetUser(request.Context(), id)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, user)
}
