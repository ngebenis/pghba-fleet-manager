# PGHBA Fleet Manager

🇬🇧 **English** | [🇮🇩 Bahasa Indonesia](README.md)

PostgreSQL `pg_hba.conf` centralized management system — manage access
rules across a whole fleet of PostgreSQL instances from a single web UI,
instead of hand-editing `pg_hba.conf` on every host.

This repository is a hardened, public-ready snapshot of the original
PGHBA Manager project: default-password enforcement, startup secret
validation, SSH host key pinning, login rate limiting, and a CI pipeline
that scans dependencies on every push. See [Security](#security) below.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Buy me a cup of coffee.

---

## Screenshots

### Dashboard

![Dashboard](docs/screenshots/dashboard.png)

> The screenshots below (Database Hosts, PG Instances, Access Rules, Temporary
> Access, Backup & Restore, User Management, Audit Log, Settings) are
> **illustrative mockups** built from this repo's actual frontend
> components and color scheme, populated with dummy/sample data — not a
> live deployment. Real screenshots will replace them as they become
> available.

<details>
<summary>Database Hosts</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>PG Instances</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>Access Rules</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>Temporary Access</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>Backup & Restore</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>User Management</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>Audit Log</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>Settings</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## Quick Start (Docker)

```bash
# 1. Clone and configure
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. Create .env from template (used by docker-compose.yml)
cp backend/.env.example .env

# 3. Generate real secrets — the app refuses to start without them
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. Edit .env with your real values, then:
docker compose up -d

# App available at: http://localhost:8080
```

**First login**: `admin` / `Admin@1234`. This account is force-flagged to
change its password on first login — the server blocks every other
action until it's changed, so there's no "forgot to change the default
password" window.

---

## Development Setup

```bash
# Start only the database
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # edit as needed
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (proxies /api to :8080)
```

---

## Security

- **No insecure defaults can run.** `main.go` calls `config.Validate()`
  at startup, which refuses to start if `JWT_SECRET` or `ENCRYPTION_KEY`
  is empty or still the published placeholder — a misconfigured
  deployment fails loudly instead of running with a forgeable JWT secret
  or a known AES key.
- **SSH host keys are pinned**, not ignored. `services.ConnectToHost`
  trusts a managed host's SSH key on first connection and persists its
  fingerprint (`db_hosts.ssh_host_key_fingerprint`); any later connection
  with a different key is refused instead of silently trusted. Clearing
  a fingerprint after a legitimate host rebuild goes through a dedicated,
  audited endpoint (`POST /api/hosts/:id/reset-host-key`), never the
  generic host-edit form.
- **Login is rate-limited**: 10 failed attempts from one IP within 15
  minutes locks that IP out for 15 minutes.
- **Deactivating or demoting a user takes effect immediately** — every
  authenticated request re-checks status/role against the database
  instead of trusting what a JWT claimed up to 8 hours ago.
- **CI scans for known-vulnerable dependencies** on every push/PR and
  weekly (`.github/workflows/security.yml`): `govulncheck` for the Go
  backend, `npm audit` (production dependencies) for the frontend.
- Backup export, CORS, and HTTP response headers are hardened — see
  `backend/internal/handlers/export.go` (path-segment sanitization),
  `backend/internal/server/router.go` (CORS defaults to no cross-origin
  access unless `CORS_ALLOWED_ORIGINS` is set), and
  `backend/internal/middleware/security_headers.go`.
- **Rotating secrets after a leak**: `backend/cmd/reencrypt-ssh-keys` and
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) walk through
  rotating `JWT_SECRET`/`ENCRYPTION_KEY` and re-encrypting every stored
  SSH key without data loss.

---

