-- name: GetMachineSyncState :one
SELECT
  m.id,
  COALESCE(ms.rules_hash, '') AS rules_hash,
  COALESCE(ms.desired_targets, '[]'::JSONB) AS desired_targets,
  COALESCE(ms.applied_targets, '[]'::JSONB) AS applied_targets,
  COALESCE(ms.pending_targets, '[]'::JSONB) AS pending_targets,
  COALESCE(ms.pending_payload, '[]'::JSONB) AS pending_payload,
  COALESCE(ms.pending_payload_rule_count, 0)::INT8 AS pending_payload_rule_count,
  COALESCE(ms.pending_full_sync, FALSE) AS pending_full_sync,
  ms.pending_preflight_at,
  COALESCE(ms.desired_binary_rule_count, 0)::INT4 AS desired_binary_rule_count,
  COALESCE(ms.desired_certificate_rule_count, 0)::INT4 AS desired_certificate_rule_count,
  COALESCE(ms.desired_teamid_rule_count, 0)::INT4 AS desired_teamid_rule_count,
  COALESCE(ms.desired_signingid_rule_count, 0)::INT4 AS desired_signingid_rule_count,
  COALESCE(ms.desired_cdhash_rule_count, 0)::INT4 AS desired_cdhash_rule_count,
  COALESCE(ms.binary_rule_count, 0)::INT4 AS binary_rule_count,
  COALESCE(ms.certificate_rule_count, 0)::INT4 AS certificate_rule_count,
  COALESCE(ms.teamid_rule_count, 0)::INT4 AS teamid_rule_count,
  COALESCE(ms.signingid_rule_count, 0)::INT4 AS signingid_rule_count,
  COALESCE(ms.cdhash_rule_count, 0)::INT4 AS cdhash_rule_count,
  COALESCE(ms.rules_received, 0)::INT4 AS rules_received,
  COALESCE(ms.rules_processed, 0)::INT4 AS rules_processed,
  ms.last_rule_sync_attempt_at,
  ms.last_rule_sync_success_at,
  ms.last_clean_sync_at,
  ms.last_reported_counts_match_at
FROM machines AS m
LEFT JOIN machine_sync_states AS ms
  ON ms.machine_id = m.id
WHERE m.id = sqlc.arg(machine_id);

-- name: UpsertMachineSyncState :exec
INSERT INTO machine_sync_states (
  machine_id,
  rules_hash,
  desired_targets,
  applied_targets,
  pending_targets,
  pending_payload,
  pending_payload_rule_count,
  pending_full_sync,
  pending_preflight_at,
  desired_binary_rule_count,
  desired_certificate_rule_count,
  desired_teamid_rule_count,
  desired_signingid_rule_count,
  desired_cdhash_rule_count,
  binary_rule_count,
  certificate_rule_count,
  teamid_rule_count,
  signingid_rule_count,
  cdhash_rule_count,
  rules_received,
  rules_processed,
  last_rule_sync_attempt_at,
  last_rule_sync_success_at,
  last_reported_counts_match_at
)
VALUES (
  sqlc.arg(machine_id),
  sqlc.arg(rules_hash),
  sqlc.arg(desired_targets),
  sqlc.arg(applied_targets),
  sqlc.arg(pending_targets),
  sqlc.arg(pending_payload),
  sqlc.arg(pending_payload_rule_count),
  sqlc.arg(pending_full_sync),
  sqlc.arg(pending_preflight_at),
  sqlc.arg(desired_binary_rule_count),
  sqlc.arg(desired_certificate_rule_count),
  sqlc.arg(desired_teamid_rule_count),
  sqlc.arg(desired_signingid_rule_count),
  sqlc.arg(desired_cdhash_rule_count),
  sqlc.arg(binary_rule_count),
  sqlc.arg(certificate_rule_count),
  sqlc.arg(teamid_rule_count),
  sqlc.arg(signingid_rule_count),
  sqlc.arg(cdhash_rule_count),
  sqlc.arg(rules_received),
  sqlc.arg(rules_processed),
  sqlc.arg(last_rule_sync_attempt_at),
  sqlc.arg(last_rule_sync_success_at),
  sqlc.arg(last_reported_counts_match_at)
)
ON CONFLICT (machine_id) DO UPDATE
SET
  rules_hash = EXCLUDED.rules_hash,
  desired_targets = EXCLUDED.desired_targets,
  applied_targets = EXCLUDED.applied_targets,
  pending_targets = EXCLUDED.pending_targets,
  pending_payload = EXCLUDED.pending_payload,
  pending_payload_rule_count = EXCLUDED.pending_payload_rule_count,
  pending_full_sync = EXCLUDED.pending_full_sync,
  pending_preflight_at = EXCLUDED.pending_preflight_at,
  desired_binary_rule_count = EXCLUDED.desired_binary_rule_count,
  desired_certificate_rule_count = EXCLUDED.desired_certificate_rule_count,
  desired_teamid_rule_count = EXCLUDED.desired_teamid_rule_count,
  desired_signingid_rule_count = EXCLUDED.desired_signingid_rule_count,
  desired_cdhash_rule_count = EXCLUDED.desired_cdhash_rule_count,
  binary_rule_count = EXCLUDED.binary_rule_count,
  certificate_rule_count = EXCLUDED.certificate_rule_count,
  teamid_rule_count = EXCLUDED.teamid_rule_count,
  signingid_rule_count = EXCLUDED.signingid_rule_count,
  cdhash_rule_count = EXCLUDED.cdhash_rule_count,
  rules_received = EXCLUDED.rules_received,
  rules_processed = EXCLUDED.rules_processed,
  last_rule_sync_attempt_at = EXCLUDED.last_rule_sync_attempt_at,
  last_rule_sync_success_at = EXCLUDED.last_rule_sync_success_at,
  last_clean_sync_at = CASE
    WHEN machine_sync_states.desired_targets IS DISTINCT FROM EXCLUDED.desired_targets THEN NULL
    ELSE machine_sync_states.last_clean_sync_at
  END,
  last_reported_counts_match_at = CASE
    WHEN machine_sync_states.desired_targets IS DISTINCT FROM EXCLUDED.desired_targets THEN NULL
    ELSE EXCLUDED.last_reported_counts_match_at
  END,
  updated_at = NOW();

