package httpapi

import (
	"net/http"

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

	items, total, err := handler.groupMemberships.ListGroupMemberships(
		request.Context(),
		domain.GroupMembershipListOptions{
			ListOptions: listOptions,
			GroupID:     params.GroupId,
			UserID:      params.UserId,
			MachineID:   params.MachineId,
		},
	)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	mapped, err := mapSlice(items, mapGroupMembership)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
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
