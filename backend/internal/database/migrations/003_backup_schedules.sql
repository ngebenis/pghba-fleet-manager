-- ============================================================
-- Backup Schedules — multiple, independently editable auto-backup
-- schedules (name, cron, storage target, scope), managed from the
-- Settings page instead of a single global AUTO_BACKUP_* env policy.
--
-- NOTE: this file is only auto-applied by Postgres' docker
-- entrypoint on a brand-new (empty) data volume. On an existing
-- deployment, run it manually against the database once:
--   psql -h <host> -U <user> -d <db> -f 003_backup_schedules.sql
-- ============================================================

CREATE TABLE IF NOT EXISTS backup_schedules (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             VARCHAR(128) NOT NULL,
    cron_expression  VARCHAR(64)  NOT NULL,
    target           VARCHAR(16)  NOT NULL DEFAULT 'local',  -- local | s3 | both
    scope_type       VARCHAR(16)  NOT NULL DEFAULT 'all',    -- all | host | instance
    scope_id         UUID,
    enabled          BOOLEAN NOT NULL DEFAULT true,
    created_by_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    last_run_at      TIMESTAMPTZ,
    last_run_status  VARCHAR(16),
    last_run_message VARCHAR(256),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_backup_schedules_enabled ON backup_schedules(enabled);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_backup_schedules_updated_at') THEN
        CREATE TRIGGER trg_backup_schedules_updated_at
            BEFORE UPDATE ON backup_schedules
            FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
    END IF;
END;
$$;

-- New audit event types for schedule CRUD. Each ADD VALUE commits on its
-- own statement, so it's safe to use in a later statement in this same
-- migration run (psql -f autocommits per statement by default).
ALTER TYPE audit_event ADD VALUE IF NOT EXISTS 'AddBackupSchedule';
ALTER TYPE audit_event ADD VALUE IF NOT EXISTS 'EditBackupSchedule';
ALTER TYPE audit_event ADD VALUE IF NOT EXISTS 'DeleteBackupSchedule';
