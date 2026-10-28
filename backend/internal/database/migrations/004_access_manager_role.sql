-- ============================================================
-- Add "AccessManager" role — full CRUD on Access Rules and
-- Temporary Access (including write-back to the server) and
-- Dashboard view, with no access to Hosts, Instances, Backups,
-- Users, Settings, or Audit Log.
--
-- NOTE: this file is only auto-applied by Postgres' docker
-- entrypoint on a brand-new (empty) data volume. On an existing
-- deployment, run it manually against the database once:
--   psql -h <host> -U <user> -d <db> -f 004_access_manager_role.sql
-- ============================================================

ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'AccessManager';
