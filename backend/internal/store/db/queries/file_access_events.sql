-- name: CreateFileAccessEvent :one
INSERT INTO file_access_events (
  id,
  machine_id,
  executable_id,
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
  $7,
  $8,
  $9
)
RETURNING
  id,
  machine_id,
  executable_id,
  rule_version,
  rule_name,
  target,
  decision,
  process_chain,
  occurred_at,
  created_at;

-- name: GetFileAccessEvent :one
SELECT
  fe.id,
  fe.machine_id,
  fe.executable_id,
  fe.rule_version,
  fe.rule_name,
  fe.target,
  fe.decision,
  COALESCE(x.file_name, '') AS file_name,
  COALESCE(x.file_sha256, '') AS file_sha256,
  COALESCE(x.signing_id, '') AS signing_id,
  COALESCE(x.team_id, '') AS team_id,
  COALESCE(x.cdhash, '') AS cdhash,
  fe.process_chain,
  fe.occurred_at,
  fe.created_at
FROM file_access_events AS fe
LEFT JOIN executables AS x ON x.id = fe.executable_id
WHERE fe.id = $1;

-- name: DeleteFileAccessEvent :exec
DELETE FROM file_access_events
WHERE id = $1;

-- name: DeleteFileAccessEventsBefore :execrows
DELETE FROM file_access_events
WHERE created_at < $1;
