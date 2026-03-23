// Package snapshot prepares and reloads the frozen state for a single Santa
// sync cycle.
//
// Santa sync is two-phase: preflight freezes the desired state, rule download
// serves that exact snapshot, and postflight acknowledges the result after the
// client reports it processed the payload for that snapshot.
package snapshot

import (
	"cmp"
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

// PendingSnapshot is the frozen rule payload prepared during preflight and
// reused for the rest of the sync cycle.
type PendingSnapshot struct {
	FullSync         bool
	Payload          []model.SyncRule
	PayloadRuleCount int64
}

// PreparePendingSnapshot resolves the machine's desired rules, freezes the
// resulting payload, and stores it as the pending snapshot for the current
// sync cycle.
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

	resolvedRules, err := ruleResolver.ResolveMachineRuleTargets(ctx, machineID)
	if err != nil {
		return PendingSnapshot{}, fmt.Errorf("resolve machine rule targets: %w", err)
	}

	pendingTargets := pendingRuleTargets(resolvedRules)
	snapshot, write := planPendingSnapshot(state, pendingTargets, request, preparedAt, machineID)

	if err = store.ReplacePendingSnapshot(ctx, write); err != nil {
		return PendingSnapshot{}, fmt.Errorf("replace pending snapshot: %w", err)
	}

	return snapshot, nil
}

// LoadPendingSnapshot reloads the frozen snapshot created during preflight so
// rule download can serve one stable payload for the full sync cycle.
func LoadPendingSnapshot(
	ctx context.Context,
	store model.DataStore,
	machineID uuid.UUID,
) (PendingSnapshot, model.MachineSyncState, error) {
	state, err := store.GetMachineSyncState(ctx, machineID)
	if err != nil {
		return PendingSnapshot{}, model.MachineSyncState{}, fmt.Errorf("get machine sync state: %w", err)
	}

	if state.PendingPreflightAt == nil {
		return PendingSnapshot{}, model.MachineSyncState{}, ErrPendingSnapshotNotFound
	}

	snapshot := PendingSnapshot{
		FullSync:         state.PendingFullSync,
		Payload:          slices.Clone(state.PendingPayload),
		PayloadRuleCount: state.PendingPayloadRuleCount,
	}

	return snapshot, state, nil
}

func planPendingSnapshot(
	state model.MachineSyncState,
	pendingTargets []model.PendingRuleTarget,
	request *syncv1.PreflightRequest,
	preparedAt time.Time,
	machineID uuid.UUID,
) (PendingSnapshot, model.PendingSnapshotWrite) {
	pendingTargets = sortPendingRuleTargets(pendingTargets)

	desiredTargets := toAppliedRuleTargets(pendingTargets)
	appliedTargets := sortAppliedRuleTargets(slices.Clone(state.AppliedTargets))

	desiredCounts := pendingRuleTargetCounts(pendingTargets)
	reportedCounts := preflightRuleCounts(request)
	reportedCountsMatch := desiredCounts == reportedCounts

	reportedCountsMatchAt := state.LastReportedCountsMatchAt
	if reportedCountsMatch {
		reportedCountsMatchAt = &preparedAt
	}

	targetsMatch := slices.Equal(desiredTargets, appliedTargets)
	fullSync := request.GetRequestCleanSync() || (targetsMatch && !reportedCountsMatch)

	payload := buildIncrementalPayload(pendingTargets, appliedTargets)
	if fullSync {
		payload = buildFullSyncPayload(pendingTargets)
	}

	payloadRuleCount := int64(len(payload))
	snapshot := PendingSnapshot{
		FullSync:         fullSync,
		Payload:          payload,
		PayloadRuleCount: payloadRuleCount,
	}

	write := model.PendingSnapshotWrite{
		MachineID:                   machineID,
		RulesHash:                   state.RulesHash,
		DesiredTargets:              slices.Clone(desiredTargets),
		AppliedTargets:              slices.Clone(appliedTargets),
		SentTargets:                 desiredTargets,
		PendingPayload:              slices.Clone(payload),
		PendingPayloadRuleCount:     payloadRuleCount,
		PendingFullSync:             fullSync,
		PendingPreflightAt:          preparedAt,
		DesiredBinaryRuleCount:      desiredCounts.Binary,
		DesiredCertificateRuleCount: desiredCounts.Certificate,
		DesiredTeamIDRuleCount:      desiredCounts.TeamID,
		DesiredSigningIDRuleCount:   desiredCounts.SigningID,
		DesiredCDHashRuleCount:      desiredCounts.CDHash,
		BinaryRuleCount:             ClampRuleCount(request.GetBinaryRuleCount()),
		CertificateRuleCount:        ClampRuleCount(request.GetCertificateRuleCount()),
		TeamIDRuleCount:             ClampRuleCount(request.GetTeamidRuleCount()),
		SigningIDRuleCount:          ClampRuleCount(request.GetSigningidRuleCount()),
		CDHashRuleCount:             ClampRuleCount(request.GetCdhashRuleCount()),
		RulesReceived:               state.RulesReceived,
		RulesProcessed:              state.RulesProcessed,
		LastRuleSyncAttemptAt:       state.LastRuleSyncAttemptAt,
		LastRuleSyncSuccessAt:       state.LastRuleSyncSuccessAt,
		LastReportedCountsMatchAt:   reportedCountsMatchAt,
	}

	return snapshot, write
}

