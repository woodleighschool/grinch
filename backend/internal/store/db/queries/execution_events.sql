-- name: CreateExecutionEvent :one
INSERT INTO execution_events (
  machine_id,
  executable_id,
  decision,
  file_path,
  executing_user,
  logged_in_users,
  current_sessions,
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
  $8
)
RETURNING
  id,
  machine_id,
  executable_id,
  decision,
  file_path,
  executing_user,
  logged_in_users,
  current_sessions,
  occurred_at,
  created_at;

-- name: GetExecutionEvent :one
SELECT
  ee.id,
  ee.machine_id,
  ee.executable_id,
  ee.decision,
  ee.file_path,
  x.file_name,
  x.file_sha256,
  x.file_bundle_id,
  x.file_bundle_path,
  x.signing_id,
  x.team_id,
  x.cdhash,
  ee.executing_user,
  ee.logged_in_users,
  ee.current_sessions,
  x.signing_chain,
  x.entitlements,
  ee.occurred_at,
  ee.created_at
FROM execution_events AS ee
JOIN executables AS x ON x.id = ee.executable_id
WHERE ee.id = $1;

-- name: DeleteExecutionEvent :exec
DELETE FROM execution_events
WHERE id = $1;

-- name: DeleteExecutionEventsBefore :execrows
DELETE FROM execution_events
WHERE created_at < $1;
