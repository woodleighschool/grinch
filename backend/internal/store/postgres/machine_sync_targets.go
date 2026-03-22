package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/store/db"
)

func (store *Store) SyncAllMachineDesiredRuleTargets(ctx context.Context) error {
	machineIDs, err := store.Queries().ListMachineIDs(ctx)
	if err != nil {
		return err
	}

	return store.syncMachineDesiredRuleTargets(ctx, machineIDs)
}

func (store *Store) SyncMachineDesiredRuleTargets(ctx context.Context, machineID uuid.UUID) error {
	return store.syncMachineDesiredRuleTargets(ctx, []uuid.UUID{machineID})
}

func (store *Store) SyncMachineDesiredRuleTargetsByPrimaryUserID(
	ctx context.Context,
	userID uuid.UUID,
) error {
	machineIDs, err := store.Queries().ListMachineIDsByPrimaryUserID(ctx, userID)
	if err != nil {
		return err
	}

	return store.syncMachineDesiredRuleTargets(ctx, machineIDs)
}

func (store *Store) syncMachineDesiredRuleTargets(ctx context.Context, machineIDs []uuid.UUID) error {
	for _, machineID := range machineIDs {
		targets, err := store.Queries().ListResolvedRulesForMachine(ctx, machineID)
		if err != nil {
			return err
		}

		desiredTargets, counts, err := marshalDesiredRuleTargets(targets)
		if err != nil {
			return err
		}

		err = store.Queries().UpsertMachineDesiredTargets(ctx, db.UpsertMachineDesiredTargetsParams{
			MachineID:                   machineID,
			DesiredTargets:              desiredTargets,
			DesiredBinaryRuleCount:      counts.Binary,
			DesiredCertificateRuleCount: counts.Certificate,
			DesiredTeamIDRuleCount:      counts.TeamID,
			DesiredSigningIDRuleCount:   counts.SigningID,
			DesiredCDHashRuleCount:      counts.CDHash,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func marshalDesiredRuleTargets(rows []db.ListResolvedRulesForMachineRow) ([]byte, domain.ExecutionRuleCounts, error) {
	targets := make([]domain.AppliedRuleTarget, 0, len(rows))
	rules := make([]domain.MachineRuleTarget, 0, len(rows))

	for _, row := range rows {
		ruleType, err := domain.ParseRuleType(row.RuleType)
		if err != nil {
			return nil, domain.ExecutionRuleCounts{}, err
		}
		policy, err := domain.ParseRulePolicy(row.Policy.String)
		if err != nil {
			return nil, domain.ExecutionRuleCounts{}, err
		}

		target := domain.MachineRuleTarget{
			RuleType:      ruleType,
			Identifier:    row.Identifier,
			Policy:        policy,
			CustomMessage: row.CustomMessage,
			CustomURL:     row.CustomURL,
			CELExpression: row.CelExpression,
		}
		targets = append(targets, domain.AppliedRuleTarget{
			RuleType:    target.RuleType,
			Identifier:  target.Identifier,
			PayloadHash: domain.MachineRuleTargetPayloadHash(target),
		})
		rules = append(rules, target)
	}

	encoded, err := json.Marshal(targets)
	if err != nil {
		return nil, domain.ExecutionRuleCounts{}, err
	}

	return encoded, domain.CountExecutionRules(rules), nil
}
