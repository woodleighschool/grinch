-- name: GetRuleByID :one
SELECT * FROM rules WHERE id = $1;

-- name: ListRulesByIDs :many
SELECT * FROM rules WHERE id = ANY($1::uuid[]);

-- name: CreateRule :one
INSERT INTO rules (name, description, identifier, rule_type, custom_msg, custom_url, notification_app_name)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateRuleByID :one
UPDATE rules SET
  name = $2,
  description = $3,
  identifier = $4,
  rule_type = $5,
  custom_msg = $6,
  custom_url = $7,
  notification_app_name = $8
WHERE id = $1
RETURNING *;

-- name: DeleteRuleByID :exec
DELETE FROM rules WHERE id = $1;
