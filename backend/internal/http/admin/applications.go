package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/rules"
	"github.com/woodleighschool/grinch/internal/store"
	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

type applicationDTO struct {
	ID            uuid.UUID                     `json:"id"`
	Name          string                        `json:"name"`
	RuleType      string                        `json:"rule_type"`
	Identifier    string                        `json:"identifier"`
	Description   string                        `json:"description,omitempty"`
	BlockMessage  string                        `json:"block_message,omitempty"`
	CelEnabled    bool                          `json:"cel_enabled"`
	CelExpression string                        `json:"cel_expression,omitempty"`
	Enabled       bool                          `json:"enabled"`
	CreatedAt     time.Time                     `json:"created_at,omitempty"`
	UpdatedAt     time.Time                     `json:"updated_at,omitempty"`
	Stats         applicationAssignmentStatsDTO `json:"assignment_stats"`
}

type applicationScopeDTO struct {
	ID            uuid.UUID `json:"id"`
	ApplicationID uuid.UUID `json:"application_id"`
	TargetType    string    `json:"target_type"`
	TargetID      uuid.UUID `json:"target_id"`
	Action        string    `json:"action"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
}

type applicationScopeRelationshipDTO struct {
	applicationScopeDTO
	TargetDisplayName    string      `json:"target_display_name,omitempty"`
	TargetDescription    string      `json:"target_description,omitempty"`
	TargetUPN            string      `json:"target_upn,omitempty"`
	EffectiveMemberIDs   []uuid.UUID `json:"effective_member_ids"`
	EffectiveMemberCount int         `json:"effective_member_count"`
	EffectiveMembers     []userDTO   `json:"effective_members,omitempty"`
}

type applicationDetailResponse struct {
	Application applicationDTO                    `json:"application"`
	Scopes      []applicationScopeRelationshipDTO `json:"scopes"`
}

type applicationAssignmentStatsDTO struct {
	AllowScopes    int `json:"allow_scopes"`
	BlockScopes    int `json:"block_scopes"`
	CelScopes      int `json:"cel_scopes"`
	TotalScopes    int `json:"total_scopes"`
	AllowUsers     int `json:"allow_users"`
	BlockUsers     int `json:"block_users"`
	CelUsers       int `json:"cel_users"`
	TotalUsers     int `json:"total_users"`
	TotalMachines  int `json:"total_machines"`
	SyncedMachines int `json:"synced_machines"`
}

type createApplicationRequest struct {
	Name          string `json:"name"`
	RuleType      string `json:"rule_type"`
	Identifier    string `json:"identifier"`
	Description   string `json:"description"`
	BlockMessage  string `json:"block_message"`
	CelEnabled    bool   `json:"cel_enabled"`
	CelExpression string `json:"cel_expression"`
}

type updateApplicationRequest struct {
	Name          *string `json:"name"`
	RuleType      *string `json:"rule_type"`
	Identifier    *string `json:"identifier"`
	Description   *string `json:"description"`
	BlockMessage  *string `json:"block_message"`
	CelEnabled    *bool   `json:"cel_enabled"`
	CelExpression *string `json:"cel_expression"`
	Enabled       *bool   `json:"enabled"`
}

type createScopeRequest struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Action     string `json:"action"`
}

func (h Handler) applicationsRoutes(r chi.Router) {
	r.Get("/", h.listApplications)
	r.Get("/check", h.checkApplication)
	r.Post("/validate", h.validateApplication)
	r.Post("/", h.createApplication)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.applicationDetails)
		r.Patch("/", h.updateApplication)
		r.Delete("/", h.deleteApplication)

		r.Route("/scopes", func(r chi.Router) {
			r.Get("/", h.listApplicationScopes)
			r.Post("/", h.createApplicationScope)
			r.Post("/validate", h.validateScopeForApplication)
			r.Delete("/{scopeID}", h.deleteApplicationScope)
		})
	})
}

func (h Handler) listApplications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := store.RuleFilter{
		Search:     r.URL.Query().Get("search"),
		RuleType:   r.URL.Query().Get("rule_type"),
		Identifier: r.URL.Query().Get("identifier"),
	}
	if enabledParam := strings.TrimSpace(r.URL.Query().Get("enabled")); enabledParam != "" {
		enabledVal, err := strconv.ParseBool(enabledParam)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid enabled value")
			return
		}
		filter.Enabled = &enabledVal
	}
	ruleset, err := h.Store.FilterRules(ctx, filter)
	if err != nil {
		h.Logger.Error("list applications", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to list applications")
		return
	}
	statsRows, err := h.Store.ListApplicationAssignmentStats(ctx)
	if err != nil {
		h.Logger.Error("list application stats", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to list applications")
		return
	}
	statsByRule := make(map[uuid.UUID]applicationAssignmentStatsDTO, len(statsRows))
	for _, row := range statsRows {
		statsByRule[row.RuleID] = mapAssignmentStats(row)
	}
	resp := make([]applicationDTO, 0, len(ruleset))
	for _, rule := range ruleset {
		dto := mapApplication(rule)
		if stats, ok := statsByRule[rule.ID]; ok {
			dto.Stats = stats
		}
		resp = append(resp, dto)
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h Handler) checkApplication(w http.ResponseWriter, r *http.Request) {
	identifier := strings.TrimSpace(r.URL.Query().Get("identifier"))
	if identifier == "" {
		respondError(w, http.StatusBadRequest, "identifier is required")
		return
	}
	rule, err := h.Store.GetRuleByTarget(r.Context(), identifier)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "application not found")
			return
		}
		h.Logger.Error("check application", "err", err)
		respondError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	respondJSON(w, http.StatusOK, mapApplication(rule))
}

func (h Handler) createApplication(w http.ResponseWriter, r *http.Request) {
	var body createApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}

	validated, fieldErrs, existing, err := h.validateApplicationInput(r.Context(), applicationValidationInput(body), nil)
	if err != nil {
		h.Logger.Error("validate application", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to create application")
		return
	}
	if len(fieldErrs) > 0 {
		code := errCodeValidationFailed
		message := "Application validation failed"
		if existing != nil {
			code = errCodeDuplicateIdentifier
			message = fmt.Sprintf("The identifier \"%s\" already belongs to \"%s\"", validated.Identifier, existing.Name)
		}
		respondValidationError(w, http.StatusUnprocessableEntity, code, message, fieldErrs, existing)
		return
	}

	metaBytes, err := h.Store.MarshalRuleMetadata(map[string]any{
		"description":    validated.Description,
		"block_message":  validated.BlockMessage,
		"cel_enabled":    validated.CelEnabled,
		"cel_expression": validated.CelExpression,
	})
	if err != nil {
		h.Logger.Error("marshal rule metadata", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to save application")
		return
	}
	rule, err := h.Store.CreateRule(r.Context(), sqlc.CreateRuleParams{
		ID:       uuid.New(),
		Name:     validated.Name,
		Type:     strings.ToLower(validated.RuleType),
		Target:   validated.Identifier,
		Scope:    string(rules.RuleScopeGroup),
		Enabled:  true,
		Metadata: metaBytes,
	})
	if err != nil {
		if isUniqueConstraintError(err, "uniq_rules_identifier_lower") {
			var existingApp *applicationDTO
			if dup, lookupErr := h.Store.GetRuleByTarget(r.Context(), validated.Identifier); lookupErr == nil {
				dto := mapApplication(dup)
				existingApp = &dto
			}
			respondValidationError(w, http.StatusUnprocessableEntity, errCodeDuplicateIdentifier, "The identifier is already in use", fieldErrors{
				"identifier": "Identifier must be unique",
			}, existingApp)
			return
		}
		h.Logger.Error("create application", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to create application")
		return
	}
	respondJSON(w, http.StatusCreated, mapApplication(rule))
}

func (h Handler) updateApplication(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid application id")
		return
	}
	var body updateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	current, err := h.Store.GetRule(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "application not found")
			return
		}
		h.Logger.Error("get application", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to load application")
		return
	}
	meta, _ := rules.ParseMetadata(current.Metadata)
	payload := applicationValidationInput{
		Name:          current.Name,
		RuleType:      current.Type,
		Identifier:    current.Target,
		Description:   meta.Description,
		BlockMessage:  meta.BlockMessage,
		CelEnabled:    meta.CelEnabled,
		CelExpression: meta.CelExpression,
	}
	if body.Name != nil {
		payload.Name = *body.Name
	}
	if body.RuleType != nil {
		payload.RuleType = *body.RuleType
	}
	if body.Identifier != nil {
		payload.Identifier = *body.Identifier
	}
	if body.Description != nil {
		payload.Description = *body.Description
	}
	if body.BlockMessage != nil {
		payload.BlockMessage = *body.BlockMessage
	}
	if body.CelEnabled != nil {
		payload.CelEnabled = *body.CelEnabled
	}
	if body.CelExpression != nil {
		payload.CelExpression = *body.CelExpression
	}
	validated, fieldErrs, existing, err := h.validateApplicationInput(r.Context(), payload, &current.ID)
	if err != nil {
		h.Logger.Error("validate application update", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to update application")
		return
	}
	if len(fieldErrs) > 0 {
		code := errCodeValidationFailed
		message := "Application validation failed"
		if existing != nil {
			code = errCodeDuplicateIdentifier
			message = fmt.Sprintf("The identifier \"%s\" already belongs to \"%s\"", validated.Identifier, existing.Name)
		}
		respondValidationError(w, http.StatusUnprocessableEntity, code, message, fieldErrs, existing)
		return
	}
	if meta.CelEnabled && !validated.CelEnabled {
		respondValidationError(w, http.StatusUnprocessableEntity, errCodeValidationFailed, "Application validation failed", fieldErrors{
			"cel_enabled": "CEL mode cannot be disabled once enabled",
		}, nil)
		return
	}
	meta.Description = validated.Description
	meta.BlockMessage = validated.BlockMessage
	meta.CelEnabled = validated.CelEnabled
	meta.CelExpression = validated.CelExpression
	metaBytes, err := h.Store.MarshalRuleMetadata(meta)
	if err != nil {
		h.Logger.Error("marshal metadata", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to update application")
		return
	}
	params := sqlc.UpdateRuleParams{
		ID:       current.ID,
		Name:     validated.Name,
		Type:     strings.ToLower(validated.RuleType),
		Target:   validated.Identifier,
		Scope:    current.Scope,
		Enabled:  current.Enabled,
		Metadata: metaBytes,
	}
	if body.Enabled != nil {
		params.Enabled = *body.Enabled
	}
	updated, err := h.Store.UpdateRule(r.Context(), params)
	if err != nil {
		h.Logger.Error("update application", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to update application")
		return
	}
	if body.Enabled != nil && !*body.Enabled && current.Enabled {
		h.requestCleanSyncAll(r.Context())
	}
	respondJSON(w, http.StatusOK, mapApplication(updated))
}

func (h Handler) deleteApplication(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid application id")
		return
	}
	if err := h.Store.DeleteRule(r.Context(), id); err != nil {
		h.Logger.Error("delete application", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to delete application")
		return
	}
	h.requestCleanSyncAll(r.Context())
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) applicationDetails(w http.ResponseWriter, r *http.Request) {
	appID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid application id")
		return
	}
	includeMembers := strings.EqualFold(r.URL.Query().Get("include_members"), "true")
	ctx := r.Context()
	rule, err := h.Store.GetRule(ctx, appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "application not found")
			return
		}
		h.Logger.Error("get application", "err", err, "app", appID)
		respondError(w, http.StatusInternalServerError, "failed to load application")
		return
	}
	scopes, err := h.Store.ListRuleScopes(ctx, appID)
	if err != nil {
		h.Logger.Error("list scopes", "err", err, "app", appID)
		respondError(w, http.StatusInternalServerError, "failed to load scopes")
		return
	}
	resolved, err := h.resolveApplicationScopes(ctx, scopes, includeMembers)
	if err != nil {
		h.Logger.Error("resolve scopes", "err", err, "app", appID)
		respondError(w, http.StatusInternalServerError, "failed to resolve scopes")
		return
	}
	appDTO := mapApplication(rule)
	appDTO.Stats = summariseScopeStats(resolved)
	resp := applicationDetailResponse{
		Application: appDTO,
		Scopes:      resolved,
	}
	respondJSON(w, http.StatusOK, resp)
}

func (h Handler) listApplicationScopes(w http.ResponseWriter, r *http.Request) {
	appID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid application id")
		return
	}
	includeMembers := strings.EqualFold(r.URL.Query().Get("include_members"), "true")
	scopes, err := h.Store.ListRuleScopes(r.Context(), appID)
	if err != nil {
		h.Logger.Error("list scopes", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to list scopes")
		return
	}
	resolved, err := h.resolveApplicationScopes(r.Context(), scopes, includeMembers)
	if err != nil {
		h.Logger.Error("resolve scopes", "err", err, "app", appID)
		respondError(w, http.StatusInternalServerError, "failed to resolve scopes")
		return
	}
	respondJSON(w, http.StatusOK, resolved)
}

func (h Handler) createApplicationScope(w http.ResponseWriter, r *http.Request) {
	appID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid application id")
		return
	}
	rule, err := h.Store.GetRule(r.Context(), appID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "application not found")
		} else {
			h.Logger.Error("get application before scope create", "err", err)
			respondError(w, http.StatusInternalServerError, "failed to verify application")
		}
		return
	}
	var body createScopeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}

	validated, fieldErrs, duplicate, err := h.validateScopeInput(r.Context(), rule, body)
	if err != nil {
		h.Logger.Error("validate scope", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to create scope")
		return
	}
	if len(fieldErrs) > 0 {
		code := errCodeValidationFailed
		message := "Scope validation failed"
		if duplicate {
			code = errCodeDuplicateScope
			message = "The selected user or group already has an assignment for this application"
		}
		respondValidationError(w, http.StatusUnprocessableEntity, code, message, fieldErrs, nil)
		return
	}

	scope, err := h.Store.CreateRuleScope(r.Context(), sqlc.InsertRuleScopeParams{
		ID:         uuid.New(),
		RuleID:     appID,
		TargetType: validated.TargetType,
		TargetID:   validated.TargetID,
		Action:     validated.Action,
	})
	if err != nil {
		if isUniqueConstraintError(err, "uniq_rule_scopes_per_target") {
			respondValidationError(w, http.StatusUnprocessableEntity, errCodeDuplicateScope, "The selected user or group already has an assignment for this application", fieldErrors{
				"target_id": "Selected target already has an assignment",
			}, nil)
			return
		}
		h.Logger.Error("create scope", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to create scope")
		return
	}
	go h.recompileRuleAssignments(context.WithoutCancel(r.Context()), rule)
	respondJSON(w, http.StatusCreated, mapScope(scope))
}

func (h Handler) deleteApplicationScope(w http.ResponseWriter, r *http.Request) {
	appID, err := parseUUIDParam(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid application id")
		return
	}
	scopeID, err := parseUUIDParam(r, "scopeID")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid scope id")
		return
	}
	scope, err := h.Store.GetRuleScope(r.Context(), scopeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.Logger.Error("get scope", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to delete scope")
		return
	}
	if scope.RuleID != appID {
		respondError(w, http.StatusBadRequest, "scope does not belong to application")
		return
	}
	if err := h.Store.DeleteRuleScope(r.Context(), scopeID); err != nil {
		h.Logger.Error("delete scope", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to delete scope")
		return
	}
	h.requestCleanSyncForScope(r.Context(), scope)
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) resolveApplicationScopes(ctx context.Context, scopes []sqlc.RuleScope, includeMembers bool) ([]applicationScopeRelationshipDTO, error) {
	resp := make([]applicationScopeRelationshipDTO, 0, len(scopes))
	userCache := make(map[uuid.UUID]sqlc.User)
	groupCache := make(map[uuid.UUID]sqlc.Group)
	groupMemberIDs := make(map[uuid.UUID][]uuid.UUID)
	groupMemberDetails := make(map[uuid.UUID][]sqlc.User)

	for _, scope := range scopes {
		dto := applicationScopeRelationshipDTO{
			applicationScopeDTO: mapScope(scope),
			EffectiveMemberIDs:  []uuid.UUID{},
		}
		switch scope.TargetType {
		case "user":
			user, err := h.getCachedUser(ctx, userCache, scope.TargetID)
			if err != nil {
				return nil, err
			}
			dto.TargetDisplayName = user.DisplayName
			dto.TargetUPN = user.Upn
			dto.EffectiveMemberIDs = []uuid.UUID{user.ID}
			dto.EffectiveMemberCount = 1
			if includeMembers {
				dto.EffectiveMembers = []userDTO{mapUserDTO(user)}
			}
		case "group":
			group, err := h.getCachedGroup(ctx, groupCache, scope.TargetID)
			if err != nil {
				return nil, err
			}
			dto.TargetDisplayName = group.DisplayName
			dto.TargetDescription = group.Description.String
			memberIDs, ok := groupMemberIDs[group.ID]
			if !ok {
				memberIDs, err = h.Store.ListGroupMemberIDs(ctx, group.ID)
				if err != nil {
					return nil, err
				}
				groupMemberIDs[group.ID] = memberIDs
			}
			if len(memberIDs) > 0 {
				dto.EffectiveMemberIDs = append([]uuid.UUID(nil), memberIDs...)
			} else {
				dto.EffectiveMemberIDs = []uuid.UUID{}
			}
			dto.EffectiveMemberCount = len(memberIDs)
			if includeMembers {
				members, ok := groupMemberDetails[group.ID]
				if !ok {
					members, err = h.Store.ListGroupMembers(ctx, group.ID)
					if err != nil {
						return nil, err
					}
					groupMemberDetails[group.ID] = members
				}
				dto.EffectiveMembers = mapUserList(members)
			}
		default:
			dto.EffectiveMemberCount = 0
		}
		resp = append(resp, dto)
	}
	return resp, nil
}

func (h Handler) getCachedUser(ctx context.Context, cache map[uuid.UUID]sqlc.User, userID uuid.UUID) (sqlc.User, error) {
	if user, ok := cache[userID]; ok {
		return user, nil
	}
	user, err := h.Store.GetUser(ctx, userID)
	if err != nil {
		return sqlc.User{}, err
	}
	cache[userID] = user
	return user, nil
}

func (h Handler) getCachedGroup(ctx context.Context, cache map[uuid.UUID]sqlc.Group, groupID uuid.UUID) (sqlc.Group, error) {
	if group, ok := cache[groupID]; ok {
		return group, nil
	}
	group, err := h.Store.GetGroup(ctx, groupID)
	if err != nil {
		return sqlc.Group{}, err
	}
	cache[groupID] = group
	return group, nil
}

func mapApplication(rule sqlc.Rule) applicationDTO {
	meta, _ := rules.ParseMetadata(rule.Metadata)
	var createdAt, updatedAt time.Time
	if rule.CreatedAt.Valid {
		createdAt = rule.CreatedAt.Time
	}
	if rule.UpdatedAt.Valid {
		updatedAt = rule.UpdatedAt.Time
	}
	return applicationDTO{
		ID:            rule.ID,
		Name:          rule.Name,
		RuleType:      strings.ToUpper(rule.Type),
		Identifier:    rule.Target,
		Description:   meta.Description,
		BlockMessage:  meta.BlockMessage,
		CelEnabled:    meta.CelEnabled,
		CelExpression: meta.CelExpression,
		Enabled:       rule.Enabled,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

func mapAssignmentStats(row sqlc.ListApplicationAssignmentStatsRow) applicationAssignmentStatsDTO {
	return applicationAssignmentStatsDTO{
		AllowScopes:    int(row.AllowScopes),
		BlockScopes:    int(row.BlockScopes),
		CelScopes:      int(row.CelScopes),
		TotalScopes:    int(row.TotalScopes),
		AllowUsers:     int(row.AllowUsers),
		BlockUsers:     int(row.BlockUsers),
		CelUsers:       int(row.CelUsers),
		TotalUsers:     int(row.TotalUsers),
		TotalMachines:  int(row.TotalMachines),
		SyncedMachines: int(row.SyncedMachines),
	}
}

func summariseScopeStats(scopes []applicationScopeRelationshipDTO) applicationAssignmentStatsDTO {
	stats := applicationAssignmentStatsDTO{}
	allowUsers := make(map[uuid.UUID]struct{})
	blockUsers := make(map[uuid.UUID]struct{})
	celUsers := make(map[uuid.UUID]struct{})
	totalUsers := make(map[uuid.UUID]struct{})

	for _, scope := range scopes {
		switch scope.Action {
		case "allow":
			stats.AllowScopes++
		case "block":
			stats.BlockScopes++
		case "cel":
			stats.CelScopes++
		}
		stats.TotalScopes++

		for _, id := range scope.EffectiveMemberIDs {
			totalUsers[id] = struct{}{}
			switch scope.Action {
			case "allow":
				allowUsers[id] = struct{}{}
			case "block":
				blockUsers[id] = struct{}{}
			case "cel":
				celUsers[id] = struct{}{}
			}
		}
	}

	stats.AllowUsers = len(allowUsers)
	stats.BlockUsers = len(blockUsers)
	stats.CelUsers = len(celUsers)
	stats.TotalUsers = len(totalUsers)
	stats.TotalMachines = 0
	stats.SyncedMachines = 0
	return stats
}

func mapScope(scope sqlc.RuleScope) applicationScopeDTO {
	var created time.Time
	if scope.CreatedAt.Valid {
		created = scope.CreatedAt.Time
	}
	return applicationScopeDTO{
		ID:            scope.ID,
		ApplicationID: scope.RuleID,
		TargetType:    scope.TargetType,
		TargetID:      scope.TargetID,
		Action:        scope.Action,
		CreatedAt:     created,
	}
}

func (h Handler) requestCleanSyncAll(ctx context.Context) {
	if err := h.Store.RequestCleanSyncAllMachines(ctx); err != nil {
		h.Logger.Warn("request global clean sync", "err", err)
	}
}

func (h Handler) requestCleanSyncForScope(ctx context.Context, scope sqlc.RuleScope) {
	var err error
	switch scope.TargetType {
	case "user":
		err = h.Store.RequestCleanSyncForUser(ctx, scope.TargetID)
	case "group":
		err = h.Store.RequestCleanSyncForGroup(ctx, scope.TargetID)
	default:
		err = h.Store.RequestCleanSyncAllMachines(ctx)
	}
	if err != nil {
		h.Logger.Warn("request scoped clean sync", "scope", scope.ID, "err", err)
	}
}
