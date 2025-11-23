-- name: ListRuleScopes :many
SELECT *
FROM rule_scopes
WHERE rule_id = $1
ORDER BY created_at ASC;

-- name: ListAllRuleScopes :many
SELECT *
FROM rule_scopes;

-- name: GetRuleScope :one
SELECT *
FROM rule_scopes
WHERE id = $1;

-- name: GetRuleScopeByTarget :one
SELECT *
FROM rule_scopes
WHERE rule_id = $1
  AND target_type = $2
  AND target_id = $3
LIMIT 1;

-- name: InsertRuleScope :one
INSERT INTO rule_scopes (id, rule_id, target_type, target_id, action)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteRuleScope :exec
DELETE FROM rule_scopes
WHERE id = $1;

-- name: ListRulesByGroupTarget :many
SELECT DISTINCT rule_id
FROM rule_scopes
WHERE target_type = 'group' AND target_id = $1;
