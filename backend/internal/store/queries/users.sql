-- name: UpsertUser :one
INSERT INTO users (id, upn, display_name, object_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id)
DO UPDATE SET
  upn = EXCLUDED.upn,
  display_name = EXCLUDED.display_name,
  object_id = EXCLUDED.object_id,
  updated_at = NOW()
RETURNING *;

-- name: ListUsers :many
SELECT *
FROM users
WHERE CASE
        WHEN sqlc.arg(search)::text = '' THEN TRUE
        ELSE (
            to_tsvector('simple', coalesce(display_name, '') || ' ' || coalesce(upn, '')) @@ websearch_to_tsquery('simple', sqlc.arg(search)::text)
            OR display_name ILIKE '%' || sqlc.arg(search)::text || '%'
            OR upn ILIKE '%' || sqlc.arg(search)::text || '%'
        )
     END
ORDER BY display_name;

-- name: GetUser :one
SELECT id, upn, display_name, object_id, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUPN :one
SELECT id, upn, display_name, object_id, created_at, updated_at
FROM users
WHERE LOWER(upn) = LOWER($1);

-- name: GetUserByLogin :one
SELECT id, upn, display_name, object_id, created_at, updated_at
FROM users
WHERE LOWER(split_part(upn, '@', 1)) = LOWER($1);

-- name: GetUserGroups :many
SELECT g.id,
       g.display_name,
       g.description,
       g.object_id,
       g.created_at,
       g.updated_at
FROM groups g
JOIN group_members gm ON g.id = gm.group_id
WHERE gm.user_id = $1
ORDER BY g.display_name;

-- name: GetUserMachines :many
SELECT *
FROM machines
WHERE user_id = $1
ORDER BY hostname;

-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1;

-- name: DeleteUserByUPN :exec
DELETE
FROM users
WHERE LOWER(upn) = LOWER($1);
