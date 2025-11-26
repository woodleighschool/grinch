package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/rules"
	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

// ruleDTO represents a raw Santa rule, mainly for debugging endpoints.
type ruleDTO struct {
	ID       uuid.UUID      `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Target   string         `json:"target"`
	Scope    string         `json:"scope"`
	Enabled  bool           `json:"enabled"`
	Metadata map[string]any `json:"metadata"`
}

// rulesRoutes exposes legacy debugging endpoints for raw rules.
func (h Handler) rulesRoutes(r chi.Router) {
	r.Get("/", h.listRules)
	r.Post("/", h.createRule)
	r.Put("/{id}", h.updateRule)
	r.Delete("/{id}", h.deleteRule)
}

// listRules streams every rule, primarily for troubleshooting.
func (h Handler) listRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.Store.ListRules(r.Context())
	if err != nil {
		h.Logger.Error("list rules", "err", err)
		respondError(w, http.StatusInternalServerError, "failed to list rules")
		return
	}
	resp := make([]ruleDTO, 0, len(rules))
	for _, rule := range rules {
		resp = append(resp, mapRule(rule))
	}
	respondJSON(w, http.StatusOK, resp)
}

// createRule stores a debugging rule and recompiles assignments out-of-band.
func (h Handler) createRule(w http.ResponseWriter, r *http.Request) {
	var body ruleDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.ID == uuid.Nil {
		body.ID = uuid.New()
	}
	payload, err := json.Marshal(body.Metadata)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid metadata")
		return
	}
	rule, err := h.Store.CreateRule(r.Context(), sqlc.CreateRuleParams{
		ID:       body.ID,
		Name:     body.Name,
		Type:     body.Type,
		Target:   body.Target,
		Scope:    body.Scope,
		Enabled:  body.Enabled,
		Metadata: payload,
	})
	if err != nil {
		h.Logger.Error("create rule", "err", err)
		respondError(w, http.StatusInternalServerError, "create failed")
		return
	}
	go h.recompileRuleAssignments(context.WithoutCancel(r.Context()), rule)
	respondJSON(w, http.StatusCreated, mapRule(rule))
}

// updateRule mutates an existing rule; primarily used in development.
func (h Handler) updateRule(w http.ResponseWriter, r *http.Request) {
	var body ruleDTO
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	payload, err := json.Marshal(body.Metadata)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid metadata")
		return
	}
	rule, err := h.Store.UpdateRule(r.Context(), sqlc.UpdateRuleParams{
		ID:       id,
		Name:     body.Name,
		Type:     body.Type,
		Target:   body.Target,
		Scope:    body.Scope,
		Enabled:  body.Enabled,
		Metadata: payload,
	})
	if err != nil {
		h.Logger.Error("update rule", "err", err)
		respondError(w, http.StatusInternalServerError, "update failed")
		return
	}
	go h.recompileRuleAssignments(context.WithoutCancel(r.Context()), rule)
	respondJSON(w, http.StatusOK, mapRule(rule))
}

// deleteRule removes a rule; legacy endpoint retained for compatibility.
func (h Handler) deleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.Store.DeleteRule(r.Context(), id); err != nil {
		h.Logger.Error("delete rule", "err", err)
		respondError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// recompileRuleAssignments rebuilds compiled assignments for a single rule.
func (h Handler) recompileRuleAssignments(ctx context.Context, rule sqlc.Rule) {
	meta, err := rules.ParseMetadata(rule.Metadata)
	if err != nil {
		h.Logger.Warn("skip rule metadata", "rule", rule.ID, "err", err)
		return
	}
	ruleScopes, err := h.Store.ListRuleScopes(ctx, rule.ID)
	if err != nil {
		h.Logger.Error("list rule scopes", "rule", rule.ID, "err", err)
		return
	}
	memberMap := map[uuid.UUID][]uuid.UUID{}
	for _, groupID := range rules.CollectGroupIDs(meta, ruleScopes) {
		members, err := h.Store.ListGroupMemberIDs(ctx, groupID)
		if err != nil {
			h.Logger.Warn("list group members", "group", groupID, "err", err)
			continue
		}
		memberMap[groupID] = members
	}
	assignments := h.Compiler.CompileAssignments(rule, ruleScopes, meta, memberMap)
	if err := h.Store.ReplaceRuleAssignments(ctx, rule.ID, assignments); err != nil {
		h.Logger.Error("replace rule assignments", "rule", rule.ID, "err", err)
	}
}

// recompileGroupRules refreshes compiled assignments for every rule targeting the group.
func (h Handler) recompileGroupRules(ctx context.Context, groupID uuid.UUID) {
	ruleIDs, err := h.Store.ListRulesByGroupTarget(ctx, groupID)
	if err != nil {
		h.Logger.Error("list rules by group target", "group", groupID, "err", err)
		return
	}
	for _, ruleID := range ruleIDs {
		rule, err := h.Store.GetRule(ctx, ruleID)
		if err != nil {
			h.Logger.Error("get rule for recompilation", "rule", ruleID, "err", err)
			continue
		}
		h.recompileRuleAssignments(ctx, rule)
	}
}

// mapRule flattens the sqlc rule row for the legacy API.
func mapRule(r sqlc.Rule) ruleDTO {
	return ruleDTO{
		ID:       r.ID,
		Name:     r.Name,
		Type:     r.Type,
		Target:   r.Target,
		Scope:    r.Scope,
		Enabled:  r.Enabled,
		Metadata: rules.SerialiseMetadata(r.Metadata),
	}
}
