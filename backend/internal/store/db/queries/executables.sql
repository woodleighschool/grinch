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
  e.id,
  e.file_sha256,
  e.file_name,
  e.file_bundle_id,
  e.file_bundle_path,
  e.signing_id,
  e.team_id,
  e.cdhash,
  COALESCE(event_counts.occurrences, 0)::INT4 AS occurrences,
  e.entitlements,
  e.signing_chain,
  e.created_at
FROM executables AS e
LEFT JOIN (
  SELECT
    executable_id,
    COUNT(*)::INT4 AS occurrences
  FROM execution_events
  GROUP BY executable_id
) AS event_counts
  ON event_counts.executable_id = e.id
WHERE e.id = sqlc.arg(id);
