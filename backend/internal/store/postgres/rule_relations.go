package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
)

func (store *Store) ListMachineRules(
	ctx context.Context,
	options domain.MachineRuleListOptions,
) ([]domain.MachineRule, int32, error) {
	orderBy, err := orderBy(options.Sort, options.Order, map[string]string{
		"id":         "machine_rule_id",
		"rule_id":    "rule_id",
		"rule_type":  "r.rule_type",
		"identifier": "r.identifier",
		"policy":     "wi.policy",
		"applied":    "applied",
	}, []string{"r.rule_type ASC", "r.identifier ASC", "r.id ASC"})
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(machineRuleListQuery, orderBy)
	rows, err := store.Pool().Query(
		ctx,
		query,
		*options.MachineID,
		searchPattern(options.Search),
		options.Limit,
		options.Offset,
	)
	if err != nil {
		return nil, 0, err
	}

	return collectRows(rows, scanMachineRule)
}

const machineRuleListQuery = `
WITH effective_groups AS (
  SELECT gm.group_id
  FROM group_memberships AS gm
  WHERE gm.member_kind = 'machine'
    AND gm.member_id = $1
  UNION
  SELECT gm.group_id
  FROM group_memberships AS gm
  JOIN users AS u
    ON u.id = gm.member_id
    AND gm.member_kind = 'user'
  JOIN machines AS m
    ON m.primary_user = u.upn
  WHERE m.machine_id = $1
    AND m.primary_user <> ''
),
machine_user AS (
  SELECT u.id
  FROM machines AS m
  JOIN users AS u
    ON u.upn = m.primary_user
  WHERE m.machine_id = $1
    AND m.primary_user <> ''
),
matching_targets AS (
  SELECT
    rt.rule_id,
    rt.subject_kind,
    rt.subject_id,
    rt.assignment,
    rt.priority,
    rt.policy,
    rt.cel_expression
  FROM rule_targets AS rt
  WHERE rt.subject_kind = 'all_devices'
    OR (
      rt.subject_kind = 'all_users'
      AND EXISTS (SELECT 1 FROM machine_user)
    )
    OR (
      rt.subject_kind = 'group'
      AND EXISTS (
        SELECT 1
        FROM effective_groups AS eg
        WHERE eg.group_id = rt.subject_id
      )
    )
),
matching_excludes AS (
  SELECT DISTINCT mt.rule_id
  FROM matching_targets AS mt
  WHERE mt.assignment = 'exclude'
),
matching_includes AS (
  SELECT
    mt.rule_id,
    mt.policy,
    mt.cel_expression,
    ROW_NUMBER() OVER (
      PARTITION BY mt.rule_id
      ORDER BY mt.priority ASC, mt.subject_kind ASC, mt.subject_id ASC NULLS FIRST
    ) AS include_rank
  FROM matching_targets AS mt
  WHERE mt.assignment = 'include'
),
winning_includes AS (
  SELECT
    rule_id,
    policy,
    cel_expression
  FROM matching_includes
  WHERE include_rank = 1
)
SELECT
  r.rule_type || '|' || r.identifier AS machine_rule_id,
  $1::uuid AS machine_id,
  r.id AS rule_id,
  wi.policy,
  EXISTS (
    SELECT 1
    FROM jsonb_array_elements(COALESCE(ms.applied_targets, '[]'::JSONB)) AS applied(target)
    WHERE applied.target->>'rule_type' = r.rule_type
      AND applied.target->>'identifier' = r.identifier
      AND applied.target->>'payload_hash' = encode(
        digest(
          concat_ws(
            E'\x1f',
            r.rule_type,
            r.identifier,
            wi.policy,
            r.custom_message,
            r.custom_url,
            wi.cel_expression
          ),
          'sha256'
        ),
        'hex'
      )
  ) AS applied,
  COUNT(*) OVER()::INT4 AS total
FROM rules AS r
JOIN winning_includes AS wi
  ON wi.rule_id = r.id
LEFT JOIN matching_excludes AS me
  ON me.rule_id = r.id
LEFT JOIN machine_sync_states AS ms
  ON ms.machine_id = $1
WHERE me.rule_id IS NULL
  AND r.enabled = TRUE
  AND ($2 = '' OR
    r.name ILIKE $2 OR
    r.identifier ILIKE $2 OR
    r.rule_type ILIKE $2)
ORDER BY %s
LIMIT NULLIF($3::INT, 0)
OFFSET $4
`

func scanMachineRule(rows pgx.Rows) (domain.MachineRule, int32, error) {
	var (
		item       domain.MachineRule
		policyText string
		total      int32
	)

	if scanErr := rows.Scan(
		&item.ID,
		&item.MachineID,
		&item.RuleID,
		&policyText,
		&item.Applied,
		&total,
	); scanErr != nil {
		return domain.MachineRule{}, 0, scanErr
	}

	policy, err := domain.ParseRulePolicy(policyText)
	if err != nil {
		return domain.MachineRule{}, 0, err
	}
	item.Policy = policy

	return item, total, nil
}

func (store *Store) ListRuleMachines(
	ctx context.Context,
	options domain.RuleMachineListOptions,
) ([]domain.RuleMachine, int32, error) {
	orderBy, err := orderBy(options.Sort, options.Order, map[string]string{
		"id":         "m.machine_id",
		"machine_id": "m.machine_id",
		"policy":     "wi.policy",
		"applied":    "applied",
	}, []string{"m.machine_id ASC"})
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(ruleMachineListQuery, orderBy)
	rows, err := store.Pool().Query(
		ctx,
		query,
		searchPattern(options.Search),
		*options.RuleID,
		options.Limit,
		options.Offset,
	)
	if err != nil {
		return nil, 0, err
	}

	return collectRows(rows, scanRuleMachine)
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
  m.machine_id,
  wi.policy,
  EXISTS (
    SELECT 1
    FROM jsonb_array_elements(COALESCE(ms.applied_targets, '[]'::JSONB)) AS applied(target)
    WHERE applied.target->>'rule_type' = r.rule_type
      AND applied.target->>'identifier' = r.identifier
      AND applied.target->>'payload_hash' = encode(
        digest(
          concat_ws(
            E'\x1f',
            r.rule_type,
            r.identifier,
            wi.policy,
            r.custom_message,
            r.custom_url,
            wi.cel_expression
          ),
          'sha256'
        ),
        'hex'
      )
  ) AS applied,
  COUNT(*) OVER()::INT4 AS total
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
ORDER BY %s
LIMIT NULLIF($3::INT, 0)
OFFSET $4
`

func scanRuleMachine(rows pgx.Rows) (domain.RuleMachine, int32, error) {
	var (
		item       domain.RuleMachine
		policyText string
		total      int32
	)

	if scanErr := rows.Scan(
		&item.RuleID,
		&item.MachineID,
		&policyText,
		&item.Applied,
		&total,
	); scanErr != nil {
		return domain.RuleMachine{}, 0, scanErr
	}

	policy, err := domain.ParseRulePolicy(policyText)
	if err != nil {
		return domain.RuleMachine{}, 0, err
	}
	item.ID = item.MachineID.String()
	item.Policy = policy

	return item, total, nil
}
