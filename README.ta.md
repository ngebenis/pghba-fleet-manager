# PGHBA Fleet Manager

🇮🇩 [ID](README.md) · 🇬🇧 [EN](README.en.md) · 🇨🇳 [ZH](README.zh.md) · 🇪🇸 [ES](README.es.md) · 🇸🇦 [AR](README.ar.md) · 🇫🇷 [FR](README.fr.md) · 🇵🇹 [PT](README.pt.md) · 🇧🇩 [BN](README.bn.md) · 🇷🇺 [RU](README.ru.md) · 🇮🇳 [HI](README.hi.md) · 🇹🇭 [TH](README.th.md) · 🇰🇭 [KM](README.km.md) · 🇵🇭 [TL](README.tl.md) · 🇻🇳 [VI](README.vi.md) · 🇯🇵 [JA](README.ja.md) · 🇰🇷 [KO](README.ko.md) · 🇵🇱 [PL](README.pl.md) · 🇮🇹 [IT](README.it.md) · 🇹🇷 [TR](README.tr.md) · 🇮🇳 **TA** · 🇮🇷 [FA](README.fa.md) · 🇺🇦 [UK](README.uk.md) · 🇩🇪 [DE](README.de.md) · 🇳🇱 [NL](README.nl.md)

PostgreSQL `pg_hba.conf` மைய மேலாண்மை அமைப்பு — ஒவ்வொரு ஹோஸ்டிலும்
`pg_hba.conf`-ஐ கைமுறையாகத் திருத்துவதற்குப் பதிலாக, ஒரே வலை UI-இலிருந்து
முழு PostgreSQL ஹோஸ்ட் தொகுப்பின் அணுகல் விதிகளை நிர்வகிக்கவும்.

