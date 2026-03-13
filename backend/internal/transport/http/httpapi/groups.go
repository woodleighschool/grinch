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

	mapped := make([]Group, 0, len(items))
	for _, item := range items {
		output, mapErr := mapGroup(item)
		if mapErr != nil {
			writeClassifiedError(writer, mapErr, apiErrorOptions{})
			return
		}
		mapped = append(mapped, output)
	}

	writeJSON(writer, http.StatusOK, GroupListResponse{
		Rows:  mapped,
		Total: total,
	})
}

func (handler *Server) CreateGroup(writer http.ResponseWriter, request *http.Request) {
	var body CreateGroupJSONRequestBody
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

	mapped, mapErr := mapGroup(group)
	if mapErr != nil {
		writeClassifiedError(writer, mapErr, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusCreated, mapped)
}

func (handler *Server) GetGroup(writer http.ResponseWriter, request *http.Request, id Id) {
	group, err := handler.admin.GetGroup(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "group not found"})
		return
	}

	mapped, mapErr := mapGroup(group)
	if mapErr != nil {
		writeClassifiedError(writer, mapErr, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, mapped)
}

func (handler *Server) PatchGroup(writer http.ResponseWriter, request *http.Request, id Id) {
	var body PatchGroupJSONRequestBody
	decodeErr := decodeJSONBody(request, &body)
	if decodeErr != nil {
		writeClassifiedError(writer, decodeErr, apiErrorOptions{})
		return
	}

	input := decodeGroupPatchRequest(body)

	updated, err := handler.admin.PatchGroup(request.Context(), id, input.name, input.description)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "group not found"})
		return
	}

	mapped, mapErr := mapGroup(updated)
	if mapErr != nil {
		writeClassifiedError(writer, mapErr, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, mapped)
}

func (handler *Server) DeleteGroup(writer http.ResponseWriter, request *http.Request, id Id) {
	deleteErr := handler.admin.DeleteGroup(request.Context(), id)
	if deleteErr != nil {
		writeClassifiedError(writer, deleteErr, apiErrorOptions{NotFoundMessage: "group not found"})
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func decodeGroupWriteRequest(body CreateGroupJSONRequestBody) (groupWriteRequest, error) {
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

func decodeGroupPatchRequest(body PatchGroupJSONRequestBody) groupWriteRequest {
	result := groupWriteRequest{}
	if body.Name != nil {
		name := strings.TrimSpace(*body.Name)
		result.name = &name
	}
	if body.Description != nil {
		description := *body.Description
		result.description = &description
	}
	return result
}
