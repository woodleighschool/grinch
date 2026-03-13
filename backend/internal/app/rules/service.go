package rules

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
)

type RuleCreateInput struct {
	CustomMessage string
	CustomURL     string
	Description   string
	Identifier    string
	Name          string
	RuleType      domain.RuleType
}

type RulePatchInput struct {
	CustomMessage *string
	CustomURL     *string
	Description   *string
	Identifier    *string
	Name          *string
	RuleType      *domain.RuleType
}

type Service struct {
	store Store
}

type Store interface {
	ListRules(context.Context, domain.RuleListOptions) ([]domain.RuleSummary, int32, error)
	GetRule(context.Context, uuid.UUID) (domain.Rule, error)
	CreateRule(context.Context, RuleCreateInput) (domain.Rule, error)
	PatchRule(context.Context, uuid.UUID, RulePatchInput) (domain.Rule, error)
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

func (service *Service) CreateRule(ctx context.Context, input RuleCreateInput) (domain.Rule, error) {
	normalized := normalizeCreateInput(input)
	if validationErr := validateCreateInput(normalized); validationErr != nil {
		return domain.Rule{}, validationErr
	}

	return service.store.CreateRule(ctx, normalized)
}

func (service *Service) PatchRule(ctx context.Context, id uuid.UUID, input RulePatchInput) (domain.Rule, error) {
	normalized := normalizePatchInput(input)
	if validationErr := validatePatchInput(normalized); validationErr != nil {
		return domain.Rule{}, validationErr
	}

	return service.store.PatchRule(ctx, id, normalized)
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

func normalizeCreateInput(input RuleCreateInput) RuleCreateInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Identifier = strings.TrimSpace(input.Identifier)
	input.CustomMessage = strings.TrimSpace(input.CustomMessage)
	input.CustomURL = strings.TrimSpace(input.CustomURL)
	return input
}

func normalizePatchInput(input RulePatchInput) RulePatchInput {
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		input.Name = &value
	}
	if input.Description != nil {
		value := strings.TrimSpace(*input.Description)
		input.Description = &value
	}
	if input.Identifier != nil {
		value := strings.TrimSpace(*input.Identifier)
		input.Identifier = &value
	}
	if input.CustomMessage != nil {
		value := strings.TrimSpace(*input.CustomMessage)
		input.CustomMessage = &value
	}
	if input.CustomURL != nil {
		value := strings.TrimSpace(*input.CustomURL)
		input.CustomURL = &value
	}
	return input
}

func validateCreateInput(input RuleCreateInput) *domain.ValidationError {
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

	if !err.HasFieldErrors() {
		return nil
	}
	return err
}

func validatePatchInput(input RulePatchInput) *domain.ValidationError {
	err := &domain.ValidationError{
		Code:   "validation_error",
		Detail: "Rule is invalid.",
	}

	if input.Name != nil && *input.Name == "" {
		err.Add("name", "must not be empty", "required")
	}
	if input.Identifier != nil && *input.Identifier == "" {
		err.Add("identifier", "must not be empty", "required")
	}

	if !err.HasFieldErrors() {
		return nil
	}
	return err
}
