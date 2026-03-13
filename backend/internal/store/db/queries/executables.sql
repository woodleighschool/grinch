-- name: GetOrCreateEventExecutable :one
WITH inserted AS (
  INSERT INTO executables (
    id,
    source,
    file_sha256,
    file_name,
    file_path,
    file_bundle_id,
    file_bundle_path,
    signing_id,
    team_id,
    cdhash,
    entitlements,
    signing_chain
  )
  VALUES (
    $1,
    'event',
    $2,
    $3,
    '',
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
  )
  ON CONFLICT (file_sha256, file_name) WHERE source = 'event' DO NOTHING
  RETURNING
    id,
    source,
    file_sha256,
    file_name,
    file_path,
    file_bundle_id,
    file_bundle_path,
    signing_id,
    team_id,
    cdhash,
    entitlements,
    signing_chain,
    created_at
)
SELECT * FROM inserted
UNION ALL
SELECT
  id,
  source,
  file_sha256,
  file_name,
  file_path,
  file_bundle_id,
  file_bundle_path,
  signing_id,
  team_id,
  cdhash,
  entitlements,
  signing_chain,
  created_at
FROM executables
WHERE source = 'event'
  AND file_sha256 = $2
  AND file_name = $3
  AND NOT EXISTS (SELECT 1 FROM inserted)
LIMIT 1;

-- name: GetOrCreateProcessExecutable :one
WITH inserted AS (
  INSERT INTO executables (
    id,
    source,
    file_sha256,
    file_name,
    file_path,
    file_bundle_id,
    file_bundle_path,
    signing_id,
    team_id,
    cdhash,
    entitlements,
    signing_chain
  )
  VALUES (
    $1,
    'process',
    $2,
    '',
    $3,
    '',
    '',
    $4,
    $5,
    $6,
    '{}'::JSONB,
    $7
  )
  ON CONFLICT (file_sha256, file_path, signing_id, team_id, cdhash, signing_chain)
    WHERE source = 'process' DO NOTHING
  RETURNING
    id,
    source,
    file_sha256,
    file_name,
    file_path,
    file_bundle_id,
    file_bundle_path,
    signing_id,
    team_id,
    cdhash,
    entitlements,
    signing_chain,
    created_at
)
SELECT * FROM inserted
UNION ALL
SELECT
  id,
  source,
  file_sha256,
  file_name,
  file_path,
  file_bundle_id,
  file_bundle_path,
  signing_id,
  team_id,
  cdhash,
  entitlements,
  signing_chain,
  created_at
FROM executables
WHERE source = 'process'
  AND file_sha256 = $2
  AND file_path = $3
  AND signing_id = $4
  AND team_id = $5
  AND cdhash = $6
  AND signing_chain = $7
  AND NOT EXISTS (SELECT 1 FROM inserted)
LIMIT 1;

-- name: GetExecutable :one
SELECT
  id,
  source,
  file_sha256,
  file_name,
  file_path,
  file_bundle_id,
  file_bundle_path,
  signing_id,
  team_id,
  cdhash,
  entitlements,
  signing_chain,
  created_at
FROM executables
WHERE id = $1;

-- name: GetExecutableNamesByIds :many
SELECT
  id,
  file_name
FROM executables
WHERE id = ANY(sqlc.arg(ids)::UUID[]);
