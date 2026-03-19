package httpapi

import (
	"net/http"
	"strings"

	"github.com/woodleighschool/grinch/internal/domain"
)

type groupWriteRequest struct {
	description *string
	name        *string
}

type groupWriteRequestBody struct {
	ID          *string `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Source      *string `json:"source,omitempty"`
	MemberCount *int32  `json:"member_count,omitempty"`
	CreatedAt   *string `json:"created_at,omitempty"`
	UpdatedAt   *string `json:"updated_at,omitempty"`
}

func (handler *Server) ListGroups(writer http.ResponseWriter, request *http.Request, params ListGroupsParams) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	items, total, err := handler.admin.ListGroups(request.Context(), domain.GroupListOptions{
		ListOptions: listOptions,
	})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, GroupListResponse{
		Rows:  items,
		Total: total,
	})
}

func (handler *Server) CreateGroup(writer http.ResponseWriter, request *http.Request) {
	var body groupWriteRequestBody
	decodeErr := decodeJSONBody(request, &body)
	if decodeErr != nil {
		writeClassifiedError(writer, decodeErr, apiErrorOptions{})
		return
	}

	input, inputErr := decodeGroupWriteRequest(body)
	if inputErr != nil {
		// Groups write straight to the store, so this transport check owns the
		// minimal required-field validation for the request body shape.
		writeClassifiedError(writer, &domain.ValidationError{
			Code:   "validation_error",
			Detail: "Group is invalid.",
			FieldErrors: []domain.FieldError{{
				Field:   "name",
				Message: inputErr.Error(),
				Code:    "required",
			}},
		}, apiErrorOptions{})
		return
	}

	group, err := handler.admin.CreateLocalGroup(request.Context(), *input.name, optionalStringValue(input.description))
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusCreated, group)
}

func (handler *Server) GetGroup(writer http.ResponseWriter, request *http.Request, id Id) {
	group, err := handler.admin.GetGroup(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "group not found"})
		return
	}

	writeJSON(writer, http.StatusOK, group)
}

func (handler *Server) UpdateGroup(writer http.ResponseWriter, request *http.Request, id Id) {
	var body groupWriteRequestBody
	decodeErr := decodeJSONBody(request, &body)
	if decodeErr != nil {
		writeClassifiedError(writer, decodeErr, apiErrorOptions{})
		return
	}

	input, inputErr := decodeGroupWriteRequest(body)
	if inputErr != nil {
		writeClassifiedError(writer, &domain.ValidationError{
			Code:   "validation_error",
			Detail: "Group is invalid.",
			FieldErrors: []domain.FieldError{{
				Field:   "name",
				Message: inputErr.Error(),
				Code:    "required",
			}},
		}, apiErrorOptions{})
		return
	}

	updated, err := handler.admin.UpdateGroup(
		request.Context(),
		id,
		*input.name,
		optionalStringValue(input.description),
	)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "group not found"})
		return
	}

	writeJSON(writer, http.StatusOK, updated)
}

func (handler *Server) DeleteGroup(writer http.ResponseWriter, request *http.Request, id Id) {
	deleteErr := handler.admin.DeleteGroup(request.Context(), id)
	if deleteErr != nil {
		writeClassifiedError(writer, deleteErr, apiErrorOptions{NotFoundMessage: "group not found"})
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func decodeGroupWriteRequest(body groupWriteRequestBody) (groupWriteRequest, error) {
	name := strings.TrimSpace(body.Name)
	if name == "" {
		return groupWriteRequest{}, badRequestError("name is required")
	}

	description := optionalStringValue(body.Description)
	return groupWriteRequest{
		description: &description,
		name:        &name,
	}, nil
}
