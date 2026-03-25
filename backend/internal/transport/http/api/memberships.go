package apihttp

import (
	"net/http"

	appmemberships "github.com/woodleighschool/grinch/internal/app/memberships"
	"github.com/woodleighschool/grinch/internal/domain"
)

func (s *Server) ListMemberships(
	w http.ResponseWriter,
	r *http.Request,
	params ListMembershipsParams,
) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order, nil)
	if err != nil {
		writeError(w, err)
		return
	}

	items, total, err := s.memberships.ListMemberships(
		r.Context(),
		domain.MembershipListOptions{
			ListOptions: listOptions,
			GroupID:     params.GroupId,
			UserID:      params.UserId,
			MachineID:   params.MachineId,
		},
	)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, MembershipListResponse{
		Rows:  items,
		Total: total,
	})
}

func (s *Server) CreateMembership(w http.ResponseWriter, r *http.Request) {
	var body CreateMembershipJSONRequestBody
	if err := decodeJSONBody(r, &body); err != nil {
		writeError(w, err)
		return
	}

	membership, err := s.memberships.CreateMembership(
		r.Context(),
		appmemberships.CreateInput{
			GroupID:    body.GroupId,
			MemberKind: body.MemberKind,
			MemberID:   body.MemberId,
		},
	)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, membership)
}

func (s *Server) GetMembership(w http.ResponseWriter, r *http.Request, id MembershipId) {
	membership, err := s.memberships.GetMembership(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, membership)
}

func (s *Server) DeleteMembership(w http.ResponseWriter, r *http.Request, id MembershipId) {
	if err := s.memberships.DeleteMembership(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	writeNoContent(w)
}
