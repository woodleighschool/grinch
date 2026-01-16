-- name: GetEventByID :one
SELECT * FROM events WHERE id = $1;

-- name: CreateEvent :one
INSERT INTO events (
  machine_id, decision,
  file_path, file_sha256, file_name,
  executing_user, execution_time,
  logged_in_users, current_sessions,
  file_bundle_id, file_bundle_path, file_bundle_executable_rel_path,
  file_bundle_name, file_bundle_version, file_bundle_version_string,
  file_bundle_hash, file_bundle_hash_millis, file_bundle_binary_count,
  pid, ppid, parent_name,
  team_id, signing_id, cdhash, cs_flags, signing_status,
  secure_signing_time, signing_time
) VALUES (
  $1, $2,
  $3, $4, $5,
  $6, $7,
  $8, $9,
  $10, $11, $12,
  $13, $14, $15,
  $16, $17, $18,
  $19, $20, $21,
  $22, $23, $24, $25, $26,
  $27, $28
)
RETURNING *;

-- name: ListSigningChainEntriesByEventID :many
SELECT
  esc.event_id,
  esc.ordinal,
  c.sha256,
  c.cn,
  c.org,
  c.ou,
  c.valid_from,
  c.valid_until
FROM event_signing_chain esc
JOIN certificates c ON esc.certificate_sha256 = c.sha256
WHERE esc.event_id = $1
ORDER BY esc.ordinal;

-- name: CreateSigningChainEntry :exec
INSERT INTO event_signing_chain (event_id, ordinal, certificate_sha256)
VALUES ($1, $2, $3);

-- name: UpsertCertificate :exec
INSERT INTO certificates (sha256, cn, org, ou, valid_from, valid_until)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (sha256) DO UPDATE SET
  cn = EXCLUDED.cn,
  org = EXCLUDED.org,
  ou = EXCLUDED.ou,
  valid_from = EXCLUDED.valid_from,
  valid_until = EXCLUDED.valid_until;

-- name: PruneEventsBefore :one
WITH deleted AS (
  DELETE FROM events
  WHERE execution_time IS NOT NULL AND execution_time < $1
  RETURNING 1
)
SELECT count(*) FROM deleted;

-- name: ListEntitlementsByEventID :many
SELECT
  ee.event_id,
  ee.ordinal,
  e.id,
  e.key,
  e.value
FROM event_entitlements ee
JOIN entitlements e ON ee.entitlement_id = e.id
WHERE ee.event_id = $1
ORDER BY ee.ordinal;

-- name: CreateEventEntitlement :exec
INSERT INTO event_entitlements (event_id, ordinal, entitlement_id)
VALUES ($1, $2, $3);

-- name: UpsertEntitlement :one
INSERT INTO entitlements (key, value)
VALUES ($1, $2)
ON CONFLICT (key, value) DO UPDATE SET
  key = EXCLUDED.key
RETURNING id;
