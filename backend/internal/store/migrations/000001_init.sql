-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS machines (
  machine_id UUID PRIMARY KEY,
  serial_number TEXT NOT NULL DEFAULT '',
  hostname TEXT NOT NULL DEFAULT '',
  model_identifier TEXT NOT NULL DEFAULT '',
  os_version TEXT NOT NULL DEFAULT '',
  os_build TEXT NOT NULL DEFAULT '',
  santa_version TEXT NOT NULL DEFAULT '',
  primary_user TEXT NOT NULL DEFAULT '',
  primary_user_groups_raw JSONB NOT NULL DEFAULT '[]'::JSONB,
  last_seen_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS machines_last_seen_at_idx ON machines (last_seen_at DESC);
CREATE INDEX IF NOT EXISTS machines_primary_user_idx ON machines (primary_user);

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  upn TEXT NOT NULL DEFAULT '',
  display_name TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT users_source_check CHECK (source IN ('local', 'entra'))
);

CREATE INDEX IF NOT EXISTS users_source_idx ON users (source);
CREATE INDEX IF NOT EXISTS users_upn_idx ON users (upn);
CREATE UNIQUE INDEX IF NOT EXISTS users_upn_unique_idx ON users (upn) WHERE upn <> '';

CREATE TABLE IF NOT EXISTS groups (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT groups_source_check CHECK (source IN ('local', 'entra'))
);

CREATE INDEX IF NOT EXISTS groups_source_idx ON groups (source);
CREATE INDEX IF NOT EXISTS groups_name_idx ON groups (name);

CREATE TABLE IF NOT EXISTS group_memberships (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id UUID NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
  member_kind TEXT NOT NULL,
  member_id UUID NOT NULL,
  origin TEXT NOT NULL DEFAULT 'explicit',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT group_memberships_member_kind_check CHECK (member_kind IN ('user', 'machine')),
  CONSTRAINT group_memberships_origin_check CHECK (origin IN ('explicit', 'synced')),
  CONSTRAINT group_memberships_unique UNIQUE (group_id, member_kind, member_id)
);

CREATE INDEX IF NOT EXISTS group_memberships_group_idx ON group_memberships (group_id);
CREATE INDEX IF NOT EXISTS group_memberships_member_lookup_idx ON group_memberships (member_kind, member_id);
CREATE INDEX IF NOT EXISTS group_memberships_origin_idx ON group_memberships (origin);

CREATE TABLE IF NOT EXISTS rules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  rule_type TEXT NOT NULL,
  identifier TEXT NOT NULL,
  identifier_key TEXT GENERATED ALWAYS AS (lower(btrim(identifier))) STORED,
  custom_message TEXT NOT NULL DEFAULT '',
  custom_url TEXT NOT NULL DEFAULT '',
  enabled BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT rules_rule_type_check CHECK (rule_type IN ('binary', 'certificate', 'team_id', 'signing_id', 'cd_hash')),
  CONSTRAINT rules_identifier_key_check CHECK (identifier_key <> ''),
  CONSTRAINT rules_type_identifier_key_unique UNIQUE (rule_type, identifier_key)
);

CREATE TABLE IF NOT EXISTS rule_targets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  rule_id UUID NOT NULL REFERENCES rules (id) ON DELETE CASCADE,
  subject_kind TEXT NOT NULL DEFAULT 'group',
  subject_id UUID NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
  assignment TEXT NOT NULL,
  priority INTEGER NULL,
  policy TEXT NULL,
  cel_expression TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT rule_targets_subject_kind_check CHECK (subject_kind IN ('group')),
  CONSTRAINT rule_targets_assignment_check CHECK (assignment IN ('include', 'exclude')),
  CONSTRAINT rule_targets_policy_check CHECK (policy IS NULL OR policy IN ('allowlist', 'blocklist', 'silent_blocklist', 'cel')),
  CONSTRAINT rule_targets_unique UNIQUE (rule_id, subject_kind, subject_id),
  CONSTRAINT rule_targets_include_requires_fields CHECK (
    (
      assignment = 'include'
      AND priority IS NOT NULL
      AND policy IS NOT NULL
    )
    OR (
      assignment = 'exclude'
      AND priority IS NULL
      AND policy IS NULL
    )
  ),
  CONSTRAINT rule_targets_cel_expression_check CHECK (
    (policy = 'cel' AND cel_expression <> '')
    OR (policy IS DISTINCT FROM 'cel' AND cel_expression = '')
  )
);

CREATE INDEX IF NOT EXISTS rule_targets_rule_priority_idx
  ON rule_targets (rule_id, priority, subject_id)
  WHERE assignment = 'include';
CREATE INDEX IF NOT EXISTS rule_targets_subject_idx ON rule_targets (subject_kind, subject_id);

