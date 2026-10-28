# PGHBA Fleet Manager

🇮🇩 [ID](README.md) · 🇬🇧 [EN](README.en.md) · 🇨🇳 [ZH](README.zh.md) · 🇪🇸 [ES](README.es.md) · 🇸🇦 [AR](README.ar.md) · 🇫🇷 [FR](README.fr.md) · 🇵🇹 [PT](README.pt.md) · 🇧🇩 [BN](README.bn.md) · 🇷🇺 [RU](README.ru.md) · 🇮🇳 [HI](README.hi.md) · 🇹🇭 [TH](README.th.md) · 🇰🇭 [KM](README.km.md) · 🇵🇭 **TL** · 🇻🇳 [VI](README.vi.md) · 🇯🇵 [JA](README.ja.md) · 🇰🇷 [KO](README.ko.md) · 🇵🇱 [PL](README.pl.md) · 🇮🇹 [IT](README.it.md) · 🇹🇷 [TR](README.tr.md) · 🇮🇳 [TA](README.ta.md) · 🇮🇷 [FA](README.fa.md) · 🇺🇦 [UK](README.uk.md) · 🇩🇪 [DE](README.de.md) · 🇳🇱 [NL](README.nl.md)

Sistema ng sentralisadong pamamahala para sa PostgreSQL `pg_hba.conf` —
pamahalaan ang mga access rule sa buong fleet ng PostgreSQL instance mula
sa iisang web UI, sa halip na mano-manong i-edit ang `pg_hba.conf` sa
bawat host.

Ang repository na ito ay isang hardened, handang-handa-sa-publiko na
snapshot ng orihinal na PGHBA Manager project: pagpapatupad ng
default-password, startup secret validation, SSH host key pinning, login
rate limiting, at isang CI pipeline na nag-i-scan ng mga dependency sa
bawat push. Tingnan ang [Seguridad](#seguridad) sa ibaba.

🌐 **24 wika na handa nang gamitin** — palitan anumang oras mula sa
dropdown sa top bar: English, Bahasa Indonesia, 中文, Español, العربية,
Français, Português, বাংলা, Русский, हिन्दी, ไทย, ខ្មែរ, Tagalog, Tiếng
Việt, 日本語, 한국어, Polski, Italiano, Türkçe, தமிழ், فارسی, Українська,
Deutsch, Nederlands. Awtomatikong nire-render nang right-to-left ang
Arabic at Persian.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Mag-abot ka sa akin ng isang tasa ng kape.

---

## Mga Screenshot

### Dashboard

![Dashboard](docs/screenshots/dashboard.png)

> Ang mga screenshot sa ibaba (Database Hosts, PG Instances, Access Rules,
> Temporary Access, Backup & Restore, User Management, Audit Log,
> Settings) ay **mga mockup na paglalarawan lamang** na ginawa mula sa
> aktwal na frontend components at color scheme ng repository na ito,
> gamit ang dummy/sample na data — hindi ito mula sa live na deployment.
> Papalitan ito ng tunay na screenshot sa hinaharap.

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

## Mabilisang Simula (Docker)

```bash
# 1. I-clone at i-configure
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. Gumawa ng .env mula sa template (ginagamit ng docker-compose.yml)
cp backend/.env.example .env

# 3. Bumuo ng tunay na secrets — tatanggi ang app na mag-start kung wala ito
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. I-edit ang .env gamit ang iyong tunay na value, pagkatapos:
docker compose up -d

# Available ang app sa: http://localhost:8080
```

**Unang login**: `admin` / `Admin@1234`. Napipilitan ang account na ito na
baguhin ang password sa unang login — hinaharangan ng server ang lahat ng
ibang aksyon hanggang mabago ito, kaya walang puwang para sa "nakalimutan
palitan ang default password".

---

## Development Setup

```bash
# Simulan lang ang database
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # i-edit kung kailangan
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (nag-proxy ng /api sa :8080)
```

---

## Seguridad

- **Walang insecure na default na maaaring tumakbo.** Tinatawag ng
  `main.go` ang `config.Validate()` sa startup, na tatanggi mag-start
  kung ang `JWT_SECRET` o `ENCRYPTION_KEY` ay blangko o nananatiling
  nakalathalang placeholder — ang maling-naka-configure na deployment ay
  mabilis na mabibigo sa halip na tumakbo gamit ang isang JWT secret na
  madaling i-forge o isang kilalang AES key.
- **Naka-pin ang mga SSH host key**, hindi ito binabalewala.
  Pinagkakatiwalaan ng `services.ConnectToHost` ang SSH key ng isang
  pinamamahalaang host sa unang koneksyon at itinatago ang fingerprint
  nito (`db_hosts.ssh_host_key_fingerprint`); ang susunod na koneksyon
  gamit ang ibang key ay tatanggihan sa halip na tahimik na
  pagkakatiwalaan. Ang pag-clear ng fingerprint matapos ang lehitimong
  pag-rebuild ng host ay dumadaan sa isang dedikado at naka-audit na
  endpoint (`POST /api/hosts/:id/reset-host-key`), hindi sa karaniwang
  host-edit form.
- **May rate limit ang login**: 10 beses na pagkabigo mula sa isang IP sa
  loob ng 15 minuto ay magsasara sa IP na iyon ng 15 minuto.
- **Agad na magkakabisa ang pag-deactivate o pag-demote ng user** — bawat
  authenticated na request ay sinusuri muli ang status/role laban sa
  database sa halip na magtiwala sa claim ng JWT na maaaring 8 oras nang
  luma.
- **Sinusuri ng CI ang mga kilalang mahinang dependency** sa bawat
  push/PR at linggo-linggo (`.github/workflows/security.yml`):
  `govulncheck` para sa Go backend, `npm audit` (production dependencies)
  para sa frontend.
- **Naka-embed at self-applying ang mga database migration.** Bawat SQL
  file sa `backend/internal/database/migrations/` ay naka-bake sa binary
  (`go:embed`) at awtomatikong inilalapat — sinusubaybayan sa isang
  `schema_migrations` table — sa bawat startup, kaya hindi na kailanman
  tahimik na makakaligtaan ang deploy sa isang pagbabago sa schema tulad
  ng maaaring mangyari sa manu-manong pinapatakbong migration script.
- Ang backup export, CORS, at HTTP response headers ay pinatibay — tingnan
  ang `backend/internal/handlers/export.go` (path-segment sanitization),
  `backend/internal/server/router.go` (default na walang cross-origin
  access ang CORS maliban kung naka-set ang `CORS_ALLOWED_ORIGINS`), at
  `backend/internal/middleware/security_headers.go`.
- **Pag-rotate ng secrets pagkatapos ng isang leak**:
  `backend/cmd/reencrypt-ssh-keys` at
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) ang gabay sa
  pag-rotate ng `JWT_SECRET`/`ENCRYPTION_KEY` at muling pag-encrypt ng
  lahat ng nakaimbak na SSH key nang walang pagkawala ng data.

