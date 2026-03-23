-- name: UpsertGroup :one
INSERT INTO groups (
  id,
  name,
  description,
  source
)
VALUES (
  sqlc.arg(id),
  sqlc.arg(name),
  sqlc.arg(description),
  sqlc.arg(source)
)
ON CONFLICT (id) DO UPDATE
SET
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
  SELECT g.source
  FROM groups AS g
  WHERE g.id = sqlc.arg(id)
),
updated AS (
  UPDATE groups AS g
  SET
    name = sqlc.arg(name),
    description = sqlc.arg(description),
    updated_at = NOW()
  WHERE g.id = sqlc.arg(id)
    AND g.source = 'local'
  RETURNING
    g.id,
    g.name,
    g.description,
    g.source,
    (
      SELECT COUNT(*)::INT4
      FROM group_memberships AS gm
      WHERE gm.group_id = g.id
    ) AS member_count,
    g.created_at,
    g.updated_at
)
SELECT
  CASE
    WHEN EXISTS (SELECT 1 FROM updated) THEN 'updated'
    WHEN EXISTS (SELECT 1 FROM matched WHERE source <> 'local') THEN 'read_only'
    ELSE 'not_found'
  END AS status,
  u.id,
  u.name,
  u.description,
  u.source,
  COALESCE(u.member_count, 0)::INT4 AS member_count,
  u.created_at,
  u.updated_at
FROM (VALUES (1)) AS marker(dummy)
LEFT JOIN updated AS u
  ON TRUE;

-- name: GetGroup :one
SELECT
  g.id,
  g.name,
  g.description,
  g.source,
  (
    SELECT COUNT(*)::INT4
    FROM group_memberships AS gm
    WHERE gm.group_id = g.id
  ) AS member_count,
  g.created_at,
  g.updated_at
FROM groups AS g
WHERE g.id = sqlc.arg(id);

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
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: DeleteGroup :one
WITH matched AS (
  SELECT g.source
  FROM groups AS g
  WHERE g.id = sqlc.arg(id)
),
deleted AS (
  DELETE FROM groups AS g
  WHERE g.id = sqlc.arg(id)
    AND g.source = 'local'
  RETURNING 1
)
SELECT
  CASE
    WHEN EXISTS (SELECT 1 FROM deleted) THEN 'deleted'
    WHEN EXISTS (SELECT 1 FROM matched WHERE source <> 'local') THEN 'read_only'
    ELSE 'not_found'
  END AS status;
