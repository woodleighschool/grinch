-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserIDByUPN :one
SELECT id FROM users WHERE upn = $1;

-- name: UpsertUser :exec
INSERT INTO users (id, upn, display_name)
VALUES ($1, $2, $3)
ON CONFLICT (upn) DO UPDATE SET display_name = EXCLUDED.display_name;
