-- name: ListMembershipsByUser :many
SELECT * FROM group_memberships
WHERE user_id = $1
ORDER BY group_id ASC
LIMIT $2 OFFSET $3;

-- name: ListMembershipsByGroup :many
SELECT * FROM group_memberships
WHERE group_id = $1
ORDER BY user_id ASC
LIMIT $2 OFFSET $3;

-- name: CountMembershipsByUser :one
SELECT COUNT(*)::bigint FROM group_memberships WHERE user_id = $1;

-- name: CountMembershipsByGroup :one
SELECT COUNT(*)::bigint FROM group_memberships WHERE group_id = $1;

-- name: ListGroupIDsByUserID :many
SELECT group_id FROM group_memberships WHERE user_id = $1;

-- name: DeleteMembershipsByGroupID :exec
DELETE FROM group_memberships WHERE group_id = $1;

-- name: CreateGroupMembership :exec
INSERT INTO group_memberships (group_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
