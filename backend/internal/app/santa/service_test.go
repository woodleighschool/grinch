package santa_test

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/app/santa"
	"github.com/woodleighschool/grinch/internal/domain"
	santamodel "github.com/woodleighschool/grinch/internal/santa/model"
)

type testStore struct {
	resolver    *testRuleResolver
	syncStates  map[uuid.UUID]santamodel.MachineSyncState
	upsertErr   error
	lastUpsert  santamodel.MachineUpsert
	upsertCalls int
}

type testRuleResolver struct {
	resolvedRules []domain.MachineResolvedRule
}

func (s *testStore) UpsertMachine(_ context.Context, machine santamodel.MachineUpsert) error {
	s.upsertCalls++
	s.lastUpsert = machine

	state := s.ensureSyncState(machine.MachineID)
	s.syncStates[machine.MachineID] = state

	return s.upsertErr
}

func (s *testStore) GetMachineSyncState(
	_ context.Context,
	machineID uuid.UUID,
) (santamodel.MachineSyncState, error) {
	state := s.ensureSyncState(machineID)
	return state, nil
}

func (s *testStore) UpdateMachineDesiredTargets(_ context.Context, machineID uuid.UUID) error {
	state := s.ensureSyncState(machineID)

	desiredTargets := make([]santamodel.AppliedRuleTarget, 0, len(s.resolver.resolvedRules))
	for _, rule := range s.resolver.resolvedRules {
		desiredTargets = append(desiredTargets, santamodel.AppliedRuleTarget{
			RuleType:    rule.RuleType,
			Identifier:  rule.Identifier,
			PayloadHash: domain.MachineRuleTargetPayloadHash(rule.MachineRuleTarget),
		})
	}

	state.DesiredTargets = desiredTargets
	s.syncStates[machineID] = state

	return nil
}

func (s *testStore) ReplacePendingSnapshot(_ context.Context, pending santamodel.PendingSnapshotWrite) error {
	s.ensureSyncState(pending.MachineID)

	s.syncStates[pending.MachineID] = santamodel.MachineSyncState{
		MachineID:                   pending.MachineID,
		RulesHash:                   pending.RulesHash,
		DesiredTargets:              slices.Clone(pending.DesiredTargets),
		AppliedTargets:              slices.Clone(pending.AppliedTargets),
		SentTargets:                 slices.Clone(pending.SentTargets),
		PendingPayload:              slices.Clone(pending.PendingPayload),
		PendingPayloadRuleCount:     pending.PendingPayloadRuleCount,
		PendingFullSync:             pending.PendingFullSync,
		PendingPreflightAt:          &pending.PendingPreflightAt,
		DesiredBinaryRuleCount:      pending.DesiredBinaryRuleCount,
		DesiredCertificateRuleCount: pending.DesiredCertificateRuleCount,
		DesiredTeamIDRuleCount:      pending.DesiredTeamIDRuleCount,
		DesiredSigningIDRuleCount:   pending.DesiredSigningIDRuleCount,
		DesiredCDHashRuleCount:      pending.DesiredCDHashRuleCount,
		BinaryRuleCount:             pending.BinaryRuleCount,
		CertificateRuleCount:        pending.CertificateRuleCount,
		TeamIDRuleCount:             pending.TeamIDRuleCount,
		SigningIDRuleCount:          pending.SigningIDRuleCount,
		CDHashRuleCount:             pending.CDHashRuleCount,
		RulesReceived:               pending.RulesReceived,
		RulesProcessed:              pending.RulesProcessed,
		LastRuleSyncAttemptAt:       pending.LastRuleSyncAttemptAt,
		LastRuleSyncSuccessAt:       pending.LastRuleSyncSuccessAt,
		LastReportedCountsMatchAt:   pending.LastReportedCountsMatchAt,
	}

	return nil
}

