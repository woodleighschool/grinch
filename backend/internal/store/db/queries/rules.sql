-- name: CreateRule :one
INSERT INTO rules (
  id,
  name,
  description,
  rule_type,
  identifier,
  custom_message,
  custom_url,
  enabled
)
VALUES (
  sqlc.arg(id),
  sqlc.arg(name),
  sqlc.arg(description),
  sqlc.arg(rule_type),
  sqlc.arg(identifier),
  sqlc.arg(custom_message),
  sqlc.arg(custom_url),
  sqlc.arg(enabled)
)
RETURNING
  id,
  name,
  description,
  rule_type,
  identifier,
  custom_message,
  custom_url,
  enabled,
  created_at,
  updated_at;

-- name: GetRule :one
SELECT
  id,
  name,
  description,
  rule_type,
  identifier,
  custom_message,
  custom_url,
  enabled,
  created_at,
  updated_at
FROM rules
WHERE id = sqlc.arg(id);

-- name: UpdateRule :one
UPDATE rules
SET
  name = sqlc.arg(name),
  description = sqlc.arg(description),
  rule_type = sqlc.arg(rule_type),
  identifier = sqlc.arg(identifier),
  custom_message = sqlc.arg(custom_message),
  custom_url = sqlc.arg(custom_url),
  enabled = sqlc.arg(enabled)
WHERE id = sqlc.arg(id)
RETURNING
  id,
  name,
  description,
  rule_type,
  identifier,
  custom_message,
  custom_url,
  enabled,
  created_at,
  updated_at;

-- name: DeleteRule :one
DELETE FROM rules
WHERE id = sqlc.arg(id)
RETURNING id;

-- name: ListResolvedRulesForMachine :many
WITH machine_user AS (
  SELECT u.id
  FROM machines AS m
  JOIN users AS u
    ON u.upn = NULLIF(m.primary_user, '')
  WHERE m.id = sqlc.arg(machine_id)
),
effective_groups AS (
  SELECT gmm.group_id
  FROM group_machine_memberships AS gmm
  WHERE gmm.machine_id = sqlc.arg(machine_id)

  UNION

  SELECT gum.group_id
  FROM machines AS m
  JOIN users AS u
    ON u.upn = NULLIF(m.primary_user, '')
  JOIN group_user_memberships AS gum
    ON gum.user_id = u.id
  WHERE m.id = sqlc.arg(machine_id)
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
    mt.subject_kind,
    mt.subject_id,
    mt.priority,
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
  r.id,
  r.name,
  r.rule_type,
  r.identifier,
  r.custom_message,
  r.custom_url,
  wi.policy,
  wi.cel_expression
FROM rules AS r
JOIN winning_includes AS wi
  ON wi.rule_id = r.id
LEFT JOIN matching_excludes AS me
  ON me.rule_id = r.id
WHERE me.rule_id IS NULL
  AND r.enabled = TRUE
ORDER BY r.rule_type ASC, r.identifier ASC, r.id ASC;

-- name: CreateRuleTarget :exec
INSERT INTO rule_targets (
  id,
  rule_id,
  subject_kind,
  subject_id,
  assignment,
  priority,
  policy,
  cel_expression
)
VALUES (
  sqlc.arg(id),
  sqlc.arg(rule_id),
  sqlc.arg(subject_kind),
  sqlc.arg(subject_id),
  sqlc.arg(assignment),
  sqlc.arg(priority),
  sqlc.arg(policy),
  sqlc.arg(cel_expression)
);

-- name: ListRuleTargetsByRule :many
SELECT
  rt.subject_kind,
  rt.subject_id,
  rt.assignment,
  rt.priority,
  rt.policy,
  rt.cel_expression,
  CASE
    WHEN rt.subject_kind = 'group' THEN COALESCE(g.name, '')
    WHEN rt.subject_kind = 'all_devices' THEN 'All Devices'
    WHEN rt.subject_kind = 'all_users' THEN 'All Users'
    ELSE ''
  END AS subject_name
FROM rule_targets AS rt
LEFT JOIN groups AS g
  ON rt.subject_kind = 'group'
  AND g.id = rt.subject_id
WHERE rt.rule_id = sqlc.arg(rule_id)
ORDER BY
  CASE WHEN rt.assignment = 'include' THEN 0 ELSE 1 END ASC,
  rt.priority ASC NULLS LAST,
  rt.subject_kind ASC,
  rt.subject_id ASC;

-- name: DeleteRuleTargetsByRule :exec
DELETE FROM rule_targets
WHERE rule_id = sqlc.arg(rule_id);
