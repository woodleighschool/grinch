package httpapi

import (
	"net/http"

	"github.com/google/uuid"

	appruletargets "github.com/woodleighschool/grinch/internal/app/ruletargets"
	"github.com/woodleighschool/grinch/internal/domain"
)

func (handler *Server) ListRuleTargets(
	writer http.ResponseWriter,
	request *http.Request,
	params ListRuleTargetsParams,
) {
	listOptions, err := parseListOptions(params.Limit, params.Offset, params.Search, params.Sort, params.Order)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	var ruleID *uuid.UUID
	if params.RuleId != nil {
		value := *params.RuleId
		ruleID = &value
	}
	var subjectKind *domain.RuleTargetSubjectKind
	if params.SubjectKind != nil {
		value, valueErr := toDomainRuleTargetSubjectKind(*params.SubjectKind)
		if valueErr != nil {
			writeClassifiedError(writer, badRequestError("invalid subject_kind"), apiErrorOptions{})
			return
		}
		subjectKind = &value
	}
	var subjectID *uuid.UUID
	if params.SubjectId != nil {
		value := *params.SubjectId
		subjectID = &value
	}
	var assignment *domain.RuleTargetAssignment
	if params.Assignment != nil {
		value, valueErr := toDomainRuleTargetAssignment(*params.Assignment)
		if valueErr != nil {
			writeClassifiedError(writer, badRequestError("invalid assignment"), apiErrorOptions{})
			return
		}
		assignment = &value
	}
	var policy *domain.RulePolicy
	if params.Policy != nil {
		value, valueErr := toDomainRulePolicy(*params.Policy)
		if valueErr != nil {
			writeClassifiedError(writer, badRequestError("invalid policy"), apiErrorOptions{})
			return
		}
		policy = &value
	}

	items, total, err := handler.ruleTargets.ListRuleTargets(request.Context(), domain.RuleTargetListOptions{
		ListOptions: listOptions,
		RuleID:      ruleID,
		SubjectKind: subjectKind,
		SubjectID:   subjectID,
		Assignment:  assignment,
		Policy:      policy,
	})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	mapped := make([]RuleTargetSummary, 0, len(items))
	for _, item := range items {
		output, mapErr := mapRuleTargetSummary(item)
		if mapErr != nil {
			writeClassifiedError(writer, mapErr, apiErrorOptions{})
			return
		}
		mapped = append(mapped, output)
	}

	writeJSON(writer, http.StatusOK, RuleTargetListResponse{
		Rows:  mapped,
		Total: total,
	})
}

func (handler *Server) CreateRuleTarget(writer http.ResponseWriter, request *http.Request) {
	var body CreateRuleTargetJSONRequestBody
	if err := decodeJSONBody(request, &body); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	input, err := decodeRuleTargetCreate(body)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	target, err := handler.ruleTargets.CreateRuleTarget(request.Context(), input)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "rule target dependencies not found"})
		return
	}

	mapped, err := mapRuleTarget(target)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusCreated, mapped)
}

func (handler *Server) GetRuleTarget(writer http.ResponseWriter, request *http.Request, id Id) {
	target, err := handler.ruleTargets.GetRuleTarget(request.Context(), id)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "rule target not found"})
		return
	}

	mapped, err := mapRuleTarget(target)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, mapped)
}

func (handler *Server) PatchRuleTarget(writer http.ResponseWriter, request *http.Request, id Id) {
	var body PatchRuleTargetJSONRequestBody
	if err := decodeJSONBody(request, &body); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	input, err := decodeRuleTargetPatch(body)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	target, err := handler.ruleTargets.PatchRuleTarget(request.Context(), id, input)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "rule target not found"})
		return
	}

	mapped, err := mapRuleTarget(target)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	writeJSON(writer, http.StatusOK, mapped)
}

func (handler *Server) DeleteRuleTarget(writer http.ResponseWriter, request *http.Request, id Id) {
	if err := handler.ruleTargets.DeleteRuleTarget(request.Context(), id); err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{NotFoundMessage: "rule target not found"})
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func decodeRuleTargetCreate(body CreateRuleTargetJSONRequestBody) (appruletargets.WriteInput, error) {
	assignment, err := toDomainRuleTargetAssignment(body.Assignment)
	if err != nil {
		return appruletargets.WriteInput{}, badRequestError("invalid assignment")
	}

	input := appruletargets.WriteInput{
		RuleID:     body.RuleId,
		SubjectID:  body.SubjectId,
		Assignment: assignment,
		Priority:   body.Priority,
	}
	if body.Policy != nil {
		policy, policyErr := toDomainRulePolicy(*body.Policy)
		if policyErr != nil {
			return appruletargets.WriteInput{}, badRequestError("invalid policy")
		}
		input.Policy = &policy
	}
	if body.CelExpression != nil {
		input.CELExpression = *body.CelExpression
	}
	return input, nil
}

func decodeRuleTargetPatch(body PatchRuleTargetJSONRequestBody) (appruletargets.PatchInput, error) {
	input := appruletargets.PatchInput{}
	input.SubjectID = body.SubjectId
	if body.Assignment != nil {
		assignment, err := toDomainRuleTargetAssignment(*body.Assignment)
		if err != nil {
			return appruletargets.PatchInput{}, badRequestError("invalid assignment")
		}
		input.Assignment = &assignment
		if assignment == domain.RuleTargetAssignmentExclude {
			input.Priority = new(*int32)
			input.Policy = new(*domain.RulePolicy)
			empty := ""
			input.CELExpression = &empty
		}
	}
	if body.Priority != nil {
		input.Priority = &body.Priority
	}
	if body.Policy != nil {
		policy, err := toDomainRulePolicy(*body.Policy)
		if err != nil {
			return appruletargets.PatchInput{}, badRequestError("invalid policy")
		}
		policyValue := &policy
		input.Policy = &policyValue
	}
	if body.CelExpression != nil {
		input.CELExpression = body.CelExpression
	}
	return input, nil
}
