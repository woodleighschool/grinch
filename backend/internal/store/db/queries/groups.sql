-- name: UpsertGroup :one
INSERT INTO groups (
  id,
  name,
  description,
  source
)
VALUES (
  $1,
  $2,
  $3,
  $4
)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  source = EXCLUDED.source,
  updated_at = NOW()
RETURNING
  id,
  name,
  description,
  source,
  0::INT4 AS member_count,
  created_at,
  updated_at;

-- name: UpdateGroup :one
WITH matched AS (
  SELECT source
  FROM groups
  WHERE groups.id = $1
),
updated AS (
  UPDATE groups
  SET
    name = $2,
    description = $3,
    updated_at = NOW()
  WHERE groups.id = $1
    AND groups.source = 'local'
  RETURNING
    id,
    name,
    description,
    source,
    (
      SELECT COUNT(*)::INT4
      FROM group_memberships
      WHERE group_id = groups.id
    ) AS member_count,
    created_at,
    updated_at
)
SELECT
  CASE
    WHEN EXISTS (SELECT 1 FROM updated) THEN 'updated'
    WHEN EXISTS (SELECT 1 FROM matched WHERE source <> 'local') THEN 'read_only'
    ELSE 'not_found'
  END AS status,
  updated.id,
  updated.name,
  updated.description,
  updated.source,
  COALESCE(updated.member_count, 0)::INT4 AS member_count,
  updated.created_at,
  updated.updated_at
FROM (SELECT 1) AS marker
LEFT JOIN updated ON TRUE;

-- name: GetGroup :one
SELECT
  g.id,
  g.name,
  g.description,
  g.source,
  COALESCE(member_counts.member_count, 0)::INT4 AS member_count,
  g.created_at,
  g.updated_at
FROM groups AS g
LEFT JOIN (
  SELECT group_id, COUNT(*)::INT4 AS member_count
  FROM group_memberships
  GROUP BY group_id
) AS member_counts
  ON member_counts.group_id = g.id
WHERE g.id = $1;

-- name: ListGroups :many
SELECT
  id,
  name,
  description,
  source,
  created_at,
  updated_at
FROM groups
ORDER BY name ASC, id ASC
LIMIT $1
OFFSET $2;

-- name: DeleteGroup :one
WITH matched AS (
  SELECT source
  FROM groups
  WHERE groups.id = $1
),
deleted AS (
  DELETE FROM groups
  WHERE groups.id = $1
    AND groups.source = 'local'
  RETURNING 1
)
SELECT CASE
  WHEN EXISTS (SELECT 1 FROM deleted) THEN 'deleted'
  WHEN EXISTS (SELECT 1 FROM matched WHERE source <> 'local') THEN 'read_only'
  ELSE 'not_found'
END AS status;
