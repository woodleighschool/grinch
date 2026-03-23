package apihttp

import (
	"net/http"

	appmemberships "github.com/woodleighschool/grinch/internal/app/memberships"
	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListMemberships(
	writer http.ResponseWriter,
	request *http.Request,
	params ListMembershipsParams,
) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order, nil)
	if err != nil {
		writeError(writer, err)
		return
	}

	items, total, err := handler.memberships.ListMemberships(
		request.Context(),
		domain.MembershipListOptions{
			ListOptions: listOptions,
			GroupID:     params.GroupId,
			UserID:      params.UserId,
			MachineID:   params.MachineId,
		},
	)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, MembershipListResponse{
		Rows:  items,
		Total: total,
	})
}

func (handler *Server) CreateMembership(writer http.ResponseWriter, request *http.Request) {
	var body CreateMembershipJSONRequestBody
	if err := decodeJSONBody(request, &body); err != nil {
		writeError(writer, err)
		return
	}

	membership, err := handler.memberships.CreateMembership(
		request.Context(),
		appmemberships.CreateInput{
			GroupID:    body.GroupId,
			MemberKind: body.MemberKind,
			MemberID:   body.MemberId,
		},
	)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusCreated, membership)
}

func (handler *Server) GetMembership(writer http.ResponseWriter, request *http.Request, id MembershipId) {
	membership, err := handler.memberships.GetMembership(request.Context(), id)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, membership)
}

func (handler *Server) DeleteMembership(writer http.ResponseWriter, request *http.Request, id MembershipId) {
	if err := handler.memberships.DeleteMembership(request.Context(), id); err != nil {
		writeError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