func (s *testStore) RecordPostflight(_ context.Context, write santamodel.PostflightWrite) error {
	state := s.ensureSyncState(write.MachineID)
	state.RulesHash = strings.TrimSpace(write.RulesHash)
	state.RulesReceived = write.RulesReceived
	state.RulesProcessed = write.RulesProcessed
	state.LastRuleSyncAttemptAt = &write.LastRuleSyncAttemptAt
	s.syncStates[write.MachineID] = state

	return nil
}

func (s *testStore) PromotePendingSnapshot(
	_ context.Context,
	machineID uuid.UUID,
	completedAt time.Time,
) error {
	state := s.ensureSyncState(machineID)
	pendingFullSync := state.PendingFullSync

	state.AppliedTargets = slices.Clone(state.SentTargets)
	state.SentTargets = nil
	state.PendingPayload = nil
	state.PendingPayloadRuleCount = 0
	state.PendingFullSync = false
	state.PendingPreflightAt = nil
	state.LastRuleSyncSuccessAt = &completedAt
	if pendingFullSync {
		state.LastCleanSyncAt = &completedAt
	}

	s.syncStates[machineID] = state
	return nil
}

func (s *testStore) IngestEvents(
	context.Context,
	uuid.UUID,
	[]santamodel.ExecutionEventWrite,
	[]santamodel.FileAccessEventWrite,
) error {
	return nil
}

func (r *testRuleResolver) ResolveMachineRuleTargets(
	context.Context,
	uuid.UUID,
) ([]domain.MachineResolvedRule, error) {
	return r.resolvedRules, nil
}

func (s *testStore) ensureSyncState(machineID uuid.UUID) santamodel.MachineSyncState {
	if s.syncStates == nil {
		s.syncStates = make(map[uuid.UUID]santamodel.MachineSyncState)
	}

	state, ok := s.syncStates[machineID]
	if !ok {
		state = santamodel.MachineSyncState{MachineID: machineID}
		s.syncStates[machineID] = state
	}

	return state
}

func newTestService(store *testStore, resolver *testRuleResolver) *santa.Service {
	store.resolver = resolver
	return santa.New(newTestLogger(), store, nil, resolver)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHandlePreflight_UpsertsMachineAndReturnsSyncSettings(t *testing.T) {
	machineID := uuid.New()
	store := &testStore{}
	service := newTestService(store, &testRuleResolver{})

	req := syncv1.PreflightRequest_builder{
		MachineId:         machineID.String(),
		ClientMode:        syncv1.ClientMode_LOCKDOWN,
		SerialNumber:      "SER123",
		Hostname:          "macbook-01",
		ModelIdentifier:   "MacBookPro18,3",
		OsVersion:         "14.5",
		OsBuild:           "23F79",
		SantaVersion:      "2026.1",
		PrimaryUser:       "user@example.com",
		PrimaryUserGroups: []string{"g1", "g2"},
		RequestCleanSync:  true,
	}.Build()

	before := time.Now().UTC()
	resp, err := service.HandlePreflight(context.Background(), machineID, req)
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}

	if store.upsertCalls != 1 {
		t.Fatalf("upsertCalls = %d, want 1", store.upsertCalls)
	}
	if store.lastUpsert.MachineID != machineID {
		t.Fatalf("MachineID = %q, want %q", store.lastUpsert.MachineID, machineID)
	}
	if store.lastUpsert.LastSeenAt.Before(before) || store.lastUpsert.LastSeenAt.After(after) {
		t.Fatalf("LastSeenAt out of expected range: %s", store.lastUpsert.LastSeenAt)
	}

	if groups := store.lastUpsert.PrimaryUserGroups; len(groups) != 2 || groups[0] != "g1" || groups[1] != "g2" {
		t.Fatalf("PrimaryUserGroups = %#v, want [g1 g2]", groups)
	}
	if resp.GetSyncType() != syncv1.SyncType_CLEAN {
		t.Fatalf("SyncType = %v, want CLEAN", resp.GetSyncType())
	}
	if store.lastUpsert.ClientMode != domain.MachineClientModeLockdown {
		t.Fatalf("ClientMode = %q, want lockdown", store.lastUpsert.ClientMode)
	}
}

