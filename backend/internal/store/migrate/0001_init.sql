-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- updated_at trigger
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at := NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-----------------------------------------------------------------------
-- Core entities
-----------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  upn          TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  object_id    TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS groups (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  display_name TEXT NOT NULL,
  description  TEXT,
  object_id    TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS group_members (
  group_id   UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  user_id    UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (group_id, user_id)
);

-----------------------------------------------------------------------
-- Machines
-----------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS machines (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  serial                TEXT NOT NULL UNIQUE,
  machine_identifier    TEXT NOT NULL UNIQUE,
  hostname              TEXT NOT NULL,
  user_id               UUID REFERENCES users(id),
  client_version        TEXT,
  client_mode           TEXT NOT NULL DEFAULT 'monitor',
  primary_user          TEXT,
  clean_sync_requested  BOOLEAN NOT NULL DEFAULT FALSE,
  sync_cursor           TEXT,
  rule_cursor           TEXT,
  last_seen             TIMESTAMPTZ,
  last_preflight_at     TIMESTAMPTZ,
  last_postflight_at    TIMESTAMPTZ,
  last_rules_received   INTEGER,
  last_rules_processed  INTEGER,
  last_preflight_payload JSONB,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (last_rules_received  IS NULL OR last_rules_received  >= 0),
  CHECK (last_rules_processed IS NULL OR last_rules_processed >= 0)
);

-----------------------------------------------------------------------
-- Rules and assignments
-----------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS rules (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT NOT NULL,
  type       TEXT NOT NULL,
  target     TEXT NOT NULL,
  scope      TEXT NOT NULL,
  enabled    BOOLEAN NOT NULL DEFAULT TRUE,
  metadata   JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rule_scopes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  rule_id     UUID NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
  target_type TEXT NOT NULL CHECK (target_type IN ('group', 'user')),
  target_id   UUID NOT NULL,
  action      TEXT NOT NULL CHECK (action IN ('allow', 'block')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rule_assignments (
  id          BIGSERIAL PRIMARY KEY,
  rule_id     UUID NOT NULL REFERENCES rules(id)       ON DELETE CASCADE,
  scope_id    UUID NOT NULL REFERENCES rule_scopes(id) ON DELETE CASCADE,
  target_type TEXT NOT NULL CHECK (target_type IN ('group', 'user')),
  action      TEXT NOT NULL CHECK (action IN ('allow', 'block')),
  user_id     UUID REFERENCES users(id)   ON DELETE CASCADE,
  group_id    UUID REFERENCES groups(id)  ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-----------------------------------------------------------------------
-- Files & events
-----------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS files (
  sha256        TEXT PRIMARY KEY,
  name          TEXT NOT NULL,
  signing_id    TEXT,
  cdhash        TEXT,
  signing_chain JSONB,
  entitlements  JSONB,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS events (
  id          BIGSERIAL PRIMARY KEY,
  machine_id  UUID NOT NULL REFERENCES machines(id) ON DELETE CASCADE,
  user_id     UUID REFERENCES users(id) ON DELETE SET NULL,
  kind        TEXT NOT NULL,
  payload     JSONB NOT NULL,
  file_sha256 TEXT REFERENCES files(sha256),
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cursors (
  machine_id       UUID PRIMARY KEY REFERENCES machines(id) ON DELETE CASCADE,
  last_rule_cursor TEXT,
  last_event_cursor TEXT,
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-----------------------------------------------------------------------
-- Triggers (updated_at)
-----------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_users_updated_at') THEN
    CREATE TRIGGER trg_users_updated_at
      BEFORE UPDATE ON users
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_groups_updated_at') THEN
    CREATE TRIGGER trg_groups_updated_at
      BEFORE UPDATE ON groups
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_machines_updated_at') THEN
    CREATE TRIGGER trg_machines_updated_at
      BEFORE UPDATE ON machines
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_rules_updated_at') THEN
    CREATE TRIGGER trg_rules_updated_at
      BEFORE UPDATE ON rules
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_cursors_updated_at') THEN
    CREATE TRIGGER trg_cursors_updated_at
      BEFORE UPDATE ON cursors
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
END$$;

-----------------------------------------------------------------------
-- Indexes
-----------------------------------------------------------------------
-- Users
CREATE INDEX IF NOT EXISTS idx_users_upn_lower
  ON users (LOWER(upn));
CREATE INDEX IF NOT EXISTS idx_users_search_vector
  ON users USING GIN (to_tsvector('simple', coalesce(display_name, '') || ' ' || coalesce(upn, '')));

-- Groups
CREATE INDEX IF NOT EXISTS idx_groups_display_name_lower
  ON groups (LOWER(display_name));
CREATE INDEX IF NOT EXISTS idx_groups_search_vector
  ON groups USING GIN (to_tsvector('simple', coalesce(display_name, '') || ' ' || coalesce(description, '')));

-- Group members
CREATE INDEX IF NOT EXISTS idx_group_members_user
  ON group_members (user_id);

-- Machines
CREATE INDEX IF NOT EXISTS idx_machines_serial
  ON machines (serial);
CREATE INDEX IF NOT EXISTS idx_machines_search_vector
  ON machines USING GIN (to_tsvector('simple',
    coalesce(serial, '') || ' ' || coalesce(hostname, '') || ' ' || coalesce(machine_identifier, '')
  ));

-- Rules
CREATE INDEX IF NOT EXISTS idx_rules_scope
  ON rules (scope);
CREATE INDEX IF NOT EXISTS idx_rules_type_lower
  ON rules (LOWER(type));
CREATE UNIQUE INDEX IF NOT EXISTS uniq_rules_identifier_lower
  ON rules (LOWER(target));
CREATE INDEX IF NOT EXISTS idx_rules_search_vector
  ON rules USING GIN (to_tsvector('simple',
    coalesce(name, '') || ' ' || coalesce(type, '') || ' ' || coalesce(target, '') || ' ' ||
    coalesce(metadata ->> 'description', '') || ' ' ||
    coalesce(metadata ->> 'block_message', '')
  ));

-- Rule scopes / assignments
CREATE INDEX IF NOT EXISTS idx_rule_scopes_rule_id
  ON rule_scopes (rule_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_rule_scopes_per_target
  ON rule_scopes (rule_id, target_type, target_id);

CREATE INDEX IF NOT EXISTS idx_rule_assignments_rule_scope
  ON rule_assignments (rule_id, scope_id);
CREATE INDEX IF NOT EXISTS idx_rule_assignments_user
  ON rule_assignments (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_rule_assignments_per_target
  ON rule_assignments (
    rule_id,
    COALESCE(user_id,  '00000000-0000-0000-0000-000000000000'::UUID),
    COALESCE(group_id, '00000000-0000-0000-0000-000000000000'::UUID),
    action
  );

-- Events
CREATE INDEX IF NOT EXISTS idx_events_machine_time
  ON events (machine_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_file_sha256
  ON events(file_sha256);
