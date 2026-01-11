-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = $1;

-- name: UpsertGroupByID :exec
INSERT INTO groups (id, display_name, description, member_count)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET
  display_name = EXCLUDED.display_name,
  description = EXCLUDED.description,
  member_count = EXCLUDED.member_count;
