-- name: InsertEvent :one
INSERT INTO events (machine_id, user_id, kind, payload, occurred_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListEvents :many
SELECT *
FROM events
ORDER BY occurred_at DESC
LIMIT $1 OFFSET $2;

-- name: SummarizeEvents :many
SELECT
    date_trunc('day', COALESCE(occurred_at, created_at))::timestamptz AS bucket,
    kind,
    COUNT(*)::bigint AS total
FROM events
WHERE COALESCE(occurred_at, created_at) >= NOW() - ($1::int * INTERVAL '1 day')
GROUP BY bucket, kind
ORDER BY bucket ASC, kind ASC;
