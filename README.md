# PGHBA Manager

PostgreSQL `pg_hba.conf` Centralized Management System

Manage `pg_hba.conf` across hundreds of PostgreSQL instances from a single web UI.

---

## Quick Start (Docker)

```bash
# 1. Clone and configure
git clone https://github.com/your-org/pghba-manager.git
cd pghba-manager

# 2. Create .env from template
cp backend/.env.example .env

# 3. IMPORTANT — Generate real secrets before running
openssl rand -hex 32   # → ENCRYPTION_KEY
openssl rand -base64 48 # → JWT_SECRET

# 4. Edit .env with your real values, then:
docker compose up -d

# App available at: http://localhost:8080
# Default login: admin / Admin@1234  (change on first login!)
```

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

## Project Structure

```
pghba-manager/
├── backend/
│   ├── cmd/server/main.go          # Entry point
│   └── internal/
│       ├── config/                 # Config loader (.env / env vars)
│       ├── database/               # GORM connection + AutoMigrate
│       ├── models/                 # All GORM models
│       ├── middleware/             # JWT auth, RBAC
│       ├── handlers/               # Gin HTTP handlers
│       │   ├── auth.go             # Login / logout / refresh
│       │   ├── hosts.go            # Database host CRUD
│       │   ├── rules.go            # pg_hba rule CRUD + enable/disable
│       │   ├── temp_access.go      # Temporary access grant/revoke
│       │   ├── backup.go           # Backup create/restore/download
│       │   ├── users.go            # User CRUD (SuperAdmin only)
│       │   ├── audit.go            # Audit log list
│       │   └── dashboard.go        # Stats endpoint
│       ├── services/
│       │   ├── hba.go              # pg_hba content rendering + backup + expiry
│       │   ├── audit.go            # Audit log writer
│       │   └── crypto.go           # AES-256-GCM for SSH keys / secrets
│       ├── scheduler/              # Cron — expire temp rules every N minutes
│       └── server/router.go        # Gin router, CORS, all routes
│
├── frontend/
│   └── src/
│       ├── api/index.ts            # Axios API client (all endpoints)
│       ├── types/index.ts          # TypeScript interfaces (mirrors Go models)
│       ├── stores/auth.ts          # Pinia auth store
│       ├── router/index.ts         # Vue Router with auth guards
│       ├── components/
│       │   └── layout/AppLayout.vue   # Sidebar + topbar shell
│       └── views/
│           ├── LoginView.vue
│           ├── DashboardView.vue
│           ├── RulesView.vue          # Full CRUD + enable/disable toggle
│           ├── TempAccessView.vue     # Grant / revoke + countdown timer
│           ├── HostsView.vue
│           ├── InstancesView.vue
│           ├── BackupView.vue
│           ├── UsersView.vue
│           ├── AuditView.vue
│           └── SettingsView.vue
│
├── database/migrations/
│   └── 001_init.sql                # Full schema with enums, indexes, triggers
│
├── Dockerfile                      # Multi-stage: Vue build + Go build → alpine
├── docker-compose.yml              # Production
└── docker-compose.dev.yml          # Development (DB only)
```

---

## API Reference

```
POST   /api/auth/login
POST   /api/auth/logout
POST   /api/auth/refresh
GET    /api/auth/me

GET    /api/dashboard/stats

GET    /api/hosts
POST   /api/hosts
PUT    /api/hosts/:id
DELETE /api/hosts/:id

GET    /api/instances

GET    /api/rules
POST   /api/rules
PUT    /api/rules/:id
DELETE /api/rules/:id
POST   /api/rules/:id/enable
POST   /api/rules/:id/disable

GET    /api/temporary-access
POST   /api/temporary-access
POST   /api/temporary-access/:id/revoke

GET    /api/backups
POST   /api/backups
POST   /api/backups/restore
GET    /api/backups/:id/download

GET    /api/users
POST   /api/users
PUT    /api/users/:id
POST   /api/users/:id/reset-password
DELETE /api/users/:id

GET    /api/audit
```

---

## RBAC Permissions

| Feature            | SuperAdmin | DBA | Operator | Viewer |
|--------------------|:----------:|:---:|:--------:|:------:|
| Manage Hosts       | ✓          | ✓   |          |        |
| Manage Instances   | ✓          | ✓   |          |        |
| Add/Edit/Delete Rule | ✓        | ✓   |          |        |
| Enable/Disable Rule | ✓         | ✓   | ✓        |        |
| Grant Temp Access  | ✓          | ✓   | ✓        |        |
| View Rules         | ✓          | ✓   | ✓        | ✓      |
| Backup & Restore   | ✓          | ✓   |          |        |
| User Management    | ✓          |     |          |        |
| View Audit Log     | ✓          | ✓   |          |        |

---

## Security Checklist (Before Production)

- [ ] Generate fresh `JWT_SECRET` (min 48 chars)
- [ ] Generate fresh `ENCRYPTION_KEY` (`openssl rand -hex 32`)
- [ ] Change default admin password on first login
- [ ] Set `DB_PASSWORD` to a strong password
- [ ] Enable TLS (`DB_SSL_MODE=require`)
- [ ] Put behind NGINX/Traefik with HTTPS
- [ ] Restrict port 5432 to internal network only
