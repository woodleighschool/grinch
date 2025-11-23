-- name: ListFiles :many
SELECT
    sha256,
    name,
    signing_id,
    cdhash,
    signing_chain,
    entitlements,
    created_at,
    updated_at,
    (SELECT COUNT(*) FROM events WHERE events.file_sha256 = files.sha256) AS event_count
FROM files
ORDER BY updated_at DESC
LIMIT $1 OFFSET $2;

-- name: GetFile :one
SELECT * FROM files WHERE sha256 = $1;