CREATE TABLE IF NOT EXISTS machine_rule_sync_states (
  machine_id UUID PRIMARY KEY REFERENCES machines (machine_id) ON DELETE CASCADE,
  request_clean_sync BOOLEAN NOT NULL DEFAULT FALSE,
  last_client_rules_hash TEXT NOT NULL DEFAULT '',
  acknowledged_targets JSONB NOT NULL DEFAULT '[]'::JSONB,
  pending_targets JSONB NOT NULL DEFAULT '[]'::JSONB,
  pending_expected_rules_hash TEXT NOT NULL DEFAULT '',
  pending_payload_rule_count BIGINT NOT NULL DEFAULT 0,
  pending_sync_type TEXT NOT NULL DEFAULT '',
  pending_preflight_at TIMESTAMPTZ NULL,
  last_postflight_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT machine_rule_sync_states_pending_payload_rule_count_check CHECK (pending_payload_rule_count >= 0),
  CONSTRAINT machine_rule_sync_states_pending_sync_type_check CHECK (
    pending_sync_type IN ('', 'normal', 'clean_rules')
  )
);

CREATE TABLE IF NOT EXISTS executables (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source TEXT NOT NULL,
  file_sha256 TEXT NOT NULL DEFAULT '',
  file_name TEXT NOT NULL DEFAULT '',
  file_path TEXT NOT NULL DEFAULT '',
  file_bundle_id TEXT NOT NULL DEFAULT '',
  file_bundle_path TEXT NOT NULL DEFAULT '',
  signing_id TEXT NOT NULL DEFAULT '',
  team_id TEXT NOT NULL DEFAULT '',
  cdhash TEXT NOT NULL DEFAULT '',
  entitlements JSONB NOT NULL DEFAULT '{}'::JSONB,
  signing_chain JSONB NOT NULL DEFAULT '[]'::JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT executables_source_check CHECK (source IN ('event', 'process')),
  CONSTRAINT executables_file_sha256_check CHECK (file_sha256 <> '')
);

CREATE INDEX IF NOT EXISTS executables_file_name_idx ON executables (file_name);
CREATE INDEX IF NOT EXISTS executables_file_path_idx ON executables (file_path);
CREATE INDEX IF NOT EXISTS executables_signing_id_idx ON executables (signing_id);
CREATE INDEX IF NOT EXISTS executables_team_id_idx ON executables (team_id);
CREATE INDEX IF NOT EXISTS executables_cdhash_idx ON executables (cdhash);
CREATE INDEX IF NOT EXISTS executables_sha_idx ON executables (file_sha256);
CREATE UNIQUE INDEX IF NOT EXISTS executables_event_identity_idx
  ON executables (file_sha256, file_name)
  WHERE source = 'event';
CREATE UNIQUE INDEX IF NOT EXISTS executables_process_identity_idx
  ON executables (file_sha256, file_path, signing_id, team_id, cdhash, signing_chain)
  WHERE source = 'process';

CREATE TABLE IF NOT EXISTS execution_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  machine_id UUID NOT NULL REFERENCES machines (machine_id) ON DELETE CASCADE,
  executable_id UUID NOT NULL REFERENCES executables (id) ON DELETE CASCADE,
  decision TEXT NOT NULL,
  file_path TEXT NOT NULL DEFAULT '',
  executing_user TEXT NOT NULL DEFAULT '',
  logged_in_users TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  current_sessions TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  occurred_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT execution_events_decision_check CHECK (
    decision IN (
      'unknown',
      'allow_unknown',
      'allow_binary',
      'allow_certificate',
      'allow_scope',
      'allow_team_id',
      'allow_signing_id',
      'allow_cd_hash',
      'block_unknown',
      'block_binary',
      'block_certificate',
      'block_scope',
      'block_team_id',
      'block_signing_id',
      'block_cd_hash',
      'bundle_binary'
    )
  )
);

CREATE INDEX IF NOT EXISTS execution_events_machine_created_idx
  ON execution_events (machine_id, created_at DESC);
CREATE INDEX IF NOT EXISTS execution_events_decision_created_idx
  ON execution_events (decision, created_at DESC);
CREATE INDEX IF NOT EXISTS execution_events_executable_created_idx
  ON execution_events (executable_id, created_at DESC);

CREATE TABLE IF NOT EXISTS file_access_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  machine_id UUID NOT NULL REFERENCES machines (machine_id) ON DELETE CASCADE,
  executable_id UUID NULL REFERENCES executables (id) ON DELETE SET NULL,
  rule_version TEXT NOT NULL DEFAULT '',
  rule_name TEXT NOT NULL DEFAULT '',
  target TEXT NOT NULL DEFAULT '',
  decision TEXT NOT NULL,
  process_chain JSONB NOT NULL DEFAULT '[]'::JSONB,
  occurred_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT file_access_events_decision_check CHECK (
    decision IN ('unknown', 'denied', 'denied_invalid_signature', 'audit_only')
  )
);

CREATE INDEX IF NOT EXISTS file_access_events_machine_created_idx
  ON file_access_events (machine_id, created_at DESC);
CREATE INDEX IF NOT EXISTS file_access_events_decision_created_idx
  ON file_access_events (decision, created_at DESC);
CREATE INDEX IF NOT EXISTS file_access_events_executable_created_idx
  ON file_access_events (executable_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS file_access_events;
DROP TABLE IF EXISTS execution_events;
DROP TABLE IF EXISTS executables;
DROP TABLE IF EXISTS machine_rule_sync_states;
DROP TABLE IF EXISTS rule_targets;
DROP TABLE IF EXISTS rules;
DROP TABLE IF EXISTS group_memberships;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS machines;
