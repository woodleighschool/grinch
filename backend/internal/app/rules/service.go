package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type Service struct {
	store Store
}

type Store interface {
	ListRules(context.Context, domain.RuleListOptions) ([]domain.RuleSummary, int32, error)
	GetRule(context.Context, uuid.UUID) (domain.Rule, error)
	CreateRule(context.Context, domain.RuleWriteInput) (domain.Rule, error)
	UpdateRule(context.Context, uuid.UUID, domain.RuleWriteInput) (domain.Rule, error)
	DeleteRule(context.Context, uuid.UUID) error
	ListResolvedMachineRules(context.Context, uuid.UUID) ([]domain.MachineResolvedRule, error)
	SyncAllMachineDesiredRuleTargets(context.Context) error
}

func New(store Store) *Service {
	return &Service{store: store}
}

func (service *Service) ListRules(
	ctx context.Context,
	options domain.RuleListOptions,
) ([]domain.RuleSummary, int32, error) {
	return service.store.ListRules(ctx, options)
}

func (service *Service) GetRule(ctx context.Context, id uuid.UUID) (domain.Rule, error) {
	return service.store.GetRule(ctx, id)
}

func (service *Service) CreateRule(ctx context.Context, input domain.RuleWriteInput) (domain.Rule, error) {
	normalized := normalizeInput(input)
	if validationErr := validateInput(normalized); validationErr != nil {
		return domain.Rule{}, validationErr
	}

	rule, err := service.store.CreateRule(ctx, normalized)
	if err != nil {
		return domain.Rule{}, err
	}
	syncErr := service.store.SyncAllMachineDesiredRuleTargets(ctx)
	if syncErr != nil {
		return domain.Rule{}, syncErr
	}

	return rule, nil
}

func (service *Service) UpdateRule(
	ctx context.Context,
	id uuid.UUID,
	input domain.RuleWriteInput,
) (domain.Rule, error) {
	normalized := normalizeInput(input)
	if validationErr := validateInput(normalized); validationErr != nil {
		return domain.Rule{}, validationErr
	}

	rule, err := service.store.UpdateRule(ctx, id, normalized)
	if err != nil {
		return domain.Rule{}, err
	}
	syncErr := service.store.SyncAllMachineDesiredRuleTargets(ctx)
	if syncErr != nil {
		return domain.Rule{}, syncErr
	}

	return rule, nil
}

func (service *Service) DeleteRule(ctx context.Context, id uuid.UUID) error {
	if err := service.store.DeleteRule(ctx, id); err != nil {
		return err
	}

	return service.store.SyncAllMachineDesiredRuleTargets(ctx)
}

func (service *Service) ResolveMachineRuleTargets(
	ctx context.Context,
	machineID uuid.UUID,
) ([]domain.MachineResolvedRule, error) {
	return service.store.ListResolvedMachineRules(ctx, machineID)
}

func normalizeInput(input domain.RuleWriteInput) domain.RuleWriteInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Identifier = strings.TrimSpace(input.Identifier)
	input.CustomMessage = strings.TrimSpace(input.CustomMessage)
	input.CustomURL = strings.TrimSpace(input.CustomURL)
	for index := range input.Targets.Include {
		target := &input.Targets.Include[index]
		target.CELExpression = strings.TrimSpace(target.CELExpression)
		if target.Policy != domain.RulePolicyCEL {
			target.CELExpression = ""
		}
	}
	return input
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