func TestHandlePreflight_ReturnsNormalWhenManagedCountsMatch(t *testing.T) {
	machineID := uuid.New()
	acknowledgedRule := domain.MachineRuleTarget{
		RuleType:   domain.RuleTypeBinary,
		Identifier: "com.example.existing",
		Policy:     domain.RulePolicyAllowlist,
	}
	acknowledgedTarget := appliedTargetFromRuleTarget(acknowledgedRule)

	store := &testStore{
		syncStates: map[uuid.UUID]santamodel.MachineSyncState{
			machineID: {
				MachineID:      machineID,
				AppliedTargets: []santamodel.AppliedRuleTarget{acknowledgedTarget},
			},
		},
	}

	service := newTestService(store, &testRuleResolver{
		resolvedRules: []domain.MachineResolvedRule{
			resolvedRule(uuid.New(), "Existing", acknowledgedRule),
			resolvedRule(uuid.New(), "Cert", domain.MachineRuleTarget{
				RuleType:   domain.RuleTypeCertificate,
				Identifier: "ABCD",
				Policy:     domain.RulePolicyBlocklist,
			}),
		},
	})

	resp, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId:            machineID.String(),
			BinaryRuleCount:      1,
			CertificateRuleCount: 1,
		}.Build(),
	)
	if err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}
	if resp.GetSyncType() != syncv1.SyncType_NORMAL {
		t.Fatalf("SyncType = %v, want NORMAL", resp.GetSyncType())
	}
}

func TestHandlePreflight_ReturnsNormalWhenDesiredRulesChangedEvenIfManagedCountsDiverge(t *testing.T) {
	machineID := uuid.New()
	store := &testStore{
		syncStates: map[uuid.UUID]santamodel.MachineSyncState{
			machineID: {MachineID: machineID},
		},
	}

	service := newTestService(store, &testRuleResolver{
		resolvedRules: []domain.MachineResolvedRule{
			resolvedRule(uuid.New(), "Binary", domain.MachineRuleTarget{
				RuleType:   domain.RuleTypeBinary,
				Identifier: "com.example.binary",
				Policy:     domain.RulePolicyAllowlist,
			}),
		},
	})

	resp, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId:       machineID.String(),
			BinaryRuleCount: 0,
		}.Build(),
	)
	if err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}
	if resp.GetSyncType() != syncv1.SyncType_NORMAL {
		t.Fatalf("SyncType = %v, want NORMAL", resp.GetSyncType())
	}
}

func TestHandlePreflight_ReturnsCleanWhenAppliedMatchesDesiredAndReportedCountsDiverge(t *testing.T) {
	machineID := uuid.New()
	ruleTarget := domain.MachineRuleTarget{
		RuleType:   domain.RuleTypeBinary,
		Identifier: "com.example.binary",
		Policy:     domain.RulePolicyAllowlist,
	}
	appliedTarget := appliedTargetFromRuleTarget(ruleTarget)

	store := &testStore{
		syncStates: map[uuid.UUID]santamodel.MachineSyncState{
			machineID: {
				MachineID:      machineID,
				AppliedTargets: []santamodel.AppliedRuleTarget{appliedTarget},
			},
		},
	}

	service := newTestService(store, &testRuleResolver{
		resolvedRules: []domain.MachineResolvedRule{
			resolvedRule(uuid.New(), "Binary", ruleTarget),
		},
	})

	resp, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId:       machineID.String(),
			BinaryRuleCount: 0,
		}.Build(),
	)
	if err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}
	if resp.GetSyncType() != syncv1.SyncType_CLEAN {
		t.Fatalf("SyncType = %v, want CLEAN", resp.GetSyncType())
	}
}