func buildIncrementalPayload(
	pendingTargets []model.PendingRuleTarget,
	appliedTargets []model.AppliedRuleTarget,
) []model.SyncRule {
	appliedByKey := make(map[string]model.AppliedRuleTarget, len(appliedTargets))
	for _, target := range appliedTargets {
		appliedByKey[appliedRuleTargetKey(target)] = target
	}

	currentKeys := make(map[string]struct{}, len(pendingTargets))
	payload := make([]model.SyncRule, 0, len(pendingTargets)+len(appliedTargets))

	for _, target := range pendingTargets {
		key := domain.MachineRuleTargetKey(target.MachineRuleTarget)
		currentKeys[key] = struct{}{}

		appliedTarget, ok := appliedByKey[key]
		if !ok || appliedTarget.PayloadHash != target.PayloadHash {
			payload = append(payload, model.SyncRule{
				MachineRuleTarget: target.MachineRuleTarget,
			})
		}
	}

	for _, target := range appliedTargets {
		key := appliedRuleTargetKey(target)
		if _, ok := currentKeys[key]; ok {
			continue
		}

		payload = append(payload, model.SyncRule{
			MachineRuleTarget: domain.MachineRuleTarget{
				RuleType:   target.RuleType,
				Identifier: target.Identifier,
			},
			Removed: true,
		})
	}

	return sortSyncRules(payload)
}

func buildFullSyncPayload(targets []model.PendingRuleTarget) []model.SyncRule {
	rules := make([]model.SyncRule, 0, len(targets))
	for _, target := range targets {
		rules = append(rules, model.SyncRule{
			MachineRuleTarget: target.MachineRuleTarget,
		})
	}

	return sortSyncRules(rules)
}

func pendingRuleTargets(resolvedRules []domain.MachineResolvedRule) []model.PendingRuleTarget {
	targets := make([]model.PendingRuleTarget, 0, len(resolvedRules))
	for _, rule := range resolvedRules {
		targets = append(targets, model.PendingRuleTarget{
			MachineRuleTarget: rule.MachineRuleTarget,
			PayloadHash:       domain.MachineRuleTargetPayloadHash(rule.MachineRuleTarget),
		})
	}

	return sortPendingRuleTargets(targets)
}

func toAppliedRuleTargets(targets []model.PendingRuleTarget) []model.AppliedRuleTarget {
	applied := make([]model.AppliedRuleTarget, 0, len(targets))
	for _, target := range targets {
		applied = append(applied, model.AppliedRuleTarget{
			RuleType:    target.RuleType,
			Identifier:  target.Identifier,
			PayloadHash: target.PayloadHash,
		})
	}

	return sortAppliedRuleTargets(applied)
}

func pendingRuleTargetCounts(targets []model.PendingRuleTarget) domain.ExecutionRuleCounts {
	rules := make([]domain.MachineRuleTarget, 0, len(targets))
	for _, target := range targets {
		rules = append(rules, target.MachineRuleTarget)
	}

	return domain.CountExecutionRules(rules)
}

func preflightRuleCounts(request *syncv1.PreflightRequest) domain.ExecutionRuleCounts {
	return domain.ExecutionRuleCounts{
		Binary:      ClampRuleCount(request.GetBinaryRuleCount()),
		Certificate: ClampRuleCount(request.GetCertificateRuleCount()),
		TeamID:      ClampRuleCount(request.GetTeamidRuleCount()),
		SigningID:   ClampRuleCount(request.GetSigningidRuleCount()),
		CDHash:      ClampRuleCount(request.GetCdhashRuleCount()),
	}
}

func appliedRuleTargetKey(target model.AppliedRuleTarget) string {
	return domain.MachineRuleTargetKey(domain.MachineRuleTarget{
		RuleType:   target.RuleType,
		Identifier: target.Identifier,
	})
}

type withKey[T any] struct {
	key string
	val T
}

func sortByPrecomputedKey[T any](items []T, keyFn func(T) string) []T {
	keyed := make([]withKey[T], len(items))
	for i, item := range items {
		keyed[i] = withKey[T]{key: keyFn(item), val: item}
	}

	slices.SortFunc(keyed, func(a, b withKey[T]) int {
		return cmp.Compare(a.key, b.key)
	})

	for i, k := range keyed {
		items[i] = k.val
	}

	return items
}

func sortPendingRuleTargets(targets []model.PendingRuleTarget) []model.PendingRuleTarget {
	return sortByPrecomputedKey(targets, func(t model.PendingRuleTarget) string {
		return domain.MachineRuleTargetKey(t.MachineRuleTarget)
	})
}

func sortAppliedRuleTargets(targets []model.AppliedRuleTarget) []model.AppliedRuleTarget {
	return sortByPrecomputedKey(targets, appliedRuleTargetKey)
}

func sortSyncRules(rules []model.SyncRule) []model.SyncRule {
	return sortByPrecomputedKey(rules, func(r model.SyncRule) string {
		return domain.MachineRuleTargetKey(r.MachineRuleTarget)
	})
}
