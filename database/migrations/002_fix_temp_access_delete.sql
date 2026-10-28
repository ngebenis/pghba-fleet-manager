-- ============================================================
-- Fix: distinguish rules created brand-new by "Grant Temp Access"
-- from pre-existing rules later marked temporary via "Set Temporary".
--
-- Without this flag, the scheduler used "rule_order = 0" as a proxy
-- for "safe to delete on expiry" — but every rule created through the
-- UI defaults to rule_order 0, so pre-existing permanent rules that
-- were ever set temporary could be permanently deleted (not just
-- disabled) once TEMP_ACCESS_ACTION=delete and they expired.
--
-- NOTE: this file is only auto-applied by Postgres' docker
-- entrypoint on a brand-new (empty) data volume. On an existing
-- deployment, run it manually against the database once:
--   psql -h <host> -U <user> -d <db> -f 002_fix_temp_access_delete.sql
-- ============================================================

ALTER TABLE hba_rules
    ADD COLUMN IF NOT EXISTS created_for_temp_access BOOLEAN NOT NULL DEFAULT false;
