-- name: UpsertGroup :one
INSERT INTO groups (id, display_name, description, object_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id)
DO UPDATE SET
  display_name = EXCLUDED.display_name,
  description = EXCLUDED.description,
  object_id = EXCLUDED.object_id,
  updated_at = NOW()
RETURNING *;

-- name: ListGroups :many
SELECT *
FROM groups
WHERE CASE
        WHEN sqlc.arg(search)::text = '' THEN TRUE
        ELSE (
            to_tsvector('simple', coalesce(display_name, '') || ' ' || coalesce(description, '')) @@ websearch_to_tsquery('simple', sqlc.arg(search)::text)
            OR display_name ILIKE '%' || sqlc.arg(search)::text || '%'
            OR description ILIKE '%' || sqlc.arg(search)::text || '%'
        )
     END
ORDER BY display_name;

-- name: DeleteGroup :exec
DELETE FROM groups WHERE id = $1;

-- name: AddGroupMember :exec
INSERT INTO group_members (group_id, user_id)
SELECT $1, $2
WHERE EXISTS (SELECT 1 FROM users WHERE id = $2)
ON CONFLICT DO NOTHING;

-- name: DeleteGroupMembers :exec
DELETE FROM group_members
WHERE group_id = $1;

-- name: ListGroupMemberIDs :many
SELECT user_id
FROM group_members
WHERE group_id = $1;

-- name: ListGroupMembers :many
SELECT u.*
FROM group_members gm
JOIN users u ON gm.user_id = u.id
WHERE gm.group_id = $1
ORDER BY LOWER(COALESCE(u.display_name, u.upn)), u.upn;

-- name: GetGroup :one
SELECT *
FROM groups
WHERE id = $1;
