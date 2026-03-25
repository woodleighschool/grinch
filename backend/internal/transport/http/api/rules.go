package apihttp

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/domain"
)

type ruleWriteRequestBody struct {
	Name          string             `json:"name"`
	Description   *string            `json:"description,omitempty"`
	RuleType      domain.RuleType    `json:"rule_type"`
	Identifier    string             `json:"identifier"`
	CustomMessage *string            `json:"custom_message,omitempty"`
	CustomURL     *string            `json:"custom_url,omitempty"`
	Enabled       *bool              `json:"enabled,omitempty"`
	Targets       domain.RuleTargets `json:"targets"`
}

func (s *Server) ListRules(w http.ResponseWriter, r *http.Request, params ListRulesParams) {
	listOptions, err := parseListOptions(
		params.Limit,
		params.Offset,
		params.Search,
		params.Sort,
		params.Order,
		params.Ids,
	)
	if err != nil {
		writeError(w, err)
		return
	}

	ruleTypes, err := parseOptionalValues(params.RuleType, domain.ParseRuleType)
	if err != nil {
		writeError(w, err)
		return
	}

	items, total, err := s.rules.ListRules(r.Context(), domain.RuleListOptions{
		ListOptions: listOptions,
		Enabled:     cloneBools(params.Enabled),
		RuleTypes:   ruleTypes,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, RuleListResponse{
		Rows:  items,
		Total: total,
	})
}

func (s *Server) CreateRule(w http.ResponseWriter, r *http.Request) {
	var body ruleWriteRequestBody
	if err := decodeJSONBody(r, &body); err != nil {
		writeError(w, err)
		return
	}

	rule, err := s.rules.CreateRule(r.Context(), decodeRuleWriteRequest(body))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, rule)
}

func (s *Server) GetRule(w http.ResponseWriter, r *http.Request, id Id) {
	rule, err := s.rules.GetRule(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, rule)
}

func (s *Server) UpdateRule(w http.ResponseWriter, r *http.Request, id Id) {
	var body ruleWriteRequestBody
	if err := decodeJSONBody(r, &body); err != nil {
		writeError(w, err)
		return
	}

	updated, err := s.rules.UpdateRule(r.Context(), id, decodeRuleWriteRequest(body))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) DeleteRule(w http.ResponseWriter, r *http.Request, id Id) {
	if err := s.rules.DeleteRule(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	writeNoContent(w)
}

func decodeRuleWriteRequest(body ruleWriteRequestBody) domain.RuleWriteInput {
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	include := make([]domain.IncludeRuleTargetWriteInput, 0, len(body.Targets.Include))
	for _, t := range body.Targets.Include {
		include = append(include, domain.IncludeRuleTargetWriteInput{
			SubjectKind:   t.SubjectKind,
			SubjectID:     t.SubjectID,
			Policy:        t.Policy,
			CELExpression: t.CELExpression,
		})
	}

	exclude := make([]domain.ExcludedGroupWriteInput, 0, len(body.Targets.Exclude))
	for _, t := range body.Targets.Exclude {
		exclude = append(exclude, domain.ExcludedGroupWriteInput{GroupID: t.GroupID})
	}

	return domain.RuleWriteInput{
		CustomMessage: optionalString(body.CustomMessage),
		CustomURL:     optionalString(body.CustomURL),
		Description:   optionalString(body.Description),
		Enabled:       enabled,
		Identifier:    body.Identifier,
		Name:          body.Name,
		RuleType:      body.RuleType,
		Targets: domain.RuleTargetsWriteInput{
			Include: include,
			Exclude: exclude,
		},
	}
}
