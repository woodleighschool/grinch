package admin

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/store/sqlc"
)

type ruleDTO struct {
	ID       uuid.UUID      `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Target   string         `json:"target"`
	Scope    string         `json:"scope"`
	Enabled  bool           `json:"enabled"`
	Metadata map[string]any `json:"metadata"`
}

func (h Handler) rulesRoutes(r chi.Router) {
	r.Get("/", h.listRules)
	r.Post("/", h.createRule)
	r.Put("/{id}", h.updateRule)
	r.Delete("/{id}", h.deleteRule)
}

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
	respondJSON(w, http.StatusCreated, mapRule(rule))
}

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
	respondJSON(w, http.StatusOK, mapRule(rule))
}

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
	respondJSON(w, http.StatusNoContent, nil)
}

func mapRule(rule sqlc.Rule) ruleDTO {
	var metadata map[string]any
	_ = json.Unmarshal(rule.Metadata, &metadata)
	return ruleDTO{
		ID:       rule.ID,
		Name:     rule.Name,
		Type:     rule.Type,
		Target:   rule.Target,
		Scope:    rule.Scope,
		Enabled:  rule.Enabled,
		Metadata: metadata,
	}
}
