-- name: CreateUserMembership :one
INSERT INTO group_user_memberships (
  id,
  group_id,
  user_id,
  origin
)
VALUES (
  sqlc.arg(id),
  sqlc.arg(group_id),
  sqlc.arg(user_id),
  sqlc.arg(origin)
)
RETURNING
  id,
  group_id,
  user_id,
  origin,
  created_at,
  updated_at;

-- name: CreateMachineMembership :one
INSERT INTO group_machine_memberships (
  id,
  group_id,
  machine_id,
  origin
)
VALUES (
  sqlc.arg(id),
  sqlc.arg(group_id),
  sqlc.arg(machine_id),
  sqlc.arg(origin)
)
RETURNING
  id,
  group_id,
  machine_id,
  origin,
  created_at,
  updated_at;

-- name: AddSyncedUserMembership :exec
INSERT INTO group_user_memberships (
  id,
  group_id,
  user_id,
  origin
)
VALUES (
  sqlc.arg(id),
  sqlc.arg(group_id),
  sqlc.arg(user_id),
  'synced'
)
ON CONFLICT (group_id, user_id) DO UPDATE
SET
  origin = 'synced';

-- name: GetPersistedMembershipView :one
SELECT
  gm.id,
  g.id AS group_id,
  g.name AS group_name,
  g.source AS group_source,
  gm.member_kind,
  gm.member_id,
  COALESCE(
    CASE
      WHEN gm.member_kind = 'user' THEN NULLIF(u.display_name, '')
      ELSE NULLIF(m.hostname, '')
    END,
    ''
  )::TEXT AS member_name,
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
  AND m.id = gm.member_id
WHERE gm.id = sqlc.arg(id);

-- name: DeleteUserMembership :execrows
DELETE FROM group_user_memberships
WHERE id = sqlc.arg(id);

-- name: DeleteMachineMembership :execrows
DELETE FROM group_machine_memberships
WHERE id = sqlc.arg(id);

-- name: ListEffectiveGroupIDsForMachine :many
SELECT gmm.group_id
FROM group_machine_memberships AS gmm
WHERE gmm.machine_id = sqlc.arg(machine_id)

UNION

SELECT gum.group_id
FROM machines AS m
JOIN users AS u
  ON u.upn = NULLIF(m.primary_user, '')
JOIN group_user_memberships AS gum
  ON gum.user_id = u.id
WHERE m.id = sqlc.arg(machine_id)

ORDER BY group_id ASC;