இந்த களஞ்சியம் அசல் PGHBA Manager திட்டத்தின் பாதுகாப்பு வலுவூட்டப்பட்ட,
பொது வெளியீட்டுக்குத் தயாரான ஸ்னாப்ஷாட் ஆகும்: இயல்புநிலை கடவுச்சொல்
அமலாக்கம், தொடக்கத்தில் ரகசிய சரிபார்ப்பு, SSH ஹோஸ்ட் கீ பின்னிங், உள்நுழைவு
வீத வரம்பு, மற்றும் ஒவ்வொரு புஷ்ஷிலும் சார்புகளை ஸ்கேன் செய்யும் CI
பைப்லைன். கீழே உள்ள [பாதுகாப்பு](#பாதுகாப்பு) பகுதியைப் பார்க்கவும்.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ தயவுசெய்து எனக்கு ஒரு கப் காபி வாங்கிக் கொடுங்கள்.

---

## ஸ்கிரீன்ஷாட்கள்

### டாஷ்போர்டு

![Dashboard](docs/screenshots/dashboard.png)

> கீழே உள்ள ஸ்கிரீன்ஷாட்கள் (தரவுத்தள ஹோஸ்ட்கள், PG இன்ஸ்டன்ஸ்கள், அணுகல்
> விதிகள், தற்காலிக அணுகல், காப்புப்பிரதி & மீட்டமைப்பு, பயனர் மேலாண்மை,
> தணிக்கை பதிவு, அமைப்புகள்) இந்த களஞ்சியத்தின் உண்மையான frontend
> கூறுகள் மற்றும் வண்ணத் திட்டத்தைப் பயன்படுத்தி உருவாக்கப்பட்ட, மாதிரி/
> டம்மி தரவுகளால் நிரப்பப்பட்ட **விளக்கப் பட மாதிரிகள்** — நேரடி
> டெப்ளாய்மென்ட் அல்ல. உண்மையான ஸ்கிரீன்ஷாட்கள் கிடைக்கும்போது இவற்றுக்குப்
> பதிலாக மாற்றப்படும்.

<details>
<summary>தரவுத்தள ஹோஸ்ட்கள்</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>PG இன்ஸ்டன்ஸ்கள்</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>அணுகல் விதிகள்</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>தற்காலிக அணுகல்</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>காப்புப்பிரதி & மீட்டமைப்பு</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>பயனர் மேலாண்மை</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>தணிக்கை பதிவு</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>அமைப்புகள்</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## விரைவு தொடக்கம் (Docker)

```bash
# 1. குளோன் செய்து கட்டமைக்கவும்
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. வார்ப்புருவிலிருந்து .env உருவாக்கவும் (docker-compose.yml பயன்படுத்துகிறது)
cp backend/.env.example .env

# 3. உண்மையான ரகசியங்களை உருவாக்கவும் — இவை இல்லாமல் ஆப் தொடங்காது
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. உங்கள் உண்மையான மதிப்புகளுடன் .env-ஐத் திருத்தவும், பிறகு:
docker compose up -d

# ஆப் இங்கு கிடைக்கும்: http://localhost:8080
```

**முதல் உள்நுழைவு**: `admin` / `Admin@1234`. இந்தக் கணக்கு முதல்
உள்நுழைவில் கடவுச்சொல்லை மாற்றும்படி கட்டாயப்படுத்தப்பட்டுள்ளது —
கடவுச்சொல் மாற்றப்படும் வரை சர்வர் மற்ற அனைத்து செயல்களையும் தடுக்கும்,
எனவே "இயல்புநிலை கடவுச்சொல்லை மாற்ற மறந்துவிட்டேன்" என்ற நிலை ஏற்படாது.

---

## மேம்பாட்டு அமைவு

```bash
# தரவுத்தளத்தை மட்டும் தொடங்கவும்
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # தேவைக்கேற்ப திருத்தவும்
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (/api-ஐ :8080-க்கு proxy செய்கிறது)
```

---

## பாதுகாப்பு

- **பாதுகாப்பற்ற இயல்புநிலைகள் இயங்க முடியாது.** `main.go`, தொடக்கத்தில்
  `config.Validate()`-ஐ அழைக்கிறது, இது `JWT_SECRET` அல்லது
  `ENCRYPTION_KEY` காலியாக இருந்தாலோ அல்லது இன்னும் வெளியிடப்பட்ட
  placeholder-ஆக இருந்தாலோ தொடங்க மறுக்கும் — தவறாக கட்டமைக்கப்பட்ட
  டெப்ளாய்மென்ட், போலியாக்கக்கூடிய JWT ரகசியம் அல்லது அறியப்பட்ட AES
  கீயுடன் இயங்குவதற்குப் பதிலாக உரத்த குரலில் தோல்வியடையும்.
- **SSH ஹோஸ்ட் கீகள் புறக்கணிக்கப்படாமல் பின் செய்யப்படுகின்றன.**
  `services.ConnectToHost`, முதல் இணைப்பில் நிர்வகிக்கப்படும் ஹோஸ்டின்
  SSH கீயை நம்பி அதன் கைரேகையை (`db_hosts.ssh_host_key_fingerprint`)
  நிரந்தரமாகச் சேமிக்கிறது; வேறு கீயுடனான பிந்தைய இணைப்புகள் அமைதியாக
  நம்பப்படுவதற்குப் பதிலாக நிராகரிக்கப்படும். சரியான ஹோஸ்ட் மறுகட்டமைப்புக்குப்
  பிறகு கைரேகையை அழிப்பது, பொதுவான ஹோஸ்ட்-திருத்தும் படிவத்தின் வழியாக
  அல்லாமல் தனியான, தணிக்கை செய்யப்பட்ட முனையத்தின் (`POST
  /api/hosts/:id/reset-host-key`) வழியாகச் செல்கிறது.
- **உள்நுழைவு வீத வரம்பிற்கு உட்பட்டது**: 15 நிமிடங்களில் ஒரே IP-யிலிருந்து
  10 தோல்வியுற்ற முயற்சிகள் அந்த IP-ஐ 15 நிமிடங்களுக்குப் பூட்டும்.
- **ஒரு பயனரை முடக்குவது அல்லது பதவி இறக்குவது உடனடியாக அமலாகும்** —
  ஒவ்வொரு அங்கீகரிக்கப்பட்ட கோரிக்கையும், 8 மணி நேரம் வரை பழையதாக
  இருக்கக்கூடிய JWT உரிமைகோரலை நம்புவதற்குப் பதிலாக தரவுத்தளத்திற்கு
  எதிராக நிலை/பங்கை மீண்டும் சரிபார்க்கிறது.
- **CI ஒவ்வொரு push/PR-இலும் மற்றும் வாராந்திரமாகவும்**
  (`.github/workflows/security.yml`) அறியப்பட்ட பாதிப்புள்ள சார்புகளை
  ஸ்கேன் செய்கிறது: Go backend-க்கு `govulncheck`, frontend-க்கு `npm
  audit` (production சார்புகள்).
- காப்புப்பிரதி ஏற்றுமதி, CORS, மற்றும் HTTP பதில் தலைப்புகள்
  பலப்படுத்தப்பட்டுள்ளன — பார்க்கவும் `backend/internal/handlers/export.go`
  (பாதை-பிரிவு சுத்திகரிப்பு), `backend/internal/server/router.go`
  (`CORS_ALLOWED_ORIGINS` அமைக்கப்படாவிட்டால் CORS இயல்புநிலையாக
  குறுக்கு-தோற்ற அணுகலை அனுமதிக்காது), மற்றும்
  `backend/internal/middleware/security_headers.go`.
- **கசிவுக்குப் பிறகு ரகசியங்களைச் சுழற்றுதல்**:
  `backend/cmd/reencrypt-ssh-keys` மற்றும்
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) `JWT_SECRET`/
  `ENCRYPTION_KEY`-ஐ சுழற்றுவதையும், சேமிக்கப்பட்ட அனைத்து SSH
  கீகளையும் தரவு இழப்பு இல்லாமல் மீண்டும் குறியாக்கம் செய்வதையும்
  படிப்படியாக விளக்குகிறது.

