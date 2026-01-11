-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE EXTENSION IF NOT EXISTS "citext";

--------------------------------------------------------------------------------
-- Entra Users
--------------------------------------------------------------------------------
CREATE TABLE users (
  id UUID PRIMARY KEY,
  upn CITEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL DEFAULT ''
);

--------------------------------------------------------------------------------
-- Entra Groups
--------------------------------------------------------------------------------
CREATE TABLE groups (
  id UUID PRIMARY KEY,
  display_name TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  member_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE group_memberships (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  group_id UUID NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  UNIQUE (group_id, user_id)
);

CREATE INDEX idx_memberships_user ON group_memberships (user_id);

--------------------------------------------------------------------------------
-- Policies
-- (Somewhat) Mirrors PreflightResponse
--------------------------------------------------------------------------------
CREATE TABLE policies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  priority INTEGER NOT NULL UNIQUE,
  settings_version INTEGER NOT NULL DEFAULT 1,
  rules_version INTEGER NOT NULL DEFAULT 1,
  set_client_mode INTEGER NOT NULL,
  set_batch_size INTEGER NOT NULL DEFAULT 50,
  set_enable_bundles BOOLEAN NOT NULL,
  set_enable_transitive_rules BOOLEAN NOT NULL,
  set_enable_all_event_upload BOOLEAN NOT NULL,
  set_disable_unknown_event_upload BOOLEAN NOT NULL,
  set_full_sync_interval_seconds INTEGER NOT NULL,
  set_push_notification_full_sync_interval_seconds INTEGER NOT NULL,
  set_push_notification_global_rule_sync_deadline_seconds INTEGER NOT NULL,
  set_allowed_path_regex TEXT NOT NULL,
  set_blocked_path_regex TEXT NOT NULL,
  set_block_usb_mount BOOLEAN NOT NULL,
  set_remount_usb_mode TEXT[] NOT NULL,
  set_override_file_access_action INTEGER NOT NULL
);

CREATE INDEX idx_policies_priority ON policies (priority, id);

--------------------------------------------------------------------------------
-- Machines
-- (Somewhat) Mirrors PreflightRequest
--------------------------------------------------------------------------------
CREATE TABLE machines (
  id UUID PRIMARY KEY,
  serial_number TEXT NOT NULL DEFAULT '',
  hostname TEXT NOT NULL DEFAULT '',
  model_identifier TEXT NOT NULL DEFAULT '',
  os_version TEXT NOT NULL DEFAULT '',
  os_build TEXT NOT NULL DEFAULT '',
  santa_version TEXT NOT NULL DEFAULT '',
  primary_user TEXT,
  primary_user_groups TEXT[] NOT NULL DEFAULT '{}',
  push_notification_token TEXT,
  sip_status INTEGER NOT NULL DEFAULT 0,
  client_mode INTEGER NOT NULL DEFAULT 0,
  request_clean_sync BOOLEAN NOT NULL DEFAULT FALSE,
  push_notification_sync BOOLEAN NOT NULL DEFAULT FALSE,
  binary_rule_count INTEGER NOT NULL DEFAULT 0,
  certificate_rule_count INTEGER NOT NULL DEFAULT 0,
  compiler_rule_count INTEGER NOT NULL DEFAULT 0,
  transitive_rule_count INTEGER NOT NULL DEFAULT 0,
  teamid_rule_count INTEGER NOT NULL DEFAULT 0,
  signingid_rule_count INTEGER NOT NULL DEFAULT 0,
  cdhash_rule_count INTEGER NOT NULL DEFAULT 0,
  rules_hash TEXT,

  -- Server side tracking
  user_id UUID REFERENCES users (id) ON DELETE SET NULL,
  last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  policy_id UUID REFERENCES policies (id) ON DELETE SET NULL,
  applied_policy_id UUID REFERENCES policies (id) ON DELETE SET NULL,
  applied_settings_version INTEGER,
  applied_rules_version INTEGER,
  policy_status SMALLINT NOT NULL DEFAULT 0
);

CREATE INDEX idx_machines_user ON machines (user_id);

CREATE INDEX idx_machines_seen ON machines (last_seen DESC);

CREATE INDEX idx_machines_policy ON machines (policy_id);

