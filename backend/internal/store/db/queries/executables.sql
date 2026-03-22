-- name: GetOrCreateExecutable :one
WITH inserted AS (
  INSERT INTO executables (
    file_sha256,
    file_name,
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
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
  )
  ON CONFLICT (file_sha256, file_name) DO NOTHING
  RETURNING
    id,
    file_sha256,
    file_name,
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
  file_sha256,
  file_name,
  file_bundle_id,
  file_bundle_path,
  signing_id,
  team_id,
  cdhash,
  entitlements,
  signing_chain,
  created_at
FROM executables
WHERE file_sha256 = $1
  AND file_name = $2
  AND NOT EXISTS (SELECT 1 FROM inserted)
LIMIT 1;

-- name: GetExecutable :one
SELECT
  id,
  file_sha256,
  file_name,
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
