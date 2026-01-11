export interface Rule {
  id: string;
  name: string;
  description: string;
  rule_type: number;
  identifier: string;
  custom_msg: string;
  custom_url: string;
  notification_app_name: string;
}

export interface PolicyAttachment {
  rule_id: string;
  action: number;
  cel_expr?: string | null;
}

export interface PolicyTarget {
  id: string;
  policy_id: string;
  kind: string;
  ref_id: string | null;
}

export interface Policy {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  priority: number;
  settings_version: number;
  rules_version: number;

  set_client_mode: number;
  set_batch_size: number;

  set_enable_bundles: boolean;
  set_enable_transitive_rules: boolean;
  set_enable_all_event_upload: boolean;
  set_disable_unknown_event_upload: boolean;
  set_full_sync_interval_seconds: number;
  set_push_notification_full_sync_interval_seconds: number;
  set_push_notification_global_rule_sync_deadline_seconds: number;
  set_allowed_path_regex: string;
  set_blocked_path_regex: string;
  set_block_usb_mount: boolean;
  set_remount_usb_mode: string[];
  set_override_file_access_action: number;

  attachments: PolicyAttachment[];
  targets: PolicyTarget[];
}

export interface Machine {
  id: string;
  serial_number: string;
  hostname: string;
  model: string;
  os_version: string;
  os_build: string;
  santa_version: string;
  primary_user: string | null;
  primary_user_groups: string[];
  push_token: string | null;
  sip_status: number;
  client_mode: number;
  request_clean_sync: boolean;
  push_notification_sync: boolean;
  user_id: string | null;
  last_seen: string | null;
  policy_id: string | null;
  policy_status: number;
  binary_rule_count: number;
  certificate_rule_count: number;
  compiler_rule_count: number;
  transitive_rule_count: number;
  team_id_rule_count: number;
  signing_id_rule_count: number;
  cdhash_rule_count: number;
  rules_hash: string | null;
  applied_policy_id: string | null;
  applied_settings_version: number | null;
  applied_rules_version: number | null;
}

export interface EventRecord {
  id: string;
  machine_id: string;
  decision: number;
  file_sha256: string;
  file_path: string;
  file_name: string;
  executing_user: string;
  execution_time: string | null;
  logged_in_users: string[];
  current_sessions: string[];
  file_bundle_id: string;
  file_bundle_path: string;
  file_bundle_executable_rel_path: string;
  file_bundle_name: string;
  file_bundle_version: string;
  file_bundle_version_string: string;
  file_bundle_hash: string;
  file_bundle_hash_millis: number;
  file_bundle_binary_count: number;
  pid: number;
  ppid: number;
  parent_name: string;
  team_id: string;
  signing_id: string;
  cdhash: string;
  cs_flags: number;
  signing_status: number;
  secure_signing_time: string | null;
  signing_time: string | null;
  signing_chain: EventCertificate[];
  entitlements: EventEntitlement[];
}

export interface EventCertificate {
  sha256: string;
  cn: string;
  org: string;
  ou: string;
  valid_from: string | null;
  valid_until: string | null;
}

export interface EventEntitlement {
  key: string;
  value: string;
}

export interface User {
  id: string;
  upn: string;
  display_name: string;
}

export interface Group {
  id: string;
  display_name: string;
  description: string;
  member_count: number;
}

export interface Membership {
  id: string;
  group_id: string;
  user_id: string;
}
