package santa_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/app/santa"
	"github.com/woodleighschool/grinch/internal/app/santa/snapshot"
	"github.com/woodleighschool/grinch/internal/config"
	"github.com/woodleighschool/grinch/internal/domain"
)

type fakeDataStore struct {
	deletedEvents  int64
	deleteCutoff   time.Time
	deleteErr      error
	ruleSyncStates map[uuid.UUID]santa.MachineRuleSyncState
	upsertErr      error
	upsertMachine  santa.MachineUpsert
	upsertCalls    int
}

type fakeRuleResolver struct {
	resolvedTargets []domain.MachineRuleTarget
}

func (store *fakeDataStore) UpsertMachine(_ context.Context, machine santa.MachineUpsert) error {
	store.upsertCalls++
	store.upsertMachine = machine
	if store.ruleSyncStates == nil {
		store.ruleSyncStates = make(map[uuid.UUID]santa.MachineRuleSyncState)
	}
	if _, exists := store.ruleSyncStates[machine.MachineID]; !exists {
		store.ruleSyncStates[machine.MachineID] = santa.MachineRuleSyncState{MachineID: machine.MachineID}
	}
	return store.upsertErr
}

func (store *fakeDataStore) GetMachineRuleSyncState(
	_ context.Context,
	machineID uuid.UUID,
) (santa.MachineRuleSyncState, error) {
	if store.ruleSyncStates == nil {
		store.ruleSyncStates = make(map[uuid.UUID]santa.MachineRuleSyncState)
	}
	state, exists := store.ruleSyncStates[machineID]
	if !exists {
		state = santa.MachineRuleSyncState{MachineID: machineID}
		store.ruleSyncStates[machineID] = state
	}
	return state, nil
}

func (store *fakeDataStore) ReplacePendingSnapshot(_ context.Context, pending santa.PendingSnapshotWrite) error {
	if store.ruleSyncStates == nil {
		store.ruleSyncStates = make(map[uuid.UUID]santa.MachineRuleSyncState)
	}

	store.ruleSyncStates[pending.MachineID] = santa.MachineRuleSyncState{
		MachineID:                pending.MachineID,
		RequestCleanSync:         pending.RequestCleanSync,
		LastClientRulesHash:      pending.LastClientRulesHash,
		AcknowledgedTargets:      cloneTargets(pending.AcknowledgedTargets),
		PendingTargets:           cloneTargets(pending.PendingTargets),
		PendingExpectedRulesHash: pending.PendingExpectedRulesHash,
		PendingPayloadRuleCount:  pending.PendingPayloadRuleCount,
		PendingSyncType:          pending.PendingSyncType,
		PendingPreflightAt:       &pending.PendingPreflightAt,
		LastPostflightAt:         pending.LastPostflightAt,
	}
	return nil
}

func (store *fakeDataStore) PromotePendingSnapshot(
	_ context.Context,
	machineID uuid.UUID,
	clientRulesHash string,
	completedAt time.Time,
) error {
	state := store.ruleSyncStates[machineID]
	state.AcknowledgedTargets = cloneTargets(state.PendingTargets)
	state.PendingTargets = nil
	state.PendingExpectedRulesHash = ""
	state.PendingPayloadRuleCount = 0
	state.PendingSyncType = santa.RuleSyncTypeNone
	state.PendingPreflightAt = nil
	state.RequestCleanSync = false
	state.LastClientRulesHash = strings.TrimSpace(clientRulesHash)
	state.LastPostflightAt = &completedAt
	store.ruleSyncStates[machineID] = state
	return nil
}

func (store *fakeDataStore) IngestEvents(
	context.Context,
	uuid.UUID,
	[]*syncv1.Event,
	[]*syncv1.FileAccessEvent,
	map[domain.EventDecision]struct{},
) (int, error) {
	return 0, nil
}

func (store *fakeDataStore) DeleteEventsBefore(_ context.Context, cutoff time.Time) (int64, error) {
	store.deleteCutoff = cutoff
	return store.deletedEvents, store.deleteErr
}

func (resolver *fakeRuleResolver) ResolveMachineRuleTargets(
	context.Context,
	uuid.UUID,
) ([]domain.MachineRuleTarget, error) {
	return resolver.resolvedTargets, nil
}

func newService(store *fakeDataStore, resolver *fakeRuleResolver) *santa.Service {
	return santa.New(testLogger(), store, config.EventsConfig{}, resolver)
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
	if response.GetSyncType() != syncv1.SyncType_CLEAN_RULES {
		t.Fatalf("SyncType = %v, want CLEAN_RULES", response.GetSyncType())
	}
}

