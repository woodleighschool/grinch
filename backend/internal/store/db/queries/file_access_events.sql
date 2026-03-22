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
  $1,
  $2,
  $3,
  $4,
  $5,
  $6,
  $7
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
WHERE id = $1;

-- name: DeleteFileAccessEvent :exec
DELETE FROM file_access_events
WHERE id = $1;

-- name: DeleteFileAccessEventsBefore :execrows
DELETE FROM file_access_events
WHERE created_at < $1;
