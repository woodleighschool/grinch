package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type WriteInput struct {
	CustomMessage string
	CustomURL     string
	Description   string
	Enabled       bool
	Identifier    string
	Name          string
	RuleType      domain.RuleType
	Targets       TargetsWriteInput
}

type TargetsWriteInput struct {
	Include []IncludeTargetWriteInput
	Exclude []ExcludedGroupWriteInput
}

type IncludeTargetWriteInput struct {
	SubjectKind   domain.RuleTargetSubjectKind
	SubjectID     *uuid.UUID
	Policy        domain.RulePolicy
	CELExpression string
}

type ExcludedGroupWriteInput struct {
	GroupID uuid.UUID
}

type Service struct {
	store Store
}

type Store interface {
	ListRules(context.Context, domain.RuleListOptions) ([]domain.RuleSummary, int32, error)
	GetRule(context.Context, uuid.UUID) (domain.Rule, error)
	CreateRule(context.Context, WriteInput) (domain.Rule, error)
	UpdateRule(context.Context, uuid.UUID, WriteInput) (domain.Rule, error)
	DeleteRule(context.Context, uuid.UUID) error
	ListResolvedMachineRules(context.Context, uuid.UUID) ([]domain.MachineResolvedRule, error)
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

func (service *Service) CreateRule(ctx context.Context, input WriteInput) (domain.Rule, error) {
	normalized := normalizeInput(input)
	if validationErr := validateInput(normalized); validationErr != nil {
		return domain.Rule{}, validationErr
	}

	return service.store.CreateRule(ctx, normalized)
}

func (service *Service) UpdateRule(ctx context.Context, id uuid.UUID, input WriteInput) (domain.Rule, error) {
	normalized := normalizeInput(input)
	if validationErr := validateInput(normalized); validationErr != nil {
		return domain.Rule{}, validationErr
	}

	return service.store.UpdateRule(ctx, id, normalized)
}

func (service *Service) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return service.store.DeleteRule(ctx, id)
}

func (service *Service) ResolveMachineRuleTargets(
	ctx context.Context,
	machineID uuid.UUID,
) ([]domain.MachineRuleTarget, error) {
	resolved, err := service.store.ListResolvedMachineRules(ctx, machineID)
	if err != nil {
		return nil, err
	}

	targets := make([]domain.MachineRuleTarget, 0, len(resolved))
	for _, rule := range resolved {
		targets = append(targets, rule.MachineRuleTarget)
	}

	return targets, nil
}

func normalizeInput(input WriteInput) WriteInput {
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

func validateInput(input WriteInput) *domain.ValidationError {
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

func validateIncludeTarget(index int, target IncludeTargetWriteInput, err *domain.ValidationError) {
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

func validateExcludedGroup(index int, group ExcludedGroupWriteInput, err *domain.ValidationError) {
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
