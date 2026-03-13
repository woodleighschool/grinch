-- name: ConvertMissingEntraUsersToLocal :exec
UPDATE users
SET
  source = 'local',
  updated_at = NOW()
WHERE source = 'entra'
  AND NOT (id = ANY(sqlc.arg(user_ids)::UUID[]));

-- name: ConvertMissingEntraGroupsToLocal :exec
UPDATE groups
SET
  source = 'local',
  updated_at = NOW()
WHERE source = 'entra'
  AND NOT (id = ANY(sqlc.arg(group_ids)::UUID[]));

-- name: DeleteUserMembersForEntraGroups :exec
DELETE FROM group_memberships AS gm
USING groups AS g
WHERE gm.group_id = g.id
  AND g.source = 'entra'
  AND gm.member_kind = 'user';
