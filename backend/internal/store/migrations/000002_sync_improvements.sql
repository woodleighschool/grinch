-- +goose Up

-- rule_targets: make subject_id nullable to support all_devices / all_users global targets
ALTER TABLE rule_targets ALTER COLUMN subject_id DROP NOT NULL;

-- rule_targets: drop timestamps (not needed on this join-like table)
ALTER TABLE rule_targets DROP COLUMN created_at;
ALTER TABLE rule_targets DROP COLUMN updated_at;

-- rule_targets: expand subject_kind to include global target kinds
ALTER TABLE rule_targets DROP CONSTRAINT rule_targets_subject_kind_check;
ALTER TABLE rule_targets ADD CONSTRAINT rule_targets_subject_kind_check
  CHECK (subject_kind IN ('group', 'all_devices', 'all_users'));

-- rule_targets: enforce subject_id presence/absence based on subject_kind
ALTER TABLE rule_targets ADD CONSTRAINT rule_targets_subject_id_check CHECK (
  (subject_kind = 'group' AND subject_id IS NOT NULL)
  OR (subject_kind IN ('all_devices', 'all_users') AND subject_id IS NULL)
);

-- rule_targets: tighten exclude arm to group targets only; replace old unique with partial indexes
ALTER TABLE rule_targets DROP CONSTRAINT rule_targets_unique;
ALTER TABLE rule_targets DROP CONSTRAINT rule_targets_include_requires_fields;
ALTER TABLE rule_targets ADD CONSTRAINT rule_targets_include_requires_fields CHECK (
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
);

CREATE UNIQUE INDEX rule_targets_group_unique_idx
  ON rule_targets (rule_id, subject_id)
  WHERE subject_kind = 'group';

CREATE UNIQUE INDEX rule_targets_global_unique_idx
  ON rule_targets (rule_id, subject_kind)
  WHERE subject_kind IN ('all_devices', 'all_users');

-- machine_rule_sync_states: rename and reshape
ALTER TABLE machine_rule_sync_states RENAME TO machine_sync_states;

ALTER TABLE machine_sync_states RENAME COLUMN last_client_rules_hash TO rules_hash;
ALTER TABLE machine_sync_states RENAME COLUMN acknowledged_targets TO applied_targets;
ALTER TABLE machine_sync_states RENAME COLUMN pending_expected_rules_hash TO expected_rules_hash;
ALTER TABLE machine_sync_states RENAME COLUMN last_postflight_at TO last_rule_sync_success_at;

ALTER TABLE machine_sync_states ADD COLUMN pending_full_sync BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE machine_sync_states
SET pending_full_sync = pending_sync_type = 'clean_rules';
ALTER TABLE machine_sync_states DROP COLUMN request_clean_sync;
ALTER TABLE machine_sync_states DROP COLUMN pending_sync_type;

ALTER TABLE machine_sync_states ADD COLUMN client_mode TEXT NOT NULL DEFAULT 'unknown';
ALTER TABLE machine_sync_states ADD COLUMN binary_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN certificate_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN compiler_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN transitive_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN teamid_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN signingid_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN cdhash_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN rules_received INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN rules_processed INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN last_rule_sync_attempt_at TIMESTAMPTZ NULL;

ALTER TABLE machine_sync_states DROP CONSTRAINT machine_rule_sync_states_pending_payload_rule_count_check;
ALTER TABLE machine_sync_states ADD CONSTRAINT machine_sync_states_pending_payload_rule_count_check
  CHECK (pending_payload_rule_count >= 0);
ALTER TABLE machine_sync_states ADD CONSTRAINT machine_sync_states_client_mode_check
  CHECK (client_mode IN ('unknown', 'monitor', 'lockdown', 'standalone'));

-- machine_sync_states: replace hash-based sync tracking with payload/target snapshot columns
ALTER TABLE machine_sync_states DROP COLUMN expected_rules_hash;
ALTER TABLE machine_sync_states ADD COLUMN pending_payload JSONB NOT NULL DEFAULT '[]'::JSONB;
ALTER TABLE machine_sync_states ADD COLUMN desired_targets JSONB NOT NULL DEFAULT '[]'::JSONB;
ALTER TABLE machine_sync_states ADD COLUMN desired_binary_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN desired_certificate_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN desired_teamid_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN desired_signingid_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN desired_cdhash_rule_count INT NOT NULL DEFAULT 0;
ALTER TABLE machine_sync_states ADD COLUMN last_clean_sync_at TIMESTAMPTZ NULL;
ALTER TABLE machine_sync_states ADD COLUMN last_reported_counts_match_at TIMESTAMPTZ NULL;

-- rules: replace generated identifier_key column with direct unique constraint
ALTER TABLE rules DROP CONSTRAINT rules_type_identifier_key_unique;
ALTER TABLE rules DROP CONSTRAINT rules_identifier_key_check;
ALTER TABLE rules DROP COLUMN identifier_key;
ALTER TABLE rules ADD CONSTRAINT rules_identifier_check CHECK (btrim(identifier) <> '');
ALTER TABLE rules ADD CONSTRAINT rules_type_identifier_unique UNIQUE (rule_type, identifier);

-- machine_sync_states: add sync status function
CREATE OR REPLACE FUNCTION machine_rule_sync_status(
  pending_preflight_at TIMESTAMPTZ,
  desired_targets JSONB,
  applied_targets JSONB,
  desired_binary_rule_count INT,
  binary_rule_count INT,
  desired_certificate_rule_count INT,
  certificate_rule_count INT,
  desired_teamid_rule_count INT,
  teamid_rule_count INT,
  desired_signingid_rule_count INT,
  signingid_rule_count INT,
  desired_cdhash_rule_count INT,
  cdhash_rule_count INT,
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

-- file_access_events: decouple from executables table
DROP INDEX IF EXISTS file_access_events_executable_created_idx;
ALTER TABLE file_access_events DROP COLUMN IF EXISTS executable_id;

-- executables: remove source column (always 'event' after file access events decoupling)
DROP INDEX IF EXISTS executables_process_identity_idx;
DROP INDEX IF EXISTS executables_event_identity_idx;
CREATE UNIQUE INDEX executables_identity_idx ON executables (file_sha256, file_name);
ALTER TABLE executables DROP CONSTRAINT executables_source_check;
ALTER TABLE executables DROP COLUMN source;

-- executables: remove file_path column (always '', not part of unique index)
ALTER TABLE executables DROP COLUMN file_path;
