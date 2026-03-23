package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/domain"
)

var (
	errMachineRuleListMachineIDRequired = errors.New("machine id is required")
	errRuleMachineListRuleIDRequired    = errors.New("rule id is required")

	machineRuleListSortColumns = map[string]string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"id":         "machine_rule_id",
		"rule_id":    "rule_id",
		"rule_type":  "r.rule_type",
		"identifier": "r.identifier",
		"policy":     "wi.policy",
		"applied":    "applied",
	}

	machineRuleListDefaultOrder = []string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"r.rule_type ASC",
		"r.identifier ASC",
		"r.id ASC",
	}

	ruleMachineListSortColumns = map[string]string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"id":         "m.id",
		"machine_id": "m.id",
		"policy":     "wi.policy",
		"applied":    "applied",
	}

	ruleMachineListDefaultOrder = []string{ //nolint:gochecknoglobals // package-level lookup table, not mutable state
		"m.id ASC",
	}
)

func (s *Store) ListMachineRules(
	ctx context.Context,
	opts domain.MachineRuleListOptions,
) ([]domain.MachineRule, int32, error) {
	if opts.MachineID == nil {
		return nil, 0, errMachineRuleListMachineIDRequired
	}

	orderBy, err := orderBy(
		opts.Sort,
		opts.Order,
		machineRuleListSortColumns,
		machineRuleListDefaultOrder,
	)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(machineRuleListQuery, orderBy)
	rows, err := s.Pool().Query(
		ctx,
		query,
		*opts.MachineID,
		searchPattern(opts.Search),
		opts.Limit,
		opts.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list machine rules: %w", err)
	}

	return collectRows(rows, scanMachineRuleRow)
}

func (s *Store) ListRuleMachines(
	ctx context.Context,
	opts domain.RuleMachineListOptions,
) ([]domain.RuleMachine, int32, error) {
	if opts.RuleID == nil {
		return nil, 0, errRuleMachineListRuleIDRequired
	}

	orderBy, err := orderBy(
		opts.Sort,
		opts.Order,
		ruleMachineListSortColumns,
		ruleMachineListDefaultOrder,
	)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(ruleMachineListQuery, orderBy)
	rows, err := s.Pool().Query(
		ctx,
		query,
		searchPattern(opts.Search),
		*opts.RuleID,
		opts.Limit,
		opts.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list rule machines: %w", err)
	}

	return collectRows(rows, scanRuleMachineRow)
}

func scanMachineRuleRow(rows pgx.Rows) (domain.MachineRule, int32, error) {
	var (
		item       domain.MachineRule
		policyText string
		total      int32
	)

	if err := rows.Scan(
		&item.ID,
		&item.MachineID,
		&item.RuleID,
		&policyText,
		&item.Applied,
		&total,
	); err != nil {
		return domain.MachineRule{}, 0, err
	}

	policy, err := domain.ParseRulePolicy(policyText)
	if err != nil {
		return domain.MachineRule{}, 0, fmt.Errorf("parse machine rule policy: %w", err)
	}

	item.Policy = policy

	return item, total, nil
}

func scanRuleMachineRow(rows pgx.Rows) (domain.RuleMachine, int32, error) {
	var (
		item       domain.RuleMachine
		policyText string
		total      int32
	)

	if err := rows.Scan(
		&item.RuleID,
		&item.MachineID,
		&policyText,
		&item.Applied,
		&total,
	); err != nil {
		return domain.RuleMachine{}, 0, err
	}

	policy, err := domain.ParseRulePolicy(policyText)
	if err != nil {
		return domain.RuleMachine{}, 0, fmt.Errorf("parse rule machine policy: %w", err)
	}

	item.ID = item.MachineID.String()
	item.Policy = policy

	return item, total, nil
}

const machineRuleListQuery = `
WITH effective_groups AS (
  SELECT gmm.group_id
  FROM group_machine_memberships AS gmm
  WHERE gmm.machine_id = $1

  UNION

  SELECT gum.group_id
  FROM machines AS m
  JOIN users AS u
    ON u.upn = NULLIF(m.primary_user, '')
  JOIN group_user_memberships AS gum
    ON gum.user_id = u.id
  WHERE m.id = $1
),
machine_user AS (
  SELECT u.id
  FROM machines AS m
  JOIN users AS u
    ON u.upn = NULLIF(m.primary_user, '')
  WHERE m.id = $1
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
    WHERE applied.target->>'rule_type' = r.rule_type::text
      AND applied.target->>'identifier' = r.identifier
      AND applied.target->>'payload_hash' = encode(
        digest(
          concat_ws(
            E'\x1f',
            r.rule_type::text,
            r.identifier,
            wi.policy::text,
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
  AND (
    $2 = ''
    OR r.name ILIKE $2
    OR r.identifier ILIKE $2
    OR r.rule_type::text ILIKE $2
  )
ORDER BY %s
LIMIT NULLIF($3::INT, 0)
OFFSET $4
`

const ruleMachineListQuery = `
WITH machine_users AS (
  SELECT
    m.id AS machine_id,
    u.id AS user_id
  FROM machines AS m
  LEFT JOIN users AS u
    ON u.upn = NULLIF(m.primary_user, '')
),
effective_groups AS (
  SELECT
    gmm.machine_id,
    gmm.group_id
  FROM group_machine_memberships AS gmm

  UNION

  SELECT
    mu.machine_id,
    gum.group_id
  FROM machine_users AS mu
  JOIN group_user_memberships AS gum
    ON gum.user_id = mu.user_id
),
matching_targets AS (
  SELECT
    m.id AS machine_id,
    rt.subject_kind,
    rt.subject_id,
    rt.assignment,
    rt.priority,
    rt.policy,
    rt.cel_expression
  FROM machines AS m
  LEFT JOIN machine_users AS mu
    ON mu.machine_id = m.id
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
        WHERE eg.machine_id = m.id
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
  m.id AS machine_id,
  wi.policy,
  EXISTS (
    SELECT 1
    FROM jsonb_array_elements(COALESCE(ms.applied_targets, '[]'::JSONB)) AS applied(target)
    WHERE applied.target->>'rule_type' = r.rule_type::text
      AND applied.target->>'identifier' = r.identifier
      AND applied.target->>'payload_hash' = encode(
        digest(
          concat_ws(
            E'\x1f',
            r.rule_type::text,
            r.identifier,
            wi.policy::text,
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
FROM winning_includes AS wi
JOIN machines AS m
  ON m.id = wi.machine_id
JOIN rules AS r
  ON r.id = $2
LEFT JOIN matching_excludes AS me
  ON me.machine_id = m.id
LEFT JOIN machine_sync_states AS ms
  ON ms.machine_id = m.id
WHERE r.enabled = TRUE
  AND me.machine_id IS NULL
  AND (
    $1 = ''
    OR m.hostname ILIKE $1
    OR m.serial_number ILIKE $1
  )
ORDER BY %s
LIMIT NULLIF($3::INT, 0)
OFFSET $4
`
