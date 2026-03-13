-- name: CreateRuleTarget :one
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
  rule_id,
  subject_kind,
  subject_id,
  assignment,
  priority,
  policy,
  cel_expression,
  created_at,
  updated_at;

-- name: GetRuleTarget :one
SELECT
  id,
  rule_id,
  subject_kind,
  subject_id,
  assignment,
  priority,
  policy,
  cel_expression,
  created_at,
  updated_at
FROM rule_targets
WHERE id = $1;

-- name: ListRuleTargetsByRule :many
SELECT
  id,
  rule_id,
  subject_kind,
  subject_id,
  assignment,
  priority,
  policy,
  cel_expression,
  created_at,
  updated_at
FROM rule_targets
WHERE rule_id = $1
ORDER BY
  CASE WHEN assignment = 'include' THEN 0 ELSE 1 END ASC,
  priority ASC NULLS LAST,
  subject_id ASC;

-- name: UpdateRuleTarget :one
UPDATE rule_targets
SET
  subject_kind = $2,
  subject_id = $3,
  assignment = $4,
  priority = $5,
  policy = $6,
  cel_expression = $7,
  updated_at = NOW()
WHERE id = $1
RETURNING
  id,
  rule_id,
  subject_kind,
  subject_id,
  assignment,
  priority,
  policy,
  cel_expression,
  created_at,
  updated_at;

-- name: DeleteRuleTarget :one
DELETE FROM rule_targets
WHERE id = $1
RETURNING
  id,
  rule_id,
  subject_kind,
  subject_id,
  assignment,
  priority,
  policy,
  cel_expression,
  created_at,
  updated_at;
