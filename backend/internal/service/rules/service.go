package rules

import (
	"context"

	"github.com/google/uuid"

	corerules "github.com/woodleighschool/grinch/internal/core/rules"
	"github.com/woodleighschool/grinch/internal/listing"
)

// RuleStore defines persistence operations for rules.
type RuleStore interface {
	Create(ctx context.Context, r corerules.Rule) (corerules.Rule, error)
	Update(ctx context.Context, r corerules.Rule) (corerules.Rule, error)
	Get(ctx context.Context, id uuid.UUID) (corerules.Rule, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetMany(ctx context.Context, ids []uuid.UUID) ([]corerules.Rule, error)
	List(ctx context.Context, query listing.Query) ([]corerules.Rule, listing.Page, error)
}

// RuleService owns rule lifecycle and coordinates policy bumps when content changes.
type RuleService struct {
	store  RuleStore
	bumper PolicyRuleVersionBumper
}

// PolicyRuleVersionBumper bumps policy versions when rules change.
type PolicyRuleVersionBumper interface {
	UpdatePolicyRulesVersionByRuleID(ctx context.Context, ruleID uuid.UUID) error
}

// NewRuleService constructs a RuleService.
func NewRuleService(store RuleStore, bumper PolicyRuleVersionBumper) *RuleService {
	return &RuleService{store: store, bumper: bumper}
}

// Get returns a rule by ID.
func (s *RuleService) Get(ctx context.Context, id uuid.UUID) (corerules.Rule, error) {
	return s.store.Get(ctx, id)
}

// GetMany returns rules matching the given IDs.
func (s *RuleService) GetMany(ctx context.Context, ids []uuid.UUID) ([]corerules.Rule, error) {
	return s.store.GetMany(ctx, ids)
}

// List returns rules matching the query.
func (s *RuleService) List(ctx context.Context, query listing.Query) ([]corerules.Rule, listing.Page, error) {
	return s.store.List(ctx, query)
}

// Create validates and persists a rule.
func (s *RuleService) Create(ctx context.Context, r corerules.Rule) (corerules.Rule, error) {
	return s.store.Create(ctx, r)
}

// Update validates and persists a rule and bumps policy versions when rule content changes.
func (s *RuleService) Update(ctx context.Context, r corerules.Rule) (corerules.Rule, error) {
	existing, err := s.store.Get(ctx, r.ID)
	if err != nil {
		return corerules.Rule{}, err
	}

	updated, err := s.store.Update(ctx, r)
	if err != nil {
		return corerules.Rule{}, err
	}

	if s.bumper != nil && ruleContentChanged(existing, r) {
		if err = s.bumper.UpdatePolicyRulesVersionByRuleID(ctx, r.ID); err != nil {
			return corerules.Rule{}, err
		}
	}

	return updated, nil
}

// Delete removes a rule and bumps policy versions for policies that reference it.
func (s *RuleService) Delete(ctx context.Context, id uuid.UUID) error {
	if s.bumper != nil {
		if err := s.bumper.UpdatePolicyRulesVersionByRuleID(ctx, id); err != nil {
			return err
		}
	}
	return s.store.Delete(ctx, id)
}

func ruleContentChanged(existing, updated corerules.Rule) bool {
	return existing.Description != updated.Description ||
		existing.Identifier != updated.Identifier ||
		existing.RuleType != updated.RuleType ||
		existing.CustomMsg != updated.CustomMsg ||
		existing.CustomURL != updated.CustomURL ||
		existing.NotificationAppName != updated.NotificationAppName
}