func TestHandleRuleDownload_ReturnsChangedRulesAndRemovalsDuringNormalSync(t *testing.T) {
	machineID := uuid.New()
	removedTarget := appliedTargetFromRuleTarget(domain.MachineRuleTarget{
		RuleType:   domain.RuleTypeBinary,
		Identifier: "com.example.removed",
		Policy:     domain.RulePolicyAllowlist,
	})

	store := &testStore{
		syncStates: map[uuid.UUID]santamodel.MachineSyncState{
			machineID: {
				MachineID:      machineID,
				AppliedTargets: []santamodel.AppliedRuleTarget{removedTarget},
			},
		},
	}

	service := newTestService(store, &testRuleResolver{
		resolvedRules: []domain.MachineResolvedRule{
			resolvedRule(uuid.New(), "Cert", domain.MachineRuleTarget{
				RuleType:   domain.RuleTypeCertificate,
				Identifier: "ABCD",
				Policy:     domain.RulePolicyBlocklist,
			}),
		},
	})

	if _, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId:            machineID.String(),
			BinaryRuleCount:      0,
			CertificateRuleCount: 1,
		}.Build(),
	); err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}

	resp, err := service.HandleRuleDownload(
		context.Background(),
		machineID,
		syncv1.RuleDownloadRequest_builder{
			MachineId: machineID.String(),
		}.Build(),
	)
	if err != nil {
		t.Fatalf("HandleRuleDownload() error = %v", err)
	}

	if len(resp.GetRules()) != 2 {
		t.Fatalf("len(response.Rules) = %d, want 2", len(resp.GetRules()))
	}
	if resp.GetRules()[0].GetPolicy() != syncv1.Policy_REMOVE {
		t.Fatalf("rules[0] policy = %v, want REMOVE", resp.GetRules()[0].GetPolicy())
	}
	if resp.GetRules()[0].GetIdentifier() != "com.example.removed" {
		t.Fatalf("rules[0] identifier = %q, want com.example.removed", resp.GetRules()[0].GetIdentifier())
	}
	if resp.GetRules()[1].GetPolicy() != syncv1.Policy_BLOCKLIST {
		t.Fatalf("rules[1] policy = %v, want BLOCKLIST", resp.GetRules()[1].GetPolicy())
	}
}

func TestHandlePostflight_PromotesPendingSnapshotOnMatchingProcessedCount(t *testing.T) {
	machineID := uuid.New()
	ruleTarget := domain.MachineRuleTarget{
		RuleType:   domain.RuleTypeBinary,
		Identifier: "com.example.binary",
		Policy:     domain.RulePolicyAllowlist,
	}

	store := &testStore{}
	service := newTestService(store, &testRuleResolver{
		resolvedRules: []domain.MachineResolvedRule{
			resolvedRule(uuid.New(), "Binary", ruleTarget),
		},
	})

	if _, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId: machineID.String(),
		}.Build(),
	); err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}

	if _, err := service.HandlePostflight(
		context.Background(),
		machineID,
		syncv1.PostflightRequest_builder{
			MachineId:      machineID.String(),
			SyncType:       syncv1.SyncType_NORMAL,
			RulesHash:      "ignored-client-rules-hash",
			RulesProcessed: 1,
		}.Build(),
	); err != nil {
		t.Fatalf("HandlePostflight() error = %v", err)
	}

	state := store.syncStates[machineID]
	if len(state.AppliedTargets) != 1 {
		t.Fatalf("len(AppliedTargets) = %d, want 1", len(state.AppliedTargets))
	}
	if state.PendingFullSync {
		t.Fatalf("PendingFullSync = true, want false")
	}
	if state.LastRuleSyncAttemptAt == nil {
		t.Fatal("LastRuleSyncAttemptAt = nil, want timestamp")
	}
	if state.LastRuleSyncSuccessAt == nil {
		t.Fatal("LastRuleSyncSuccessAt = nil, want timestamp")
	}
}