func TestHandlePreflight_ReturnsNormalWhenClientMatchesAcknowledgedHash(t *testing.T) {
	machineID := uuid.New()
	acknowledged := storedTarget(domain.MachineRuleTarget{
		RuleType:      domain.RuleTypeBinary,
		Identifier:    "com.example.existing",
		IdentifierKey: "com.example.existing",
		Policy:        domain.RulePolicyAllowlist,
	})
	store := &fakeDataStore{
		ruleSyncStates: map[uuid.UUID]santa.MachineRuleSyncState{
			machineID: {
				MachineID:           machineID,
				AcknowledgedTargets: []santa.StoredRuleTarget{acknowledged},
			},
		},
	}
	service := newService(store, &fakeRuleResolver{
		resolvedTargets: []domain.MachineRuleTarget{
			acknowledged.MachineRuleTarget,
			{
				RuleType:      domain.RuleTypeCertificate,
				Identifier:    "ABCD",
				IdentifierKey: "abcd",
				Policy:        domain.RulePolicyBlocklist,
			},
		},
	})

	response, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId: machineID.String(),
			RulesHash: snapshot.SantaRulesHash([]santa.StoredRuleTarget{acknowledged}),
		}.Build(),
	)
	if err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}
	if response.GetSyncType() != syncv1.SyncType_NORMAL {
		t.Fatalf("SyncType = %v, want NORMAL", response.GetSyncType())
	}
}

func TestHandlePreflight_ReturnsCleanRulesWhenClientHashDiverges(t *testing.T) {
	machineID := uuid.New()
	acknowledged := storedTarget(domain.MachineRuleTarget{
		RuleType:      domain.RuleTypeBinary,
		Identifier:    "com.example.binary",
		IdentifierKey: "com.example.binary",
		Policy:        domain.RulePolicyAllowlist,
	})
	store := &fakeDataStore{
		ruleSyncStates: map[uuid.UUID]santa.MachineRuleSyncState{
			machineID: {
				MachineID:           machineID,
				AcknowledgedTargets: []santa.StoredRuleTarget{acknowledged},
			},
		},
	}

	service := newService(store, &fakeRuleResolver{resolvedTargets: []domain.MachineRuleTarget{}})
	response, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId: machineID.String(),
			RulesHash: "different-client-state",
		}.Build(),
	)
	if err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}
	if response.GetSyncType() != syncv1.SyncType_CLEAN_RULES {
		t.Fatalf("SyncType = %v, want CLEAN_RULES", response.GetSyncType())
	}
}

func TestHandleRuleDownload_ReturnsChangedRulesAndRemovalsDuringNormalSync(t *testing.T) {
	machineID := uuid.New()
	removed := storedTarget(domain.MachineRuleTarget{
		RuleType:      domain.RuleTypeBinary,
		Identifier:    "com.example.removed",
		IdentifierKey: "com.example.removed",
		Policy:        domain.RulePolicyAllowlist,
	})
	store := &fakeDataStore{
		ruleSyncStates: map[uuid.UUID]santa.MachineRuleSyncState{
			machineID: {
				MachineID:           machineID,
				AcknowledgedTargets: []santa.StoredRuleTarget{removed},
			},
		},
	}
	service := newService(store, &fakeRuleResolver{
		resolvedTargets: []domain.MachineRuleTarget{
			{
				RuleType:      domain.RuleTypeCertificate,
				Identifier:    "ABCD",
				IdentifierKey: "abcd",
				Policy:        domain.RulePolicyBlocklist,
			},
		},
	})

	ackHash := snapshot.SantaRulesHash([]santa.StoredRuleTarget{removed})
	if _, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{
			MachineId: machineID.String(),
			RulesHash: ackHash,
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

func TestHandlePostflight_PromotesPendingSnapshotOnMatchingHashAndProcessedCount(t *testing.T) {
	machineID := uuid.New()
	target := domain.MachineRuleTarget{
		RuleType:      domain.RuleTypeBinary,
		Identifier:    "com.example.binary",
		IdentifierKey: "com.example.binary",
		Policy:        domain.RulePolicyAllowlist,
	}
	store := &fakeDataStore{}
	service := newService(store, &fakeRuleResolver{resolvedTargets: []domain.MachineRuleTarget{target}})

	if _, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{MachineId: machineID.String()}.Build(),
	); err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}

	pendingState := store.ruleSyncStates[machineID]
	if _, err := service.HandlePostflight(
		context.Background(),
		machineID,
		syncv1.PostflightRequest_builder{
			MachineId:      machineID.String(),
			SyncType:       syncv1.SyncType_NORMAL,
			RulesHash:      pendingState.PendingExpectedRulesHash,
			RulesProcessed: 1,
		}.Build(),
	); err != nil {
		t.Fatalf("HandlePostflight() error = %v", err)
	}

	updated := store.ruleSyncStates[machineID]
	if len(updated.AcknowledgedTargets) != 1 {
		t.Fatalf("len(AcknowledgedTargets) = %d, want 1", len(updated.AcknowledgedTargets))
	}
	if updated.PendingExpectedRulesHash != "" {
		t.Fatalf("PendingExpectedRulesHash = %q, want empty", updated.PendingExpectedRulesHash)
	}
	if updated.PendingSyncType != santa.RuleSyncTypeNone {
		t.Fatalf("PendingSyncType = %q, want empty", updated.PendingSyncType)
	}
}

