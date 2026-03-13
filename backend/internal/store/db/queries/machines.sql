-- name: UpsertMachine :one
INSERT INTO machines (
  machine_id,
  serial_number,
  hostname,
  model_identifier,
  os_version,
  os_build,
  santa_version,
  primary_user,
  primary_user_groups_raw,
  last_seen_at
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
  $9,
  $10
)
ON CONFLICT (machine_id) DO UPDATE SET
  serial_number = EXCLUDED.serial_number,
  hostname = EXCLUDED.hostname,
  model_identifier = EXCLUDED.model_identifier,
  os_version = EXCLUDED.os_version,
  os_build = EXCLUDED.os_build,
  santa_version = EXCLUDED.santa_version,
  primary_user = EXCLUDED.primary_user,
  primary_user_groups_raw = EXCLUDED.primary_user_groups_raw,
  last_seen_at = EXCLUDED.last_seen_at,
  updated_at = NOW()
RETURNING
  machine_id,
  serial_number,
  hostname,
  model_identifier,
  os_version,
  os_build,
  santa_version,
  primary_user,
  primary_user_groups_raw,
  last_seen_at,
  created_at,
  updated_at;

-- name: GetMachine :one
SELECT
  m.machine_id,
  m.serial_number,
  m.hostname,
  m.model_identifier,
  m.os_version,
  m.os_build,
  m.santa_version,
  m.primary_user,
  m.primary_user_groups_raw,
  COALESCE(rs.request_clean_sync, FALSE) AS request_clean_sync,
  m.last_seen_at,
  m.created_at,
  m.updated_at,
  u.id AS primary_user_id
FROM machines AS m
LEFT JOIN machine_rule_sync_states AS rs
  ON rs.machine_id = m.machine_id
LEFT JOIN users AS u
  ON u.upn = m.primary_user
  AND m.primary_user <> ''
WHERE m.machine_id = $1;

-- name: ListMachines :many
SELECT
  m.machine_id,
  m.serial_number,
  m.hostname,
  m.model_identifier,
  m.os_version,
  m.os_build,
  m.santa_version,
  m.primary_user,
  m.primary_user_groups_raw,
  m.last_seen_at,
  m.created_at,
  m.updated_at,
  u.id AS primary_user_id
FROM machines AS m
LEFT JOIN users AS u
  ON u.upn = m.primary_user
  AND m.primary_user <> ''
ORDER BY m.last_seen_at DESC
LIMIT $1
OFFSET $2;

-- name: DeleteMachine :exec
DELETE FROM machines
WHERE machine_id = $1;