-- name: UpsertMachineDesiredTargets :exec
INSERT INTO machine_sync_states (
  machine_id,
  desired_targets,
  desired_binary_rule_count,
  desired_certificate_rule_count,
  desired_teamid_rule_count,
  desired_signingid_rule_count,
  desired_cdhash_rule_count
)
VALUES (
  sqlc.arg(machine_id),
  sqlc.arg(desired_targets),
  sqlc.arg(desired_binary_rule_count),
  sqlc.arg(desired_certificate_rule_count),
  sqlc.arg(desired_teamid_rule_count),
  sqlc.arg(desired_signingid_rule_count),
  sqlc.arg(desired_cdhash_rule_count)
)
ON CONFLICT (machine_id) DO UPDATE
SET
  desired_targets = EXCLUDED.desired_targets,
  desired_binary_rule_count = EXCLUDED.desired_binary_rule_count,
  desired_certificate_rule_count = EXCLUDED.desired_certificate_rule_count,
  desired_teamid_rule_count = EXCLUDED.desired_teamid_rule_count,
  desired_signingid_rule_count = EXCLUDED.desired_signingid_rule_count,
  desired_cdhash_rule_count = EXCLUDED.desired_cdhash_rule_count,
  last_clean_sync_at = CASE
    WHEN machine_sync_states.desired_targets IS DISTINCT FROM EXCLUDED.desired_targets THEN NULL
    ELSE machine_sync_states.last_clean_sync_at
  END,
  last_reported_counts_match_at = CASE
    WHEN machine_sync_states.desired_targets IS DISTINCT FROM EXCLUDED.desired_targets THEN NULL
    ELSE machine_sync_states.last_reported_counts_match_at
  END,
  updated_at = NOW();

-- name: RecordMachineSyncPostflight :execrows
UPDATE machine_sync_states
SET
  rules_hash = sqlc.arg(rules_hash),
  rules_received = sqlc.arg(rules_received),
  rules_processed = sqlc.arg(rules_processed),
  last_rule_sync_attempt_at = sqlc.arg(last_rule_sync_attempt_at),
  updated_at = NOW()
WHERE machine_id = sqlc.arg(machine_id);

-- name: PromoteMachineSyncPendingSnapshot :execrows
UPDATE machine_sync_states
SET
  applied_targets = pending_targets,
  pending_targets = '[]'::JSONB,
  pending_payload = '[]'::JSONB,
  pending_payload_rule_count = 0,
  pending_full_sync = FALSE,
  pending_preflight_at = NULL,
  last_rule_sync_success_at = sqlc.arg(last_rule_sync_success_at),
  last_clean_sync_at = CASE
    WHEN pending_full_sync THEN sqlc.arg(last_rule_sync_success_at)
    ELSE last_clean_sync_at
  END,
  updated_at = NOW()
WHERE machine_id = sqlc.arg(machine_id);
