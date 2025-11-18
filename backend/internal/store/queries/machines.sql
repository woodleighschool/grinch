-- name: UpsertMachine :one
INSERT INTO machines (id, machine_identifier, serial, hostname, user_id, client_version)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (machine_identifier)
DO UPDATE SET
  serial = EXCLUDED.serial,
  hostname = EXCLUDED.hostname,
  user_id = EXCLUDED.user_id,
  client_version = EXCLUDED.client_version,
  updated_at = NOW()
RETURNING *;

-- name: GetMachine :one
SELECT * FROM machines
WHERE id = $1;

-- name: GetMachineByIdentifier :one
SELECT * FROM machines WHERE machine_identifier = $1;

-- name: ListMachines :many
SELECT *
FROM machines
WHERE CASE
        WHEN sqlc.arg(search)::text = '' THEN TRUE
        ELSE (
            to_tsvector(
                'simple',
                coalesce(serial, '') || ' ' ||
                coalesce(hostname, '') || ' ' ||
                coalesce(machine_identifier, '')
            ) @@ websearch_to_tsquery('simple', sqlc.arg(search)::text)
            OR serial ILIKE '%' || sqlc.arg(search)::text || '%'
            OR hostname ILIKE '%' || sqlc.arg(search)::text || '%'
            OR machine_identifier ILIKE '%' || sqlc.arg(search)::text || '%'
        )
     END
ORDER BY updated_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateMachineSyncState :one
UPDATE machines
SET last_seen = $2,
    sync_cursor = $3,
    rule_cursor = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: RequestCleanSyncAllMachines :exec
UPDATE machines SET clean_sync_requested = true;

-- name: RequestCleanSyncForUser :exec
UPDATE machines
SET clean_sync_requested = true
WHERE user_id = $1;

-- name: RequestCleanSyncForGroup :exec
UPDATE machines
SET clean_sync_requested = true
WHERE user_id IN (
    SELECT user_id FROM group_members WHERE group_id = $1
);

-- name: UpdateMachinePreflightState :one
UPDATE machines
SET primary_user = $2,
    client_mode = $3,
    clean_sync_requested = $4,
    last_preflight_at = NOW(),
    last_preflight_payload = $5,
    last_seen = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateMachinePostflightState :one
UPDATE machines
SET last_postflight_at = NOW(),
    last_rules_received = $2,
    last_rules_processed = $3,
    clean_sync_requested = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
