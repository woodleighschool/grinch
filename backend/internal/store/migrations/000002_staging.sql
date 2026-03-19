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

-- machine_rule_sync_states: rename and reshape machine_sync_states
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

-- +goose Down
DROP INDEX IF EXISTS rule_targets_global_unique_idx;
DROP INDEX IF EXISTS rule_targets_group_unique_idx;

ALTER TABLE rule_targets DROP CONSTRAINT rule_targets_include_requires_fields;
ALTER TABLE rule_targets ADD CONSTRAINT rule_targets_include_requires_fields CHECK (
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
);
ALTER TABLE rule_targets ADD CONSTRAINT rule_targets_unique UNIQUE (rule_id, subject_kind, subject_id);

ALTER TABLE rule_targets DROP CONSTRAINT rule_targets_subject_id_check;

ALTER TABLE rule_targets DROP CONSTRAINT rule_targets_subject_kind_check;
ALTER TABLE rule_targets ADD CONSTRAINT rule_targets_subject_kind_check
  CHECK (subject_kind IN ('group'));

ALTER TABLE machine_sync_states DROP CONSTRAINT machine_sync_states_client_mode_check;
ALTER TABLE machine_sync_states DROP CONSTRAINT machine_sync_states_pending_payload_rule_count_check;
ALTER TABLE machine_sync_states ADD CONSTRAINT machine_rule_sync_states_pending_payload_rule_count_check
  CHECK (pending_payload_rule_count >= 0);

ALTER TABLE machine_sync_states DROP COLUMN last_rule_sync_attempt_at;
ALTER TABLE machine_sync_states DROP COLUMN rules_processed;
ALTER TABLE machine_sync_states DROP COLUMN rules_received;
ALTER TABLE machine_sync_states DROP COLUMN cdhash_rule_count;
ALTER TABLE machine_sync_states DROP COLUMN signingid_rule_count;
ALTER TABLE machine_sync_states DROP COLUMN teamid_rule_count;
ALTER TABLE machine_sync_states DROP COLUMN transitive_rule_count;
ALTER TABLE machine_sync_states DROP COLUMN compiler_rule_count;
ALTER TABLE machine_sync_states DROP COLUMN certificate_rule_count;
ALTER TABLE machine_sync_states DROP COLUMN binary_rule_count;
ALTER TABLE machine_sync_states DROP COLUMN client_mode;

ALTER TABLE machine_sync_states ADD COLUMN pending_sync_type TEXT NOT NULL DEFAULT '';
UPDATE machine_sync_states
SET pending_sync_type = CASE
  WHEN expected_rules_hash = '' THEN ''
  WHEN pending_full_sync THEN 'clean_rules'
  ELSE 'normal'
END;
ALTER TABLE machine_sync_states DROP COLUMN pending_full_sync;
ALTER TABLE machine_sync_states ADD CONSTRAINT machine_rule_sync_states_pending_sync_type_check
  CHECK (pending_sync_type IN ('', 'normal', 'clean_rules'));

ALTER TABLE machine_sync_states ADD COLUMN request_clean_sync BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE machine_sync_states RENAME COLUMN last_rule_sync_success_at TO last_postflight_at;
ALTER TABLE machine_sync_states RENAME COLUMN expected_rules_hash TO pending_expected_rules_hash;
ALTER TABLE machine_sync_states RENAME COLUMN applied_targets TO acknowledged_targets;
ALTER TABLE machine_sync_states RENAME COLUMN rules_hash TO last_client_rules_hash;
ALTER TABLE machine_sync_states RENAME TO machine_rule_sync_states;

ALTER TABLE rule_targets ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE rule_targets ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE rule_targets ALTER COLUMN subject_id SET NOT NULL;
