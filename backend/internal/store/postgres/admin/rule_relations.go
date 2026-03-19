package admin

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	santasnapshot "github.com/woodleighschool/grinch/internal/app/santa/snapshot"
	"github.com/woodleighschool/grinch/internal/domain"
	pgutil "github.com/woodleighschool/grinch/internal/store/postgres/shared"
)

type storedRuleTargetRow struct {
	RuleID       *uuid.UUID `json:"rule_id,omitempty"`
	RuleName     string     `json:"rule_name"`
	RuleType     string     `json:"rule_type"`
	Identifier   string     `json:"identifier"`
	IdentifierKey string    `json:"identifier_key"`
	Policy       string     `json:"policy"`
	CustomMessage string    `json:"custom_message"`
	CustomURL    string     `json:"custom_url"`
	CELExpression string    `json:"cel_expression"`
	PayloadHash  string     `json:"payload_hash"`
}

type appliedRuleTarget struct {
	domain.MachineRuleTarget

	RuleID      *uuid.UUID
	RuleName    string
	PayloadHash string
}

type machineRuleRow struct {
	domain.MachineRule

	name       string
	ruleType   domain.RuleType
	identifier string
}

func (store *Store) ListMachineRules(
	ctx context.Context,
	options domain.MachineRuleListOptions,
) ([]domain.MachineRule, int32, error) {
	if options.MachineID == nil {
		return nil, 0, nil
	}

	appliedTargets, err := store.listAppliedRuleTargets(ctx, *options.MachineID)
	if err != nil {
		return nil, 0, err
	}
	appliedByKey := make(map[string]appliedRuleTarget, len(appliedTargets))
	for _, target := range appliedTargets {
		appliedByKey[santasnapshot.RuleTargetKey(target.MachineRuleTarget)] = target
	}

	desiredRows, err := store.store.Queries().ListResolvedRulesForMachine(ctx, *options.MachineID)
	if err != nil {
		return nil, 0, err
	}

	items := make([]machineRuleRow, 0, len(desiredRows)+len(appliedTargets))
	seen := make(map[string]struct{}, len(desiredRows))
	for _, row := range desiredRows {
		ruleType, ruleTypeErr := pgutil.ToRuleType(row.RuleType)
		if ruleTypeErr != nil {
			return nil, 0, ruleTypeErr
		}
		policy, policyErr := pgutil.ToRulePolicy(row.Policy.String)
		if policyErr != nil {
			return nil, 0, policyErr
		}

		target := domain.MachineRuleTarget{
			RuleType:      ruleType,
			Identifier:    row.Identifier,
			IdentifierKey: row.IdentifierKey.String,
			Policy:        policy,
			CustomMessage: row.CustomMessage,
			CustomURL:     row.CustomUrl,
			CELExpression: row.CelExpression,
		}
		key := santasnapshot.RuleTargetKey(target)
		applied, exists := appliedByKey[key]
		items = append(items, machineRuleRow{
			MachineRule: domain.MachineRule{
				ID:        key,
				MachineID: *options.MachineID,
				RuleID:    &row.ID,
				Policy:    policy,
				Applied:   exists && applied.PayloadHash == santasnapshot.RuleTargetPayloadHash(target),
			},
			name:       row.Name,
			ruleType:   ruleType,
			identifier: row.Identifier,
		})
		seen[key] = struct{}{}
	}

	for _, target := range appliedTargets {
		key := santasnapshot.RuleTargetKey(target.MachineRuleTarget)
		if _, exists := seen[key]; exists {
			continue
		}

		items = append(items, machineRuleRow{
			MachineRule: domain.MachineRule{
				ID:        key,
				MachineID: *options.MachineID,
				RuleID:    target.RuleID,
				Policy:    target.Policy,
				Applied:   true,
			},
			name:       target.RuleName,
			ruleType:   target.RuleType,
			identifier: target.Identifier,
		})
	}

	filtered := filterMachineRules(items, options.Search)
	sortMachineRules(filtered, options.Sort, options.Order)

	total := int32(len(filtered))
	start, end := paginate(len(filtered), options.Offset, options.Limit)
	rows := make([]domain.MachineRule, 0, end-start)
	for _, item := range filtered[start:end] {
		rows = append(rows, item.MachineRule)
	}
	return rows, total, nil
}