## Project Structure

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Entry point
│   │   └── reencrypt-ssh-keys/         # Secret-rotation tool (see Security)
│   └── internal/
│       ├── config/                     # Config loader + startup secret validation
│       ├── database/                   # GORM connection
│       ├── models/                     # All GORM models
│       ├── middleware/                 # JWT auth, RBAC, rate limiting, security headers
│       ├── handlers/                   # Gin HTTP handlers
│       │   ├── auth.go                 # Login / logout / refresh / change-password
│       │   ├── hosts.go                # Host CRUD, SSH validate, deploy-key, reset-host-key
│       │   ├── instances.go            # Instance discovery/scan
│       │   ├── rules.go                # pg_hba rule CRUD + enable/disable/reorder/clone
│       │   ├── temp_access.go          # Temporary access grant/revoke
│       │   ├── backup.go, backup_schedules.go   # Backup create/restore/download + scheduling
│       │   ├── export.go               # Config/backup export (local + S3)
│       │   ├── healthcheck.go          # Host health checks + notifications
│       │   ├── users.go                # User CRUD (SuperAdmin only)
│       │   ├── audit.go                # Audit log list + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH client, host key pinning (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # pg_hba content rendering + atomic write + expiry
│       │   ├── crypto.go               # AES-256-GCM for SSH keys
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — expire temp rules every N minutes
│       └── server/router.go            # Gin router, CORS, all routes
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API client
│       ├── types/index.ts              # TypeScript interfaces (mirrors Go models)
│       ├── stores/auth.ts              # Pinia auth store
│       ├── router/index.ts             # Vue Router with auth + force-password-change guards
│       ├── components/layout/AppLayout.vue
│       └── views/                      # One view per feature area
│
├── database/migrations/
│   ├── 001_init.sql                    # Full schema with enums, indexes, triggers
│   └── 002–007_*.sql                   # RBAC roles, backup schedules, password/host-key hardening
│
├── docs/rotate-secrets.md              # Secret rotation runbook
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-stage: Vue build + Go build → alpine
├── docker-compose.yml                  # Production
└── docker-compose.dev.yml              # Development (DB only)
```

---

## API Reference

The full route list lives in `backend/internal/server/router.go`. Core
endpoints:

```
POST   /api/auth/login
POST   /api/auth/logout
POST   /api/auth/refresh
POST   /api/auth/change-password
GET    /api/auth/me

GET    /api/dashboard/stats

GET    /api/hosts
POST   /api/hosts
PUT    /api/hosts/:id
DELETE /api/hosts/:id
POST   /api/hosts/validate              # SSH pre-flight check before saving
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # Clear pinned SSH host key fingerprint

GET    /api/instances
POST   /api/instances/discover
POST   /api/instances/:id/scan

GET    /api/rules
POST   /api/rules
PUT    /api/rules/:id
DELETE /api/rules/:id
POST   /api/rules/:id/enable
POST   /api/rules/:id/disable
POST   /api/rules/reorder
POST   /api/rules/clone

GET    /api/temporary-access
POST   /api/temporary-access
POST   /api/temporary-access/:id/revoke

GET    /api/backups
POST   /api/backups
POST   /api/backups/:id/restore
GET    /api/backups/:id/download
GET    /api/backup-schedules

GET    /api/users
POST   /api/users
PUT    /api/users/:id
POST   /api/users/:id/reset-password
DELETE /api/users/:id

GET    /api/audit
GET    /api/audit/diff/:id
```

---

## RBAC Permissions

| Feature               | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Manage Hosts           | ✓          | ✓   |          |        |               |                    |
| Manage Instances       | ✓          | ✓   |          |        |               |                    |
| Add/Edit/Delete Rule   | ✓          | ✓   |          |        | ✓             |                    |
| Enable/Disable Rule    | ✓          | ✓   | ✓        |        | ✓             |                    |
| View Rules             | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Grant/Revoke Temp Access | ✓        | ✓   | ✓        |        | ✓             | ✓                  |
| Backup & Restore       | ✓          | ✓   |          |        |               |                    |
| User Management        | ✓          |     |          |        |               |                    |
| View Audit Log         | ✓          | ✓   |          |        |               |                    |

`AccessManager` and `TempAccessManager` are scoped roles for teams that
should only manage access rules / temporary access without full DBA
privileges over hosts, instances, backups, or users.

---

## Production Checklist

- [ ] Set real `JWT_SECRET`/`ENCRYPTION_KEY` — the app won't start otherwise
- [ ] Log in once and change the default admin password (enforced on first login)
- [ ] Set a strong `DB_PASSWORD`
- [ ] Enable TLS (`DB_SSL_MODE=require`)
- [ ] Put the app behind NGINX/Traefik with HTTPS
- [ ] Restrict Postgres's port to the internal network only
- [ ] Set `CORS_ALLOWED_ORIGINS` only if the frontend is deployed on a different origin than the API