---

## Istraktura ng Proyekto

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Entry point
│   │   └── reencrypt-ssh-keys/         # Tool para sa pag-rotate ng secret (tingnan ang Seguridad)
│   └── internal/
│       ├── config/                     # Config loader + startup secret validation
│       ├── database/                   # Koneksyon ng GORM + embedded SQL migration runner
│       │   └── migrations/             # 001 (baseline, sa pamamagitan ng initdb) + 002–008, awtomatikong inilalapat sa startup
│       ├── models/                     # Lahat ng GORM model
│       ├── middleware/                 # JWT auth, RBAC, rate limiting, security headers
│       ├── handlers/                   # Gin HTTP handlers
│       │   ├── auth.go                 # Login / logout / refresh / change-password
│       │   ├── hosts.go                # Host CRUD, SSH validate, deploy-key, reset-host-key
│       │   ├── instances.go            # Instance discovery/scan
│       │   ├── rules.go                # pg_hba rule CRUD + enable/disable/reorder/clone
│       │   ├── temp_access.go          # Temporary access grant/revoke
│       │   ├── backup.go, backup_schedules.go   # Backup create/restore/download + scheduling
│       │   ├── export.go               # Config/backup export (local + S3)
│       │   ├── healthcheck.go          # Host health check + notifications
│       │   ├── users.go                # User CRUD (SuperAdmin lang)
│       │   ├── audit.go                # Audit log list + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH client, host key pinning (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # pg_hba content rendering + atomic write + expiry
│       │   ├── crypto.go               # AES-256-GCM para sa SSH key
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — nag-e-expire ng temp rules kada N minuto
│       └── server/router.go            # Gin router, CORS, lahat ng route
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API client
│       ├── types/index.ts              # TypeScript interfaces (salamin ng Go models)
│       ├── i18n/                       # 24 locale file + AVAILABLE_LOCALES registry (RTL-aware)
│       ├── stores/auth.ts              # Pinia auth store
│       ├── router/index.ts             # Vue Router na may auth + force-password-change guards
│       ├── components/layout/AppLayout.vue
│       └── views/                      # Isang view bawat feature area
│
├── docs/rotate-secrets.md              # Runbook ng secret rotation
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-stage: Vue build + Go build → alpine
├── docker-compose.yml                  # Production
└── docker-compose.dev.yml              # Development (DB lang)
```

---

## API Reference

Ang kumpletong listahan ng route ay nasa
`backend/internal/server/router.go`. Pangunahing endpoint:

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
POST   /api/hosts/validate              # SSH pre-flight check bago i-save
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # I-clear ang naka-pin na SSH host key fingerprint

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

## Mga RBAC Permission

| Feature                | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|-------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Pamahalaan ang Hosts    | ✓          | ✓   |          |        |               |                    |
| Pamahalaan ang Instances | ✓         | ✓   |          |        |               |                    |
| Idagdag/Edit/Burahin ang Rule | ✓    | ✓   |          |        | ✓             |                    |
| I-enable/Disable ang Rule | ✓        | ✓   | ✓        |        | ✓             |                    |
| Tingnan ang Rules       | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Magbigay/Bawiin ang Temp Access | ✓  | ✓   | ✓        |        | ✓             | ✓                  |
| Backup at Pagbawi       | ✓          | ✓   |          |        |               |                    |
| Pamamahala ng User      | ✓          |     |          |        |               |                    |
| Tingnan ang Audit Log   | ✓          | ✓   |          |        |               |                    |

Ang `AccessManager` at `TempAccessManager` ay mga limitadong role para sa
mga team na kailangan lamang mamahala ng access rules / temporary access
nang wala ang buong DBA privilege sa hosts, instances, backups, o users.

---

## Production Checklist

- [ ] I-set ang tunay na `JWT_SECRET`/`ENCRYPTION_KEY` — hindi mag-s-start ang app kung wala ito
- [ ] Mag-login nang isang beses at palitan ang default admin password (ipinapatupad sa unang login)
- [ ] Mag-set ng malakas na `DB_PASSWORD`
- [ ] I-enable ang TLS (`DB_SSL_MODE=require`)
- [ ] Ilagay ang app sa likod ng NGINX/Traefik na may HTTPS
- [ ] Limitahan ang port ng Postgres sa internal network lamang
- [ ] I-set ang `CORS_ALLOWED_ORIGINS` lamang kung iba ang origin ng frontend kumpara sa API
