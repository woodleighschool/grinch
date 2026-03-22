// Package snapshot freezes one Santa sync cycle into a pending desired state.
//
// Santa sync is inherently two-phase: preflight decides what the machine
// should end up with, rule download must serve that same frozen state, and
// postflight acknowledges the frozen result only after the client reports it
// processed the full payload we sent for that frozen snapshot.
package snapshot

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/santa/model"
)

var ErrPendingSnapshotNotFound = errors.New("pending rule sync snapshot not found")

type PendingSnapshot struct {
	FullSync         bool
	Payload          []model.SyncRule
	PayloadRuleCount int64
}

func PreparePendingSnapshot(
	ctx context.Context,
	store model.DataStore,
	ruleResolver model.RuleResolver,
	machineID uuid.UUID,
	request *syncv1.PreflightRequest,
	preparedAt time.Time,
) (PendingSnapshot, error) {
	state, err := store.GetMachineSyncState(ctx, machineID)
	if err != nil {
		return PendingSnapshot{}, fmt.Errorf("get machine sync state: %w", err)
	}

	ruleTargets, err := ruleResolver.ResolveMachineRuleTargets(ctx, machineID)
	if err != nil {
		return PendingSnapshot{}, fmt.Errorf("resolve machine rule targets: %w", err)
	}

	currentTargets := withRulePayloadHashes(ruleTargets)
	snapshot, pendingWrite := buildPendingSnapshot(
		state,
		currentTargets,
		request,
		preparedAt,
		machineID,
	)

	if storeErr := store.ReplacePendingSnapshot(ctx, pendingWrite); storeErr != nil {
		return PendingSnapshot{}, fmt.Errorf("store pending rule snapshot: %w", storeErr)
	}

	return snapshot, nil
}

// PendingSnapshotForMachine reloads the frozen state created at preflight time
// so rule download serves one stable snapshot for the whole sync cycle.
func PendingSnapshotForMachine(
	ctx context.Context,
	store model.DataStore,
	machineID uuid.UUID,
) (PendingSnapshot, model.MachineSyncState, error) {
	state, err := store.GetMachineSyncState(ctx, machineID)
	if err != nil {
		return PendingSnapshot{}, model.MachineSyncState{}, fmt.Errorf("get machine sync state: %w", err)
	}
	if state.PendingPreflightAt == nil {
		return PendingSnapshot{}, model.MachineSyncState{}, fmt.Errorf("%w", ErrPendingSnapshotNotFound)
	}

	return PendingSnapshot{
		FullSync:         state.PendingFullSync,
		Payload:          slices.Clone(state.PendingPayload),
		PayloadRuleCount: state.PendingPayloadRuleCount,
	}, state, nil
}

func buildPendingSnapshot(
	state model.MachineSyncState,
	current []model.PendingRuleTarget,
	request *syncv1.PreflightRequest,
	preparedAt time.Time,
	machineID uuid.UUID,
) (PendingSnapshot, model.PendingSnapshotWrite) {
	desiredTargets := toAppliedRuleTargets(current)
	desiredCounts := countPendingRuleTargets(current)
	applied := slices.Clone(state.AppliedTargets)
	targetsMatch := appliedRuleTargetsMatch(desiredTargets, applied)
	payload := diffSnapshot(current, applied)
	reportedCountsMatchAt := state.LastReportedCountsMatchAt
	reportedCountsMatch := managedRuleCountsMatch(desiredCounts, request)
	if reportedCountsMatch {
		reportedCountsMatchAt = &preparedAt
	}
	fullSync := request.GetRequestCleanSync() || (targetsMatch && !reportedCountsMatch)
	if fullSync {
		payload = fullSyncRules(current)
	}

	payloadRuleCount := int64(len(payload))
	snapshot := PendingSnapshot{
		FullSync:         fullSync,
		Payload:          payload,
		PayloadRuleCount: payloadRuleCount,
	}

	return snapshot, model.PendingSnapshotWrite{
		MachineID:                   machineID,
		RulesHash:                   state.RulesHash,
		DesiredTargets:              desiredTargets,
		AppliedTargets:              applied,
		PendingTargets:              desiredTargets,
		PendingPayload:              slices.Clone(payload),
		PendingPayloadRuleCount:     payloadRuleCount,
		PendingFullSync:             fullSync,
		PendingPreflightAt:          preparedAt,
		DesiredBinaryRuleCount:      desiredCounts.Binary,
		DesiredCertificateRuleCount: desiredCounts.Certificate,
		DesiredTeamIDRuleCount:      desiredCounts.TeamID,
		DesiredSigningIDRuleCount:   desiredCounts.SigningID,
		DesiredCDHashRuleCount:      desiredCounts.CDHash,
		ClientMode:                  MapClientMode(request.GetClientMode()),
		BinaryRuleCount:             SafeCount(request.GetBinaryRuleCount()),
		CertificateRuleCount:        SafeCount(request.GetCertificateRuleCount()),
		CompilerRuleCount:           SafeCount(request.GetCompilerRuleCount()),
		TransitiveRuleCount:         SafeCount(request.GetTransitiveRuleCount()),
		TeamIDRuleCount:             SafeCount(request.GetTeamidRuleCount()),
		SigningIDRuleCount:          SafeCount(request.GetSigningidRuleCount()),
		CDHashRuleCount:             SafeCount(request.GetCdhashRuleCount()),
		RulesReceived:               state.RulesReceived,
		RulesProcessed:              state.RulesProcessed,
		LastRuleSyncAttemptAt:       state.LastRuleSyncAttemptAt,
		LastRuleSyncSuccessAt:       state.LastRuleSyncSuccessAt,
		LastReportedCountsMatchAt:   reportedCountsMatchAt,
	}
}

