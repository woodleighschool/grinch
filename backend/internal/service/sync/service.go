package sync

import (
	"context"
	"fmt"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	coreerrors "github.com/woodleighschool/grinch/internal/core/errors"
	coreevents "github.com/woodleighschool/grinch/internal/core/events"
	coremachines "github.com/woodleighschool/grinch/internal/core/machines"
	"github.com/woodleighschool/grinch/internal/core/policies"
	corerules "github.com/woodleighschool/grinch/internal/core/rules"
)

// MachineCatalog exposes the minimal machine operations needed for sync.
type MachineCatalog interface {
	Get(ctx context.Context, id uuid.UUID) (coremachines.Machine, error)
	Upsert(ctx context.Context, machine coremachines.Machine) (coremachines.Machine, error)
}

// PolicyCatalog exposes the policy data required for sync stages.
type PolicyCatalog interface {
	Get(ctx context.Context, id uuid.UUID) (policies.Policy, error)
	ListPolicyRuleAttachmentsForSyncByPolicyID(
		ctx context.Context,
		policyID uuid.UUID,
		limit, offset int,
	) ([]policies.PolicyAttachment, error)
}

// RuleCatalog provides bulk rule lookup.
type RuleCatalog interface {
	GetMany(ctx context.Context, ids []uuid.UUID) ([]corerules.Rule, error)
}

// EventRecorder persists incoming execution events.
type EventRecorder interface {
	InsertBatch(ctx context.Context, events []coreevents.Event) error
}

// Service coordinates Santa sync operations across the domain services.
type Service struct {
	machines MachineCatalog
	policies PolicyCatalog
	rules    RuleCatalog
	events   EventRecorder
	pageSize int
}

const defaultRulePageSize = 200

// NewService builds a sync service with its dependencies.
func NewService(m MachineCatalog, p PolicyCatalog, r RuleCatalog, e EventRecorder) *Service {
	return &Service{
		machines: m,
		policies: p,
		rules:    r,
		events:   e,
		pageSize: defaultRulePageSize,
	}
}

// RulePageSize returns the configured page size for rule sync pagination.
func (s *Service) RulePageSize() int {
	return s.pageSize
}

// EventUpload ingests execution events for a machine and refreshes its heartbeat.
func (s *Service) EventUpload(ctx context.Context, machineID uuid.UUID, events []coreevents.Event) error {
	if len(events) > 0 {
		if err := s.events.InsertBatch(ctx, events); err != nil {
			return fmt.Errorf("eventupload: insert events: %w", err)
		}
	}

	machine, err := s.machines.Get(ctx, machineID)
	if err != nil {
		return fmt.Errorf("eventupload: get machine: %w", err)
	}

	machine.LastSeen = nowUTC()
	_, err = s.machines.Upsert(ctx, machine)
	if err != nil {
		return fmt.Errorf("eventupload: upsert machine: %w", err)
	}

	return nil
}

// Preflight processes the Santa preflight stage and returns sync parameters for the client.
func (s *Service) Preflight(
	ctx context.Context,
	machineID uuid.UUID,
	req *syncv1.PreflightRequest,
) (*syncv1.PreflightResponse, error) {
	existing, err := s.machines.Get(ctx, machineID)
	if err != nil && !coreerrors.IsCode(err, coreerrors.CodeNotFound) {
		return nil, fmt.Errorf("preflight: get machine: %w", err)
	}

	builder := newMachineBuilder(machineID, existing, nowUTC())
	machine := builder.fromPreflight(req)

	if machine.PolicyID == nil {
		reset := builder.clearApplied(machine)
		if _, err = s.machines.Upsert(ctx, reset); err != nil {
			return nil, fmt.Errorf("preflight: upsert machine: %w", err)
		}
		return &syncv1.PreflightResponse{}, nil
	}

	policy, err := s.policies.Get(ctx, *machine.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("preflight: get policy: %w", err)
	}

	machine.PolicyStatus = policies.ComputeStatus(policies.AssignmentState{
		AppliedPolicyID:        machine.AppliedPolicyID,
		AppliedSettingsVersion: machine.AppliedSettingsVersion,
		AppliedRulesVersion:    machine.AppliedRulesVersion,
	}, policy)
	machine.AppliedSettingsVersion = &policy.SettingsVersion

	resolved, err := s.machines.Upsert(ctx, machine)
	if err != nil {
		return nil, fmt.Errorf("preflight: upsert machine: %w", err)
	}

	resp := buildPreflightResponse(policy)
	resp.SyncType = determineSyncType(resolved, policy, req)
	return resp, nil
}

