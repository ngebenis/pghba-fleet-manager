-- ============================================================
-- hba_rules.address: CIDR -> VARCHAR(64)
--
-- The column was created as CIDR (NULL for local-type rules, per the
-- 001_init.sql comment), but the Go model has always declared it as a
-- plain `size:64` string with no nullable wrapper type. Importing or
-- creating a "local" rule — whose Address is the Go zero value "" — made
-- GORM insert the literal empty string into this CIDR column, which
-- Postgres rejects with "invalid input syntax for type inet: \"\"". That
-- INSERT failure was silently swallowed by importHBAForInstance, so
-- every "local" line in a server's real pg_hba.conf (the three default
-- Unix-socket rules Postgres ships with) quietly never made it into the
-- database on import, no matter how many times you re-imported.
--
-- Nothing in this codebase relies on Postgres' native CIDR semantics
-- (no <<=, no inet_same_family — the CIDR-aware comparisons in
-- conflicts.go are done in Go), so there's no capability lost by
-- storing it as plain text instead.
-- ============================================================

ALTER TABLE hba_rules ALTER COLUMN address TYPE VARCHAR(64) USING COALESCE(address::text, '');
ALTER TABLE hba_rules ALTER COLUMN address SET DEFAULT '';
ALTER TABLE hba_rules ALTER COLUMN address SET NOT NULL;
