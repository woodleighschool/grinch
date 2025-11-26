package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

// groupDTO mirrors the fields exposed by the admin UI.
type groupDTO struct {
	ID          uuid.UUID   `json:"id"`
	DisplayName string      `json:"displayName"`
	Description string      `json:"description"`
	Members     []uuid.UUID `json:"members"`
}

// groupsRoutes registers membership + group management endpoints.
func (h Handler) groupsRoutes(r chi.Router) {
	r.Get("/", h.listGroups)
	r.Post("/", h.upsertGroup)
	r.Delete("/{id}", h.deleteGroup)
	r.Get("/{id}/members", h.groupEffectiveMembers)
	r.Get("/{id}/effective-members", h.groupEffectiveMembers)
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

// upsertGroup creates or updates a group record and synchronises membership.
func (h Handler) upsertGroup(w http.ResponseWriter, r *http.Request) {
	var body groupDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.ID == uuid.Nil {
		body.ID = uuid.New()
	}
	group, err := h.Store.UpsertGroup(r.Context(), sqlc.UpsertGroupParams{
		ID:          body.ID,
		DisplayName: body.DisplayName,
		Description: sqlNullString(body.Description),
	})
	if err != nil {
		h.Logger.Error("upsert group", "err", err)
		respondError(w, http.StatusInternalServerError, "save failed")
		return
	}
	if len(body.Members) > 0 {
		if err := h.Store.ReplaceGroupMembers(r.Context(), body.ID, body.Members); err != nil {
			h.Logger.Error("sync members", "err", err)
			respondError(w, http.StatusInternalServerError, "members update failed")
			return
		}
		go h.recompileGroupRules(context.WithoutCancel(r.Context()), body.ID)
	}
	respondJSON(w, http.StatusOK, groupDTO{
		ID:          group.ID,
		DisplayName: group.DisplayName,
		Description: group.Description.String,
		Members:     body.Members,
	})
}

// deleteGroup removes a directory group entirely.
func (h Handler) deleteGroup(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.Store.DeleteGroup(r.Context(), id); err != nil {
		h.Logger.Error("delete group", "err", err)
		respondError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	respondJSON(w, http.StatusNoContent, nil)
}

// sqlNullString converts optional text fields into pgtype.Text.
func sqlNullString(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
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
