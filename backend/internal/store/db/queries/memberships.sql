-- name: CreateGroupMembership :one
INSERT INTO group_memberships (
  id,
  group_id,
  member_kind,
  member_id,
  origin
)
VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
)
RETURNING
  id,
  group_id,
  member_kind,
  member_id,
  origin,
  created_at,
  updated_at;

-- name: AddSyncedGroupMembership :exec
INSERT INTO group_memberships (
  id,
  group_id,
  member_kind,
  member_id,
  origin
)
VALUES (
  $1,
  $2,
  $3,
  $4,
  'synced'
)
ON CONFLICT (group_id, member_kind, member_id) DO UPDATE SET
  origin = 'synced',
  updated_at = NOW();

-- name: GetGroupMembership :one
SELECT
  id,
  group_id,
  member_kind,
  member_id,
  origin,
  created_at,
  updated_at
FROM group_memberships
WHERE id = $1;

-- name: GetPersistedGroupMembershipView :one
SELECT
  gm.id,
  g.id AS group_id,
  g.name AS group_name,
  g.source AS group_source,
  gm.member_kind,
  gm.member_id,
  CASE
    WHEN gm.member_kind = 'user' THEN NULLIF(u.display_name, '')
    ELSE NULLIF(m.hostname, '')
  END AS member_name,
  gm.created_at,
  gm.updated_at
FROM group_memberships AS gm
JOIN groups AS g
  ON g.id = gm.group_id
LEFT JOIN users AS u
  ON gm.member_kind = 'user'
  AND u.id = gm.member_id
LEFT JOIN machines AS m
  ON gm.member_kind = 'machine'
  AND m.machine_id = gm.member_id
WHERE gm.id = $1;

-- name: DeleteGroupMembership :one
DELETE FROM group_memberships
WHERE id = $1
RETURNING
  id,
  group_id,
  member_kind,
  member_id,
  origin,
  created_at,
  updated_at;

-- name: ListEffectiveGroupIDsForMachine :many
SELECT gm.group_id
FROM group_memberships AS gm
WHERE gm.member_kind = 'machine'
  AND gm.member_id = $1
UNION
SELECT gm.group_id
FROM group_memberships AS gm
JOIN users AS u
  ON u.id = gm.member_id
  AND gm.member_kind = 'user'
JOIN machines AS m
  ON m.primary_user = u.upn
WHERE m.machine_id = $1
  AND m.primary_user <> '';
