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

func (handler *Server) ListRules(writer http.ResponseWriter, request *http.Request, params ListRulesParams) {
	listOptions, err := parseListOptions(
		params.Limit,
		params.Offset,
		params.Search,
		params.Sort,
		params.Order,
		params.Ids,
	)
	if err != nil {
		writeError(writer, err)
		return
	}

	ruleTypes, err := parseOptionalValues(params.RuleType, domain.ParseRuleType)
	if err != nil {
		writeError(writer, err)
		return
	}

	items, total, err := handler.rules.ListRules(request.Context(), domain.RuleListOptions{
		ListOptions: listOptions,
		Enabled:     cloneBools(params.Enabled),
		RuleTypes:   ruleTypes,
	})
	if err != nil {
		writeError(writer, err)
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
		writeError(writer, err)
		return
	}

	rule, err := handler.rules.CreateRule(request.Context(), decodeRuleWriteRequest(body))
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusCreated, rule)
}

func (handler *Server) GetRule(writer http.ResponseWriter, request *http.Request, id Id) {
	rule, err := handler.rules.GetRule(request.Context(), id)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, rule)
}

func (handler *Server) UpdateRule(writer http.ResponseWriter, request *http.Request, id Id) {
	var body ruleWriteRequestBody
	if err := decodeJSONBody(request, &body); err != nil {
		writeError(writer, err)
		return
	}

	updated, err := handler.rules.UpdateRule(request.Context(), id, decodeRuleWriteRequest(body))
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJSON(writer, http.StatusOK, updated)
}

func (handler *Server) DeleteRule(writer http.ResponseWriter, request *http.Request, id Id) {
	if err := handler.rules.DeleteRule(request.Context(), id); err != nil {
		writeError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
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
