package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

// userDTO is the public representation of a directory user.
type userDTO struct {
	ID          uuid.UUID `json:"id"`
	UPN         string    `json:"upn"`
	DisplayName string    `json:"displayName"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

// userDetailResponse bundles the related objects for the detail drawer.
type userDetailResponse struct {
	User         userDTO      `json:"user"`
	Groups       []groupDTO   `json:"groups"`
	Devices      []deviceDTO  `json:"devices"`
	RecentBlocks []eventDTO   `json:"recent_blocks"`
	Policies     []userPolicy `json:"policies"`
}

// userPolicy describes one effective assignment for a user.
type userPolicy struct {
	ScopeID         string `json:"scope_id"`
	ApplicationID   string `json:"application_id"`
	ApplicationName string `json:"application_name"`
	RuleType        string `json:"rule_type"`
	Identifier      string `json:"identifier"`
	Action          string `json:"action"`
	TargetType      string `json:"target_type"`
	TargetID        string `json:"target_id"`
	TargetName      string `json:"target_name"`
	ViaGroup        bool   `json:"via_group"`
	CreatedAt       string `json:"created_at"`
}

// usersRoutes registers the CRUD + policy endpoints.
func (h Handler) usersRoutes(r chi.Router) {
	r.Get("/", h.listUsers)
	r.Post("/", h.upsertUser)
	r.Get("/{id}", h.userDetails)
	r.Get("/{id}/policies", h.userEffectivePolicies)
	r.Get("/{id}/effective-policies", h.userEffectivePolicies)
}

// listUsers surfaces fuzzy search results for the admin UI.
func (h Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	users, err := h.Store.ListUsers(ctx, search)
	if err != nil {
		h.Logger.Error("list users", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	resp := make([]userDTO, 0, len(users))
	for _, u := range users {
		resp = append(resp, mapUserDTO(u))
	}
	respondJSON(w, http.StatusOK, resp)
}

// upsertUser creates or updates a user from the admin UI.
func (h Handler) upsertUser(w http.ResponseWriter, r *http.Request) {
	var body userDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.ID == uuid.Nil {
		body.ID = uuid.New()
	}
	user, err := h.Store.UpsertUser(r.Context(), sqlc.UpsertUserParams{
		ID:          body.ID,
		Upn:         body.UPN,
		DisplayName: body.DisplayName,
	})
	if err != nil {
		h.Logger.Error("upsert user", "err", err)
		respondError(w, http.StatusInternalServerError, "save failed")
		return
	}
	respondJSON(w, http.StatusOK, mapUserDTO(user))
}

// userDetails aggregates the most recent view of a user.
func (h Handler) userDetails(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	ctx := r.Context()
	user, err := h.Store.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		h.Logger.Error("get user", "err", err, "user", userID)
		respondError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	groups, err := h.Store.GetUserGroups(ctx, userID)
	if err != nil {
		h.Logger.Error("get user groups", "err", err, "user", userID)
		respondError(w, http.StatusInternalServerError, "failed to load groups")
		return
	}
	machines, err := h.Store.GetUserMachines(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		h.Logger.Error("get user machines", "err", err, "user", userID)
		respondError(w, http.StatusInternalServerError, "failed to load devices")
		return
	}
	events, err := h.Store.ListBlocksByUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			h.Logger.Error("get user block events", "err", err, "user", userID)
			respondError(w, http.StatusInternalServerError, "failed to load blocked events")
			return
		}
		events = nil
	}
	assignments, err := h.Store.ListUserAssignments(ctx, userID)
	if err != nil {
		h.Logger.Error("list user policies", "err", err, "user", userID)
		respondError(w, http.StatusInternalServerError, "failed to load policies")
		return
	}

	resp := userDetailResponse{
		User:         mapUserDTO(user),
		Groups:       mapGroups(groups),
		Devices:      mapDevices(machines),
		RecentBlocks: mapUserBlocks(events),
		Policies:     mapUserPolicies(assignments, user),
	}
	respondJSON(w, http.StatusOK, resp)
}

func mapUserDTO(u sqlc.User) userDTO {
	var created, updated time.Time
	if u.CreatedAt.Valid {
		created = u.CreatedAt.Time
	}
	if u.UpdatedAt.Valid {
		updated = u.UpdatedAt.Time
	}
	return userDTO{
		ID:          u.ID,
		UPN:         u.Upn,
		DisplayName: u.DisplayName,
		CreatedAt:   created,
		UpdatedAt:   updated,
	}
}

func mapUserList(users []sqlc.User) []userDTO {
	resp := make([]userDTO, 0, len(users))
	for _, u := range users {
		resp = append(resp, mapUserDTO(u))
	}
	return resp
}

func mapGroups(groups []sqlc.Group) []groupDTO {
	resp := make([]groupDTO, 0, len(groups))
	for _, g := range groups {
		resp = append(resp, groupDTO{
			ID:          g.ID,
			DisplayName: g.DisplayName,
			Description: g.Description.String,
		})
	}
	return resp
}

func mapDevices(machines []sqlc.Machine) []deviceDTO {
	resp := make([]deviceDTO, 0, len(machines))
	for _, m := range machines {
		resp = append(resp, mapDevice(m))
	}
	return resp
}

func mapUserBlocks(events []sqlc.ListBlocksByUserRow) []eventDTO {
	resp := make([]eventDTO, 0, len(events))
	for _, e := range events {
		resp = append(resp, mapUserBlock(e))
	}
	return resp
}

func mapUserPolicies(rows []sqlc.ListUserAssignmentsRow, user sqlc.User) []userPolicy {
	resp := make([]userPolicy, 0, len(rows))
	for _, row := range rows {
		var created string
		if row.ScopeCreatedAt.Valid {
			created = row.ScopeCreatedAt.Time.Format(time.RFC3339)
		}
		var targetID, targetName string
		viaGroup := row.TargetType == "group"
		if viaGroup {
			targetID = row.GroupID.String()
			targetName = row.GroupName.String
		} else {
			targetID = row.UserID.String()
			targetName = user.DisplayName
		}
		resp = append(resp, userPolicy{
			ScopeID:         row.ScopeID.String(),
			ApplicationID:   row.RuleID.String(),
			ApplicationName: row.RuleName,
			RuleType:        strings.ToUpper(row.RuleType),
			Identifier:      row.RuleTarget,
			Action:          strings.ToUpper(row.Action),
			TargetType:      row.TargetType,
			TargetID:        targetID,
			TargetName:      targetName,
			ViaGroup:        viaGroup,
			CreatedAt:       created,
		})
	}
	return resp
}

// userEffectivePolicies is a lightweight endpoint used for popovers.
func (h Handler) userEffectivePolicies(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	ctx := r.Context()
	user, err := h.Store.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		h.Logger.Error("get user effective policies", "err", err, "user", userID)
		respondError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	assignments, err := h.Store.ListUserAssignments(ctx, userID)
	if err != nil {
		h.Logger.Error("list user effective policies", "err", err, "user", userID)
		respondError(w, http.StatusInternalServerError, "failed to load policies")
		return
	}
	resp := struct {
		User     userDTO      `json:"user"`
		Policies []userPolicy `json:"policies"`
	}{
		User:     mapUserDTO(user),
		Policies: mapUserPolicies(assignments, user),
	}
	respondJSON(w, http.StatusOK, resp)
}
