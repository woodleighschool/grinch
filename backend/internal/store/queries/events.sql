-- name: InsertEvent :one
INSERT INTO events (machine_id, user_id, kind, payload, occurred_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListEventSummaries :many
SELECT events.id, events.occurred_at, events.kind, events.payload, machines.hostname, machines.id AS machineId, users.upn, users.id AS userId
FROM events
INNER JOIN machines ON events.machine_id = machines.id
LEFT JOIN users ON events.user_id = users.id
ORDER BY occurred_at DESC
LIMIT $1 OFFSET $2;

-- name: ListBlocksByUser :many
SELECT events.id, events.occurred_at, events.kind, events.payload, machines.hostname, machines.id AS machineId, users.upn, users.id AS userId 
FROM events
INNER JOIN machines on events.machine_id = machines.id
INNER JOIN users on events.user_id = users.id
WHERE events.user_id = $1 AND events.kind LIKE 'BLOCK%'
ORDER BY occurred_at DESC
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
