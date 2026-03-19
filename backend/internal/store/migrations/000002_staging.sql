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

-- machine_rule_sync_states: rename clean sync 'clean_rules' to 'clean'
UPDATE machine_rule_sync_states
SET pending_sync_type = 'clean'
WHERE pending_sync_type = 'clean_rules';

ALTER TABLE machine_rule_sync_states DROP CONSTRAINT machine_rule_sync_states_pending_sync_type_check;
ALTER TABLE machine_rule_sync_states ADD CONSTRAINT machine_rule_sync_states_pending_sync_type_check
  CHECK (pending_sync_type IN ('', 'normal', 'clean'));

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

UPDATE machine_rule_sync_states
SET pending_sync_type = 'clean_rules'
WHERE pending_sync_type = 'clean';

ALTER TABLE machine_rule_sync_states DROP CONSTRAINT machine_rule_sync_states_pending_sync_type_check;
ALTER TABLE machine_rule_sync_states ADD CONSTRAINT machine_rule_sync_states_pending_sync_type_check
  CHECK (pending_sync_type IN ('', 'normal', 'clean_rules'));

ALTER TABLE rule_targets ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE rule_targets ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE rule_targets ALTER COLUMN subject_id SET NOT NULL;