func (store *Store) ListRuleMachines(
	ctx context.Context,
	options domain.RuleMachineListOptions,
) ([]domain.RuleMachine, int32, error) {
	if options.RuleID == nil {
		return nil, 0, nil
	}

	rows, err := store.store.Pool().Query(ctx, ruleMachineListQuery, pgutil.SearchPattern(options.Search), *options.RuleID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.RuleMachine, 0)
	for rows.Next() {
		item, scanErr := scanRuleMachine(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}

	sortRuleMachines(items, options.Sort, options.Order)
	total := int32(len(items))
	start, end := paginate(len(items), options.Offset, options.Limit)
	return items[start:end], total, nil
}

const ruleMachineListQuery = `
WITH machine_users AS (
  SELECT
    m.machine_id,
    u.id AS user_id
  FROM machines AS m
  LEFT JOIN users AS u
    ON u.upn = m.primary_user
    AND m.primary_user <> ''
),
effective_groups AS (
  SELECT
    m.machine_id,
    gm.group_id
  FROM machines AS m
  JOIN group_memberships AS gm
    ON gm.member_kind = 'machine'
    AND gm.member_id = m.machine_id
  UNION
  SELECT
    mu.machine_id,
    gm.group_id
  FROM machine_users AS mu
  JOIN group_memberships AS gm
    ON gm.member_kind = 'user'
    AND gm.member_id = mu.user_id
),
matching_targets AS (
  SELECT
    m.machine_id,
    rt.subject_kind,
    rt.subject_id,
    rt.assignment,
    rt.priority,
    rt.policy,
    rt.cel_expression
  FROM machines AS m
  LEFT JOIN machine_users AS mu
    ON mu.machine_id = m.machine_id
  JOIN rule_targets AS rt
    ON rt.rule_id = $2
  WHERE rt.subject_kind = 'all_devices'
    OR (
      rt.subject_kind = 'all_users'
      AND mu.user_id IS NOT NULL
    )
    OR (
      rt.subject_kind = 'group'
      AND EXISTS (
        SELECT 1
        FROM effective_groups AS eg
        WHERE eg.machine_id = m.machine_id
          AND eg.group_id = rt.subject_id
      )
    )
),
matching_excludes AS (
  SELECT DISTINCT mt.machine_id
  FROM matching_targets AS mt
  WHERE mt.assignment = 'exclude'
),
matching_includes AS (
  SELECT
    mt.machine_id,
    mt.policy,
    mt.cel_expression,
    ROW_NUMBER() OVER (
      PARTITION BY mt.machine_id
      ORDER BY mt.priority ASC, mt.subject_kind ASC, mt.subject_id ASC NULLS FIRST
    ) AS include_rank
  FROM matching_targets AS mt
  WHERE mt.assignment = 'include'
),
winning_includes AS (
  SELECT
    machine_id,
    policy,
    cel_expression
  FROM matching_includes
  WHERE include_rank = 1
)
SELECT
  r.id AS rule_id,
  r.rule_type,
  r.identifier,
  r.identifier_key,
  r.custom_message,
  r.custom_url,
  wi.policy,
  wi.cel_expression,
  m.machine_id,
  COALESCE(ms.applied_targets, '[]'::JSONB) AS applied_targets
FROM rules AS r
JOIN winning_includes AS wi
  ON TRUE
JOIN machines AS m
  ON m.machine_id = wi.machine_id
LEFT JOIN matching_excludes AS me
  ON me.machine_id = m.machine_id
LEFT JOIN machine_sync_states AS ms
  ON ms.machine_id = m.machine_id
WHERE r.id = $2
  AND r.enabled = TRUE
  AND me.machine_id IS NULL
  AND ($1 = '' OR
    m.hostname ILIKE $1 OR
    m.serial_number ILIKE $1)
`

func scanRuleMachine(rows pgx.Rows) (domain.RuleMachine, error) {
	var (
		item           domain.RuleMachine
		ruleType       string
		identifier     string
		identifierKey  string
		customMessage  string
		customURL      string
		policyValue    string
		celExpression  string
		appliedTargets []byte
	)

	if scanErr := rows.Scan(
		&item.RuleID,
		&ruleType,
		&identifier,
		&identifierKey,
		&customMessage,
		&customURL,
		&policyValue,
		&celExpression,
		&item.MachineID,
		&appliedTargets,
	); scanErr != nil {
		return domain.RuleMachine{}, scanErr
	}

	policy, policyErr := pgutil.ToRulePolicy(policyValue)
	if policyErr != nil {
		return domain.RuleMachine{}, policyErr
	}
	ruleTypeValue, ruleTypeErr := pgutil.ToRuleType(ruleType)
	if ruleTypeErr != nil {
		return domain.RuleMachine{}, ruleTypeErr
	}

	item.ID = item.MachineID.String()
	item.Policy = policy

	applied, appliedErr := unmarshalAppliedRuleTargets(appliedTargets)
	if appliedErr != nil {
		return domain.RuleMachine{}, appliedErr
	}
	target := domain.MachineRuleTarget{
		RuleType:      ruleTypeValue,
		Identifier:    identifier,
		IdentifierKey: identifierKey,
		Policy:        policy,
		CustomMessage: customMessage,
		CustomURL:     customURL,
		CELExpression: celExpression,
	}
	appliedByKey := make(map[string]appliedRuleTarget, len(applied))
	for _, target := range applied {
		appliedByKey[santasnapshot.RuleTargetKey(target.MachineRuleTarget)] = target
	}
	appliedTarget, exists := appliedByKey[santasnapshot.RuleTargetKey(target)]
	item.Applied = exists && appliedTarget.PayloadHash == santasnapshot.RuleTargetPayloadHash(target)

	return item, nil
}

func (store *Store) listAppliedRuleTargets(ctx context.Context, machineID uuid.UUID) ([]appliedRuleTarget, error) {
	row, err := store.store.Queries().GetMachineSyncState(ctx, machineID)
	if err != nil {
		return nil, err
	}
	return unmarshalAppliedRuleTargets(row.AppliedTargets)
}

func unmarshalAppliedRuleTargets(value []byte) ([]appliedRuleTarget, error) {
	if len(value) == 0 {
		return nil, nil
	}

	var rows []storedRuleTargetRow
	if err := json.Unmarshal(value, &rows); err != nil {
		return nil, err
	}

	targets := make([]appliedRuleTarget, 0, len(rows))
	for _, row := range rows {
		ruleType, err := domain.ParseRuleType(row.RuleType)
		if err != nil {
			return nil, err
		}
		policy, err := domain.ParseRulePolicy(row.Policy)
		if err != nil {
			return nil, err
		}

		targets = append(targets, appliedRuleTarget{
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

func filterMachineRules(items []machineRuleRow, search string) []machineRuleRow {
	search = strings.TrimSpace(strings.ToLower(search))
	if search == "" {
		return items
	}

	filtered := make([]machineRuleRow, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.name), search) ||
			strings.Contains(strings.ToLower(item.identifier), search) ||
			strings.Contains(strings.ToLower(string(item.ruleType)), search) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func sortMachineRules(items []machineRuleRow, sortField string, sortOrder string) {
	desc := strings.EqualFold(sortOrder, "desc")
	slices.SortFunc(items, func(left machineRuleRow, right machineRuleRow) int {
		var result int
		switch sortField {
		case "id":
			result = strings.Compare(left.ID, right.ID)
		case "rule_id":
			result = strings.Compare(uuidString(left.RuleID), uuidString(right.RuleID))
		case "rule_type":
			result = strings.Compare(string(left.ruleType), string(right.ruleType))
		case "identifier":
			result = strings.Compare(left.identifier, right.identifier)
		case "policy":
			result = strings.Compare(string(left.Policy), string(right.Policy))
		case "applied":
			switch {
			case left.Applied == right.Applied:
				result = strings.Compare(left.ID, right.ID)
			case left.Applied:
				result = 1
			default:
				result = -1
			}
		default:
			result = strings.Compare(left.ID, right.ID)
		}

		if desc {
			return -result
		}
		return result
	})
}

func sortRuleMachines(items []domain.RuleMachine, sortField string, sortOrder string) {
	desc := strings.EqualFold(sortOrder, "desc")
	slices.SortFunc(items, func(left domain.RuleMachine, right domain.RuleMachine) int {
		var result int
		switch sortField {
		case "machine_id", "id":
			result = strings.Compare(left.MachineID.String(), right.MachineID.String())
		case "policy":
			result = strings.Compare(string(left.Policy), string(right.Policy))
		case "applied":
			switch {
			case left.Applied == right.Applied:
				result = strings.Compare(left.MachineID.String(), right.MachineID.String())
			case left.Applied:
				result = 1
			default:
				result = -1
			}
		default:
			result = strings.Compare(left.MachineID.String(), right.MachineID.String())
		}

		if desc {
			return -result
		}
		return result
	})
}

func uuidString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func paginate(length int, offset int32, limit int32) (int, int) {
	start := int(offset)
	if start > length {
		start = length
	}
	end := length
	if limit > 0 && start+int(limit) < end {
		end = start + int(limit)
	}
	return start, end
}
