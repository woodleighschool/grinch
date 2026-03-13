package httpapi

import (
	"net/http"

	"github.com/google/uuid"

	appgroupmemberships "github.com/woodleighschool/grinch/internal/app/groupmemberships"
	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListGroupMemberships(
	writer http.ResponseWriter,
	request *http.Request,
	params ListGroupMembershipsParams,
) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	var groupID *uuid.UUID
	if params.GroupId != nil {
		value := *params.GroupId
		groupID = &value
	}

	var userID *uuid.UUID
	if params.UserId != nil {
		value := *params.UserId
		userID = &value
	}

	var machineID *uuid.UUID
	if params.MachineId != nil {
		value := *params.MachineId
		machineID = &value
	}

	items, total, err := handler.groupMemberships.ListGroupMemberships(
		request.Context(),
		domain.GroupMembershipListOptions{
			ListOptions: listOptions,
			GroupID:     groupID,
			UserID:      userID,
			MachineID:   machineID,
		},
	)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	mapped := make([]GroupMembership, 0, len(items))
	for _, item := range items {
		output, mapErr := mapGroupMembership(item)
		if mapErr != nil {
			writeClassifiedError(writer, mapErr, apiErrorOptions{})
			return
		}
		mapped = append(mapped, output)
	}

	writeJSON(writer, http.StatusOK, GroupMembershipListResponse{
		Rows:  mapped,
		Total: total,
	})
}

func (handler *Server) CreateGroupMembership(writer http.ResponseWriter, request *http.Request) {
	var body CreateGroupMembershipJSONRequestBody
	if err := decodeJSONBody(request, &body); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	memberKind, err := toDomainMemberKind(body.MemberKind)
	if err != nil {
		writeClassifiedError(writer, badRequestError("invalid member_kind"), apiErrorOptions{})
		return
	}

	membership, err := handler.groupMemberships.CreateGroupMembership(
		request.Context(),
		appgroupmemberships.CreateInput{
			GroupID:    body.GroupId,
			MemberKind: memberKind,
			MemberID:   body.MemberId,
		},
	)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "group membership dependencies not found"})
		return
	}

	mapped, err := mapGroupMembership(membership)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusCreated, mapped)
}

func (handler *Server) GetGroupMembership(writer http.ResponseWriter, request *http.Request, id GroupMembershipId) {
	membership, err := handler.groupMemberships.GetGroupMembership(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "group membership not found"})
		return
	}

	mapped, err := mapGroupMembership(membership)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, mapped)
}

func (handler *Server) DeleteGroupMembership(writer http.ResponseWriter, request *http.Request, id GroupMembershipId) {
	deleteErr := handler.groupMemberships.DeleteGroupMembership(request.Context(), id)
	if deleteErr != nil {
		writeClassifiedError(writer, deleteErr, apiErrorOptions{NotFoundMessage: "group membership not found"})
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
