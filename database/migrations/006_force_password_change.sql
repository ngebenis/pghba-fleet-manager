-- ============================================================
-- Force password change for the published default Super Admin
--
-- 001_init.sql seeds a Super Admin with a fixed bcrypt hash for
-- "Admin@1234" — both the username and password are visible to anyone who
-- reads the (now public) migration history, with nothing in the app that
-- ever forced it to be changed. Add a flag so a deployment that never
-- rotated it gets blocked into changing it on next login, while anyone
-- who already changed it (their stored hash no longer matches the seeded
-- one) is left alone.
-- ============================================================

ALTER TABLE users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT false;

UPDATE users
SET must_change_password = true
WHERE username = 'admin'
  AND password = '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/o9OiP3eUy';
