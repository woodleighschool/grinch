-- name: UpsertMachine :one
INSERT INTO machines (
  id,
  serial_number,
  hostname,
  model_identifier,
  os_version,
  os_build,
  santa_version,
  primary_user,
  primary_user_groups,
  client_mode,
  last_seen_at
)
VALUES (
  sqlc.arg(machine_id),
  sqlc.arg(serial_number),
  sqlc.arg(hostname),
  sqlc.arg(model_identifier),
  sqlc.arg(os_version),
  sqlc.arg(os_build),
  sqlc.arg(santa_version),
  sqlc.arg(primary_user),
  sqlc.arg(primary_user_groups),
  sqlc.arg(client_mode),
  sqlc.arg(last_seen_at)
)
ON CONFLICT (id) DO UPDATE
SET
  serial_number = EXCLUDED.serial_number,
  hostname = EXCLUDED.hostname,
  model_identifier = EXCLUDED.model_identifier,
  os_version = EXCLUDED.os_version,
  os_build = EXCLUDED.os_build,
  santa_version = EXCLUDED.santa_version,
  primary_user = EXCLUDED.primary_user,
  primary_user_groups = EXCLUDED.primary_user_groups,
  client_mode = EXCLUDED.client_mode,
  last_seen_at = EXCLUDED.last_seen_at,
  updated_at = NOW()
RETURNING
  id,
  serial_number,
  hostname,
  model_identifier,
  os_version,
  os_build,
  santa_version,
  primary_user,
  primary_user_groups,
  client_mode,
  last_seen_at,
  created_at,
  updated_at;

-- name: GetMachine :one
SELECT
  m.id,
  m.serial_number,
  m.hostname,
  m.model_identifier,
  m.os_version,
  m.os_build,
  m.santa_version,
  m.primary_user,
  m.primary_user_groups,
  machine_rule_sync_status(
    ms.pending_preflight_at,
    ms.desired_targets,
    ms.applied_targets,
    ms.desired_binary_rule_count,
    ms.binary_rule_count,
    ms.desired_certificate_rule_count,
    ms.certificate_rule_count,
    ms.desired_teamid_rule_count,
    ms.teamid_rule_count,
    ms.desired_signingid_rule_count,
    ms.signingid_rule_count,
    ms.desired_cdhash_rule_count,
    ms.cdhash_rule_count,
    ms.last_clean_sync_at,
    ms.last_reported_counts_match_at
  ) AS rule_sync_status,
  m.client_mode,
  COALESCE(ms.binary_rule_count, 0)::INT4 AS binary_rule_count,
  COALESCE(ms.certificate_rule_count, 0)::INT4 AS certificate_rule_count,
  COALESCE(ms.teamid_rule_count, 0)::INT4 AS teamid_rule_count,
  COALESCE(ms.signingid_rule_count, 0)::INT4 AS signingid_rule_count,
  COALESCE(ms.cdhash_rule_count, 0)::INT4 AS cdhash_rule_count,
  m.last_seen_at,
  m.created_at,
  m.updated_at,
  u.id AS primary_user_id
FROM machines AS m
LEFT JOIN machine_sync_states AS ms
  ON ms.machine_id = m.id
LEFT JOIN users AS u
  ON u.upn = NULLIF(m.primary_user, '')
WHERE m.id = sqlc.arg(machine_id);

-- name: ListMachineIDs :many
SELECT id
FROM machines
ORDER BY id ASC;

-- name: ListMachineIDsByPrimaryUserID :many
SELECT m.id
FROM machines AS m
JOIN users AS u
  ON u.upn = NULLIF(m.primary_user, '')
WHERE u.id = sqlc.arg(primary_user_id)
ORDER BY m.id ASC;

-- name: ListMachines :many
SELECT
  m.id,
  m.serial_number,
  m.hostname,
  m.model_identifier,
  m.os_version,
  m.os_build,
  m.santa_version,
  m.primary_user,
  m.primary_user_groups,
  m.last_seen_at,
  m.created_at,
  m.updated_at,
  u.id AS primary_user_id
FROM machines AS m
LEFT JOIN users AS u
  ON u.upn = NULLIF(m.primary_user, '')
ORDER BY m.last_seen_at DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: DeleteMachine :exec
DELETE FROM machines
WHERE id = sqlc.arg(machine_id);