func diffSnapshot(current []model.PendingRuleTarget, applied []model.AppliedRuleTarget) []model.SyncRule {
	appliedByKey := make(map[string]model.AppliedRuleTarget, len(applied))
	for _, target := range applied {
		appliedByKey[string(target.RuleType)+"|"+target.Identifier] = target
	}

	payload := make([]model.SyncRule, 0, len(current)+len(applied))
	currentKeys := make(map[string]struct{}, len(current))
	for _, target := range current {
		key := domain.MachineRuleTargetKey(target.MachineRuleTarget)
		currentKeys[key] = struct{}{}

		appliedTarget, exists := appliedByKey[key]
		if !exists || appliedTarget.PayloadHash != target.PayloadHash {
			payload = append(payload, model.SyncRule{MachineRuleTarget: target.MachineRuleTarget})
		}
	}

	for _, target := range applied {
		if _, exists := currentKeys[string(target.RuleType)+"|"+target.Identifier]; !exists {
			payload = append(payload, model.SyncRule{
				MachineRuleTarget: domain.MachineRuleTarget{
					RuleType:   target.RuleType,
					Identifier: target.Identifier,
				},
				Removed: true,
			})
		}
	}

	return payload
}

func fullSyncRules(targets []model.PendingRuleTarget) []model.SyncRule {
	rules := make([]model.SyncRule, 0, len(targets))
	for _, target := range targets {
		rules = append(rules, model.SyncRule{MachineRuleTarget: target.MachineRuleTarget})
	}

	return rules
}

func withRulePayloadHashes(targets []domain.MachineResolvedRule) []model.PendingRuleTarget {
	result := make([]model.PendingRuleTarget, 0, len(targets))
	for _, target := range targets {
		result = append(result, model.PendingRuleTarget{
			MachineRuleTarget: target.MachineRuleTarget,
			PayloadHash:       domain.MachineRuleTargetPayloadHash(target.MachineRuleTarget),
		})
	}

	return result
}

func toAppliedRuleTargets(targets []model.PendingRuleTarget) []model.AppliedRuleTarget {
	result := make([]model.AppliedRuleTarget, 0, len(targets))
	for _, target := range targets {
		result = append(result, model.AppliedRuleTarget{
			RuleType:    target.RuleType,
			Identifier:  target.Identifier,
			PayloadHash: target.PayloadHash,
		})
	}

	return result
}

func appliedRuleTargetsMatch(left []model.AppliedRuleTarget, right []model.AppliedRuleTarget) bool {
	return slices.Equal(left, right)
}

func managedRuleCountsMatch(expected domain.ExecutionRuleCounts, request *syncv1.PreflightRequest) bool {
	return expected.Binary == SafeCount(request.GetBinaryRuleCount()) &&
		expected.Certificate == SafeCount(request.GetCertificateRuleCount()) &&
		expected.TeamID == SafeCount(request.GetTeamidRuleCount()) &&
		expected.SigningID == SafeCount(request.GetSigningidRuleCount()) &&
		expected.CDHash == SafeCount(request.GetCdhashRuleCount())
}

func countPendingRuleTargets(targets []model.PendingRuleTarget) domain.ExecutionRuleCounts {
	rules := make([]domain.MachineRuleTarget, 0, len(targets))
	for _, target := range targets {
		rules = append(rules, target.MachineRuleTarget)
	}

	return domain.CountExecutionRules(rules)
}
