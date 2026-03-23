-- name: CreateFileAccessEvent :one
INSERT INTO file_access_events (
  machine_id,
  rule_version,
  rule_name,
  target,
  decision,
  process_chain,
  occurred_at
)
VALUES (
  sqlc.arg(machine_id),
  sqlc.arg(rule_version),
  sqlc.arg(rule_name),
  sqlc.arg(target),
  sqlc.arg(decision),
  sqlc.arg(process_chain),
  sqlc.arg(occurred_at)
)
RETURNING
  id,
  machine_id,
  rule_version,
  rule_name,
  target,
  decision,
  process_chain,
  occurred_at,
  created_at;

-- name: GetFileAccessEvent :one
SELECT
  id,
  machine_id,
  rule_version,
  rule_name,
  target,
  decision,
  process_chain,
  occurred_at,
  created_at
FROM file_access_events
WHERE id = sqlc.arg(id);

-- name: DeleteFileAccessEvent :exec
DELETE FROM file_access_events
WHERE id = sqlc.arg(id);

-- name: DeleteFileAccessEventsBefore :execrows
DELETE FROM file_access_events
WHERE created_at < sqlc.arg(before_created_at);
