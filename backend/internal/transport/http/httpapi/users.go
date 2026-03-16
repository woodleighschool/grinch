package httpapi

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListUsers(writer http.ResponseWriter, request *http.Request, params ListUsersParams) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	items, total, err := handler.admin.ListUsers(request.Context(), domain.UserListOptions{ListOptions: listOptions})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	mapped, err := mapSlice(items, mapUser)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, UserListResponse{
		Rows:  mapped,
		Total: total,
	})
}

func (handler *Server) GetUser(writer http.ResponseWriter, request *http.Request, id Id) {
	user, err := handler.admin.GetUser(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "user not found"})
		return
	}

	mapped, mapErr := mapUser(user)
	if mapErr != nil {
		writeClassifiedError(writer, mapErr, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, mapped)
}
