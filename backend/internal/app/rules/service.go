package rules

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type Store interface {
	ListRules(context.Context, domain.RuleListOptions) ([]domain.RuleSummary, int32, error)
	GetRule(context.Context, uuid.UUID) (domain.Rule, error)
	CreateRule(context.Context, domain.RuleWriteInput) (domain.Rule, error)
	UpdateRule(context.Context, uuid.UUID, domain.RuleWriteInput) (domain.Rule, error)
	DeleteRule(context.Context, uuid.UUID) error
	ListResolvedMachineRules(context.Context, uuid.UUID) ([]domain.MachineResolvedRule, error)
	UpdateAllMachineDesiredTargets(context.Context) error
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListRules(ctx context.Context, opts domain.RuleListOptions) ([]domain.RuleSummary, int32, error) {
	return s.store.ListRules(ctx, opts)
}

func (s *Service) GetRule(ctx context.Context, id uuid.UUID) (domain.Rule, error) {
	return s.store.GetRule(ctx, id)
}

func (s *Service) CreateRule(ctx context.Context, input domain.RuleWriteInput) (domain.Rule, error) {
	if err := validateInput(input); err != nil {
		return domain.Rule{}, err
	}

	rule, err := s.store.CreateRule(ctx, input)
	if err != nil {
		return domain.Rule{}, err
	}

	if err = s.store.UpdateAllMachineDesiredTargets(ctx); err != nil {
		return domain.Rule{}, err
	}

	return rule, nil
}

func (s *Service) UpdateRule(ctx context.Context, id uuid.UUID, input domain.RuleWriteInput) (domain.Rule, error) {
	if err := validateInput(input); err != nil {
		return domain.Rule{}, err
	}

	rule, err := s.store.UpdateRule(ctx, id, input)
	if err != nil {
		return domain.Rule{}, err
	}

	if err = s.store.UpdateAllMachineDesiredTargets(ctx); err != nil {
		return domain.Rule{}, err
	}

	return rule, nil
}

func (s *Service) DeleteRule(ctx context.Context, id uuid.UUID) error {
	if err := s.store.DeleteRule(ctx, id); err != nil {
		return err
	}

	return s.store.UpdateAllMachineDesiredTargets(ctx)
}

func (s *Service) ResolveMachineRuleTargets(
	ctx context.Context,
	machineID uuid.UUID,
) ([]domain.MachineResolvedRule, error) {
	return s.store.ListResolvedMachineRules(ctx, machineID)
}

func validateInput(input domain.RuleWriteInput) *domain.ValidationError {
	err := &domain.ValidationError{
		Code:   "validation_error",
		Detail: "Rule is invalid.",
	}

	if input.Name == "" {
		err.Add("name", "must not be empty", "required")
	}
	if input.Identifier == "" {
		err.Add("identifier", "must not be empty", "required")
	}
	if input.RuleType == "" {
		err.Add("rule_type", "must not be empty", "required")
	}
	for index, target := range input.Targets.Include {
		validateIncludeTarget(index, target, err)
	}
	for index, group := range input.Targets.Exclude {
		validateExcludedGroup(index, group, err)
	}

	if !err.HasFieldErrors() {
		return nil
	}
	return err
}

func validateIncludeTarget(index int, target domain.IncludeRuleTargetWriteInput, err *domain.ValidationError) {
	validateTargetSubject(fmt.Sprintf("targets.include[%d]", index), target.SubjectKind, target.SubjectID, err)
	if target.Policy == "" {
		err.Add(fmt.Sprintf("targets.include[%d].policy", index), "is required for include targets", "required")
		return
	}
	if target.Policy == domain.RulePolicyCEL && target.CELExpression == "" {
		err.Add(fmt.Sprintf("targets.include[%d].cel_expression", index), "is required when policy is cel", "required")
	}
	if target.Policy != domain.RulePolicyCEL && target.CELExpression != "" {
		err.Add(
			fmt.Sprintf("targets.include[%d].cel_expression", index),
			"must be empty unless policy is cel",
			"invalid",
		)
	}
}

func validateExcludedGroup(index int, group domain.ExcludedGroupWriteInput, err *domain.ValidationError) {
	if group.GroupID == uuid.Nil {
		err.Add(fmt.Sprintf("targets.exclude[%d].group_id", index), "is required", "required")
	}
}

func validateTargetSubject(
	prefix string,
	subjectKind domain.RuleTargetSubjectKind,
	subjectID *uuid.UUID,
	err *domain.ValidationError,
) {
	switch subjectKind {
	case domain.RuleTargetSubjectKindGroup:
		if subjectID == nil {
			err.Add(prefix+".subject_id", "is required for group targets", "required")
		}
	case domain.RuleTargetSubjectKindAllDevices, domain.RuleTargetSubjectKindAllUsers:
		if subjectID != nil {
			err.Add(prefix+".subject_id", "must be empty unless subject_kind is group", "invalid")
		}
	default:
		err.Add(prefix+".subject_kind", "must be group, all_devices, or all_users", "invalid")
	}
}