func TestHandlePostflight_LeavesPendingSnapshotOnHashMismatch(t *testing.T) {
	machineID := uuid.New()
	target := domain.MachineRuleTarget{
		RuleType:      domain.RuleTypeBinary,
		Identifier:    "com.example.binary",
		IdentifierKey: "com.example.binary",
		Policy:        domain.RulePolicyAllowlist,
	}
	store := &fakeDataStore{}
	service := newService(store, &fakeRuleResolver{resolvedTargets: []domain.MachineRuleTarget{target}})

	if _, err := service.HandlePreflight(
		context.Background(),
		machineID,
		syncv1.PreflightRequest_builder{MachineId: machineID.String()}.Build(),
	); err != nil {
		t.Fatalf("HandlePreflight() error = %v", err)
	}

	before := store.ruleSyncStates[machineID]
	if _, err := service.HandlePostflight(
		context.Background(),
		machineID,
		syncv1.PostflightRequest_builder{
			MachineId:      machineID.String(),
			SyncType:       syncv1.SyncType_NORMAL,
			RulesHash:      "wrong-hash",
			RulesProcessed: 1,
		}.Build(),
	); err != nil {
		t.Fatalf("HandlePostflight() error = %v", err)
	}

	after := store.ruleSyncStates[machineID]
	if len(after.AcknowledgedTargets) != 0 {
		t.Fatalf("len(AcknowledgedTargets) = %d, want 0", len(after.AcknowledgedTargets))
	}
	if after.PendingExpectedRulesHash != before.PendingExpectedRulesHash {
		t.Fatalf(
			"PendingExpectedRulesHash = %q, want %q",
			after.PendingExpectedRulesHash,
			before.PendingExpectedRulesHash,
		)
	}
}

func TestSantaRulesHash_IgnoresCustomFieldsAndUsesStableOrdering(t *testing.T) {
	base := []santa.StoredRuleTarget{
		storedTarget(domain.MachineRuleTarget{
			RuleType:      domain.RuleTypeCertificate,
			Identifier:    "BBBB",
			IdentifierKey: "bbbb",
			Policy:        domain.RulePolicyBlocklist,
			CustomMessage: "ignored-message",
			CustomURL:     "https://ignored.example",
		}),
		storedTarget(domain.MachineRuleTarget{
			RuleType:      domain.RuleTypeBinary,
			Identifier:    "AAAA",
			IdentifierKey: "aaaa",
			Policy:        domain.RulePolicyCEL,
			CELExpression: "machine.serial == 'abc'",
		}),
	}
	reordered := []santa.StoredRuleTarget{
		storedTarget(domain.MachineRuleTarget{
			RuleType:      domain.RuleTypeBinary,
			Identifier:    "AAAA",
			IdentifierKey: "aaaa",
			Policy:        domain.RulePolicyCEL,
			CELExpression: "machine.serial == 'abc'",
		}),
		storedTarget(domain.MachineRuleTarget{
			RuleType:      domain.RuleTypeCertificate,
			Identifier:    "BBBB",
			IdentifierKey: "bbbb",
			Policy:        domain.RulePolicyBlocklist,
		}),
	}

	if snapshot.SantaRulesHash(base) != snapshot.SantaRulesHash(reordered) {
		t.Fatalf("SantaRulesHash() changed for ordering/custom-field-only differences")
	}
}

func cloneTargets(targets []santa.StoredRuleTarget) []santa.StoredRuleTarget {
	if len(targets) == 0 {
		return nil
	}

	cloned := make([]santa.StoredRuleTarget, 0, len(targets))
	cloned = append(cloned, targets...)
	return cloned
}

func storedTarget(target domain.MachineRuleTarget) santa.StoredRuleTarget {
	return santa.StoredRuleTarget{
		MachineRuleTarget: target,
		PayloadHash:       testPayloadHash(target),
	}
}

func testPayloadHash(target domain.MachineRuleTarget) string {
	return strings.Join([]string{
		string(target.RuleType),
		target.IdentifierKey,
		target.Identifier,
		string(target.Policy),
		target.CustomMessage,
		target.CustomURL,
		target.CELExpression,
	}, "\x1f")
}
