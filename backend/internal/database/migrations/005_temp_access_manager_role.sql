-- ============================================================
-- Add "TempAccessManager" role — can only grant/revoke Temporary
-- Access and view the Dashboard. Narrower than "AccessManager":
-- no access to Access Rules (add/edit/delete/reorder/clone) at all.
--
-- NOTE: this file is only auto-applied by Postgres' docker
-- entrypoint on a brand-new (empty) data volume. On an existing
-- deployment, run it manually against the database once:
--   psql -h <host> -U <user> -d <db> -f 005_temp_access_manager_role.sql
-- ============================================================

ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'TempAccessManager';