---

## திட்ட அமைப்பு

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # நுழைவு புள்ளி
│   │   └── reencrypt-ssh-keys/         # ரகசிய சுழற்சி கருவி (பாதுகாப்பைப் பார்க்கவும்)
│   └── internal/
│       ├── config/                     # கட்டமைப்பு ஏற்றி + தொடக்க ரகசிய சரிபார்ப்பு
│       ├── database/                   # GORM இணைப்பு
│       ├── models/                     # அனைத்து GORM மாடல்கள்
│       ├── middleware/                 # JWT auth, RBAC, வீத வரம்பு, பாதுகாப்பு தலைப்புகள்
│       ├── handlers/                   # Gin HTTP ஹேண்ட்லர்கள்
│       │   ├── auth.go                 # உள்நுழைவு / வெளியேறு / புதுப்பி / கடவுச்சொல் மாற்று
│       │   ├── hosts.go                # ஹோஸ்ட் CRUD, SSH சரிபார்ப்பு, deploy-key, reset-host-key
│       │   ├── instances.go            # இன்ஸ்டன்ஸ் கண்டறிதல்/ஸ்கேன்
│       │   ├── rules.go                # pg_hba விதி CRUD + இயக்கு/முடக்கு/மறுவரிசைப்படுத்து/குளோன்
│       │   ├── temp_access.go          # தற்காலிக அணுகல் வழங்கு/ரத்துசெய்
│       │   ├── backup.go, backup_schedules.go   # காப்புப்பிரதி உருவாக்கு/மீட்டமை/பதிவிறக்கு + அட்டவணை
│       │   ├── export.go               # கட்டமைப்பு/காப்புப்பிரதி ஏற்றுமதி (உள்ளூர் + S3)
│       │   ├── healthcheck.go          # ஹோஸ்ட் ஆரோக்கிய சோதனைகள் + அறிவிப்புகள்
│       │   ├── users.go                # பயனர் CRUD (SuperAdmin மட்டும்)
│       │   ├── audit.go                # தணிக்கை பதிவு பட்டியல் + வேறுபாடு
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH கிளையன்ட், ஹோஸ்ட் கீ பின்னிங் (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # pg_hba உள்ளடக்க உருவாக்கம் + அணு எழுதுதல் + காலாவதி
│       │   ├── crypto.go               # SSH கீகளுக்கான AES-256-GCM
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — ஒவ்வொரு N நிமிடத்திற்கும் தற்காலிக விதிகளை காலாவதி செய்
│       └── server/router.go            # Gin router, CORS, அனைத்து routes
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API கிளையன்ட்
│       ├── types/index.ts              # TypeScript இன்டர்ஃபேஸ்கள் (Go மாடல்களை பிரதிபலிக்கும்)
│       ├── stores/auth.ts              # Pinia auth store
│       ├── router/index.ts             # auth + கட்டாய கடவுச்சொல் மாற்று காப்புடன் Vue Router
│       ├── components/layout/AppLayout.vue
│       └── views/                      # ஒவ்வொரு அம்சப் பகுதிக்கும் ஒரு view
│
├── database/migrations/
│   ├── 001_init.sql                    # enum, index, trigger-களுடன் முழு ஸ்கீமா
│   └── 002–007_*.sql                   # RBAC பங்குகள், காப்புப்பிரதி அட்டவணைகள், கடவுச்சொல்/host-key வலுவூட்டல்
│
├── docs/rotate-secrets.md              # ரகசிய சுழற்சி வழிகாட்டி
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # பல-நிலை: Vue build + Go build → alpine
├── docker-compose.yml                  # Production
└── docker-compose.dev.yml              # Development (DB மட்டும்)
```

---

## API குறிப்பு

முழு route பட்டியல் `backend/internal/server/router.go`-இல் உள்ளது.
முக்கிய முனையங்கள்:

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
POST   /api/hosts/validate              # சேமிப்பதற்கு முன் SSH முன்-சரிபார்ப்பு
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # பின் செய்யப்பட்ட SSH ஹோஸ்ட் கீ கைரேகையை அழி

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

## RBAC அனுமதிகள்

| அம்சம்                   | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|----------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| ஹோஸ்ட்களை நிர்வகி          | ✓          | ✓   |          |        |               |                    |
| இன்ஸ்டன்ஸ்களை நிர்வகி       | ✓          | ✓   |          |        |               |                    |
| விதி சேர்/திருத்து/நீக்கு    | ✓          | ✓   |          |        | ✓             |                    |
| விதியை இயக்கு/முடக்கு        | ✓          | ✓   | ✓        |        | ✓             |                    |
| விதிகளைக் காண்க             | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| தற்காலிக அணுகலை வழங்கு/ரத்துசெய் | ✓     | ✓   | ✓        |        | ✓             | ✓                  |
| காப்புப்பிரதி & மீட்டமைப்பு | ✓          | ✓   |          |        |               |                    |
| பயனர் மேலாண்மை              | ✓          |     |          |        |               |                    |
| தணிக்கை பதிவைக் காண்க       | ✓          | ✓   |          |        |               |                    |

`AccessManager` மற்றும் `TempAccessManager` ஆகியவை, ஹோஸ்ட்கள்,
இன்ஸ்டன்ஸ்கள், காப்புப்பிரதிகள் அல்லது பயனர்கள் மீதான முழு DBA
அதிகாரம் இல்லாமல் அணுகல் விதிகள் / தற்காலிக அணுகலை மட்டும் நிர்வகிக்க
வேண்டிய குழுக்களுக்கான வரையறுக்கப்பட்ட பங்குகள் ஆகும்.

---

## Production சரிபார்ப்புப் பட்டியல்

- [ ] உண்மையான `JWT_SECRET`/`ENCRYPTION_KEY`-ஐ அமைக்கவும் — இல்லையெனில் ஆப் தொடங்காது
- [ ] ஒருமுறை உள்நுழைந்து இயல்புநிலை admin கடவுச்சொல்லை மாற்றவும் (முதல் உள்நுழைவில் கட்டாயம்)
- [ ] வலுவான `DB_PASSWORD`-ஐ அமைக்கவும்
- [ ] TLS-ஐ இயக்கவும் (`DB_SSL_MODE=require`)
- [ ] ஆப்பை HTTPS உடன் NGINX/Traefik-க்குப் பின் வைக்கவும்
- [ ] Postgres-இன் போர்ட்டை உள் நெட்வொர்க்குக்கு மட்டும் கட்டுப்படுத்தவும்
- [ ] frontend, API-விட வேறு origin-இல் டெப்ளாய் செய்யப்பட்டால் மட்டும் `CORS_ALLOWED_ORIGINS`-ஐ அமைக்கவும்
