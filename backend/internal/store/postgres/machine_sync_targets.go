package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/domain"
	"github.com/woodleighschool/grinch/internal/santa/model"
	"github.com/woodleighschool/grinch/internal/store/db"
)

func (s *Store) UpdateAllMachineDesiredTargets(ctx context.Context) error {
	machineIDs, err := s.Queries().ListMachineIDs(ctx)
	if err != nil {
		return fmt.Errorf("list machine ids: %w", err)
	}

	return s.updateMachineDesiredTargets(ctx, machineIDs)
}

func (s *Store) UpdateMachineDesiredTargets(ctx context.Context, machineID uuid.UUID) error {
	return s.updateMachineDesiredTargets(ctx, []uuid.UUID{machineID})
}

func (s *Store) UpdateMachineDesiredTargetsByPrimaryUserID(
	ctx context.Context,
	userID uuid.UUID,
) error {
	machineIDs, err := s.Queries().ListMachineIDsByPrimaryUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("list machine ids by primary user id: %w", err)
	}

	return s.updateMachineDesiredTargets(ctx, machineIDs)
}

func (s *Store) updateMachineDesiredTargets(ctx context.Context, machineIDs []uuid.UUID) error {
	if len(machineIDs) == 0 {
		return nil
	}

	queries := s.Queries()

	for _, machineID := range machineIDs {
		rows, err := queries.ListResolvedRulesForMachine(ctx, machineID)
		if err != nil {
			return fmt.Errorf("list resolved rules for machine %s: %w", machineID, err)
		}

		desiredTargets, counts, err := buildDesiredRuleTargets(rows)
		if err != nil {
			return fmt.Errorf("build desired targets for machine %s: %w", machineID, err)
		}

		if err = queries.UpsertMachineDesiredTargets(ctx, db.UpsertMachineDesiredTargetsParams{
			MachineID:                   machineID,
			DesiredTargets:              desiredTargets,
			DesiredBinaryRuleCount:      counts.Binary,
			DesiredCertificateRuleCount: counts.Certificate,
			DesiredTeamIDRuleCount:      counts.TeamID,
			DesiredSigningIDRuleCount:   counts.SigningID,
			DesiredCDHashRuleCount:      counts.CDHash,
		}); err != nil {
			return fmt.Errorf("upsert desired targets for machine %s: %w", machineID, err)
		}
	}

	return nil
}

func buildDesiredRuleTargets(
	rows []db.ListResolvedRulesForMachineRow,
) ([]byte, domain.ExecutionRuleCounts, error) {
	targets := make([]model.AppliedRuleTarget, 0, len(rows))
	rules := make([]domain.MachineRuleTarget, 0, len(rows))

	for _, row := range rows {
		rule, err := mapResolvedRuleTarget(row)
		if err != nil {
			return nil, domain.ExecutionRuleCounts{}, err
		}

		targets = append(targets, model.AppliedRuleTarget{
			RuleType:    rule.RuleType,
			Identifier:  rule.Identifier,
			PayloadHash: domain.MachineRuleTargetPayloadHash(rule),
		})
		rules = append(rules, rule)
	}

	encoded, err := json.Marshal(targets)
	if err != nil {
		return nil, domain.ExecutionRuleCounts{}, fmt.Errorf("marshal desired targets: %w", err)
	}

	return encoded, domain.CountExecutionRules(rules), nil
}

func mapResolvedRuleTarget(row db.ListResolvedRulesForMachineRow) (domain.MachineRuleTarget, error) {
	ruleType, err := domain.ParseRuleType(string(row.RuleType))
	if err != nil {
		return domain.MachineRuleTarget{}, fmt.Errorf("parse rule type: %w", err)
	}

	policy, err := domain.ParseRulePolicy(string(row.Policy.RulePolicy))
	if err != nil {
		return domain.MachineRuleTarget{}, fmt.Errorf("parse rule policy: %w", err)
	}

	return domain.MachineRuleTarget{
		RuleType:      ruleType,
		Identifier:    row.Identifier,
		Policy:        policy,
		CustomMessage: row.CustomMessage,
		CustomURL:     row.CustomURL,
		CELExpression: row.CelExpression,
	}, nil
}