--------------------------------------------------------------------------------
-- Policy targets
-- Selection criteria for policy assignments
--------------------------------------------------------------------------------
CREATE TABLE policy_targets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  policy_id UUID NOT NULL REFERENCES policies (id) ON DELETE CASCADE,
  kind TEXT NOT NULL,
  user_id UUID REFERENCES users (id) ON DELETE CASCADE,
  group_id UUID REFERENCES groups (id) ON DELETE CASCADE,
  machine_id UUID REFERENCES machines (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_policy_targets ON policy_targets (
  policy_id,
  kind,
  COALESCE(user_id::TEXT, ''),
  COALESCE(group_id::TEXT, ''),
  COALESCE(machine_id::TEXT, '')
);

CREATE INDEX idx_targets_policy ON policy_targets (policy_id);

--------------------------------------------------------------------------------
-- Rules
--------------------------------------------------------------------------------
CREATE TABLE rules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  identifier TEXT NOT NULL UNIQUE,
  rule_type INTEGER NOT NULL DEFAULT 0,
  custom_msg TEXT NOT NULL DEFAULT '',
  custom_url TEXT NOT NULL DEFAULT '',
  notification_app_name TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_rules_type ON rules (rule_type, identifier);

--------------------------------------------------------------------------------
-- Policy rules
-- Attaches rules to policies with action and optional CEL.
--------------------------------------------------------------------------------
CREATE TABLE policy_rules (
  policy_id UUID NOT NULL REFERENCES policies (id) ON DELETE CASCADE,
  rule_id UUID NOT NULL REFERENCES rules (id) ON DELETE CASCADE,
  action INTEGER NOT NULL DEFAULT 0,
  cel_expr TEXT,
  PRIMARY KEY (policy_id, rule_id)
);

CREATE INDEX idx_policy_rules_rule ON policy_rules (rule_id);

--------------------------------------------------------------------------------
-- Events
--------------------------------------------------------------------------------
CREATE TABLE events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  machine_id UUID NOT NULL REFERENCES machines (id) ON DELETE CASCADE,
  decision INTEGER NOT NULL DEFAULT 0,
  file_path TEXT NOT NULL DEFAULT '',
  file_sha256 TEXT NOT NULL DEFAULT '',
  file_name TEXT NOT NULL DEFAULT '',
  executing_user TEXT NOT NULL DEFAULT '',
  execution_time TIMESTAMPTZ,
  logged_in_users TEXT[] NOT NULL DEFAULT '{}',
  current_sessions TEXT[] NOT NULL DEFAULT '{}',
  file_bundle_id TEXT NOT NULL DEFAULT '',
  file_bundle_path TEXT NOT NULL DEFAULT '',
  file_bundle_executable_rel_path TEXT NOT NULL DEFAULT '',
  file_bundle_name TEXT NOT NULL DEFAULT '',
  file_bundle_version TEXT NOT NULL DEFAULT '',
  file_bundle_version_string TEXT NOT NULL DEFAULT '',
  file_bundle_hash TEXT NOT NULL DEFAULT '',
  file_bundle_hash_millis INTEGER NOT NULL DEFAULT 0,
  file_bundle_binary_count INTEGER NOT NULL DEFAULT 0,
  pid INTEGER NOT NULL DEFAULT 0,
  ppid INTEGER NOT NULL DEFAULT 0,
  parent_name TEXT NOT NULL DEFAULT '',
  team_id TEXT NOT NULL DEFAULT '',
  signing_id TEXT NOT NULL DEFAULT '',
  cdhash TEXT NOT NULL DEFAULT '',
  cs_flags INTEGER NOT NULL DEFAULT 0,
  signing_status INTEGER NOT NULL DEFAULT 0,
  secure_signing_time TIMESTAMPTZ,
  signing_time TIMESTAMPTZ
);

CREATE INDEX idx_events_machine ON events (machine_id, execution_time DESC);

CREATE INDEX idx_events_decision ON events (decision);

CREATE INDEX idx_events_sha256 ON events (file_sha256);

--------------------------------------------------------------------------------
-- Entitlements
--------------------------------------------------------------------------------
CREATE TABLE entitlements (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  key TEXT NOT NULL DEFAULT '',
  value TEXT NOT NULL DEFAULT '',
  UNIQUE (key, value)
);

CREATE TABLE event_entitlements (
  event_id UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
  ordinal INTEGER NOT NULL,
  entitlement_id UUID NOT NULL REFERENCES entitlements (id) ON DELETE CASCADE,
  PRIMARY KEY (event_id, ordinal)
);

--------------------------------------------------------------------------------
-- Signing certificates
--------------------------------------------------------------------------------
CREATE TABLE certificates (
  sha256 TEXT PRIMARY KEY,
  cn TEXT NOT NULL DEFAULT '',
  org TEXT NOT NULL DEFAULT '',
  ou TEXT NOT NULL DEFAULT '',
  valid_from TIMESTAMPTZ,
  valid_until TIMESTAMPTZ
);

CREATE TABLE event_signing_chain (
  event_id UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
  ordinal INTEGER NOT NULL,
  certificate_sha256 TEXT NOT NULL REFERENCES certificates (sha256) ON DELETE RESTRICT,
  PRIMARY KEY (event_id, ordinal)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS event_signing_chain;

DROP TABLE IF EXISTS event_entitlements;

DROP TABLE IF EXISTS certificates;

DROP TABLE IF EXISTS entitlements;

DROP TABLE IF EXISTS events;

DROP TABLE IF EXISTS policy_rules;

DROP TABLE IF EXISTS rules;

DROP TABLE IF EXISTS policy_targets;

DROP TABLE IF EXISTS machines;

DROP TABLE IF EXISTS policies;

DROP TABLE IF EXISTS group_memberships;

DROP TABLE IF EXISTS groups;

DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS citext;

DROP EXTENSION IF EXISTS pgcrypto;

-- +goose StatementEnd