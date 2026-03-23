-- name: ConvertMissingEntraUsersToLocal :exec
UPDATE users
SET
  source = 'local'
WHERE source = 'entra'
  AND id <> ALL(sqlc.arg(user_ids)::UUID[]);

-- name: ConvertMissingEntraGroupsToLocal :exec
UPDATE groups
SET
  source = 'local'
WHERE source = 'entra'
  AND id <> ALL(sqlc.arg(group_ids)::UUID[]);

-- name: DeleteUserMembersForEntraGroups :exec
DELETE FROM group_user_memberships AS gum
USING groups AS g
WHERE g.id = gum.group_id
  AND g.source = 'entra';

-- name: BulkUpsertSyncedUserMemberships :exec
INSERT INTO group_user_memberships (
  group_id,
  user_id,
  origin
)
SELECT
  UNNEST(sqlc.arg(group_ids)::UUID[]),
  UNNEST(sqlc.arg(user_ids)::UUID[]),
  'synced'
ON CONFLICT (group_id, user_id) DO UPDATE
SET
  origin = 'synced';
