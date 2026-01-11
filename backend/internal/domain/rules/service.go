package rules

import (
	"context"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
	"github.com/woodleighschool/grinch/internal/domain/errx"
	"github.com/woodleighschool/grinch/internal/listing"
)

// PolicyVersionBumper bumps policy versions when rules change.
type PolicyVersionBumper interface {
	UpdatePolicyRulesVersionByRuleID(ctx context.Context, ruleID uuid.UUID) error
}

// Repo defines persistence operations for rules.
type Repo interface {
	Create(ctx context.Context, r Rule) (Rule, error)
	Update(ctx context.Context, r Rule) (Rule, error)
	Get(ctx context.Context, id uuid.UUID) (Rule, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetMany(ctx context.Context, ids []uuid.UUID) ([]Rule, error)
	List(ctx context.Context, query listing.Query) ([]Rule, listing.Page, error)
}

// Service provides rule operations.
type Service struct {
	repo   Repo
	bumper PolicyVersionBumper
}

// NewService constructs a Service backed by the given repo.
func NewService(repo Repo, bumper PolicyVersionBumper) Service {
	return Service{repo: repo, bumper: bumper}
}

// Get returns a rule by ID.
func (s Service) Get(ctx context.Context, id uuid.UUID) (Rule, error) {
	return s.repo.Get(ctx, id)
}

// GetMany returns rules matching the given IDs.
func (s Service) GetMany(ctx context.Context, ids []uuid.UUID) ([]Rule, error) {
	return s.repo.GetMany(ctx, ids)
}

// List returns rules matching the query.
func (s Service) List(ctx context.Context, query listing.Query) ([]Rule, listing.Page, error) {
	return s.repo.List(ctx, query)
}

// Create validates and persists a rule.
func (s Service) Create(ctx context.Context, r Rule) (Rule, error) {
	if err := validate(r); err != nil {
		return Rule{}, err
	}
	return s.repo.Create(ctx, r)
}

// Update validates and persists a rule and bumps policy versions when rule content changes.
func (s Service) Update(ctx context.Context, r Rule) (Rule, error) {
	if err := validate(r); err != nil {
		return Rule{}, err
	}

	existing, err := s.repo.Get(ctx, r.ID)
	if err != nil {
		return Rule{}, err
	}

	updated, err := s.repo.Update(ctx, r)
	if err != nil {
		return Rule{}, err
	}

	if s.bumper != nil && contentChanged(existing, r) {
		if err = s.bumper.UpdatePolicyRulesVersionByRuleID(ctx, r.ID); err != nil {
			return Rule{}, err
		}
	}

	return updated, nil
}

// Delete removes a rule and bumps policy versions for policies that reference it.
func (s Service) Delete(ctx context.Context, id uuid.UUID) error {
	if s.bumper != nil {
		if err := s.bumper.UpdatePolicyRulesVersionByRuleID(ctx, id); err != nil {
			return err
		}
	}
	return s.repo.Delete(ctx, id)
}

func contentChanged(existing, updated Rule) bool {
	return existing.Description != updated.Description ||
		existing.Identifier != updated.Identifier ||
		existing.RuleType != updated.RuleType ||
		existing.CustomMsg != updated.CustomMsg ||
		existing.CustomURL != updated.CustomURL ||
		existing.NotificationAppName != updated.NotificationAppName
}

func validate(r Rule) error {
	if err := errx.ValidateStruct(r); err != nil {
		return err
	}
	return validateIdentifier(r)
}

func validateIdentifier(r Rule) error {
	switch r.RuleType {
	case syncv1.RuleType_BINARY, syncv1.RuleType_CERTIFICATE:
		return errx.ValidateStruct(struct {
			Identifier string `validate:"sha256"`
		}{r.Identifier})

	case syncv1.RuleType_TEAMID:
		return errx.ValidateStruct(struct {
			Identifier string `validate:"teamid"`
		}{r.Identifier})

	case syncv1.RuleType_SIGNINGID:
		return errx.ValidateStruct(struct {
			Identifier string `validate:"signingid"`
		}{r.Identifier})

	case syncv1.RuleType_CDHASH:
		return errx.ValidateStruct(struct {
			Identifier string `validate:"cdhash"`
		}{r.Identifier})

	case syncv1.RuleType_RULETYPE_UNKNOWN:
		return &errx.Error{
			Code:    errx.CodeInvalid,
			Message: "Validation failed",
			Fields:  map[string]string{"rule_type": "Rule type is required"},
		}

	default:
		return &errx.Error{
			Code:    errx.CodeInvalid,
			Message: "Validation failed",
			Fields:  map[string]string{"rule_type": "Invalid rule type"},
		}
	}
}
