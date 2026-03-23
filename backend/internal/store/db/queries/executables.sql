-- name: GetOrCreateExecutable :one
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
  sqlc.arg(file_sha256),
  sqlc.arg(file_name),
  sqlc.arg(file_bundle_id),
  sqlc.arg(file_bundle_path),
  sqlc.arg(signing_id),
  sqlc.arg(team_id),
  sqlc.arg(cdhash),
  sqlc.arg(entitlements),
  sqlc.arg(signing_chain)
)
ON CONFLICT (file_sha256, file_name) DO UPDATE
SET file_name = executables.file_name
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
  created_at;

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
WHERE id = sqlc.arg(id);
