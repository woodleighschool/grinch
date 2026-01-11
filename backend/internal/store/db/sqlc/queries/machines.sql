-- name: GetMachineByID :one
SELECT * FROM machines WHERE id = $1;

-- name: UpsertMachineByID :one
INSERT INTO machines (
  id, serial_number, hostname, model_identifier, os_version, os_build, santa_version,
  primary_user, primary_user_groups, push_notification_token,
  sip_status, client_mode, request_clean_sync, push_notification_sync,
  binary_rule_count, certificate_rule_count, compiler_rule_count, transitive_rule_count,
  teamid_rule_count, signingid_rule_count, cdhash_rule_count, rules_hash,
  user_id, last_seen, policy_id, applied_policy_id, applied_settings_version, applied_rules_version, policy_status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  $8, $9, $10,
  $11, $12, $13, $14,
  $15, $16, $17, $18,
  $19, $20, $21, $22,
  $23, $24, $25, $26, $27, $28, $29
)
ON CONFLICT (id) DO UPDATE SET
  serial_number = EXCLUDED.serial_number,
  hostname = EXCLUDED.hostname,
  model_identifier = EXCLUDED.model_identifier,
  os_version = EXCLUDED.os_version,
  os_build = EXCLUDED.os_build,
  santa_version = EXCLUDED.santa_version,
  primary_user = EXCLUDED.primary_user,
  primary_user_groups = EXCLUDED.primary_user_groups,
  push_notification_token = EXCLUDED.push_notification_token,
  sip_status = EXCLUDED.sip_status,
  client_mode = EXCLUDED.client_mode,
  request_clean_sync = EXCLUDED.request_clean_sync,
  push_notification_sync = EXCLUDED.push_notification_sync,
  binary_rule_count = EXCLUDED.binary_rule_count,
  certificate_rule_count = EXCLUDED.certificate_rule_count,
  compiler_rule_count = EXCLUDED.compiler_rule_count,
  transitive_rule_count = EXCLUDED.transitive_rule_count,
  teamid_rule_count = EXCLUDED.teamid_rule_count,
  signingid_rule_count = EXCLUDED.signingid_rule_count,
  cdhash_rule_count = EXCLUDED.cdhash_rule_count,
  rules_hash = EXCLUDED.rules_hash,
  user_id = EXCLUDED.user_id,
  last_seen = EXCLUDED.last_seen,
  policy_id = EXCLUDED.policy_id,
  applied_policy_id = EXCLUDED.applied_policy_id,
  applied_settings_version = EXCLUDED.applied_settings_version,
  applied_rules_version = EXCLUDED.applied_rules_version,
  policy_status = EXCLUDED.policy_status
RETURNING *;

-- name: UpdateMachinePolicyStateByID :exec
UPDATE machines SET
  policy_id = $2,
  policy_status = $3
WHERE id = $1;