func TestHandlePostflight_LeavesPendingSnapshotOnProcessedCountMismatch(t *testing.T) {
	machineID := uuid.New()
	ruleTarget := domain.MachineRuleTarget{
		RuleType:   domain.RuleTypeBinary,
		Identifier: "com.example.binary",
		Policy:     domain.RulePolicyAllowlist,
	}

	store := &testStore{}
	service := newTestService(store, &testRuleResolver{
		resolvedRules: []domain.MachineResolvedRule{
			resolvedRule(uuid.New(), "Binary", ruleTarget),
		},
	})

	if _, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId: machineID.String(),
		}.Build(),
	); err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}

	if _, err := service.HandlePostflight(
		context.Background(),
		machineID,
		syncv1.PostflightRequest_builder{
			MachineId:      machineID.String(),
			SyncType:       syncv1.SyncType_NORMAL,
			RulesHash:      "ignored-client-rules-hash",
			RulesProcessed: 0,
		}.Build(),
	); err != nil {
		t.Fatalf("HandlePostflight() error = %v", err)
	}

	state := store.syncStates[machineID]
	if len(state.AppliedTargets) != 0 {
		t.Fatalf("len(AppliedTargets) = %d, want 0", len(state.AppliedTargets))
	}
	if state.RulesHash != "ignored-client-rules-hash" {
		t.Fatalf("RulesHash = %q, want ignored-client-rules-hash", state.RulesHash)
	}
	if state.LastRuleSyncAttemptAt == nil {
		t.Fatal("LastRuleSyncAttemptAt = nil, want timestamp")
	}
}

func TestHandleRuleDownload_AllowsEmptyPendingSnapshot(t *testing.T) {
	machineID := uuid.New()
	store := &testStore{}
	service := newTestService(store, &testRuleResolver{})

	if _, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId: machineID.String(),
		}.Build(),
	); err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}

	resp, err := service.HandleRuleDownload(
		context.Background(),
		machineID,
		syncv1.RuleDownloadRequest_builder{
			MachineId: machineID.String(),
		}.Build(),
	)
	if err != nil {
		t.Fatalf("HandleRuleDownload() error = %v", err)
	}
	if len(resp.GetRules()) != 0 {
		t.Fatalf("len(response.Rules) = %d, want 0", len(resp.GetRules()))
	}
}

func TestHandlePostflight_PromotesEmptyPendingSnapshot(t *testing.T) {
	machineID := uuid.New()
	store := &testStore{}
	service := newTestService(store, &testRuleResolver{})

	if _, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId: machineID.String(),
		}.Build(),
	); err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}

	if _, err := service.HandlePostflight(
		context.Background(),
		machineID,
		syncv1.PostflightRequest_builder{
			MachineId:      machineID.String(),
			SyncType:       syncv1.SyncType_CLEAN,
			RulesHash:      "ignored-client-rules-hash",
			RulesProcessed: 0,
		}.Build(),
	); err != nil {
		t.Fatalf("HandlePostflight() error = %v", err)
	}

	state := store.syncStates[machineID]
	if state.PendingPreflightAt != nil {
		t.Fatal("PendingPreflightAt != nil, want cleared")
	}
	if len(state.AppliedTargets) != 0 {
		t.Fatalf("len(AppliedTargets) = %d, want 0", len(state.AppliedTargets))
	}
	if state.LastRuleSyncSuccessAt == nil {
		t.Fatal("LastRuleSyncSuccessAt = nil, want timestamp")
	}
}

func appliedTargetFromRuleTarget(target domain.MachineRuleTarget) santamodel.AppliedRuleTarget {
	return santamodel.AppliedRuleTarget{
		RuleType:    target.RuleType,
		Identifier:  target.Identifier,
		PayloadHash: domain.MachineRuleTargetPayloadHash(target),
	}
}

func resolvedRule(ruleID uuid.UUID, name string, target domain.MachineRuleTarget) domain.MachineResolvedRule {
	return domain.MachineResolvedRule{
		MachineRuleTarget: target,
		RuleID:            ruleID,
		Name:              name,
	}
}
