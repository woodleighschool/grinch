package httpapi

import (
	"net/http"

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

	subjectKindValue, hasSubjectKind, err := decodeOptional(params.SubjectKind, toDomainRuleTargetSubjectKind)
	if err != nil {
		writeClassifiedError(writer, badRequestError("invalid subject_kind"), apiErrorOptions{})
		return
	}
	var subjectKind *domain.RuleTargetSubjectKind
	if hasSubjectKind {
		subjectKind = &subjectKindValue
	}

	assignmentValue, hasAssignment, err := decodeOptional(params.Assignment, toDomainRuleTargetAssignment)
	if err != nil {
		writeClassifiedError(writer, badRequestError("invalid assignment"), apiErrorOptions{})
		return
	}
	var assignment *domain.RuleTargetAssignment
	if hasAssignment {
		assignment = &assignmentValue
	}

	policyValue, hasPolicy, err := decodeOptional(params.Policy, toDomainRulePolicy)
	if err != nil {
		writeClassifiedError(writer, badRequestError("invalid policy"), apiErrorOptions{})
		return
	}
	var policy *domain.RulePolicy
	if hasPolicy {
		policy = &policyValue
	}

	items, total, err := handler.ruleTargets.ListRuleTargets(request.Context(), domain.RuleTargetListOptions{
		ListOptions: listOptions,
		RuleID:      params.RuleId,
		SubjectKind: subjectKind,
		SubjectID:   params.SubjectId,
		Assignment:  assignment,
		Policy:      policy,
	})
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
	}

	mapped, err := mapSlice(items, mapRuleTargetSummary)
	if err != nil {
		writeClassifiedError(writer, err, apiErrorOptions{})
		return
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
