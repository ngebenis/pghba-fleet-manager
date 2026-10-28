# PGHBA Fleet Manager

<!--LANGSWITCH-->

ប្រព័ន្ធគ្រប់គ្រងកណ្ដាលសម្រាប់ `pg_hba.conf` របស់ PostgreSQL — គ្រប់គ្រងច្បាប់
សិទ្ធិចូលប្រើទូទាំងក្រុមម៉ាស៊ីនលេខ PostgreSQL ទាំងមូលពីគេហទំព័រតែមួយ
ជំនួសឱ្យការកែសម្រួល `pg_hba.conf` ដោយដៃនៅលើម៉ាស៊ីននីមួយៗ

រីផូស៊ីតូរីនេះគឺជាស្នេបសត់ដែលបានពង្រឹងសុវត្ថិភាព និងត្រៀមខ្លួនសម្រាប់
សាធារណៈ នៃគម្រោង PGHBA Manager ដើម៖ ការបង្ខំឱ្យផ្លាស់ប្ដូរពាក្យសម្ងាត់
លំនាំដើម ការផ្ទៀងផ្ទាត់ secret ពេលចាប់ផ្ដើម ការភ្ជាប់កូនសោម៉ាស៊ីន SSH
ការកំណត់អត្រាចូលប្រើ និង pipeline CI ដែលស្កេនឌីផិនឌង់ស៊ីរាល់ពេល push
សូមមើល [សុវត្ថិភាព](#សុវត្ថិភាព) ខាងក្រោម

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ សូមចាក់កាហ្វេមួយកែវជូនខ្ញុំផង។

---

## រូបថតអេក្រង់

### ផ្ទាំងគ្រប់គ្រង

![Dashboard](docs/screenshots/dashboard.png)

> រូបថតអេក្រង់ខាងក្រោម (ម៉ាស៊ីនមូលដ្ឋានទិន្នន័យ, ម៉ាស៊ីនលេខ PG, ច្បាប់សិទ្ធិចូលប្រើ,
> សិទ្ធិចូលប្រើបណ្ដោះអាសន្ន, បម្រុងទុក និងស្ដារ, គ្រប់គ្រងអ្នកប្រើ, កំណត់ហេតុសវនកម្ម,
> ការកំណត់) គឺជា **គំរូបង្ហាញឧទាហរណ៍** ដែលបង្កើតឡើងពីសមាសធាតុ frontend
> និងពណ៌ពិតប្រាកដនៃរីផូស៊ីតូរីនេះ ដោយប្រើទិន្នន័យគំរូ/ទិន្នន័យក្លែងក្លាយ —
> មិនមែនជាការដាក់ឱ្យប្រើប្រាស់ជាក់ស្ដែងនោះទេ។ រូបថតអេក្រង់ពិតប្រាកដនឹងជំនួស
> ពេលមានភាពត្រៀមខ្លួន។

<details>
<summary>ម៉ាស៊ីនមូលដ្ឋានទិន្នន័យ</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>ម៉ាស៊ីនលេខ PG</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>ច្បាប់សិទ្ធិចូលប្រើ</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>សិទ្ធិចូលប្រើបណ្ដោះអាសន្ន</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>បម្រុងទុក និងស្ដារ</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>គ្រប់គ្រងអ្នកប្រើ</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>កំណត់ហេតុសវនកម្ម</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>ការកំណត់</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## ចាប់ផ្ដើមរហ័ស (Docker)

```bash
# 1. ក្លូន និងកំណត់រចនាសម្ព័ន្ធ
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. បង្កើត .env ពីគំរូ (ប្រើដោយ docker-compose.yml)
cp backend/.env.example .env

# 3. បង្កើត secret ពិតប្រាកដ — កម្មវិធីនឹងបដិសេធមិនចាប់ផ្ដើមដោយគ្មានវា
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. កែសម្រួល .env ដោយតម្លៃពិតប្រាកដរបស់អ្នក បន្ទាប់មក៖
docker compose up -d

# កម្មវិធីអាចចូលប្រើបានតាម៖ http://localhost:8080
```

**ការចូលប្រើលើកដំបូង**៖ `admin` / `Admin@1234`។ គណនីនេះត្រូវបានបង្ខំឱ្យ
ផ្លាស់ប្ដូរពាក្យសម្ងាត់នៅពេលចូលប្រើលើកដំបូង — ម៉ាស៊ីនមេនឹងទប់ស្កាត់សកម្មភាព
ផ្សេងទៀតទាំងអស់រហូតដល់វាត្រូវបានផ្លាស់ប្ដូរ ដូច្នេះមិនមានចន្លោះប្រហោង
"ភ្លេចផ្លាស់ប្ដូរពាក្យសម្ងាត់លំនាំដើម" ទេ។

---

## ការរៀបចំសម្រាប់អភិវឌ្ឍន៍

```bash
# ចាប់ផ្ដើមតែមូលដ្ឋានទិន្នន័យប៉ុណ្ណោះ
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # កែសម្រួលតាមតម្រូវការ
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (ភ្ជាប់ត /api ទៅ :8080)
```

---

## សុវត្ថិភាព

- **គ្មានតម្លៃលំនាំដើមដែលមិនមានសុវត្ថិភាពអាចដំណើរការបានទេ។** `main.go` ហៅ
  `config.Validate()` ពេលចាប់ផ្ដើម ដែលនឹងបដិសេធមិនចាប់ផ្ដើមប្រសិនបើ
  `JWT_SECRET` ឬ `ENCRYPTION_KEY` ទទេ ឬនៅតែជាតម្លៃ placeholder ដែលបានផ្សព្វផ្សាយ —
  ការដាក់ឱ្យប្រើប្រាស់ខុសឆ្គងនឹងបរាជ័យយ៉ាងច្បាស់លាស់ ជំនួសឱ្យការដំណើរការ
  ជាមួយ JWT secret ដែលអាចក្លែងបន្លំបាន ឬ AES key ដែលគេស្គាល់។
- **កូនសោម៉ាស៊ីន SSH ត្រូវបានភ្ជាប់** មិនមែនត្រូវបានមិនអើពើទេ។
  `services.ConnectToHost` ជឿទុកចិត្តលើកូនសោ SSH របស់ម៉ាស៊ីនដែលគ្រប់គ្រងនៅ
  ពេលភ្ជាប់លើកដំបូង ហើយរក្សាទុក fingerprint របស់វា
  (`db_hosts.ssh_host_key_fingerprint`)។ ការភ្ជាប់លើកក្រោយដោយកូនសោខុសគ្នា
  នឹងត្រូវបានបដិសេធ ជំនួសឱ្យការជឿទុកចិត្តដោយស្ងាត់ៗ។ ការលុប fingerprint
  បន្ទាប់ពីការកសាងម៉ាស៊ីនឡើងវិញដោយស្របច្បាប់ត្រូវឆ្លងកាត់ endpoint ពិសេសដែល
  មានការកត់ត្រាសវនកម្ម (`POST /api/hosts/:id/reset-host-key`) មិនមែនតាមរយៈ
  ទម្រង់កែសម្រួលម៉ាស៊ីនទូទៅទេ។
- **ការចូលប្រើត្រូវបានកំណត់អត្រា**៖ ការព្យាយាមបរាជ័យ 10 ដងពី IP តែមួយ
  ក្នុងរយៈពេល 15 នាទី នឹងចាក់សោ IP នោះអស់រយៈពេល 15 នាទី។
- **ការបិទ ឬបន្ថយតួនាទីអ្នកប្រើមានប្រសិទ្ធភាពភ្លាមៗ** — រាល់សំណើដែលបាន
  ផ្ទៀងផ្ទាត់អត្តសញ្ញាណនឹងពិនិត្យមើលស្ថានភាព/តួនាទីជាថ្មីជាមួយមូលដ្ឋានទិន្នន័យ
  ជំនួសឱ្យការជឿទុកចិត្តលើអ្វីដែល JWT បានអះអាងរហូតដល់ 8 ម៉ោង។
- **CI ស្កេនរកឌីផិនឌង់ស៊ីដែលងាយរងគ្រោះដែលគេស្គាល់** រាល់ពេល push/PR និងរៀងរាល់
  សប្ដាហ៍ (`.github/workflows/security.yml`)៖ `govulncheck` សម្រាប់ backend Go
  `npm audit` (ឌីផិនឌង់ស៊ីផលិតកម្ម) សម្រាប់ frontend។
- ការនាំចេញការបម្រុងទុក, CORS, និង HTTP response headers ត្រូវបានពង្រឹង — សូមមើល
  `backend/internal/handlers/export.go` (ការសម្អាត path segment),
  `backend/internal/server/router.go` (CORS លំនាំដើមមិនអនុញ្ញាតសិទ្ធិចូលប្រើ
  cross-origin ទេ លុះត្រាតែបានកំណត់ `CORS_ALLOWED_ORIGINS`), និង
  `backend/internal/middleware/security_headers.go`។
- **ការប្ដូរ secret បន្ទាប់ពីការលេចធ្លាយ**៖ `backend/cmd/reencrypt-ssh-keys`
  និង [`docs/rotate-secrets.md`](docs/rotate-secrets.md) នឹងណែនាំជំហាន
  ការប្ដូរ `JWT_SECRET`/`ENCRYPTION_KEY` និងអ៊ិនគ្រីបកូនសោ SSH ដែលបានរក្សាទុក
  ទាំងអស់ឡើងវិញដោយគ្មានការបាត់បង់ទិន្នន័យ។

---

## រចនាសម្ព័ន្ធគម្រោង

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # ចំណុចចាប់ផ្ដើម
│   │   └── reencrypt-ssh-keys/         # ឧបករណ៍ប្ដូរ secret (មើលសុវត្ថិភាព)
│   └── internal/
│       ├── config/                     # កម្មវិធីផ្ទុក config + ផ្ទៀងផ្ទាត់ secret ពេលចាប់ផ្ដើម
│       ├── database/                   # ការភ្ជាប់ GORM
│       ├── models/                     # គំរូ GORM ទាំងអស់
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

## ឯកសារយោង API

បញ្ជី route ពេញលេញនៅក្នុង `backend/internal/server/router.go`។ endpoint
ចម្បង៖

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

## សិទ្ធិ RBAC

| លក្ខណៈពិសេស             | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| គ្រប់គ្រងម៉ាស៊ីន          | ✓          | ✓   |          |        |               |                    |
| គ្រប់គ្រងម៉ាស៊ីនលេខ        | ✓          | ✓   |          |        |               |                    |
| បន្ថែម/កែសម្រួល/លុបច្បាប់   | ✓          | ✓   |          |        | ✓             |                    |
| បើក/បិទច្បាប់             | ✓          | ✓   | ✓        |        | ✓             |                    |
| មើលច្បាប់                | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| ផ្ដល់/ដកហូតសិទ្ធិចូលប្រើបណ្ដោះអាសន្ន | ✓ | ✓   | ✓        |        | ✓             | ✓                  |
| បម្រុងទុក និងស្ដារ        | ✓          | ✓   |          |        |               |                    |
| គ្រប់គ្រងអ្នកប្រើ          | ✓          |     |          |        |               |                    |
| មើលកំណត់ហេតុសវនកម្ម       | ✓          | ✓   |          |        |               |                    |

`AccessManager` និង `TempAccessManager` គឺជាតួនាទីដែលមានវិសាលភាពកំណត់
សម្រាប់ក្រុមដែលគ្រាន់តែត្រូវការគ្រប់គ្រងច្បាប់សិទ្ធិចូលប្រើ / សិទ្ធិចូលប្រើ
បណ្ដោះអាសន្ន ដោយមិនចាំបាច់មានសិទ្ធិ DBA ពេញលេញលើម៉ាស៊ីន, ម៉ាស៊ីនលេខ,
ការបម្រុងទុក ឬអ្នកប្រើឡើយ។

---

## បញ្ជីត្រួតពិនិត្យសម្រាប់ផលិតកម្ម

- [ ] កំណត់ `JWT_SECRET`/`ENCRYPTION_KEY` ពិតប្រាកដ — កម្មវិធីនឹងមិនចាប់ផ្ដើមទេបើគ្មានវា
- [ ] ចូលប្រើម្ដង ហើយផ្លាស់ប្ដូរពាក្យសម្ងាត់ admin លំនាំដើម (បង្ខំនៅពេលចូលប្រើលើកដំបូង)
- [ ] កំណត់ `DB_PASSWORD` ដែលរឹងមាំ
- [ ] បើកដំណើរការ TLS (`DB_SSL_MODE=require`)
- [ ] ដាក់កម្មវិធីនៅពីក្រោយ NGINX/Traefik ជាមួយ HTTPS
- [ ] កំណត់ច្រក Postgres ឱ្យតែសម្រាប់បណ្ដាញខាងក្នុងប៉ុណ្ណោះ
- [ ] កំណត់ `CORS_ALLOWED_ORIGINS` តែនៅពេល frontend ត្រូវបានដាក់ឱ្យប្រើលើ origin ផ្សេងពី API ប៉ុណ្ណោះ
