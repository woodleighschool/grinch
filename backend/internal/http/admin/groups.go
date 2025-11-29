package admin

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// groupDTO mirrors the fields exposed by the admin UI.
type groupDTO struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
}

// groupsRoutes registers group + membership endpoints.
func (h Handler) groupsRoutes(r chi.Router) {
	r.Get("/", h.listGroups)
	r.Get("/{id}/members", h.groupEffectiveMembers)
}

// listGroups returns directory groups with optional search filtering.
func (h Handler) listGroups(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	groups, err := h.Store.ListGroups(r.Context(), search)
	if err != nil {
		h.Logger.Error("list groups", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to list groups")
		return
	}
	resp := make([]groupDTO, 0, len(groups))
	for _, g := range groups {
		resp = append(resp, groupDTO{
			ID:          g.ID,
			DisplayName: g.DisplayName,
			Description: g.Description.String,
		})
	}
	respondJSON(w, http.StatusOK, resp)
}

// groupEffectiveMembers resolves a group's membership including derived data.
func (h Handler) groupEffectiveMembers(w http.ResponseWriter, r *http.Request) {
	groupID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid group id")
		return
	}
	ctx := r.Context()
	group, err := h.Store.GetGroup(ctx, groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "group not found")
			return
		}
		h.Logger.Error("get group", "err", err, "group", groupID)
		respondError(w, http.StatusInternalServerError, "failed to load group")
		return
	}
	memberIDs, err := h.Store.ListGroupMemberIDs(ctx, groupID)
	if err != nil {
		h.Logger.Error("list group member ids", "err", err, "group", groupID)
		respondError(w, http.StatusInternalServerError, "failed to load membership ids")
		return
	}
	var members []userDTO
	if len(memberIDs) > 0 {
		users, err := h.Store.ListGroupMembers(ctx, groupID)
		if err != nil {
			h.Logger.Error("list group members", "err", err, "group", groupID)
			respondError(w, http.StatusInternalServerError, "failed to load members")
			return
		}
		members = mapUserList(users)
	}
	resp := groupEffectiveMembersResponse{
		Group:     groupDTO{ID: group.ID, DisplayName: group.DisplayName, Description: group.Description.String},
		Members:   members,
		MemberIDs: memberIDs,
		Count:     len(memberIDs),
	}
	respondJSON(w, http.StatusOK, resp)
}

// groupEffectiveMembersResponse contains both member IDs and their richer records.
type groupEffectiveMembersResponse struct {
	Group     groupDTO    `json:"group"`
	Members   []userDTO   `json:"members"`
	MemberIDs []uuid.UUID `json:"member_ids"`
	Count     int         `json:"count"`
}
