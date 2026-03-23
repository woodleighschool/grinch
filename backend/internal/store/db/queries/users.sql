-- name: UpsertUser :one
INSERT INTO users (
  id,
  upn,
  display_name,
  source
)
VALUES (
  sqlc.arg(id),
  sqlc.arg(upn),
  sqlc.arg(display_name),
  sqlc.arg(source)
)
ON CONFLICT (id) DO UPDATE
SET
  upn = EXCLUDED.upn,
  display_name = EXCLUDED.display_name,
  source = EXCLUDED.source,
  updated_at = NOW()
RETURNING
  id,
  upn,
  display_name,
  source,
  created_at,
  updated_at;

-- name: GetUser :one
SELECT
  id,
  upn,
  display_name,
  source,
  created_at,
  updated_at
FROM users
WHERE id = sqlc.arg(id);

-- name: ListUsers :many
SELECT
  id,
  upn,
  display_name,
  source,
  created_at,
  updated_at
FROM users
ORDER BY display_name ASC, id ASC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = sqlc.arg(id);
