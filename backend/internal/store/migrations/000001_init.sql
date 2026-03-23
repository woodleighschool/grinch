-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE principal_source AS ENUM ('local', 'entra');
CREATE TYPE membership_member_kind AS ENUM ('user', 'machine');
CREATE TYPE membership_origin AS ENUM ('explicit', 'synced');
CREATE TYPE rule_type AS ENUM ('binary', 'certificate', 'team_id', 'signing_id', 'cd_hash');
CREATE TYPE rule_target_subject_kind AS ENUM ('group', 'all_devices', 'all_users');
CREATE TYPE rule_target_assignment AS ENUM ('include', 'exclude');
CREATE TYPE rule_policy AS ENUM ('allowlist', 'blocklist', 'silent_blocklist', 'cel');
CREATE TYPE santa_client_mode AS ENUM ('unknown', 'monitor', 'lockdown', 'standalone');
CREATE TYPE execution_decision AS ENUM (
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
);
CREATE TYPE file_access_decision AS ENUM (
  'unknown',
  'denied',
  'denied_invalid_signature',
  'audit_only'
);

CREATE TABLE machines (
  id UUID PRIMARY KEY,
  serial_number TEXT NOT NULL,
  hostname TEXT NOT NULL,
  model_identifier TEXT NOT NULL,
  os_version TEXT NOT NULL,
  os_build TEXT NOT NULL,
  santa_version TEXT NOT NULL,
  primary_user TEXT NOT NULL DEFAULT '',
  primary_user_groups TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  client_mode santa_client_mode NOT NULL DEFAULT 'unknown',
  last_seen_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX machines_last_seen_at_idx ON machines (last_seen_at DESC);
CREATE INDEX machines_primary_user_idx ON machines (primary_user);

CREATE TABLE users (
  id UUID PRIMARY KEY,
  upn TEXT NOT NULL DEFAULT '',
  display_name TEXT NOT NULL DEFAULT '',
  source principal_source NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX users_source_idx ON users (source);
CREATE UNIQUE INDEX users_upn_unique_idx ON users (upn) WHERE upn <> '';

CREATE TABLE groups (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  source principal_source NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT groups_name_not_blank CHECK (btrim(name) <> '')
);

CREATE INDEX groups_source_idx ON groups (source);
CREATE INDEX groups_name_idx ON groups (name);

CREATE TABLE group_user_memberships (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id UUID NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  origin membership_origin NOT NULL DEFAULT 'explicit',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT group_user_memberships_unique UNIQUE (group_id, user_id)
);

CREATE INDEX group_user_memberships_group_id_idx ON group_user_memberships (group_id);
CREATE INDEX group_user_memberships_user_id_idx ON group_user_memberships (user_id);

CREATE TABLE group_machine_memberships (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id UUID NOT NULL REFERENCES groups (id) ON DELETE CASCADE,
  machine_id UUID NOT NULL REFERENCES machines (id) ON DELETE CASCADE,
  origin membership_origin NOT NULL DEFAULT 'explicit',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT group_machine_memberships_unique UNIQUE (group_id, machine_id)
);

CREATE INDEX group_machine_memberships_group_id_idx ON group_machine_memberships (group_id);
CREATE INDEX group_machine_memberships_machine_id_idx ON group_machine_memberships (machine_id);

CREATE VIEW group_memberships AS
SELECT
  id,
  group_id,
  'user'::membership_member_kind AS member_kind,
  user_id AS member_id,
  origin,
  created_at,
  updated_at
FROM group_user_memberships

UNION ALL

SELECT
  id,
  group_id,
  'machine'::membership_member_kind AS member_kind,
  machine_id AS member_id,
  origin,
  created_at,
  updated_at
FROM group_machine_memberships;

CREATE TABLE rules (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  rule_type rule_type NOT NULL,
  identifier TEXT NOT NULL,
  custom_message TEXT NOT NULL DEFAULT '',
  custom_url TEXT NOT NULL DEFAULT '',
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT rules_name_not_blank CHECK (btrim(name) <> ''),
  CONSTRAINT rules_identifier_not_blank CHECK (btrim(identifier) <> ''),
  CONSTRAINT rules_type_identifier_unique UNIQUE (rule_type, identifier)
);

CREATE TABLE rule_targets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  rule_id UUID NOT NULL REFERENCES rules (id) ON DELETE CASCADE,
  subject_kind rule_target_subject_kind NOT NULL DEFAULT 'group',
  subject_id UUID NULL REFERENCES groups (id) ON DELETE CASCADE,
  assignment rule_target_assignment NOT NULL,
  priority INTEGER NULL,
  policy rule_policy NULL,
  cel_expression TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT rule_targets_subject_id_check CHECK (
    (subject_kind = 'group' AND subject_id IS NOT NULL)
    OR (subject_kind IN ('all_devices', 'all_users') AND subject_id IS NULL)
  ),
  CONSTRAINT rule_targets_include_requires_fields CHECK (
    (
      assignment = 'include'
      AND priority IS NOT NULL
      AND policy IS NOT NULL
    )
    OR (
      assignment = 'exclude'
      AND subject_kind = 'group'
      AND priority IS NULL
      AND policy IS NULL
    )
  ),
  CONSTRAINT rule_targets_cel_expression_check CHECK (
    (policy = 'cel' AND cel_expression <> '')
    OR (policy IS DISTINCT FROM 'cel' AND cel_expression = '')
  )
);

CREATE INDEX rule_targets_rule_id_idx ON rule_targets (rule_id);
CREATE INDEX rule_targets_include_priority_idx
  ON rule_targets (rule_id, priority)
  WHERE assignment = 'include';
CREATE INDEX rule_targets_group_subject_idx
  ON rule_targets (subject_id)
  WHERE subject_kind = 'group';
CREATE UNIQUE INDEX rule_targets_group_unique_idx
  ON rule_targets (rule_id, subject_id)
  WHERE subject_kind = 'group';
CREATE UNIQUE INDEX rule_targets_global_unique_idx
  ON rule_targets (rule_id, subject_kind)
  WHERE subject_kind IN ('all_devices', 'all_users');

CREATE TABLE machine_sync_states (
  machine_id UUID PRIMARY KEY REFERENCES machines (id) ON DELETE CASCADE,
  rules_hash TEXT NOT NULL DEFAULT '',
  applied_targets JSONB NOT NULL DEFAULT '[]'::JSONB,
  pending_targets JSONB NOT NULL DEFAULT '[]'::JSONB,
  pending_payload JSONB NOT NULL DEFAULT '[]'::JSONB,
  desired_targets JSONB NOT NULL DEFAULT '[]'::JSONB,
  pending_payload_rule_count BIGINT NOT NULL DEFAULT 0,
  pending_full_sync BOOLEAN NOT NULL DEFAULT FALSE,
  pending_preflight_at TIMESTAMPTZ NULL,
  desired_binary_rule_count INTEGER NOT NULL DEFAULT 0,
  desired_certificate_rule_count INTEGER NOT NULL DEFAULT 0,
  desired_teamid_rule_count INTEGER NOT NULL DEFAULT 0,
  desired_signingid_rule_count INTEGER NOT NULL DEFAULT 0,
  desired_cdhash_rule_count INTEGER NOT NULL DEFAULT 0,
  binary_rule_count INTEGER NOT NULL DEFAULT 0,
  certificate_rule_count INTEGER NOT NULL DEFAULT 0,
  teamid_rule_count INTEGER NOT NULL DEFAULT 0,
  signingid_rule_count INTEGER NOT NULL DEFAULT 0,
  cdhash_rule_count INTEGER NOT NULL DEFAULT 0,
  rules_received INTEGER NOT NULL DEFAULT 0,
  rules_processed INTEGER NOT NULL DEFAULT 0,
  last_rule_sync_attempt_at TIMESTAMPTZ NULL,
  last_rule_sync_success_at TIMESTAMPTZ NULL,
  last_clean_sync_at TIMESTAMPTZ NULL,
  last_reported_counts_match_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT machine_sync_states_applied_targets_is_array
    CHECK (jsonb_typeof(applied_targets) = 'array'),
  CONSTRAINT machine_sync_states_pending_targets_is_array
    CHECK (jsonb_typeof(pending_targets) = 'array'),
  CONSTRAINT machine_sync_states_pending_payload_is_array
    CHECK (jsonb_typeof(pending_payload) = 'array'),
  CONSTRAINT machine_sync_states_desired_targets_is_array
    CHECK (jsonb_typeof(desired_targets) = 'array')
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER machines_set_updated_at
  BEFORE UPDATE ON machines
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER users_set_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER groups_set_updated_at
  BEFORE UPDATE ON groups
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER group_user_memberships_set_updated_at
  BEFORE UPDATE ON group_user_memberships
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER group_machine_memberships_set_updated_at
  BEFORE UPDATE ON group_machine_memberships
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER rules_set_updated_at
  BEFORE UPDATE ON rules
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER rule_targets_set_updated_at
  BEFORE UPDATE ON rule_targets
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER machine_sync_states_set_updated_at
  BEFORE UPDATE ON machine_sync_states
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION machine_rule_sync_status(
  pending_preflight_at TIMESTAMPTZ,
  desired_targets JSONB,
  applied_targets JSONB,
  desired_binary_rule_count INTEGER,
  binary_rule_count INTEGER,
  desired_certificate_rule_count INTEGER,
  certificate_rule_count INTEGER,
  desired_teamid_rule_count INTEGER,
  teamid_rule_count INTEGER,
  desired_signingid_rule_count INTEGER,
  signingid_rule_count INTEGER,
  desired_cdhash_rule_count INTEGER,
  cdhash_rule_count INTEGER,
  last_clean_sync_at TIMESTAMPTZ,
  last_reported_counts_match_at TIMESTAMPTZ
) RETURNS TEXT
LANGUAGE SQL
IMMUTABLE
AS $$
  SELECT CASE
    WHEN pending_preflight_at IS NOT NULL THEN 'pending'
    WHEN COALESCE(desired_targets, '[]'::JSONB) IS DISTINCT FROM COALESCE(applied_targets, '[]'::JSONB) THEN 'pending'
    WHEN (
      COALESCE(desired_binary_rule_count, 0) = COALESCE(binary_rule_count, 0)
      AND COALESCE(desired_certificate_rule_count, 0) = COALESCE(certificate_rule_count, 0)
      AND COALESCE(desired_teamid_rule_count, 0) = COALESCE(teamid_rule_count, 0)
      AND COALESCE(desired_signingid_rule_count, 0) = COALESCE(signingid_rule_count, 0)
      AND COALESCE(desired_cdhash_rule_count, 0) = COALESCE(cdhash_rule_count, 0)
    ) THEN 'synced'
    WHEN last_clean_sync_at IS NOT NULL
      AND (
        last_reported_counts_match_at IS NULL
        OR last_reported_counts_match_at < last_clean_sync_at
      ) THEN 'issue'
    ELSE 'pending'
  END;
$$;
-- +goose StatementEnd

CREATE TABLE executables (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  file_sha256 TEXT NOT NULL,
  file_name TEXT NOT NULL,
  file_bundle_id TEXT NOT NULL DEFAULT '',
  file_bundle_path TEXT NOT NULL DEFAULT '',
  signing_id TEXT NOT NULL DEFAULT '',
  team_id TEXT NOT NULL DEFAULT '',
  cdhash TEXT NOT NULL DEFAULT '',
  entitlements JSONB NOT NULL DEFAULT '{}'::JSONB,
  signing_chain JSONB NOT NULL DEFAULT '[]'::JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT executables_file_sha256_not_blank CHECK (btrim(file_sha256) <> ''),
  CONSTRAINT executables_file_name_not_blank CHECK (btrim(file_name) <> ''),
  CONSTRAINT executables_entitlements_is_object CHECK (jsonb_typeof(entitlements) = 'object'),
  CONSTRAINT executables_signing_chain_is_array CHECK (jsonb_typeof(signing_chain) = 'array'),
  CONSTRAINT executables_file_sha256_name_unique UNIQUE (file_sha256, file_name)
);

CREATE INDEX executables_file_name_idx ON executables (file_name);
CREATE INDEX executables_signing_id_idx ON executables (signing_id);
CREATE INDEX executables_team_id_idx ON executables (team_id);
CREATE INDEX executables_cdhash_idx ON executables (cdhash);

CREATE TABLE execution_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  machine_id UUID NOT NULL REFERENCES machines (id) ON DELETE CASCADE,
  executable_id UUID NOT NULL REFERENCES executables (id) ON DELETE CASCADE,
  decision execution_decision NOT NULL,
  file_path TEXT NOT NULL DEFAULT '',
  executing_user TEXT NOT NULL DEFAULT '',
  logged_in_users TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  current_sessions TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
  occurred_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX execution_events_machine_created_idx
  ON execution_events (machine_id, created_at DESC);
CREATE INDEX execution_events_decision_created_idx
  ON execution_events (decision, created_at DESC);
CREATE INDEX execution_events_executable_created_idx
  ON execution_events (executable_id, created_at DESC);

CREATE TABLE file_access_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  machine_id UUID NOT NULL REFERENCES machines (id) ON DELETE CASCADE,
  rule_version TEXT NOT NULL DEFAULT '',
  rule_name TEXT NOT NULL DEFAULT '',
  target TEXT NOT NULL DEFAULT '',
  decision file_access_decision NOT NULL,
  process_chain JSONB NOT NULL DEFAULT '[]'::JSONB,
  occurred_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT file_access_events_process_chain_is_array
    CHECK (jsonb_typeof(process_chain) = 'array')
);

CREATE INDEX file_access_events_machine_created_idx
  ON file_access_events (machine_id, created_at DESC);
CREATE INDEX file_access_events_decision_created_idx
  ON file_access_events (decision, created_at DESC);