// RuleDownload emits policy rules for a machine starting at cursorOffset.
func (s *Service) RuleDownload(
	ctx context.Context,
	machineID uuid.UUID,
	cursorOffset int,
) (*syncv1.RuleDownloadResponse, error) {
	machine, err := s.machines.Get(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("ruledownload: get machine: %w", err)
	}

	machine.LastSeen = nowUTC()
	if _, err = s.machines.Upsert(ctx, machine); err != nil {
		return nil, fmt.Errorf("ruledownload: upsert machine: %w", err)
	}

	if machine.PolicyID == nil {
		return noopResponse(), nil
	}

	policy, err := s.policies.Get(ctx, *machine.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("ruledownload: get policy: %w", err)
	}

	if upToDate(machine, policy.RulesVersion) {
		noopResp, noopErr := s.ensureStatefulNoop(ctx, *machine.PolicyID)
		if noopErr != nil {
			return nil, fmt.Errorf("ruledownload: verify attachments: %w", noopErr)
		}
		return noopResp, nil
	}

	attachments, err := s.policies.ListPolicyRuleAttachmentsForSyncByPolicyID(
		ctx,
		*machine.PolicyID,
		s.pageSize,
		cursorOffset,
	)
	if err != nil {
		return nil, fmt.Errorf("ruledownload: list attachments: %w", err)
	}

	if cursorOffset == 0 && len(attachments) == 0 {
		return noopResponse(), nil
	}

	rules, err := s.loadRules(ctx, attachments)
	if err != nil {
		return nil, fmt.Errorf("ruledownload: load rules: %w", err)
	}

	return &syncv1.RuleDownloadResponse{
		Rules:  buildRuleset(attachments, rules),
		Cursor: nextCursor(cursorOffset, len(attachments), s.pageSize),
	}, nil
}

// Postflight records applied policy state after the Santa postflight stage.
func (s *Service) Postflight(
	ctx context.Context,
	machineID uuid.UUID,
	req *syncv1.PostflightRequest,
) (*syncv1.PostflightResponse, error) {
	machine, err := s.machines.Get(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("postflight: get machine: %w", err)
	}

	machine.LastSeen = nowUTC()

	if machine.PolicyID == nil {
		reset := newMachineBuilder(machineID, machine, machine.LastSeen).clearApplied(machine)
		if _, err = s.machines.Upsert(ctx, reset); err != nil {
			return nil, fmt.Errorf("postflight: upsert machine: %w", err)
		}
		return &syncv1.PostflightResponse{}, nil
	}

	policy, err := s.policies.Get(ctx, *machine.PolicyID)
	if err != nil {
		return nil, fmt.Errorf("postflight: get policy: %w", err)
	}

	if rulesHash := req.GetRulesHash(); rulesHash != "" {
		machine.AppliedRulesVersion = &policy.RulesVersion
	}

	machine.AppliedPolicyID = machine.PolicyID
	machine.PolicyStatus = computePostflightStatus(machine, policy)

	if _, err = s.machines.Upsert(ctx, machine); err != nil {
		return nil, fmt.Errorf("postflight: upsert machine: %w", err)
	}

	return &syncv1.PostflightResponse{}, nil
}
