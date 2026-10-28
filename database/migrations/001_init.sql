-- ============================================================
-- PGHBA Manager — Database Schema
-- PostgreSQL 16+
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE env_type       AS ENUM ('Development', 'UAT', 'Production');
CREATE TYPE instance_status AS ENUM ('Online', 'Offline', 'Unknown');
CREATE TYPE rule_type      AS ENUM ('local', 'host', 'hostssl', 'hostnossl', 'hostnogssenc');
CREATE TYPE auth_method    AS ENUM ('trust', 'reject', 'md5', 'scram-sha-256', 'password', 'gss', 'sspi', 'ident', 'peer', 'ldap', 'radius', 'cert', 'pam');
CREATE TYPE temp_duration  AS ENUM ('Minutes', 'Hours', 'Days', 'Weeks', 'Months');
CREATE TYPE user_role      AS ENUM ('SuperAdmin', 'DBA', 'Operator', 'Viewer');
CREATE TYPE user_status    AS ENUM ('Active', 'Inactive');
CREATE TYPE audit_event    AS ENUM (
    'Login', 'Logout', 'LoginFailed',
    'AddHost', 'EditHost', 'DeleteHost',
    'AddInstance', 'EditInstance', 'DeleteInstance',
    'AddRule', 'EditRule', 'DeleteRule',
    'EnableRule', 'DisableRule',
    'GrantTempAccess', 'RevokeTempAccess', 'ExpireTempAccess',
    'CreateBackup', 'RestoreBackup',
    'AddUser', 'EditUser', 'DeleteUser',
    'ReloadConf'
);

-- ============================================================
-- USERS
-- ============================================================

