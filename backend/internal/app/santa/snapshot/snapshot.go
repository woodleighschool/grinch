// Package snapshot freezes one Santa sync cycle into a pending desired state.
//
// Santa sync is inherently two-phase: preflight decides what the machine
// should end up with, rule download must serve that same frozen state, and
// postflight acknowledges the frozen result only after the client reports the
// final database hash it actually reached.
package snapshot

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"
	"github.com/zeebo/xxh3"

	"github.com/woodleighschool/grinch/internal/app/santa/model"
	"github.com/woodleighschool/grinch/internal/domain"
)

var ErrPendingSnapshotNotFound = errors.New("pending rule sync snapshot not found")

const (
	santaStoredRuleTypeUnknown     uint32 = 0
	santaStoredRuleTypeCDHash      uint32 = 500
	santaStoredRuleTypeBinary      uint32 = 1000
	santaStoredRuleTypeSigningID   uint32 = 2000
	santaStoredRuleTypeCertificate uint32 = 3000
	santaStoredRuleTypeTeamID      uint32 = 4000

	santaStoredRuleStateUnknown     uint32 = 0
	santaStoredRuleStateAllow       uint32 = 1
	santaStoredRuleStateBlock       uint32 = 2
	santaStoredRuleStateSilentBlock uint32 = 3
	santaStoredRuleStateCEL         uint32 = 9
)

type PendingSnapshot struct {
	FullSync          bool
	Targets           []model.StoredRuleTarget
	Payload           []model.SyncRule
	ExpectedRulesHash string
	PayloadRuleCount  int64
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

	payload := fullSyncRules(state.PendingTargets)
	if !state.PendingFullSync {
		payload = diffSnapshot(state.PendingTargets, state.AppliedTargets)
	}

	return PendingSnapshot{
		FullSync:          state.PendingFullSync,
		Targets:           cloneTargets(state.PendingTargets),
		Payload:           payload,
		ExpectedRulesHash: state.ExpectedRulesHash,
		PayloadRuleCount:  state.PendingPayloadRuleCount,
	}, state, nil
}

// PostflightMatchesSnapshot promotes only when the client reports the exact
// final rule database hash we expected and the exact number of payload rules it
// successfully processed for this frozen sync cycle.
func PostflightMatchesSnapshot(request *syncv1.PostflightRequest, state model.MachineSyncState) bool {
	if state.PendingPreflightAt == nil {
		return false
	}
	if strings.TrimSpace(request.GetRulesHash()) != state.ExpectedRulesHash {
		return false
	}
	if int64(request.GetRulesProcessed()) != state.PendingPayloadRuleCount {
		return false
	}

	return true
}

func buildPendingSnapshot(
	state model.MachineSyncState,
	current []model.StoredRuleTarget,
	request *syncv1.PreflightRequest,
	preparedAt time.Time,
	machineID uuid.UUID,
) (PendingSnapshot, model.PendingSnapshotWrite) {
	applied := cloneTargets(state.AppliedTargets)
	payload := diffSnapshot(current, applied)
	expectedRulesHash := SantaRulesHash(current)
	clientRulesHash := strings.TrimSpace(request.GetRulesHash())
	fullSync := request.GetRequestCleanSync() || clientRulesHash != SantaRulesHash(applied)
	if fullSync {
		payload = fullSyncRules(current)
	}

	payloadRuleCount := countSyncRules(payload)
	snapshot := PendingSnapshot{
		FullSync:          fullSync,
		Targets:           cloneTargets(current),
		Payload:           payload,
		ExpectedRulesHash: expectedRulesHash,
		PayloadRuleCount:  payloadRuleCount,
	}

	return snapshot, model.PendingSnapshotWrite{
		MachineID:               machineID,
		RulesHash:               state.RulesHash,
		AppliedTargets:          applied,
		PendingTargets:          cloneTargets(current),
		ExpectedRulesHash:       expectedRulesHash,
		PendingPayloadRuleCount: payloadRuleCount,
		PendingFullSync:         fullSync,
		PendingPreflightAt:      preparedAt,
		ClientMode:              machineClientModeFromProto(request.GetClientMode()),
		BinaryRuleCount:         int32(request.GetBinaryRuleCount()),
		CertificateRuleCount:    int32(request.GetCertificateRuleCount()),
		CompilerRuleCount:       int32(request.GetCompilerRuleCount()),
		TransitiveRuleCount:     int32(request.GetTransitiveRuleCount()),
		TeamIDRuleCount:         int32(request.GetTeamidRuleCount()),
		SigningIDRuleCount:      int32(request.GetSigningidRuleCount()),
		CDHashRuleCount:         int32(request.GetCdhashRuleCount()),
		RulesReceived:           state.RulesReceived,
		RulesProcessed:          state.RulesProcessed,
		LastRuleSyncAttemptAt:   state.LastRuleSyncAttemptAt,
		LastRuleSyncSuccessAt:   state.LastRuleSyncSuccessAt,
	}
}

func diffSnapshot(current []model.StoredRuleTarget, applied []model.StoredRuleTarget) []model.SyncRule {
	appliedByKey := make(map[string]model.StoredRuleTarget, len(applied))
	for _, target := range applied {
		appliedByKey[ruleTargetKey(target.MachineRuleTarget)] = target
	}

	payload := make([]model.SyncRule, 0, len(current)+len(applied))
	currentKeys := make(map[string]struct{}, len(current))
	for _, target := range current {
		key := ruleTargetKey(target.MachineRuleTarget)
		currentKeys[key] = struct{}{}

		appliedTarget, exists := appliedByKey[key]
		if !exists || appliedTarget.PayloadHash != target.PayloadHash {
			payload = append(payload, model.SyncRule{StoredRuleTarget: target})
		}
	}

	for _, target := range applied {
		if _, exists := currentKeys[ruleTargetKey(target.MachineRuleTarget)]; !exists {
			payload = append(payload, model.SyncRule{
				StoredRuleTarget: target,
				Removed:          true,
			})
		}
	}

	return payload
}

