package policies

import (
	"context"
	"slices"

	"github.com/google/uuid"

	coremachines "github.com/woodleighschool/grinch/internal/core/machines"
	corepolicies "github.com/woodleighschool/grinch/internal/core/policies"
	"github.com/woodleighschool/grinch/internal/listing"
)

const defaultPolicyPageSize = 200

// PolicyStore defines persistence operations for policies.
type PolicyStore interface {
	Create(ctx context.Context, policy corepolicies.Policy) (corepolicies.Policy, error)
	Update(ctx context.Context, policy corepolicies.Policy) (corepolicies.Policy, error)
	Get(ctx context.Context, id uuid.UUID) (corepolicies.Policy, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, query listing.Query) ([]corepolicies.PolicyListItem, listing.Page, error)
	ListEnabled(ctx context.Context) ([]corepolicies.Policy, error)
	ListPolicyTargetsByPolicyIDs(ctx context.Context, policyIDs []uuid.UUID) ([]corepolicies.PolicyTarget, error)
	ListPolicyRuleAttachmentsByPolicyID(
		ctx context.Context,
		policyID uuid.UUID,
	) ([]corepolicies.PolicyAttachment, error)
	ListPolicyRuleAttachmentsForSyncByPolicyID(
		ctx context.Context,
		policyID uuid.UUID,
		limit, offset int,
	) ([]corepolicies.PolicyAttachment, error)
	UpdatePolicyRulesVersionByRuleID(ctx context.Context, ruleID uuid.UUID) error
}

// GroupMembershipLookup resolves group memberships for policy evaluation.
type GroupMembershipLookup interface {
	GroupIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

// MachineCatalog exposes the minimal machine operations needed for reconciliation.
type MachineCatalog interface {
	List(ctx context.Context, query listing.Query) ([]coremachines.MachineListItem, listing.Page, error)
	UpdatePolicyState(ctx context.Context, id uuid.UUID, policyID *uuid.UUID, status corepolicies.Status) error
}

// PolicyService owns policy lifecycle, evaluation, and assignment reconciliation.
type PolicyService struct {
	store       PolicyStore
	memberships GroupMembershipLookup
	machines    MachineCatalog
	pageSize    int
}

// NewPolicyService constructs a PolicyService.
func NewPolicyService(store PolicyStore, memberships GroupMembershipLookup, machines MachineCatalog) *PolicyService {
	return &PolicyService{
		store:       store,
		memberships: memberships,
		machines:    machines,
		pageSize:    defaultPolicyPageSize,
	}
}

// Get returns a policy by ID.
func (s *PolicyService) Get(ctx context.Context, id uuid.UUID) (corepolicies.Policy, error) {
	return s.store.Get(ctx, id)
}

// List returns policies matching the query.
func (s *PolicyService) List(
	ctx context.Context,
	query listing.Query,
) ([]corepolicies.PolicyListItem, listing.Page, error) {
	return s.store.List(ctx, query)
}

// ListEnabled returns all enabled policies.
func (s *PolicyService) ListEnabled(ctx context.Context) ([]corepolicies.Policy, error) {
	return s.store.ListEnabled(ctx)
}

// ListPolicyRuleAttachmentsByPolicyID returns all attachments for a policy.
func (s *PolicyService) ListPolicyRuleAttachmentsByPolicyID(
	ctx context.Context,
	policyID uuid.UUID,
) ([]corepolicies.PolicyAttachment, error) {
	return s.store.ListPolicyRuleAttachmentsByPolicyID(ctx, policyID)
}

// ListPolicyRuleAttachmentsForSyncByPolicyID returns a page of attachments for sync.
func (s *PolicyService) ListPolicyRuleAttachmentsForSyncByPolicyID(
	ctx context.Context,
	policyID uuid.UUID,
	limit, offset int,
) ([]corepolicies.PolicyAttachment, error) {
	return s.store.ListPolicyRuleAttachmentsForSyncByPolicyID(ctx, policyID, limit, offset)
}

// Create validates and creates a policy.
func (s *PolicyService) Create(ctx context.Context, policy corepolicies.Policy) (corepolicies.Policy, error) {
	policy.SettingsVersion = 1
	policy.RulesVersion = 1

	created, err := s.store.Create(ctx, policy)
	if err != nil {
		return corepolicies.Policy{}, err
	}

	_ = s.RefreshAssignments(ctx) // Best effort.
	return created, nil
}

// Update validates and updates a policy.
func (s *PolicyService) Update(ctx context.Context, policy corepolicies.Policy) (corepolicies.Policy, error) {
	existing, err := s.store.Get(ctx, policy.ID)
	if err != nil {
		return corepolicies.Policy{}, err
	}

	policy.SettingsVersion = existing.SettingsVersion
	policy.RulesVersion = existing.RulesVersion

	if settingsChanged(existing, policy) {
		policy.SettingsVersion++
	}
	if attachmentsChanged(existing, policy) {
		policy.RulesVersion++
	}

	updated, err := s.store.Update(ctx, policy)
	if err != nil {
		return corepolicies.Policy{}, err
	}

	_ = s.RefreshAssignments(ctx) // Best effort.
	return updated, nil
}

// Delete removes a policy by ID.
func (s *PolicyService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.store.Delete(ctx, id); err != nil {
		return err
	}
	return s.RefreshAssignments(ctx)
}

// ResolveForMachine returns the effective policy for a machine.
func (s *PolicyService) ResolveForMachine(
	ctx context.Context,
	machine coremachines.Machine,
) (corepolicies.Policy, error) {
	userID := uuid.Nil
	if machine.UserID != nil {
		userID = *machine.UserID
	}

	groupIDs, err := s.lookupGroupIDs(ctx, userID)
	if err != nil {
		return corepolicies.Policy{}, err
	}

	enabled, err := s.store.ListEnabled(ctx)
	if err != nil {
		return corepolicies.Policy{}, err
	}
	if len(enabled) == 0 {
		return corepolicies.Policy{}, nil
	}

	policyIDs := extractPolicyIDs(enabled)

	targets, err := s.store.ListPolicyTargetsByPolicyIDs(ctx, policyIDs)
	if err != nil {
		return corepolicies.Policy{}, err
	}

	subject := corepolicies.Subject{
		MachineID: machine.ID,
		UserID:    userID,
		GroupIDs:  groupIDs,
	}

	return corepolicies.SelectPolicy(subject, enabled, targets), nil
}

// UpdatePolicyRulesVersionByRuleID increments rules versions for policies that reference a rule.
func (s *PolicyService) UpdatePolicyRulesVersionByRuleID(ctx context.Context, ruleID uuid.UUID) error {
	if err := s.store.UpdatePolicyRulesVersionByRuleID(ctx, ruleID); err != nil {
		return err
	}
	return s.RefreshAssignments(ctx)
}

// RefreshAssignments recomputes policy assignments for all machines.
func (s *PolicyService) RefreshAssignments(ctx context.Context) error {
	if s.machines == nil {
		return nil
	}

	offset := 0
	for {
		items, _, err := s.machines.List(ctx, listing.Query{
			Limit:  s.pageSize,
			Offset: offset,
		})
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}

		for _, item := range items {
			if err = s.refreshMachine(ctx, item); err != nil {
				return err
			}
		}

		if len(items) < s.pageSize {
			return nil
		}
		offset += s.pageSize
	}
}

