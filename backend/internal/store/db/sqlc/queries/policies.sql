-- name: GetPolicyByID :one
SELECT * FROM policies WHERE id = $1;

-- name: ListEnabledPolicies :many
SELECT * FROM policies WHERE enabled = TRUE ORDER BY priority DESC, id ASC;

-- name: CreatePolicy :one
INSERT INTO policies (
  name, description, enabled, priority, settings_version, rules_version,
  set_client_mode, set_batch_size,
  set_enable_bundles, set_enable_transitive_rules, set_enable_all_event_upload, set_disable_unknown_event_upload,
  set_full_sync_interval_seconds, set_push_notification_full_sync_interval_seconds, set_push_notification_global_rule_sync_deadline_seconds,
  set_allowed_path_regex, set_blocked_path_regex,
  set_block_usb_mount, set_remount_usb_mode,
  set_override_file_access_action
) VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8,
  $9, $10, $11, $12,
  $13, $14, $15,
  $16, $17,
  $18, $19,
  $20
)
RETURNING *;

-- name: UpdatePolicyByID :one
UPDATE policies SET
  name = $2,
  description = $3,
  enabled = $4,
  priority = $5,
  settings_version = $6,
  rules_version = $7,
  set_client_mode = $8,
  set_batch_size = $9,
  set_enable_bundles = $10,
  set_enable_transitive_rules = $11,
  set_enable_all_event_upload = $12,
  set_disable_unknown_event_upload = $13,
  set_full_sync_interval_seconds = $14,
  set_push_notification_full_sync_interval_seconds = $15,
  set_push_notification_global_rule_sync_deadline_seconds = $16,
  set_allowed_path_regex = $17,
  set_blocked_path_regex = $18,
  set_block_usb_mount = $19,
  set_remount_usb_mode = $20,
  set_override_file_access_action = $21
WHERE id = $1
RETURNING *;

-- name: DeletePolicyByID :exec
DELETE FROM policies WHERE id = $1;

-- name: UpdatePolicyRulesVersionByRuleID :exec
UPDATE policies SET rules_version = rules_version + 1
WHERE id IN (SELECT policy_id FROM policy_rules WHERE rule_id = $1);

-- Targets

-- name: ListPolicyTargetsByPolicyID :many
SELECT * FROM policy_targets WHERE policy_id = $1 ORDER BY kind, id;

-- name: ListPolicyTargetsByPolicyIDs :many
SELECT * FROM policy_targets WHERE policy_id = ANY($1::uuid[]) ORDER BY policy_id, kind, id;

-- name: CreatePolicyTarget :exec
INSERT INTO policy_targets (policy_id, kind, user_id, group_id, machine_id)
VALUES ($1, $2, $3, $4, $5);

-- name: DeletePolicyTargetsByPolicyID :exec
DELETE FROM policy_targets WHERE policy_id = $1;
