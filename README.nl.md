# PGHBA Fleet Manager

<!--LANGSWITCH-->

PostgreSQL `pg_hba.conf` centraal beheersysteem — beheer toegangsregels
voor een hele vloot PostgreSQL-instanties vanuit één webinterface, in
plaats van `pg_hba.conf` handmatig te bewerken op elke host.

Deze repository is een gehard, publiek bruikbaar snapshot van het
oorspronkelijke PGHBA Manager-project: afdwinging van
standaardwachtwoorden, validatie van secrets bij het opstarten,
SSH-host-key-pinning, login-rate-limiting en een CI-pipeline die bij elke
push afhankelijkheden scant. Zie [Beveiliging](#beveiliging) hieronder.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Trakteer me gerust op een kopje koffie.

---

## Screenshots

### Dashboard

![Dashboard](docs/screenshots/dashboard.png)

> De onderstaande screenshots (Database-hosts, PG-instanties,
> Toegangsregels, Tijdelijke toegang, Back-up & herstel,
> Gebruikersbeheer, Auditlogboek, Instellingen) zijn **illustratieve
> mockups**, gebouwd met de werkelijke frontend-componenten en het
> kleurenschema van deze repo, gevuld met dummy-/voorbeeldgegevens —
> geen live-implementatie. Echte screenshots vervangen ze naarmate ze
> beschikbaar komen.

<details>
<summary>Database-hosts</summary>

![Database-hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>PG-instanties</summary>

![PG-instanties](docs/screenshots/instances.png)
</details>

<details>
<summary>Toegangsregels</summary>

![Toegangsregels](docs/screenshots/rules.png)
</details>

<details>
<summary>Tijdelijke toegang</summary>

![Tijdelijke toegang](docs/screenshots/temp-access.png)
</details>

<details>
<summary>Back-up & herstel</summary>

![Back-up & herstel](docs/screenshots/backup.png)
</details>

<details>
<summary>Gebruikersbeheer</summary>

![Gebruikersbeheer](docs/screenshots/users.png)
</details>

<details>
<summary>Auditlogboek</summary>

![Auditlogboek](docs/screenshots/audit.png)
</details>

<details>
<summary>Instellingen</summary>

![Instellingen](docs/screenshots/settings.png)
</details>

---

## Snel starten (Docker)

```bash
# 1. Klonen en configureren
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. .env aanmaken vanuit template (gebruikt door docker-compose.yml)
cp backend/.env.example .env

# 3. Genereer echte secrets — de app start anders niet
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. Bewerk .env met je echte waarden, start dan:
docker compose up -d

# App beschikbaar op: http://localhost:8080
```

**Eerste keer inloggen**: `admin` / `Admin@1234`. Dit account is
gemarkeerd om het wachtwoord bij de eerste login te wijzigen — de server
blokkeert elke andere actie totdat dit is gewijzigd, dus er is geen
"vergeten het standaardwachtwoord te wijzigen"-venster.

---

## Ontwikkelomgeving

```bash
# Start alleen de database
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # pas aan indien nodig
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (proxyt /api naar :8080)
```

---

## Beveiliging

- **Er kunnen geen onveilige standaardwaarden draaien.** `main.go` roept
  bij het opstarten `config.Validate()` aan, dat weigert te starten als
  `JWT_SECRET` of `ENCRYPTION_KEY` leeg is of nog de gepubliceerde
  placeholder bevat — een verkeerd geconfigureerde implementatie faalt
  luid, in plaats van te draaien met een vervalsbaar JWT-secret of een
  bekende AES-sleutel.
- **SSH-host-keys worden gepind**, niet genegeerd.
  `services.ConnectToHost` vertrouwt de SSH-sleutel van een beheerde
  host bij de eerste verbinding en slaat de vingerafdruk op
  (`db_hosts.ssh_host_key_fingerprint`); elke latere verbinding met een
  andere sleutel wordt geweigerd in plaats van stilzwijgend vertrouwd.
  Het wissen van een vingerafdruk na een legitieme host-rebuild
  verloopt via een speciaal, geauditeerd endpoint
  (`POST /api/hosts/:id/reset-host-key`), nooit via het algemene
  host-bewerkingsformulier.
- **Login heeft een rate-limit**: 10 mislukte pogingen vanaf één IP
  binnen 15 minuten sluit dat IP 15 minuten buiten.
- **Het deactiveren of degraderen van een gebruiker heeft direct
  effect** — elke geauthenticeerde aanvraag controleert status/rol
  opnieuw tegen de database, in plaats van te vertrouwen op wat een JWT
  tot 8 uur geleden beweerde.
- **CI scant bij elke push/PR en wekelijks op bekende kwetsbare
  afhankelijkheden** (`.github/workflows/security.yml`): `govulncheck`
  voor de Go-backend, `npm audit` (productieafhankelijkheden) voor de
  frontend.
- Back-upexport, CORS en HTTP-responseheaders zijn gehard — zie
  `backend/internal/handlers/export.go` (sanitatie van padsegmenten),
  `backend/internal/server/router.go` (CORS weigert standaard
  cross-origin-toegang tenzij `CORS_ALLOWED_ORIGINS` is ingesteld), en
  `backend/internal/middleware/security_headers.go`.
- **Secrets roteren na een lek**: `backend/cmd/reencrypt-ssh-keys` en
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) leiden je door het
  roteren van `JWT_SECRET`/`ENCRYPTION_KEY` en het opnieuw versleutelen
  van alle opgeslagen SSH-sleutels zonder gegevensverlies.

---

## Projectstructuur

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Startpunt
│   │   └── reencrypt-ssh-keys/         # Secret-rotatietool (zie Beveiliging)
│   └── internal/
│       ├── config/                     # Config-loader + secret-validatie bij opstarten
│       ├── database/                   # GORM-verbinding
│       ├── models/                     # Alle GORM-modellen
│       ├── middleware/                 # JWT-auth, RBAC, rate-limiting, security-headers
│       ├── handlers/                   # Gin-HTTP-handlers
│       │   ├── auth.go                 # Login / logout / refresh / wachtwoord wijzigen
│       │   ├── hosts.go                # Host-CRUD, SSH-validatie, deploy-key, reset-host-key
│       │   ├── instances.go            # Instantie-discovery/scan
│       │   ├── rules.go                # pg_hba-regel-CRUD + enable/disable/reorder/clone
│       │   ├── temp_access.go          # Tijdelijke toegang verlenen/intrekken
│       │   ├── backup.go, backup_schedules.go   # Back-up maken/herstellen/downloaden + planning
│       │   ├── export.go               # Config-/back-upexport (lokaal + S3)
│       │   ├── healthcheck.go          # Host-healthchecks + meldingen
│       │   ├── users.go                # Gebruikers-CRUD (alleen SuperAdmin)
│       │   ├── audit.go                # Auditlogboeklijst + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH-client, host-key-pinning (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # pg_hba-inhoudsrendering + atomair schrijven + verloop
│       │   ├── crypto.go               # AES-256-GCM voor SSH-sleutels
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — tijdelijke regels elke N minuten laten verlopen
│       └── server/router.go            # Gin-router, CORS, alle routes
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API-client
│       ├── types/index.ts              # TypeScript-interfaces (weerspiegelt Go-modellen)
│       ├── stores/auth.ts              # Pinia auth-store
│       ├── router/index.ts             # Vue Router met auth- + wachtwoordwijziging-guards
│       ├── components/layout/AppLayout.vue
│       └── views/                      # Eén view per functiegebied
│
├── database/migrations/
│   ├── 001_init.sql                    # Volledig schema met enums, indexen, triggers
│   └── 002–007_*.sql                   # RBAC-rollen, back-upschema's, wachtwoord-/host-key-verharding
│
├── docs/rotate-secrets.md              # Runbook voor secret-rotatie
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-stage: Vue-build + Go-build → alpine
├── docker-compose.yml                  # Productie
└── docker-compose.dev.yml              # Ontwikkeling (alleen DB)
```

---

## API-referentie

De volledige routelijst staat in `backend/internal/server/router.go`.
Kern-endpoints:

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
POST   /api/hosts/validate              # SSH-vooraf-check voordat opgeslagen wordt
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # Gepinde SSH-host-key-vingerafdruk wissen

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

## RBAC-rechten

| Functie                | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Hosts beheren          | ✓          | ✓   |          |        |               |                    |
| Instanties beheren     | ✓          | ✓   |          |        |               |                    |
| Regel toevoegen/bewerken/verwijderen | ✓ | ✓ |        |        | ✓             |                    |
| Regel in-/uitschakelen | ✓          | ✓   | ✓        |        | ✓             |                    |
| Regels bekijken        | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Tijdelijke toegang verlenen/intrekken | ✓ | ✓ | ✓      |        | ✓             | ✓                  |
| Back-up & herstel      | ✓          | ✓   |          |        |               |                    |
| Gebruikersbeheer       | ✓          |     |          |        |               |                    |
| Auditlogboek bekijken  | ✓          | ✓   |          |        |               |                    |

`AccessManager` en `TempAccessManager` zijn beperkte rollen voor teams
die alleen toegangsregels / tijdelijke toegang moeten beheren zonder
volledige DBA-rechten over hosts, instanties, back-ups of gebruikers.

---

## Productie-checklist

- [ ] Stel echte `JWT_SECRET`/`ENCRYPTION_KEY` in — de app start anders niet
- [ ] Log eenmaal in en wijzig het standaard-adminwachtwoord (afgedwongen bij eerste login)
- [ ] Stel een sterk `DB_PASSWORD` in
- [ ] Schakel TLS in (`DB_SSL_MODE=require`)
- [ ] Zet de app achter NGINX/Traefik met HTTPS
- [ ] Beperk de Postgres-poort tot alleen het interne netwerk
- [ ] Stel `CORS_ALLOWED_ORIGINS` alleen in als de frontend op een andere origin draait dan de API
