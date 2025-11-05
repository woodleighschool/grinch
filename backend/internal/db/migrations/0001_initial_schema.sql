CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE scope_target AS ENUM ('group', 'user');
CREATE TYPE rule_action AS ENUM ('allow', 'block');
CREATE TYPE user_type AS ENUM ('local', 'cloud');
CREATE TYPE sync_type AS ENUM ('NORMAL', 'CLEAN', 'CLEAN_ALL');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id TEXT,
    principal_name TEXT NOT NULL,
    display_name TEXT,
    email TEXT,
    user_type user_type NOT NULL DEFAULT 'cloud',
    password_hash TEXT,
    is_protected_local BOOLEAN NOT NULL DEFAULT FALSE,
    synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_identifier_consistency CHECK (
        (user_type = 'cloud' AND external_id IS NOT NULL)
        OR (user_type = 'local' AND external_id IS NULL)
    ),
    CONSTRAINT users_external_id_unique UNIQUE (external_id)
);

CREATE UNIQUE INDEX idx_users_principal_name ON users (principal_name);
CREATE INDEX idx_users_user_type ON users (user_type);
CREATE INDEX idx_users_principal_password ON users (principal_name) WHERE password_hash IS NOT NULL;
CREATE INDEX idx_users_protected_local ON users (is_protected_local) WHERE is_protected_local = TRUE;

