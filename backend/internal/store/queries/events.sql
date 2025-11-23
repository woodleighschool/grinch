-- name: UpsertFile :exec
INSERT INTO files (sha256, name, signing_id, cdhash, signing_chain, entitlements)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (sha256) DO UPDATE SET
    updated_at = NOW(),
    name = EXCLUDED.name,
    signing_id = EXCLUDED.signing_id,
    cdhash = EXCLUDED.cdhash,
    signing_chain = EXCLUDED.signing_chain,
    entitlements = EXCLUDED.entitlements;

-- name: InsertEvent :one
INSERT INTO events (machine_id, user_id, kind, payload, occurred_at, file_sha256)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListEventSummaries :many
SELECT 
    events.id, 
    events.occurred_at, 
    events.kind, 
    (events.payload || jsonb_build_object(
        'signing_chain', files.signing_chain,
        'entitlements', files.entitlements
    )) AS payload, 
    machines.hostname, 
    machines.id AS machineId, 
    users.upn, 
    users.id AS userId
FROM events
INNER JOIN machines ON events.machine_id = machines.id
LEFT JOIN users ON events.user_id = users.id
LEFT JOIN files ON events.file_sha256 = files.sha256
ORDER BY events.occurred_at DESC
LIMIT $1 OFFSET $2;

-- name: ListBlocksByUser :many
SELECT 
    events.id, 
    events.occurred_at, 
    events.kind, 
    (events.payload || jsonb_build_object(
        'signing_chain', files.signing_chain,
        'entitlements', files.entitlements
    )) AS payload, 
    machines.hostname, 
    machines.id AS machineId, 
    users.upn, 
    users.id AS userId 
FROM events
INNER JOIN machines on events.machine_id = machines.id
INNER JOIN users on events.user_id = users.id
LEFT JOIN files ON events.file_sha256 = files.sha256
WHERE events.user_id = $1 AND events.kind LIKE 'BLOCK%'
ORDER BY events.occurred_at DESC
LIMIT 50;

-- name: SummariseEvents :many
SELECT
    date_trunc('day', COALESCE(occurred_at, created_at))::timestamptz AS bucket,
    kind,
    COUNT(*)::bigint AS total
FROM events
WHERE COALESCE(occurred_at, created_at) >= NOW() - ($1::int * INTERVAL '1 day')
GROUP BY bucket, kind
ORDER BY bucket ASC, kind ASC;
