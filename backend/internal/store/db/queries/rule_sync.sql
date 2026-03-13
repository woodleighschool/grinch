-- name: GetMachineRuleSyncState :one
SELECT
  m.machine_id,
  COALESCE(rs.request_clean_sync, FALSE) AS request_clean_sync,
  COALESCE(rs.last_client_rules_hash, '') AS last_client_rules_hash,
  COALESCE(rs.acknowledged_targets, '[]'::JSONB) AS acknowledged_targets,
  COALESCE(rs.pending_targets, '[]'::JSONB) AS pending_targets,
  COALESCE(rs.pending_expected_rules_hash, '') AS pending_expected_rules_hash,
  COALESCE(rs.pending_payload_rule_count, 0)::INT8 AS pending_payload_rule_count,
  COALESCE(rs.pending_sync_type, '') AS pending_sync_type,
  rs.pending_preflight_at,
  rs.last_postflight_at
FROM machines AS m
LEFT JOIN machine_rule_sync_states AS rs
  ON rs.machine_id = m.machine_id
WHERE m.machine_id = $1;

-- name: UpsertMachineRuleSyncState :one
INSERT INTO machine_rule_sync_states (
  machine_id,
  request_clean_sync,
  last_client_rules_hash,
  acknowledged_targets,
  pending_targets,
  pending_expected_rules_hash,
  pending_payload_rule_count,
  pending_sync_type,
  pending_preflight_at,
  last_postflight_at
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
  $9,
  $10
)
ON CONFLICT (machine_id) DO UPDATE SET
  request_clean_sync = EXCLUDED.request_clean_sync,
  last_client_rules_hash = EXCLUDED.last_client_rules_hash,
  acknowledged_targets = EXCLUDED.acknowledged_targets,
  pending_targets = EXCLUDED.pending_targets,
  pending_expected_rules_hash = EXCLUDED.pending_expected_rules_hash,
  pending_payload_rule_count = EXCLUDED.pending_payload_rule_count,
  pending_sync_type = EXCLUDED.pending_sync_type,
  pending_preflight_at = EXCLUDED.pending_preflight_at,
  last_postflight_at = EXCLUDED.last_postflight_at,
  updated_at = NOW()
RETURNING
  machine_id,
  request_clean_sync,
  last_client_rules_hash,
  acknowledged_targets,
  pending_targets,
  pending_expected_rules_hash,
  pending_payload_rule_count,
  pending_sync_type,
  pending_preflight_at,
  last_postflight_at,
  created_at,
  updated_at;

-- name: GetMachineAcknowledgedRuleTargetsJSON :one
SELECT COALESCE(
  (
    SELECT acknowledged_targets::TEXT
    FROM machine_rule_sync_states
    WHERE machine_id = $1
  ),
  '[]'
) AS acknowledged_targets;