CREATE TABLE users (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    username    VARCHAR(64) NOT NULL UNIQUE,
    full_name   VARCHAR(128) NOT NULL,
    email       VARCHAR(256) NOT NULL UNIQUE,
    password    TEXT        NOT NULL,           -- bcrypt hash
    role        user_role   NOT NULL DEFAULT 'Viewer',
    status      user_status NOT NULL DEFAULT 'Active',
    theme       VARCHAR(16) NOT NULL DEFAULT 'dark',
    last_login  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email    ON users(email);
CREATE INDEX idx_users_status   ON users(status);

-- ============================================================
-- REFRESH TOKENS
-- ============================================================

CREATE TABLE refresh_tokens (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT        NOT NULL UNIQUE,    -- SHA-256 of token
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked     BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id    ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- ============================================================
-- DATABASE HOSTS
-- ============================================================

CREATE TABLE db_hosts (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(128) NOT NULL UNIQUE,
    ip_address      INET        NOT NULL,
    ssh_port        INTEGER     NOT NULL DEFAULT 22 CHECK (ssh_port BETWEEN 1 AND 65535),
    ssh_user        VARCHAR(64) NOT NULL DEFAULT 'postgres',
    ssh_key         TEXT,                       -- AES-256 encrypted private key
    pg_version      VARCHAR(8)  NOT NULL,       -- e.g. "16", "14"
    environment     env_type    NOT NULL DEFAULT 'Development',
    active          BOOLEAN     NOT NULL DEFAULT TRUE,
    notes           TEXT,
    created_by      UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_db_hosts_active      ON db_hosts(active);
CREATE INDEX idx_db_hosts_environment ON db_hosts(environment);
CREATE INDEX idx_db_hosts_ip          ON db_hosts(ip_address);

-- ============================================================
-- POSTGRESQL INSTANCES
-- ============================================================

CREATE TABLE pg_instances (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    host_id         UUID        NOT NULL REFERENCES db_hosts(id) ON DELETE CASCADE,
    instance_name   VARCHAR(128) NOT NULL,
    pg_version      VARCHAR(8)  NOT NULL,
    port            INTEGER     NOT NULL CHECK (port BETWEEN 1 AND 65535),
    data_directory  TEXT        NOT NULL,
    hba_path        TEXT        NOT NULL,       -- full path to pg_hba.conf
    status          instance_status NOT NULL DEFAULT 'Unknown',
    last_scan       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (host_id, port),
    UNIQUE (host_id, instance_name)
);

CREATE INDEX idx_pg_instances_host_id ON pg_instances(host_id);
CREATE INDEX idx_pg_instances_status  ON pg_instances(status);

-- ============================================================
-- HBA RULES
-- ============================================================

CREATE TABLE hba_rules (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id     UUID        NOT NULL REFERENCES pg_instances(id) ON DELETE CASCADE,
    rule_order      INTEGER     NOT NULL DEFAULT 0,    -- order in pg_hba.conf
    rule_type       rule_type   NOT NULL DEFAULT 'host',
    database        VARCHAR(256) NOT NULL DEFAULT 'all',
    db_user         VARCHAR(256) NOT NULL DEFAULT 'all',
    address         CIDR,                              -- NULL for local type
    auth_method     auth_method NOT NULL DEFAULT 'scram-sha-256',
    auth_options    TEXT,                              -- e.g. clientcert=verify-full
    comment         TEXT,
    enabled         BOOLEAN     NOT NULL DEFAULT TRUE,
    is_temporary    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_by      UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hba_rules_instance_id  ON hba_rules(instance_id);
CREATE INDEX idx_hba_rules_enabled      ON hba_rules(enabled);
CREATE INDEX idx_hba_rules_is_temporary ON hba_rules(is_temporary);
CREATE INDEX idx_hba_rules_address      ON hba_rules(address);
CREATE INDEX idx_hba_rules_order        ON hba_rules(instance_id, rule_order);

-- ============================================================
-- TEMPORARY ACCESS
-- ============================================================

CREATE TABLE temp_access (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id         UUID        NOT NULL REFERENCES hba_rules(id) ON DELETE CASCADE,
    instance_id     UUID        NOT NULL REFERENCES pg_instances(id) ON DELETE CASCADE,
    granted_by      UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    duration_value  INTEGER     NOT NULL CHECK (duration_value > 0),
    duration_unit   temp_duration NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    reason          TEXT,
    revoked         BOOLEAN     NOT NULL DEFAULT FALSE,
    revoked_by      UUID        REFERENCES users(id) ON DELETE SET NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_temp_access_rule_id    ON temp_access(rule_id);
CREATE INDEX idx_temp_access_expires_at ON temp_access(expires_at);
CREATE INDEX idx_temp_access_revoked    ON temp_access(revoked);
CREATE INDEX idx_temp_access_granted_by ON temp_access(granted_by);

-- ============================================================
-- BACKUPS
-- ============================================================

CREATE TABLE hba_backups (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id     UUID        NOT NULL REFERENCES pg_instances(id) ON DELETE CASCADE,
    filename        VARCHAR(256) NOT NULL,
    file_path       TEXT        NOT NULL,
    file_size       BIGINT      NOT NULL DEFAULT 0,  -- bytes
    content         TEXT        NOT NULL,             -- actual pg_hba.conf content
    trigger_event   VARCHAR(64) NOT NULL,             -- 'manual', 'before_edit', 'before_delete', 'before_restore'
    note            TEXT,
    created_by      UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hba_backups_instance_id ON hba_backups(instance_id);
CREATE INDEX idx_hba_backups_created_at  ON hba_backups(created_at);

-- ============================================================
-- AUDIT LOG
-- ============================================================

CREATE TABLE audit_logs (
    id              BIGSERIAL   PRIMARY KEY,
    user_id         UUID        REFERENCES users(id) ON DELETE SET NULL,
    username        VARCHAR(64) NOT NULL,             -- denormalized for history
    event           audit_event NOT NULL,
    target          TEXT,                             -- what was affected
    instance_id     UUID        REFERENCES pg_instances(id) ON DELETE SET NULL,
    instance_label  VARCHAR(256),                     -- denormalized label
    client_ip       INET,
    detail          JSONB,                            -- structured detail payload
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id    ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_event      ON audit_logs(event);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_instance   ON audit_logs(instance_id);

-- ============================================================
-- FUNCTIONS & TRIGGERS — updated_at auto-update
-- ============================================================

CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY['users','db_hosts','pg_instances','hba_rules'] LOOP
        EXECUTE format(
            'CREATE TRIGGER trg_%s_updated_at
             BEFORE UPDATE ON %s
             FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at()',
            t, t
        );
    END LOOP;
END;
$$;

-- ============================================================
-- SEED — default Super Admin
-- password: Admin@1234  (bcrypt cost 12, change on first login)
-- ============================================================

INSERT INTO users (username, full_name, email, password, role, status) VALUES (
    'admin',
    'Administrator',
    'admin@pghba.local',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/o9OiP3eUy',
    'SuperAdmin',
    'Active'
);