func fullSyncRules(targets []model.StoredRuleTarget) []model.SyncRule {
	rules := make([]model.SyncRule, 0, len(targets))
	for _, target := range targets {
		rules = append(rules, model.SyncRule{StoredRuleTarget: target})
	}

	return rules
}

func withRulePayloadHashes(targets []domain.MachineResolvedRule) []model.StoredRuleTarget {
	result := make([]model.StoredRuleTarget, 0, len(targets))
	for _, target := range targets {
		ruleID := target.RuleID
		result = append(result, model.StoredRuleTarget{
			MachineRuleTarget: target.MachineRuleTarget,
			RuleID:            &ruleID,
			RuleName:          target.Name,
			PayloadHash:       ruleTargetPayloadHash(target.MachineRuleTarget),
		})
	}

	return result
}

func cloneTargets(targets []model.StoredRuleTarget) []model.StoredRuleTarget {
	if len(targets) == 0 {
		return nil
	}

	cloned := make([]model.StoredRuleTarget, 0, len(targets))
	cloned = append(cloned, targets...)
	return cloned
}

func countSyncRules(rules []model.SyncRule) int64 {
	var count int64
	for range rules {
		count++
	}

	return count
}

func machineClientModeFromProto(value syncv1.ClientMode) domain.MachineClientMode {
	switch value {
	case syncv1.ClientMode_MONITOR:
		return domain.MachineClientModeMonitor
	case syncv1.ClientMode_LOCKDOWN:
		return domain.MachineClientModeLockdown
	case syncv1.ClientMode_STANDALONE:
		return domain.MachineClientModeStandalone
	default:
		return domain.MachineClientModeUnknown
	}
}

func ruleTargetKey(target domain.MachineRuleTarget) string {
	return string(target.RuleType) + "|" + target.IdentifierKey
}

func RuleTargetKey(target domain.MachineRuleTarget) string {
	return ruleTargetKey(target)
}

func ruleTargetPayloadHash(target domain.MachineRuleTarget) string {
	return stableHash(
		string(target.RuleType),
		target.IdentifierKey,
		target.Identifier,
		string(target.Policy),
		target.CustomMessage,
		target.CustomURL,
		target.CELExpression,
	)
}

func RuleTargetPayloadHash(target domain.MachineRuleTarget) string {
	return ruleTargetPayloadHash(target)
}

func stableHash(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x1f")))
	return hex.EncodeToString(sum[:])
}

// SantaRulesHash mirrors Santa's dynamic execution-rule hash:
// identifier bytes, CEL bytes, stored state int32 LE, stored type int32 LE,
// iterated by (identifier ASC, stored type ASC), excluding transitive rules.
func SantaRulesHash(targets []model.StoredRuleTarget) string {
	if len(targets) == 0 {
		return ""
	}

	sortedTargets := cloneTargets(targets)
	slices.SortFunc(sortedTargets, func(left model.StoredRuleTarget, right model.StoredRuleTarget) int {
		if left.IdentifierKey != right.IdentifierKey {
			return strings.Compare(left.IdentifierKey, right.IdentifierKey)
		}

		return int(santaStoredRuleType(left.RuleType)) - int(santaStoredRuleType(right.RuleType))
	})

	var payload bytes.Buffer
	var intBuffer [4]byte
	for _, target := range sortedTargets {
		if target.Policy == domain.RulePolicyCEL {
			payload.WriteString(target.Identifier)
			payload.WriteString(target.CELExpression)
		} else {
			payload.WriteString(target.Identifier)
		}

		binary.LittleEndian.PutUint32(intBuffer[:], santaStoredRuleState(target.Policy))
		payload.Write(intBuffer[:])

		binary.LittleEndian.PutUint32(intBuffer[:], santaStoredRuleType(target.RuleType))
		payload.Write(intBuffer[:])
	}

	sum := xxh3.Hash128(payload.Bytes())
	return fmt.Sprintf("%016x%016x", sum.Hi, sum.Lo)
}

func santaStoredRuleType(value domain.RuleType) uint32 {
	switch value {
	case domain.RuleTypeCDHash:
		return santaStoredRuleTypeCDHash
	case domain.RuleTypeBinary:
		return santaStoredRuleTypeBinary
	case domain.RuleTypeSigningID:
		return santaStoredRuleTypeSigningID
	case domain.RuleTypeCertificate:
		return santaStoredRuleTypeCertificate
	case domain.RuleTypeTeamID:
		return santaStoredRuleTypeTeamID
	default:
		return santaStoredRuleTypeUnknown
	}
}

func santaStoredRuleState(value domain.RulePolicy) uint32 {
	switch value {
	case domain.RulePolicyAllowlist:
		return santaStoredRuleStateAllow
	case domain.RulePolicyBlocklist:
		return santaStoredRuleStateBlock
	case domain.RulePolicySilentBlocklist:
		return santaStoredRuleStateSilentBlock
	case domain.RulePolicyCEL:
		return santaStoredRuleStateCEL
	default:
		return santaStoredRuleStateUnknown
	}
}
