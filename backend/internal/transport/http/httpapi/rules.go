package httpapi

import (
	"net/http"

	"github.com/woodleighschool/grinch/internal/app/rules"
	"github.com/woodleighschool/grinch/internal/domain"
)

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

	mapped := make([]RuleSummary, 0, len(items))
	for _, item := range items {
		output, mapErr := mapRuleSummary(item)
		if mapErr != nil {
			writeClassifiedError(writer, mapErr, apiErrorOptions{})
			return
		}
		mapped = append(mapped, output)
	}

	writeJSON(writer, http.StatusOK, RuleListResponse{
		Rows:  mapped,
		Total: total,
	})
}

func (handler *Server) CreateRule(writer http.ResponseWriter, request *http.Request) {
	var body CreateRuleJSONRequestBody
	if err := decodeJSONBody(request, &body); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	input, err := decodeRuleWriteRequest(body)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	rule, err := handler.rules.CreateRule(request.Context(), input)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	mapped, err := mapRule(rule)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusCreated, mapped)
}

func (handler *Server) GetRule(writer http.ResponseWriter, request *http.Request, id Id) {
	rule, err := handler.rules.GetRule(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "rule not found"})
		return
	}

	mapped, err := mapRule(rule)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, mapped)
}

func (handler *Server) PatchRule(writer http.ResponseWriter, request *http.Request, id Id) {
	var body PatchRuleJSONRequestBody
	if err := decodeJSONBody(request, &body); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	input, err := decodeRulePatchRequest(body)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	updated, err := handler.rules.PatchRule(request.Context(), id, input)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{
			NotFoundMessage: "rule not found",
		})
		return
	}

	mapped, err := mapRule(updated)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, mapped)
}

func (handler *Server) DeleteRule(writer http.ResponseWriter, request *http.Request, id Id) {
	if err := handler.rules.DeleteRule(request.Context(), id); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "rule not found"})
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func decodeRuleWriteRequest(body CreateRuleJSONRequestBody) (rules.RuleCreateInput, error) {
	ruleType, err := toDomainRuleType(body.RuleType)
	if err != nil {
		return rules.RuleCreateInput{}, badRequestError("invalid rule_type")
	}

	return rules.RuleCreateInput{
		CustomMessage: optionalStringValue(body.CustomMessage),
		CustomURL:     optionalStringValue(body.CustomUrl),
		Description:   optionalStringValue(body.Description),
		Identifier:    body.Identifier,
		Name:          body.Name,
		RuleType:      ruleType,
	}, nil
}

func decodeRulePatchRequest(body PatchRuleJSONRequestBody) (rules.RulePatchInput, error) {
	result := rules.RulePatchInput{
		Name:          body.Name,
		Description:   body.Description,
		Identifier:    body.Identifier,
		CustomMessage: body.CustomMessage,
		CustomURL:     body.CustomUrl,
	}
	if body.RuleType != nil {
		ruleType, err := toDomainRuleType(*body.RuleType)
		if err != nil {
			return rules.RulePatchInput{}, badRequestError("invalid rule_type")
		}
		result.RuleType = &ruleType
	}
	return result, nil
}