CREATE TABLE local_user_metadata (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    santa_agent_machine_id TEXT,
    last_converted_check TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE group_memberships (
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX idx_group_memberships_user ON group_memberships (user_id);

CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    rule_type TEXT NOT NULL,
    identifier TEXT NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_applications_identifier ON applications (identifier);

CREATE TABLE application_scopes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    target_type scope_target NOT NULL,
    target_id UUID NOT NULL,
    action rule_action NOT NULL DEFAULT 'allow',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT application_scopes_unique_target UNIQUE (application_id, target_type, target_id)
);

CREATE INDEX idx_application_scopes_application ON application_scopes (application_id);
CREATE INDEX idx_application_scopes_target ON application_scopes (target_type, target_id);

CREATE TABLE hosts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    hostname TEXT NOT NULL,
    serial_number TEXT,
    machine_id TEXT NOT NULL UNIQUE,
    primary_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    last_seen TIMESTAMPTZ,
    os_version TEXT,
    os_build TEXT,
    model_identifier TEXT,
    santa_version TEXT,
    client_mode TEXT NOT NULL DEFAULT 'MONITOR',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hosts_primary_user ON hosts (primary_user_id);
CREATE INDEX idx_hosts_last_seen ON hosts (last_seen DESC);

CREATE TABLE blocked_events (
    id BIGSERIAL PRIMARY KEY,
    host_id UUID REFERENCES hosts(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    application_id UUID REFERENCES applications(id) ON DELETE SET NULL,
    process_path TEXT NOT NULL,
    process_hash TEXT,
    signer TEXT,
    blocked_reason TEXT,
    event_payload JSONB,
    file_sha256 TEXT,
    file_name TEXT,
    executing_user TEXT,
    execution_time TIMESTAMPTZ,
    logged_in_users TEXT[],
    current_sessions TEXT[],
    decision TEXT,
    file_bundle_id TEXT,
    file_bundle_path TEXT,
    file_bundle_name TEXT,
    file_bundle_version TEXT,
    file_bundle_hash TEXT,
    file_bundle_binary_count INTEGER,
    pid INTEGER,
    ppid INTEGER,
    parent_name TEXT,
    quarantine_data_url TEXT,
    quarantine_referer_url TEXT,
    quarantine_timestamp TIMESTAMPTZ,
    quarantine_agent_bundle_id TEXT,
    signing_chain JSONB,
    signing_id TEXT,
    team_id TEXT,
    cdhash TEXT,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_blocked_events_occurred_at ON blocked_events (occurred_at DESC);
CREATE INDEX idx_blocked_events_file_sha256 ON blocked_events (file_sha256);
CREATE INDEX idx_blocked_events_team_id ON blocked_events (team_id);
CREATE INDEX idx_blocked_events_signing_id ON blocked_events (signing_id);
CREATE INDEX idx_blocked_events_decision ON blocked_events (decision);

CREATE TABLE entra_sync_state (
    id SMALLINT PRIMARY KEY DEFAULT 1,
    delta_link TEXT,
    last_synced TIMESTAMPTZ
);

CREATE TABLE santa_sync_state (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    machine_id TEXT NOT NULL UNIQUE,
    last_sync_time TIMESTAMPTZ,
    last_sync_type sync_type NOT NULL DEFAULT 'NORMAL',
    rules_delivered INTEGER NOT NULL DEFAULT 0,
    rules_processed INTEGER NOT NULL DEFAULT 0,
    rule_count_hash TEXT,
    binary_rule_hash TEXT,
    certificate_rule_hash TEXT,
    teamid_rule_hash TEXT,
    signingid_rule_hash TEXT,
    cdhash_rule_hash TEXT,
    transitive_rule_hash TEXT,
    compiler_rule_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_santa_sync_state_machine ON santa_sync_state (machine_id);
CREATE INDEX idx_santa_sync_state_last_sync ON santa_sync_state (last_sync_time DESC NULLS LAST);

CREATE TABLE santa_rule_cursors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    machine_id TEXT NOT NULL,
    cursor_token TEXT NOT NULL UNIQUE,
    last_rule_id UUID,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT santa_rule_cursors_machine_fk
        FOREIGN KEY (machine_id) REFERENCES hosts(machine_id) ON DELETE CASCADE
);

CREATE INDEX idx_santa_rule_cursors_machine ON santa_rule_cursors (machine_id);
CREATE INDEX idx_santa_rule_cursors_expires ON santa_rule_cursors (expires_at);

CREATE TABLE santa_rule_deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    machine_id TEXT NOT NULL,
    rule_hash TEXT NOT NULL,
    rule_type TEXT NOT NULL,
    rule_identifier TEXT NOT NULL,
    policy TEXT NOT NULL,
    delivered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT santa_rule_deliveries_machine_fk
        FOREIGN KEY (machine_id) REFERENCES hosts(machine_id) ON DELETE CASCADE,
    CONSTRAINT santa_rule_deliveries_unique UNIQUE (machine_id, rule_hash)
);

CREATE INDEX idx_santa_rule_deliveries_machine ON santa_rule_deliveries (machine_id);

CREATE TABLE role_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO role_groups (name, description)
VALUES ('admin', 'Administrative access to the system')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE user_role_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_group_id UUID NOT NULL REFERENCES role_groups(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    CONSTRAINT user_role_assignments_unique UNIQUE (user_id, role_group_id)
);

CREATE INDEX idx_user_role_assignments_user ON user_role_assignments (user_id);
CREATE INDEX idx_user_role_assignments_role ON user_role_assignments (role_group_id);

CREATE TABLE application_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key TEXT NOT NULL UNIQUE,
    value JSONB NOT NULL DEFAULT 'null'::jsonb,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_application_settings_key ON application_settings (key);

INSERT INTO application_settings (key, value, description) VALUES
    ('saml.enabled', 'false'::jsonb, 'Whether SAML authentication is enabled'),
    ('saml.metadata_url', 'null'::jsonb, 'SAML Identity Provider metadata URL'),
    ('saml.entity_id', 'null'::jsonb, 'SAML Service Provider entity ID'),
    ('saml.acs_url', 'null'::jsonb, 'SAML Assertion Consumer Service URL'),
    ('saml.sp_private_key', 'null'::jsonb, 'SAML Service Provider private key (PEM format)'),
    ('saml.sp_certificate', 'null'::jsonb, 'SAML Service Provider certificate (PEM format)'),
    ('saml.name_id_format', 'null'::jsonb, 'SAML NameID format'),
    ('saml.object_id_attribute', '"http://schemas.microsoft.com/identity/claims/objectidentifier"'::jsonb, 'SAML attribute for object ID'),
    ('saml.upn_attribute', '"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/upn"'::jsonb, 'SAML attribute for UPN'),
    ('saml.email_attribute', '"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"'::jsonb, 'SAML attribute for email'),
    ('saml.display_name_attribute', '"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"'::jsonb, 'SAML attribute for display name')
ON CONFLICT (key) DO NOTHING;

CREATE OR REPLACE FUNCTION cleanup_orphaned_local_users()
RETURNS void AS $$
BEGIN
    DELETE FROM users
    WHERE user_type = 'local'
      AND id IN (
          SELECT u.id
          FROM users u
          JOIN local_user_metadata lum ON u.id = lum.user_id
          WHERE lum.first_seen_at < NOW() - INTERVAL '90 days'
            AND NOT EXISTS (SELECT 1 FROM blocked_events be WHERE be.user_id = u.id)
            AND NOT EXISTS (SELECT 1 FROM hosts h WHERE h.primary_user_id = u.id)
      );
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_expired_santa_cursors()
RETURNS void AS $$
BEGIN
    DELETE FROM santa_rule_cursors WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;
