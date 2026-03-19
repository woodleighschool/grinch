package santa

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	appsanta "github.com/woodleighschool/grinch/internal/app/santa"
	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

type syncRuleTargetRow struct {
	RuleID        *uuid.UUID `json:"rule_id,omitempty"`
	RuleName      string     `json:"rule_name"`
	RuleType      string     `json:"rule_type"`
	Identifier    string     `json:"identifier"`
	IdentifierKey string     `json:"identifier_key"`
	Policy        string     `json:"policy"`
	CustomMessage string     `json:"custom_message"`
	CustomURL     string     `json:"custom_url"`
	CELExpression string     `json:"cel_expression"`
	PayloadHash   string     `json:"payload_hash"`
}

func (store *Store) GetMachineSyncState(
	ctx context.Context,
	machineID uuid.UUID,
) (appsanta.MachineSyncState, error) {
	row, err := store.queries.GetMachineSyncState(ctx, machineID)
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

	_, err = store.queries.UpsertMachineSyncState(ctx, db.UpsertMachineSyncStateParams{
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
	return store.store.RunInTx(ctx, func(queries *db.Queries) error {
		row, err := queries.GetMachineSyncState(ctx, write.MachineID)
		if err != nil {
			return err
		}

		return upsertMachineSyncState(ctx, queries, row, db.UpsertMachineSyncStateParams{
			MachineID:               write.MachineID,
			RulesHash:               write.RulesHash,
			AppliedTargets:          row.AppliedTargets,
			PendingTargets:          row.PendingTargets,
			ExpectedRulesHash:       row.ExpectedRulesHash,
			PendingPayloadRuleCount: row.PendingPayloadRuleCount,
			PendingFullSync:         row.PendingFullSync,
			PendingPreflightAt:      row.PendingPreflightAt,
			ClientMode:              row.ClientMode,
			BinaryRuleCount:         row.BinaryRuleCount,
			CertificateRuleCount:    row.CertificateRuleCount,
			CompilerRuleCount:       row.CompilerRuleCount,
			TransitiveRuleCount:     row.TransitiveRuleCount,
			TeamidRuleCount:         row.TeamidRuleCount,
			SigningidRuleCount:      row.SigningidRuleCount,
			CdhashRuleCount:         row.CdhashRuleCount,
			RulesReceived:           write.RulesReceived,
			RulesProcessed:          write.RulesProcessed,
			LastRuleSyncAttemptAt:   &write.LastRuleSyncAttemptAt,
			LastRuleSyncSuccessAt:   row.LastRuleSyncSuccessAt,
		})
	})
}

func (store *Store) PromotePendingSnapshot(
	ctx context.Context,
	machineID uuid.UUID,
	completedAt time.Time,
) error {
	return store.store.RunInTx(ctx, func(queries *db.Queries) error {
		row, err := queries.GetMachineSyncState(ctx, machineID)
		if err != nil {
			return err
		}

		emptyTargets, err := marshalSyncRuleTargets(nil)
		if err != nil {
			return err
		}

		return upsertMachineSyncState(ctx, queries, row, db.UpsertMachineSyncStateParams{
			MachineID:               machineID,
			RulesHash:               row.RulesHash,
			AppliedTargets:          row.PendingTargets,
			PendingTargets:          emptyTargets,
			ExpectedRulesHash:       "",
			PendingPayloadRuleCount: 0,
			PendingFullSync:         false,
			PendingPreflightAt:      nil,
			ClientMode:              row.ClientMode,
			BinaryRuleCount:         row.BinaryRuleCount,
			CertificateRuleCount:    row.CertificateRuleCount,
			CompilerRuleCount:       row.CompilerRuleCount,
			TransitiveRuleCount:     row.TransitiveRuleCount,
			TeamidRuleCount:         row.TeamidRuleCount,
			SigningidRuleCount:      row.SigningidRuleCount,
			CdhashRuleCount:         row.CdhashRuleCount,
			RulesReceived:           row.RulesReceived,
			RulesProcessed:          row.RulesProcessed,
			LastRuleSyncAttemptAt:   row.LastRuleSyncAttemptAt,
			LastRuleSyncSuccessAt:   &completedAt,
		})
	})
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

	return appsanta.MachineSyncState{
		MachineID:               row.MachineID,
		RulesHash:               row.RulesHash,
		AppliedTargets:          appliedTargets,
		PendingTargets:          pendingTargets,
		ExpectedRulesHash:       row.ExpectedRulesHash,
		PendingPayloadRuleCount: row.PendingPayloadRuleCount,
		PendingFullSync:         row.PendingFullSync,
		PendingPreflightAt:      row.PendingPreflightAt,
		ClientMode:              domain.ParseMachineClientMode(row.ClientMode),
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
	rows := make([]syncRuleTargetRow, 0, len(targets))
	for _, target := range targets {
		rows = append(rows, syncRuleTargetRow{
			RuleID:        target.RuleID,
			RuleName:      target.RuleName,
			RuleType:      string(target.RuleType),
			Identifier:    target.Identifier,
			IdentifierKey: target.IdentifierKey,
			Policy:        string(target.Policy),
			CustomMessage: target.CustomMessage,
			CustomURL:     target.CustomURL,
			CELExpression: target.CELExpression,
			PayloadHash:   target.PayloadHash,
		})
	}

	return json.Marshal(rows)
}

func unmarshalSyncRuleTargets(value []byte) ([]appsanta.StoredRuleTarget, error) {
	if len(value) == 0 {
		return nil, nil
	}

	var rows []syncRuleTargetRow
	if err := json.Unmarshal(value, &rows); err != nil {
		return nil, err
	}

	targets := make([]appsanta.StoredRuleTarget, 0, len(rows))
	for _, row := range rows {
		ruleType, err := domain.ParseRuleType(row.RuleType)
		if err != nil {
			return nil, err
		}
		policy, err := domain.ParseRulePolicy(row.Policy)
		if err != nil {
			return nil, err
		}

		targets = append(targets, appsanta.StoredRuleTarget{
			MachineRuleTarget: domain.MachineRuleTarget{
				RuleType:      ruleType,
				Identifier:    row.Identifier,
				IdentifierKey: row.IdentifierKey,
				Policy:        policy,
				CustomMessage: row.CustomMessage,
				CustomURL:     row.CustomURL,
				CELExpression: row.CELExpression,
			},
			RuleID:      row.RuleID,
			RuleName:    row.RuleName,
			PayloadHash: row.PayloadHash,
		})
	}

	return targets, nil
}

func upsertMachineSyncState(
	ctx context.Context,
	queries *db.Queries,
	row db.GetMachineSyncStateRow,
	params db.UpsertMachineSyncStateParams,
) error {
	if params.AppliedTargets == nil {
		params.AppliedTargets = row.AppliedTargets
	}
	if params.PendingTargets == nil {
		params.PendingTargets = row.PendingTargets
	}
	_, err := queries.UpsertMachineSyncState(ctx, params)
	return err
}
