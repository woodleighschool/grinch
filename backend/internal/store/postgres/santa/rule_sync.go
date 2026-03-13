package santa

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	appsanta "github.com/woodleighschool/grinch/internal/app/santa"
	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

type syncRuleTargetRow struct {
	RuleType      string `json:"rule_type"`
	Identifier    string `json:"identifier"`
	IdentifierKey string `json:"identifier_key"`
	Policy        string `json:"policy"`
	CustomMessage string `json:"custom_message"`
	CustomURL     string `json:"custom_url"`
	CELExpression string `json:"cel_expression"`
	PayloadHash   string `json:"payload_hash"`
}

func (store *Store) GetMachineRuleSyncState(
	ctx context.Context,
	machineID uuid.UUID,
) (appsanta.MachineRuleSyncState, error) {
	row, err := store.store.Queries().GetMachineRuleSyncState(ctx, machineID)
	if err != nil {
		return appsanta.MachineRuleSyncState{}, err
	}

	return mapMachineRuleSyncState(row)
}

func (store *Store) ReplacePendingSnapshot(
	ctx context.Context,
	snapshot appsanta.PendingSnapshotWrite,
) error {
	acknowledgedTargets, err := marshalSyncRuleTargets(snapshot.AcknowledgedTargets)
	if err != nil {
		return err
	}
	pendingTargets, err := marshalSyncRuleTargets(snapshot.PendingTargets)
	if err != nil {
		return err
	}

	_, err = store.store.Queries().UpsertMachineRuleSyncState(ctx, db.UpsertMachineRuleSyncStateParams{
		MachineID:                snapshot.MachineID,
		RequestCleanSync:         snapshot.RequestCleanSync,
		LastClientRulesHash:      snapshot.LastClientRulesHash,
		AcknowledgedTargets:      acknowledgedTargets,
		PendingTargets:           pendingTargets,
		PendingExpectedRulesHash: snapshot.PendingExpectedRulesHash,
		PendingPayloadRuleCount:  snapshot.PendingPayloadRuleCount,
		PendingSyncType:          string(snapshot.PendingSyncType),
		PendingPreflightAt:       &snapshot.PendingPreflightAt,
		LastPostflightAt:         snapshot.LastPostflightAt,
	})
	return err
}

func (store *Store) PromotePendingSnapshot(
	ctx context.Context,
	machineID uuid.UUID,
	clientRulesHash string,
	completedAt time.Time,
) error {
	return store.store.RunInTx(ctx, func(queries *db.Queries) error {
		row, err := queries.GetMachineRuleSyncState(ctx, machineID)
		if err != nil {
			return err
		}

		pendingTargets, err := unmarshalSyncRuleTargets(row.PendingTargets)
		if err != nil {
			return err
		}

		acknowledgedTargets, err := marshalSyncRuleTargets(pendingTargets)
		if err != nil {
			return err
		}
		emptyTargets, err := marshalSyncRuleTargets(nil)
		if err != nil {
			return err
		}

		_, err = queries.UpsertMachineRuleSyncState(ctx, db.UpsertMachineRuleSyncStateParams{
			MachineID:                machineID,
			RequestCleanSync:         false,
			LastClientRulesHash:      strings.TrimSpace(clientRulesHash),
			AcknowledgedTargets:      acknowledgedTargets,
			PendingTargets:           emptyTargets,
			PendingExpectedRulesHash: "",
			PendingPayloadRuleCount:  0,
			PendingSyncType:          string(appsanta.RuleSyncTypeNone),
			PendingPreflightAt:       nil,
			LastPostflightAt:         &completedAt,
		})
		return err
	})
}

func mapMachineRuleSyncState(row db.GetMachineRuleSyncStateRow) (appsanta.MachineRuleSyncState, error) {
	acknowledgedTargets, err := unmarshalSyncRuleTargets(row.AcknowledgedTargets)
	if err != nil {
		return appsanta.MachineRuleSyncState{}, err
	}
	pendingTargets, err := unmarshalSyncRuleTargets(row.PendingTargets)
	if err != nil {
		return appsanta.MachineRuleSyncState{}, err
	}

	return appsanta.MachineRuleSyncState{
		MachineID:                row.MachineID,
		RequestCleanSync:         row.RequestCleanSync,
		LastClientRulesHash:      row.LastClientRulesHash,
		AcknowledgedTargets:      acknowledgedTargets,
		PendingTargets:           pendingTargets,
		PendingExpectedRulesHash: row.PendingExpectedRulesHash,
		PendingPayloadRuleCount:  row.PendingPayloadRuleCount,
		PendingSyncType:          appsanta.RuleSyncType(row.PendingSyncType),
		PendingPreflightAt:       row.PendingPreflightAt,
		LastPostflightAt:         row.LastPostflightAt,
	}, nil
}

func marshalSyncRuleTargets(targets []appsanta.StoredRuleTarget) ([]byte, error) {
	rows := make([]syncRuleTargetRow, 0, len(targets))
	for _, target := range targets {
		rows = append(rows, syncRuleTargetRow{
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
			PayloadHash: row.PayloadHash,
		})
	}

	return targets, nil
}