func (s *PolicyService) refreshMachine(ctx context.Context, item coremachines.MachineListItem) error {
	machine := coremachines.Machine{
		ID:                     item.ID,
		UserID:                 item.UserID,
		PolicyID:               item.PolicyID,
		AppliedPolicyID:        item.AppliedPolicyID,
		AppliedSettingsVersion: item.AppliedSettingsVersion,
		AppliedRulesVersion:    item.AppliedRulesVersion,
	}

	policy, err := s.ResolveForMachine(ctx, machine)
	if err != nil {
		return err
	}

	var desiredPolicyID *uuid.UUID
	if policy.ID != uuid.Nil {
		desiredPolicyID = &policy.ID
	}

	status := corepolicies.ComputeStatus(corepolicies.AssignmentState{
		AppliedPolicyID:        machine.AppliedPolicyID,
		AppliedSettingsVersion: machine.AppliedSettingsVersion,
		AppliedRulesVersion:    machine.AppliedRulesVersion,
	}, policy)

	if ptrEqual(item.PolicyID, desiredPolicyID) && item.PolicyStatus == status {
		return nil
	}

	return s.machines.UpdatePolicyState(ctx, item.ID, desiredPolicyID, status)
}

func (s *PolicyService) lookupGroupIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	if s.memberships == nil || userID == uuid.Nil {
		return nil, nil
	}
	return s.memberships.GroupIDsForUser(ctx, userID)
}

func extractPolicyIDs(policies []corepolicies.Policy) []uuid.UUID {
	ids := make([]uuid.UUID, len(policies))
	for i := range policies {
		ids[i] = policies[i].ID
	}
	return ids
}

func settingsChanged(existing, updated corepolicies.Policy) bool {
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
		!slices.Equal(existing.SetRemountUSBMode, updated.SetRemountUSBMode) ||
		existing.SetOverrideFileAccessAction != updated.SetOverrideFileAccessAction
}

func attachmentsChanged(existing, updated corepolicies.Policy) bool {
	if len(existing.Attachments) != len(updated.Attachments) {
		return true
	}

	existingSet := make(map[uuid.UUID]corepolicies.PolicyAttachment)
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
