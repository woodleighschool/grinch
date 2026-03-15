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
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8
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
WHERE id = $1;

-- name: ListRules :many
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
ORDER BY created_at DESC, id DESC
LIMIT $1
OFFSET $2;

-- name: UpdateRule :one
UPDATE rules
SET
  name = $2,
  description = $3,
  rule_type = $4,
  identifier = $5,
  custom_message = $6,
  custom_url = $7,
  enabled = $8,
  updated_at = NOW()
WHERE id = $1
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
WHERE id = $1
RETURNING id;

-- name: CountGroupsByIDs :one
SELECT COUNT(*)::INT4
FROM groups
WHERE id = ANY(sqlc.arg(group_ids)::UUID[]);

-- name: CountRulesByIDs :one
SELECT COUNT(*)::INT4
FROM rules
WHERE id = ANY(sqlc.arg(rule_ids)::UUID[]);

-- name: ListResolvedRulesForMachine :many
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
matching_excludes AS (
  SELECT DISTINCT rt.rule_id
  FROM rule_targets AS rt
  JOIN effective_groups AS eg
    ON eg.group_id = rt.subject_id
  WHERE rt.assignment = 'exclude'
    AND rt.subject_kind = 'group'
),
matching_includes AS (
  SELECT
    rt.rule_id,
    rt.subject_id,
    rt.priority,
    rt.policy,
    rt.cel_expression,
    ROW_NUMBER() OVER (
      PARTITION BY rt.rule_id
      ORDER BY rt.priority ASC, rt.subject_id ASC
    ) AS include_rank
  FROM rule_targets AS rt
  JOIN effective_groups AS eg
    ON eg.group_id = rt.subject_id
  WHERE rt.assignment = 'include'
    AND rt.subject_kind = 'group'
),
winning_includes AS (
  SELECT
    rule_id,
    subject_id,
    priority,
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
  r.identifier_key,
  r.custom_message,
  r.custom_url,
  wi.subject_id AS matched_group_id,
  wi.priority AS matched_priority,
  wi.policy,
  wi.cel_expression
FROM rules AS r
JOIN winning_includes AS wi
  ON wi.rule_id = r.id
LEFT JOIN matching_excludes AS me
  ON me.rule_id = r.id
WHERE me.rule_id IS NULL
  AND r.enabled = true
ORDER BY r.rule_type ASC, r.identifier_key ASC, r.id ASC;
