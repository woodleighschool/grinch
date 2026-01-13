package policies

import (
	"context"
	"reflect"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain/machines"
	"github.com/woodleighschool/grinch/internal/listing"
)

// GroupLookup resolves group memberships for policy evaluation.
type GroupLookup interface {
	GroupIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// PolicyRefresher recalculates policy assignments.
type PolicyRefresher interface {
	RefreshAll(ctx context.Context) error
}

// Repo defines persistence operations for policies.
type Repo interface {
	Create(ctx context.Context, policy Policy) (Policy, error)
	Update(ctx context.Context, policy Policy) (Policy, error)
	Get(ctx context.Context, id uuid.UUID) (Policy, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, query listing.Query) ([]ListItem, listing.Page, error)
	ListEnabled(ctx context.Context) ([]Policy, error)
	ListPolicyTargetsByPolicyIDs(ctx context.Context, policyIDs []uuid.UUID) ([]Target, error)
	ListPolicyRuleAttachmentsByPolicyID(ctx context.Context, policyID uuid.UUID) ([]Attachment, error)
	ListPolicyRuleAttachmentsForSyncByPolicyID(
		ctx context.Context,
		policyID uuid.UUID,
		limit, offset int,
	) ([]Attachment, error)
	UpdatePolicyRulesVersionByRuleID(ctx context.Context, ruleID uuid.UUID) error
}

// Service provides policy operations.
type Service struct {
	repo      Repo
	groups    GroupLookup
	refresher PolicyRefresher
}

// NewService constructs a Service.
func NewService(repo Repo, groups GroupLookup, refresher PolicyRefresher) Service {
	return Service{repo: repo, groups: groups, refresher: refresher}
}

// Get returns a policy by ID.
func (s Service) Get(ctx context.Context, id uuid.UUID) (Policy, error) {
	return s.repo.Get(ctx, id)
}

// List returns policies matching the query.
func (s Service) List(ctx context.Context, query listing.Query) ([]ListItem, listing.Page, error) {
	return s.repo.List(ctx, query)
}

// ListEnabled returns all enabled policies.
func (s Service) ListEnabled(ctx context.Context) ([]Policy, error) {
	return s.repo.ListEnabled(ctx)
}

// ListPolicyTargetsByPolicyIDs returns targets for the given policy IDs.
func (s Service) ListPolicyTargetsByPolicyIDs(ctx context.Context, policyIDs []uuid.UUID) ([]Target, error) {
	return s.repo.ListPolicyTargetsByPolicyIDs(ctx, policyIDs)
}

// ListPolicyRuleAttachmentsByPolicyID returns all attachments for a policy.
func (s Service) ListPolicyRuleAttachmentsByPolicyID(ctx context.Context, policyID uuid.UUID) ([]Attachment, error) {
	return s.repo.ListPolicyRuleAttachmentsByPolicyID(ctx, policyID)
}

// ListPolicyRuleAttachmentsForSyncByPolicyID returns a page of attachments for sync.
func (s Service) ListPolicyRuleAttachmentsForSyncByPolicyID(
	ctx context.Context,
	policyID uuid.UUID,
	limit, offset int,
) ([]Attachment, error) {
	return s.repo.ListPolicyRuleAttachmentsForSyncByPolicyID(ctx, policyID, limit, offset)
}

// Create validates and creates a policy.
func (s Service) Create(ctx context.Context, policy Policy) (Policy, error) {
	if err := validate(policy); err != nil {
		return Policy{}, err
	}

	policy.SettingsVersion = 1
	policy.RulesVersion = 1

	created, err := s.repo.Create(ctx, policy)
	if err != nil {
		return Policy{}, err
	}

	_ = s.refreshPolicies(ctx) // Best effort.
	return created, nil
}

// Update validates and updates a policy.
func (s Service) Update(ctx context.Context, policy Policy) (Policy, error) {
	if err := validate(policy); err != nil {
		return Policy{}, err
	}

	existing, err := s.repo.Get(ctx, policy.ID)
	if err != nil {
		return Policy{}, err
	}

	policy.SettingsVersion = existing.SettingsVersion
	policy.RulesVersion = existing.RulesVersion

	if settingsChanged(existing, policy) {
		policy.SettingsVersion++
	}
	if attachmentsChanged(existing, policy) {
		policy.RulesVersion++
	}

	updated, err := s.repo.Update(ctx, policy)
	if err != nil {
		return Policy{}, err
	}

	_ = s.refreshPolicies(ctx) // Best effort.
	return updated, nil
}

// Delete removes a policy by ID.
func (s Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.refreshPolicies(ctx) // Best effort.
	return nil
}

// ResolveForMachine returns the effective policy for a machine.
func (s Service) ResolveForMachine(ctx context.Context, machine machines.Machine) (Policy, error) {
	return resolveForMachine(ctx, s, machine)
}

// UpdatePolicyRulesVersionByRuleID increments rules versions for policies that reference a rule.
func (s Service) UpdatePolicyRulesVersionByRuleID(ctx context.Context, ruleID uuid.UUID) error {
	if err := s.repo.UpdatePolicyRulesVersionByRuleID(ctx, ruleID); err != nil {
		return err
	}

	_ = s.refreshPolicies(ctx) // Best effort.
	return nil
}

func (s Service) refreshPolicies(ctx context.Context) error {
	if s.refresher == nil {
		return nil
	}
	return s.refresher.RefreshAll(ctx)
}

func settingsChanged(existing, updated Policy) bool {
	return existing.Name != updated.Name ||
		existing.Description != updated.Description ||
		existing.Enabled != updated.Enabled ||
		existing.Priority != updated.Priority ||
		existing.SetClientMode != updated.SetClientMode ||
		existing.SetBatchSize != updated.SetBatchSize ||
		existing.SetEnableBundles != updated.SetEnableBundles ||
		existing.SetEnableTransitiveRules != updated.SetEnableTransitiveRules ||
		existing.SetEnableAllEventUpload != updated.SetEnableAllEventUpload ||
		existing.SetDisableUnknownEventUpload != updated.SetDisableUnknownEventUpload ||
		existing.SetFullSyncIntervalSeconds != updated.SetFullSyncIntervalSeconds ||
		existing.SetPushNotificationFullSyncIntervalSeconds != updated.SetPushNotificationFullSyncIntervalSeconds ||
		existing.SetPushNotificationGlobalRuleSyncDeadlineSeconds != updated.SetPushNotificationGlobalRuleSyncDeadlineSeconds ||
		existing.SetAllowedPathRegex != updated.SetAllowedPathRegex ||
		existing.SetBlockedPathRegex != updated.SetBlockedPathRegex ||
		existing.SetBlockUSBMount != updated.SetBlockUSBMount ||
		!reflect.DeepEqual(existing.SetRemountUSBMode, updated.SetRemountUSBMode) ||
		existing.SetOverrideFileAccessAction != updated.SetOverrideFileAccessAction
}

func attachmentsChanged(existing, updated Policy) bool {
	if len(existing.Attachments) != len(updated.Attachments) {
		return true
	}

	existingSet := make(map[uuid.UUID]Attachment)
	for _, a := range existing.Attachments {
		existingSet[a.RuleID] = a
	}

	for _, a := range updated.Attachments {
		ea, ok := existingSet[a.RuleID]
		if !ok || ea.Action != a.Action || !ptrEqual(ea.CELExpr, a.CELExpr) {
			return true
		}
	}

	return false
}

func ptrEqual[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
