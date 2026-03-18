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
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8
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
WHERE rt.rule_id = $1
ORDER BY
  CASE WHEN rt.assignment = 'include' THEN 0 ELSE 1 END ASC,
  rt.priority ASC NULLS LAST,
  rt.subject_kind ASC,
  rt.subject_id ASC;

-- name: DeleteRuleTargetsByRule :exec
DELETE FROM rule_targets
WHERE rule_id = $1;
