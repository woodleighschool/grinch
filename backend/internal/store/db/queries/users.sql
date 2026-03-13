-- name: UpsertUser :one
INSERT INTO users (
  id,
  upn,
  display_name,
  source
)
VALUES (
  $1,
  $2,
  $3,
  $4
)
ON CONFLICT (id) DO UPDATE SET
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
WHERE id = $1;

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
LIMIT $1
OFFSET $2;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
