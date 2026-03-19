package santa

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	appsanta "github.com/woodleighschool/grinch/internal/app/santa"
	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

func (store *Store) GetMachineSyncState(
	ctx context.Context,
	machineID uuid.UUID,
) (appsanta.MachineSyncState, error) {
	row, err := store.store.Queries().GetMachineSyncState(ctx, machineID)
	if err != nil {
		return appsanta.MachineSyncState{}, err
	}

	return mapMachineSyncState(row)
}

func (store *Store) ReplacePendingSnapshot(
	ctx context.Context,
	snapshot appsanta.PendingSnapshotWrite,
) error {
	appliedTargets, err := marshalSyncRuleTargets(snapshot.AppliedTargets)
	if err != nil {
		return err
	}
	pendingTargets, err := marshalSyncRuleTargets(snapshot.PendingTargets)
	if err != nil {
		return err
	}

	_, err = store.store.Queries().UpsertMachineSyncState(ctx, db.UpsertMachineSyncStateParams{
		MachineID:               snapshot.MachineID,
		RulesHash:               snapshot.RulesHash,
		AppliedTargets:          appliedTargets,
		PendingTargets:          pendingTargets,
		ExpectedRulesHash:       snapshot.ExpectedRulesHash,
		PendingPayloadRuleCount: snapshot.PendingPayloadRuleCount,
		PendingFullSync:         snapshot.PendingFullSync,
		PendingPreflightAt:      &snapshot.PendingPreflightAt,
		ClientMode:              string(snapshot.ClientMode),
		BinaryRuleCount:         snapshot.BinaryRuleCount,
		CertificateRuleCount:    snapshot.CertificateRuleCount,
		CompilerRuleCount:       snapshot.CompilerRuleCount,
		TransitiveRuleCount:     snapshot.TransitiveRuleCount,
		TeamidRuleCount:         snapshot.TeamIDRuleCount,
		SigningidRuleCount:      snapshot.SigningIDRuleCount,
		CdhashRuleCount:         snapshot.CDHashRuleCount,
		RulesReceived:           snapshot.RulesReceived,
		RulesProcessed:          snapshot.RulesProcessed,
		LastRuleSyncAttemptAt:   snapshot.LastRuleSyncAttemptAt,
		LastRuleSyncSuccessAt:   snapshot.LastRuleSyncSuccessAt,
	})
	return err
}

func (store *Store) RecordPostflight(
	ctx context.Context,
	write appsanta.PostflightWrite,
) error {
	updated, err := store.store.Queries().RecordMachineSyncPostflight(ctx, db.RecordMachineSyncPostflightParams{
		MachineID:             write.MachineID,
		RulesHash:             write.RulesHash,
		RulesReceived:         write.RulesReceived,
		RulesProcessed:        write.RulesProcessed,
		LastRuleSyncAttemptAt: &write.LastRuleSyncAttemptAt,
	})
	if err != nil {
		return err
	}
	if updated == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (store *Store) PromotePendingSnapshot(
	ctx context.Context,
	machineID uuid.UUID,
	completedAt time.Time,
) error {
	updated, err := store.store.Queries().
		PromoteMachineSyncPendingSnapshot(ctx, db.PromoteMachineSyncPendingSnapshotParams{
			MachineID:             machineID,
			LastRuleSyncSuccessAt: &completedAt,
		})
	if err != nil {
		return err
	}
	if updated == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func mapMachineSyncState(row db.GetMachineSyncStateRow) (appsanta.MachineSyncState, error) {
	appliedTargets, err := unmarshalSyncRuleTargets(row.AppliedTargets)
	if err != nil {
		return appsanta.MachineSyncState{}, err
	}
	pendingTargets, err := unmarshalSyncRuleTargets(row.PendingTargets)
	if err != nil {
		return appsanta.MachineSyncState{}, err
	}

	clientMode, err := domain.ParseMachineClientMode(row.ClientMode)
	if err != nil {
		return appsanta.MachineSyncState{}, err
	}

	return appsanta.MachineSyncState{
		MachineID:               row.MachineID,
		RulesHash:               row.RulesHash,
		AppliedTargets:          appliedTargets,
		PendingTargets:          pendingTargets,
		ExpectedRulesHash:       row.ExpectedRulesHash,
		PendingPayloadRuleCount: row.PendingPayloadRuleCount,
		PendingFullSync:         row.PendingFullSync,
		PendingPreflightAt:      row.PendingPreflightAt,
		ClientMode:              clientMode,
		BinaryRuleCount:         row.BinaryRuleCount,
		CertificateRuleCount:    row.CertificateRuleCount,
		CompilerRuleCount:       row.CompilerRuleCount,
		TransitiveRuleCount:     row.TransitiveRuleCount,
		TeamIDRuleCount:         row.TeamidRuleCount,
		SigningIDRuleCount:      row.SigningidRuleCount,
		CDHashRuleCount:         row.CdhashRuleCount,
		RulesReceived:           row.RulesReceived,
		RulesProcessed:          row.RulesProcessed,
		LastRuleSyncAttemptAt:   row.LastRuleSyncAttemptAt,
		LastRuleSyncSuccessAt:   row.LastRuleSyncSuccessAt,
	}, nil
}

func marshalSyncRuleTargets(targets []appsanta.StoredRuleTarget) ([]byte, error) {
	return json.Marshal(targets)
}

func unmarshalSyncRuleTargets(value []byte) ([]appsanta.StoredRuleTarget, error) {
	if len(value) == 0 {
		return nil, nil
	}

	var targets []appsanta.StoredRuleTarget
	if err := json.Unmarshal(value, &targets); err != nil {
		return nil, err
	}

	return targets, nil
}
