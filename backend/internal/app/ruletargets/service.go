package ruletargets

import (
	"context"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type WriteInput struct {
	RuleID        uuid.UUID
	SubjectID     uuid.UUID
	Assignment    domain.RuleTargetAssignment
	Priority      *int32
	Policy        *domain.RulePolicy
	CELExpression string
}

type PatchInput struct {
	SubjectID     *uuid.UUID
	Assignment    *domain.RuleTargetAssignment
	Priority      **int32
	Policy        **domain.RulePolicy
	CELExpression *string
}

type Store interface {
	ListRuleTargets(context.Context, domain.RuleTargetListOptions) ([]domain.RuleTargetSummary, int32, error)
	GetRuleTarget(context.Context, uuid.UUID) (domain.RuleTarget, error)
	CreateRuleTarget(context.Context, WriteInput) (domain.RuleTarget, error)
	PatchRuleTarget(context.Context, uuid.UUID, WriteInput) (domain.RuleTarget, error)
	DeleteRuleTarget(context.Context, uuid.UUID) error
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (service *Service) ListRuleTargets(
	ctx context.Context,
	options domain.RuleTargetListOptions,
) ([]domain.RuleTargetSummary, int32, error) {
	return service.store.ListRuleTargets(ctx, options)
}

func (service *Service) GetRuleTarget(ctx context.Context, id uuid.UUID) (domain.RuleTarget, error) {
	return service.store.GetRuleTarget(ctx, id)
}

func (service *Service) CreateRuleTarget(ctx context.Context, input WriteInput) (domain.RuleTarget, error) {
	if validationErr := service.validate(input); validationErr != nil {
		return domain.RuleTarget{}, validationErr
	}

	return service.store.CreateRuleTarget(ctx, input)
}

func (service *Service) PatchRuleTarget(
	ctx context.Context,
	id uuid.UUID,
	patch PatchInput,
) (domain.RuleTarget, error) {
	current, err := service.store.GetRuleTarget(ctx, id)
	if err != nil {
		return domain.RuleTarget{}, err
	}

	input := WriteInput{
		RuleID:        current.RuleID,
		SubjectID:     current.SubjectID,
		Assignment:    current.Assignment,
		Priority:      current.Priority,
		Policy:        current.Policy,
		CELExpression: current.CELExpression,
	}

	if patch.SubjectID != nil {
		input.SubjectID = *patch.SubjectID
	}
	if patch.Assignment != nil {
		input.Assignment = *patch.Assignment
	}
	if patch.Priority != nil {
		input.Priority = *patch.Priority
	}
	if patch.Policy != nil {
		input.Policy = *patch.Policy
	}
	if patch.CELExpression != nil {
		input.CELExpression = *patch.CELExpression
	}

	if validationErr := service.validate(input); validationErr != nil {
		return domain.RuleTarget{}, validationErr
	}

	return service.store.PatchRuleTarget(ctx, id, input)
}

func (service *Service) DeleteRuleTarget(ctx context.Context, id uuid.UUID) error {
	return service.store.DeleteRuleTarget(ctx, id)
}

func (service *Service) validate(input WriteInput) *domain.ValidationError {
	err := &domain.ValidationError{
		Code:   "validation_error",
		Detail: "Rule target is invalid.",
	}

	service.validateAssignment(input, err)

	if !err.HasFieldErrors() {
		return nil
	}
	return err
}

func (service *Service) validateAssignment(input WriteInput, err *domain.ValidationError) {
	switch input.Assignment {
	case domain.RuleTargetAssignmentInclude:
		service.validateIncludeTarget(input, err)
	case domain.RuleTargetAssignmentExclude:
		service.validateExcludeTarget(input, err)
	default:
		err.Add("assignment", "must be include or exclude", "invalid")
	}
}

func (service *Service) validateIncludeTarget(input WriteInput, err *domain.ValidationError) {
	if input.Priority == nil {
		err.Add("priority", "is required for include targets", "required")
	}
	if input.Policy == nil {
		err.Add("policy", "is required for include targets", "required")
		return
	}
	if *input.Policy == domain.RulePolicyCEL && input.CELExpression == "" {
		err.Add("cel_expression", "is required when policy is cel", "required")
	}
	if *input.Policy != domain.RulePolicyCEL && input.CELExpression != "" {
		err.Add("cel_expression", "must be empty unless policy is cel", "invalid")
	}
}

func (service *Service) validateExcludeTarget(input WriteInput, err *domain.ValidationError) {
	if input.Priority != nil {
		err.Add("priority", "must be empty for exclude targets", "invalid")
	}
	if input.Policy != nil {
		err.Add("policy", "must be empty for exclude targets", "invalid")
	}
	if input.CELExpression != "" {
		err.Add("cel_expression", "must be empty for exclude targets", "invalid")
	}
}
