package santa_test

import (
	"context"
	"encoding/json"
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

type fakeDataStore struct {
	ruleSyncStates map[uuid.UUID]santamodel.MachineSyncState
	upsertErr      error
	upsertMachine  santamodel.MachineUpsert
	upsertCalls    int
}

type fakeRuleResolver struct {
	resolvedTargets []domain.MachineResolvedRule
}

func (store *fakeDataStore) UpsertMachine(_ context.Context, machine santamodel.MachineUpsert) error {
	store.upsertCalls++
	store.upsertMachine = machine
	if store.ruleSyncStates == nil {
		store.ruleSyncStates = make(map[uuid.UUID]santamodel.MachineSyncState)
	}
	if _, exists := store.ruleSyncStates[machine.MachineID]; !exists {
		store.ruleSyncStates[machine.MachineID] = santamodel.MachineSyncState{MachineID: machine.MachineID}
	}
	return store.upsertErr
}

func (store *fakeDataStore) GetMachineSyncState(
	_ context.Context,
	machineID uuid.UUID,
) (santamodel.MachineSyncState, error) {
	if store.ruleSyncStates == nil {
		store.ruleSyncStates = make(map[uuid.UUID]santamodel.MachineSyncState)
	}
	state, exists := store.ruleSyncStates[machineID]
	if !exists {
		state = santamodel.MachineSyncState{MachineID: machineID}
		store.ruleSyncStates[machineID] = state
	}
	return state, nil
}

func (store *fakeDataStore) SyncMachineDesiredRuleTargets(context.Context, uuid.UUID) error {
	return nil
}

func (store *fakeDataStore) ReplacePendingSnapshot(_ context.Context, pending santamodel.PendingSnapshotWrite) error {
	if store.ruleSyncStates == nil {
		store.ruleSyncStates = make(map[uuid.UUID]santamodel.MachineSyncState)
	}

	store.ruleSyncStates[pending.MachineID] = santamodel.MachineSyncState{
		MachineID:                   pending.MachineID,
		RulesHash:                   pending.RulesHash,
		DesiredTargets:              slices.Clone(pending.DesiredTargets),
		AppliedTargets:              slices.Clone(pending.AppliedTargets),
		PendingTargets:              slices.Clone(pending.PendingTargets),
		PendingPayload:              slices.Clone(pending.PendingPayload),
		PendingPayloadRuleCount:     pending.PendingPayloadRuleCount,
		PendingFullSync:             pending.PendingFullSync,
		PendingPreflightAt:          &pending.PendingPreflightAt,
		DesiredBinaryRuleCount:      pending.DesiredBinaryRuleCount,
		DesiredCertificateRuleCount: pending.DesiredCertificateRuleCount,
		DesiredTeamIDRuleCount:      pending.DesiredTeamIDRuleCount,
		DesiredSigningIDRuleCount:   pending.DesiredSigningIDRuleCount,
		DesiredCDHashRuleCount:      pending.DesiredCDHashRuleCount,
		ClientMode:                  pending.ClientMode,
		BinaryRuleCount:             pending.BinaryRuleCount,
		CertificateRuleCount:        pending.CertificateRuleCount,
		CompilerRuleCount:           pending.CompilerRuleCount,
		TransitiveRuleCount:         pending.TransitiveRuleCount,
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

func (store *fakeDataStore) RecordPostflight(_ context.Context, write santamodel.PostflightWrite) error {
	state := store.ruleSyncStates[write.MachineID]
	state.RulesHash = strings.TrimSpace(write.RulesHash)
	state.RulesReceived = write.RulesReceived
	state.RulesProcessed = write.RulesProcessed
	state.LastRuleSyncAttemptAt = &write.LastRuleSyncAttemptAt
	store.ruleSyncStates[write.MachineID] = state
	return nil
}

func (store *fakeDataStore) PromotePendingSnapshot(
	_ context.Context,
	machineID uuid.UUID,
	completedAt time.Time,
) error {
	state := store.ruleSyncStates[machineID]
	pendingFullSync := state.PendingFullSync
	state.AppliedTargets = slices.Clone(state.PendingTargets)
	state.PendingTargets = nil
	state.PendingPayload = nil
	state.PendingPayloadRuleCount = 0
	state.PendingFullSync = false
	state.PendingPreflightAt = nil
	state.LastRuleSyncSuccessAt = &completedAt
	if pendingFullSync {
		state.LastCleanSyncAt = &completedAt
	}
	store.ruleSyncStates[machineID] = state
	return nil
}

func (store *fakeDataStore) IngestEvents(
	context.Context,
	uuid.UUID,
	[]santamodel.ExecutionEventWrite,
	[]santamodel.FileAccessEventWrite,
) error {
	return nil
}

func (resolver *fakeRuleResolver) ResolveMachineRuleTargets(
	context.Context,
	uuid.UUID,
) ([]domain.MachineResolvedRule, error) {
	return resolver.resolvedTargets, nil
}

func newService(store *fakeDataStore, resolver *fakeRuleResolver) *santa.Service {
	return santa.New(testLogger(), store, nil, resolver)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHandlePreflight_UpsertsMachineAndReturnsSyncSettings(t *testing.T) {
	machineID := uuid.New()
	store := &fakeDataStore{}
	service := newService(store, &fakeRuleResolver{})

	request := syncv1.PreflightRequest_builder{
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
	response, err := service.HandlePreflight(context.Background(), machineID, request)
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}

	if store.upsertCalls != 1 {
		t.Fatalf("upsertCalls = %d, want 1", store.upsertCalls)
	}
	if store.upsertMachine.MachineID != machineID {
		t.Fatalf("MachineID = %q, want %q", store.upsertMachine.MachineID, machineID)
	}
	if store.upsertMachine.LastSeenAt.Before(before) || store.upsertMachine.LastSeenAt.After(after) {
		t.Fatalf("LastSeenAt out of expected range: %s", store.upsertMachine.LastSeenAt)
	}

	var primaryUserGroups []string
	if unmarshalErr := json.Unmarshal(
		store.upsertMachine.PrimaryUserGroupsRaw,
		&primaryUserGroups,
	); unmarshalErr != nil {
		t.Fatalf("PrimaryUserGroupsRaw json = %v", unmarshalErr)
	}
	if len(primaryUserGroups) != 2 || primaryUserGroups[0] != "g1" || primaryUserGroups[1] != "g2" {
		t.Fatalf("PrimaryUserGroupsRaw = %#v, want [g1 g2]", primaryUserGroups)
	}
	if response.GetSyncType() != syncv1.SyncType_CLEAN {
		t.Fatalf("SyncType = %v, want CLEAN", response.GetSyncType())
	}
	state := store.ruleSyncStates[machineID]
	if state.ClientMode != domain.MachineClientModeLockdown {
		t.Fatalf("ClientMode = %q, want lockdown", state.ClientMode)
	}
}

func TestHandlePreflight_ReturnsNormalWhenManagedCountsMatch(t *testing.T) {
	machineID := uuid.New()
	acknowledgedTarget := domain.MachineRuleTarget{
		RuleType:   domain.RuleTypeBinary,
		Identifier: "com.example.existing",
		Policy:     domain.RulePolicyAllowlist,
	}
	acknowledged := storedTarget(acknowledgedTarget)
	store := &fakeDataStore{
		ruleSyncStates: map[uuid.UUID]santamodel.MachineSyncState{
			machineID: {
				MachineID:      machineID,
				AppliedTargets: []santamodel.AppliedRuleTarget{acknowledged},
			},
		},
	}
	service := newService(store, &fakeRuleResolver{
		resolvedTargets: []domain.MachineResolvedRule{
			resolvedTarget(uuid.New(), "Existing", acknowledgedTarget),
			resolvedTarget(uuid.New(), "Cert", domain.MachineRuleTarget{
				RuleType:   domain.RuleTypeCertificate,
				Identifier: "ABCD",
				Policy:     domain.RulePolicyBlocklist,
			}),
		},
	})

	response, err := service.HandlePreflight(
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
	if response.GetSyncType() != syncv1.SyncType_NORMAL {
		t.Fatalf("SyncType = %v, want NORMAL", response.GetSyncType())
	}
}

func TestHandlePreflight_ReturnsNormalWhenDesiredRulesChangedEvenIfManagedCountsDiverge(t *testing.T) {
	machineID := uuid.New()
	store := &fakeDataStore{
		ruleSyncStates: map[uuid.UUID]santamodel.MachineSyncState{
			machineID: {
				MachineID: machineID,
			},
		},
	}

	service := newService(store, &fakeRuleResolver{resolvedTargets: []domain.MachineResolvedRule{
		resolvedTarget(uuid.New(), "Binary", domain.MachineRuleTarget{
			RuleType:   domain.RuleTypeBinary,
			Identifier: "com.example.binary",
			Policy:     domain.RulePolicyAllowlist,
		}),
	}})
	response, err := service.HandlePreflight(
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
	if response.GetSyncType() != syncv1.SyncType_NORMAL {
		t.Fatalf("SyncType = %v, want NORMAL", response.GetSyncType())
	}
}

func TestHandlePreflight_ReturnsCleanWhenAppliedMatchesDesiredAndReportedCountsDiverge(t *testing.T) {
	machineID := uuid.New()
	target := domain.MachineRuleTarget{
		RuleType:   domain.RuleTypeBinary,
		Identifier: "com.example.binary",
		Policy:     domain.RulePolicyAllowlist,
	}
	applied := storedTarget(target)
	store := &fakeDataStore{
		ruleSyncStates: map[uuid.UUID]santamodel.MachineSyncState{
			machineID: {
				MachineID:      machineID,
				AppliedTargets: []santamodel.AppliedRuleTarget{applied},
			},
		},
	}

	service := newService(store, &fakeRuleResolver{resolvedTargets: []domain.MachineResolvedRule{
		resolvedTarget(uuid.New(), "Binary", target),
	}})
	response, err := service.HandlePreflight(
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
	if response.GetSyncType() != syncv1.SyncType_CLEAN {
		t.Fatalf("SyncType = %v, want CLEAN", response.GetSyncType())
	}
}

func TestHandleRuleDownload_ReturnsChangedRulesAndRemovalsDuringNormalSync(t *testing.T) {
	machineID := uuid.New()
	removed := storedTarget(domain.MachineRuleTarget{
		RuleType:   domain.RuleTypeBinary,
		Identifier: "com.example.removed",
		Policy:     domain.RulePolicyAllowlist,
	})
	store := &fakeDataStore{
		ruleSyncStates: map[uuid.UUID]santamodel.MachineSyncState{
			machineID: {
				MachineID:      machineID,
				AppliedTargets: []santamodel.AppliedRuleTarget{removed},
			},
		},
	}
	service := newService(store, &fakeRuleResolver{
		resolvedTargets: []domain.MachineResolvedRule{
			resolvedTarget(uuid.New(), "Cert", domain.MachineRuleTarget{
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

	response, err := service.HandleRuleDownload(
		context.Background(),
		machineID,
		syncv1.RuleDownloadRequest_builder{MachineId: machineID.String()}.Build(),
	)
	if err != nil {
		t.Fatalf("HandleRuleDownload() error = %v", err)
	}

	if len(response.GetRules()) != 2 {
		t.Fatalf("len(response.Rules) = %d, want 2", len(response.GetRules()))
	}
	if response.GetRules()[0].GetPolicy() != syncv1.Policy_BLOCKLIST {
		t.Fatalf("first policy = %v, want BLOCKLIST", response.GetRules()[0].GetPolicy())
	}
	if response.GetRules()[1].GetPolicy() != syncv1.Policy_REMOVE {
		t.Fatalf("second policy = %v, want REMOVE", response.GetRules()[1].GetPolicy())
	}
	if response.GetRules()[1].GetIdentifier() != "com.example.removed" {
		t.Fatalf("remove identifier = %q, want com.example.removed", response.GetRules()[1].GetIdentifier())
	}
}

func TestHandlePostflight_PromotesPendingSnapshotOnMatchingProcessedCount(t *testing.T) {
	machineID := uuid.New()
	target := domain.MachineRuleTarget{
		RuleType:   domain.RuleTypeBinary,
		Identifier: "com.example.binary",
		Policy:     domain.RulePolicyAllowlist,
	}
	store := &fakeDataStore{}
	service := newService(store, &fakeRuleResolver{resolvedTargets: []domain.MachineResolvedRule{
		resolvedTarget(uuid.New(), "Binary", target),
	}})

	if _, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{MachineId: machineID.String()}.Build(),
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

	updated := store.ruleSyncStates[machineID]
	if len(updated.AppliedTargets) != 1 {
		t.Fatalf("len(AppliedTargets) = %d, want 1", len(updated.AppliedTargets))
	}
	if updated.PendingFullSync {
		t.Fatalf("PendingFullSync = true, want false")
	}
	if updated.LastRuleSyncAttemptAt == nil {
		t.Fatal("LastRuleSyncAttemptAt = nil, want timestamp")
	}
	if updated.LastRuleSyncSuccessAt == nil {
		t.Fatal("LastRuleSyncSuccessAt = nil, want timestamp")
	}
}

func TestHandlePostflight_LeavesPendingSnapshotOnProcessedCountMismatch(t *testing.T) {
	machineID := uuid.New()
	target := domain.MachineRuleTarget{
		RuleType:   domain.RuleTypeBinary,
		Identifier: "com.example.binary",
		Policy:     domain.RulePolicyAllowlist,
	}
	store := &fakeDataStore{}
	service := newService(store, &fakeRuleResolver{resolvedTargets: []domain.MachineResolvedRule{
		resolvedTarget(uuid.New(), "Binary", target),
	}})

	if _, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{MachineId: machineID.String()}.Build(),
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

	after := store.ruleSyncStates[machineID]
	if len(after.AppliedTargets) != 0 {
		t.Fatalf("len(AppliedTargets) = %d, want 0", len(after.AppliedTargets))
	}
	if after.RulesHash != "ignored-client-rules-hash" {
		t.Fatalf("RulesHash = %q, want ignored-client-rules-hash", after.RulesHash)
	}
	if after.LastRuleSyncAttemptAt == nil {
		t.Fatal("LastRuleSyncAttemptAt = nil, want timestamp")
	}
}

func TestHandleRuleDownload_AllowsEmptyPendingSnapshot(t *testing.T) {
	machineID := uuid.New()
	store := &fakeDataStore{}
	service := newService(store, &fakeRuleResolver{})

	if _, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId: machineID.String(),
		}.Build(),
	); err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}

	response, err := service.HandleRuleDownload(
		context.Background(),
		machineID,
		syncv1.RuleDownloadRequest_builder{MachineId: machineID.String()}.Build(),
	)
	if err != nil {
		t.Fatalf("HandleRuleDownload() error = %v", err)
	}
	if len(response.GetRules()) != 0 {
		t.Fatalf("len(response.Rules) = %d, want 0", len(response.GetRules()))
	}
}

func TestHandlePostflight_PromotesEmptyPendingSnapshot(t *testing.T) {
	machineID := uuid.New()
	store := &fakeDataStore{}
	service := newService(store, &fakeRuleResolver{})

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

	updated := store.ruleSyncStates[machineID]
	if updated.PendingPreflightAt != nil {
		t.Fatal("PendingPreflightAt != nil, want cleared")
	}
	if len(updated.AppliedTargets) != 0 {
		t.Fatalf("len(AppliedTargets) = %d, want 0", len(updated.AppliedTargets))
	}
	if updated.LastRuleSyncSuccessAt == nil {
		t.Fatal("LastRuleSyncSuccessAt = nil, want timestamp")
	}
}

func storedTarget(target domain.MachineRuleTarget) santamodel.AppliedRuleTarget {
	return santamodel.AppliedRuleTarget{
		RuleType:    target.RuleType,
		Identifier:  target.Identifier,
		PayloadHash: testPayloadHash(target),
	}
}

func resolvedTarget(ruleID uuid.UUID, name string, target domain.MachineRuleTarget) domain.MachineResolvedRule {
	return domain.MachineResolvedRule{
		MachineRuleTarget: target,
		RuleID:            ruleID,
		Name:              name,
	}
}

func testPayloadHash(target domain.MachineRuleTarget) string {
	return domain.MachineRuleTargetPayloadHash(target)
}
