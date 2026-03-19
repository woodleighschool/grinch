package httpapi

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/app/rules"
	"github.com/woodleighschool/grinch/internal/domain"
)

type ruleWriteRequestBody struct {
	ID            *string            `json:"id,omitempty"`
	Name          string             `json:"name"`
	Description   *string            `json:"description,omitempty"`
	RuleType      domain.RuleType    `json:"rule_type"`
	Identifier    string             `json:"identifier"`
	CustomMessage *string            `json:"custom_message,omitempty"`
	CustomURL     *string            `json:"custom_url,omitempty"`
	Enabled       *bool              `json:"enabled,omitempty"`
	Targets       domain.RuleTargets `json:"targets"`
	CreatedAt     *string            `json:"created_at,omitempty"`
	UpdatedAt     *string            `json:"updated_at,omitempty"`
}

func (handler *Server) ListRules(writer http.ResponseWriter, request *http.Request, params ListRulesParams) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	items, total, err := handler.rules.ListRules(request.Context(), domain.RuleListOptions{ListOptions: listOptions})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, RuleListResponse{
		Rows:  items,
		Total: total,
	})
}

func (handler *Server) CreateRule(writer http.ResponseWriter, request *http.Request) {
	var body ruleWriteRequestBody
	if err := decodeJSONBody(request, &body); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	input := decodeRuleWriteRequest(body)

	rule, err := handler.rules.CreateRule(request.Context(), input)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusCreated, rule)
}

func (handler *Server) GetRule(writer http.ResponseWriter, request *http.Request, id Id) {
	rule, err := handler.rules.GetRule(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "rule not found"})
		return
	}

	writeJSON(writer, http.StatusOK, rule)
}

func (handler *Server) UpdateRule(writer http.ResponseWriter, request *http.Request, id Id) {
	var body ruleWriteRequestBody
	if err := decodeJSONBody(request, &body); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	input := decodeRuleWriteRequest(body)

	updated, err := handler.rules.UpdateRule(request.Context(), id, input)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{
			NotFoundMessage: "rule not found",
		})
		return
	}

	writeJSON(writer, http.StatusOK, updated)
}

func (handler *Server) DeleteRule(writer http.ResponseWriter, request *http.Request, id Id) {
	if err := handler.rules.DeleteRule(request.Context(), id); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "rule not found"})
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func decodeRuleWriteRequest(body ruleWriteRequestBody) rules.WriteInput {
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	return rules.WriteInput{
		CustomMessage: optionalStringValue(body.CustomMessage),
		CustomURL:     optionalStringValue(body.CustomURL),
		Description:   optionalStringValue(body.Description),
		Enabled:       enabled,
		Identifier:    body.Identifier,
		Name:          body.Name,
		RuleType:      body.RuleType,
		Targets:       decodeRuleTargets(body.Targets),
	}
}

func decodeRuleTargets(targets domain.RuleTargets) rules.TargetsWriteInput {
	include := make([]rules.IncludeTargetWriteInput, 0, len(targets.Include))
	for _, target := range targets.Include {
		include = append(include, rules.IncludeTargetWriteInput{
			SubjectKind:   target.SubjectKind,
			SubjectID:     target.SubjectID,
			Policy:        target.Policy,
			CELExpression: target.CELExpression,
		})
	}

	exclude := make([]rules.ExcludedGroupWriteInput, 0, len(targets.Exclude))
	for _, target := range targets.Exclude {
		exclude = append(exclude, rules.ExcludedGroupWriteInput{GroupID: target.GroupID})
	}

	return rules.TargetsWriteInput{Include: include, Exclude: exclude}
}
