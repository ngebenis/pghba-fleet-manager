# PGHBA Fleet Manager

🇮🇩 [ID](README.md) · 🇬🇧 [EN](README.en.md) · 🇨🇳 [ZH](README.zh.md) · 🇪🇸 [ES](README.es.md) · 🇸🇦 [AR](README.ar.md) · 🇫🇷 [FR](README.fr.md) · 🇵🇹 [PT](README.pt.md) · 🇧🇩 [BN](README.bn.md) · 🇷🇺 [RU](README.ru.md) · 🇮🇳 [HI](README.hi.md) · 🇹🇭 [TH](README.th.md) · 🇰🇭 [KM](README.km.md) · 🇵🇭 [TL](README.tl.md) · 🇻🇳 [VI](README.vi.md) · 🇯🇵 [JA](README.ja.md) · 🇰🇷 [KO](README.ko.md) · 🇵🇱 [PL](README.pl.md) · 🇮🇹 [IT](README.it.md) · 🇹🇷 [TR](README.tr.md) · 🇮🇳 [TA](README.ta.md) · 🇮🇷 [FA](README.fa.md) · 🇺🇦 [UK](README.uk.md) · 🇩🇪 **DE** · 🇳🇱 [NL](README.nl.md)

PostgreSQL `pg_hba.conf` zentrales Verwaltungssystem — verwalten Sie
Zugriffsregeln für eine ganze Flotte von PostgreSQL-Instanzen über eine
einzige Web-Oberfläche, statt `pg_hba.conf` auf jedem Host manuell zu
bearbeiten.

Dieses Repository ist ein gehärteter, öffentlich nutzbarer Snapshot des
ursprünglichen PGHBA-Manager-Projekts: Erzwingung von Passwortänderungen
bei Standardpasswörtern, Validierung von Secrets beim Start, SSH-Host-Key-
Pinning, Login-Rate-Limiting und eine CI-Pipeline, die bei jedem Push
Abhängigkeiten scannt. Siehe [Sicherheit](#sicherheit) unten.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Spendiere mir gerne eine Tasse Kaffee.

---

## Screenshots

### Dashboard

![Dashboard](docs/screenshots/dashboard.png)

> Die folgenden Screenshots (Datenbank-Hosts, PG-Instanzen,
> Zugriffsregeln, Temporärer Zugriff, Sicherung & Wiederherstellung,
> Benutzerverwaltung, Audit-Log, Einstellungen) sind **illustrative
> Mockups**, erstellt aus den tatsächlichen Frontend-Komponenten und dem
> Farbschema dieses Repos, befüllt mit Dummy-/Beispieldaten — kein
> Live-Deployment. Echte Screenshots werden sie nach und nach ersetzen.

<details>
<summary>Datenbank-Hosts</summary>

![Datenbank-Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>PG-Instanzen</summary>

![PG-Instanzen](docs/screenshots/instances.png)
</details>

<details>
<summary>Zugriffsregeln</summary>

![Zugriffsregeln](docs/screenshots/rules.png)
</details>

<details>
<summary>Temporärer Zugriff</summary>

![Temporärer Zugriff](docs/screenshots/temp-access.png)
</details>

<details>
<summary>Sicherung & Wiederherstellung</summary>

![Sicherung & Wiederherstellung](docs/screenshots/backup.png)
</details>

<details>
<summary>Benutzerverwaltung</summary>

![Benutzerverwaltung](docs/screenshots/users.png)
</details>

<details>
<summary>Audit-Log</summary>

![Audit-Log](docs/screenshots/audit.png)
</details>

<details>
<summary>Einstellungen</summary>

![Einstellungen](docs/screenshots/settings.png)
</details>

---

## Schnellstart (Docker)

```bash
# 1. Klonen und konfigurieren
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. .env aus Vorlage erstellen (wird von docker-compose.yml verwendet)
cp backend/.env.example .env

# 3. Echte Secrets generieren — die App startet sonst nicht
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. .env mit echten Werten bearbeiten, dann:
docker compose up -d

# App verfügbar unter: http://localhost:8080
```

**Erste Anmeldung**: `admin` / `Admin@1234`. Dieses Konto ist so markiert,
dass das Passwort beim ersten Login geändert werden muss — der Server
blockiert jede andere Aktion, bis es geändert wurde, sodass es kein
"Standardpasswort vergessen zu ändern"-Zeitfenster gibt.

---

## Entwicklungs-Setup

```bash
# Nur die Datenbank starten
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # bei Bedarf anpassen
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (leitet /api an :8080 weiter)
```

---

## Sicherheit

- **Keine unsicheren Standardwerte können laufen.** `main.go` ruft beim
  Start `config.Validate()` auf, das den Start verweigert, wenn
  `JWT_SECRET` oder `ENCRYPTION_KEY` leer ist oder noch dem
  veröffentlichten Platzhalter entspricht — eine falsch konfigurierte
  Bereitstellung schlägt laut fehl, statt mit einem fälschbaren
  JWT-Secret oder einem bekannten AES-Schlüssel zu laufen.
- **SSH-Host-Keys werden gepinnt**, nicht ignoriert. `services.ConnectToHost`
  vertraut dem SSH-Schlüssel eines verwalteten Hosts bei der ersten
  Verbindung und speichert dessen Fingerabdruck
  (`db_hosts.ssh_host_key_fingerprint`); jede spätere Verbindung mit
  einem anderen Schlüssel wird abgelehnt, statt stillschweigend vertraut.
  Das Löschen eines Fingerabdrucks nach einem legitimen Host-Rebuild
  erfolgt über einen dedizierten, auditierten Endpunkt
  (`POST /api/hosts/:id/reset-host-key`), niemals über das allgemeine
  Host-Bearbeitungsformular.
- **Login ist ratenbegrenzt**: 10 fehlgeschlagene Versuche von einer IP
  innerhalb von 15 Minuten sperren diese IP für 15 Minuten.
- **Deaktivieren oder Herabstufen eines Benutzers wirkt sofort** — jede
  authentifizierte Anfrage prüft Status/Rolle erneut gegen die
  Datenbank, statt dem zu vertrauen, was ein JWT bis zu 8 Stunden lang
  behauptet hat.
- **CI scannt bei jedem Push/PR und wöchentlich nach bekannten
  verwundbaren Abhängigkeiten** (`.github/workflows/security.yml`):
  `govulncheck` für das Go-Backend, `npm audit` (Produktionsabhängigkeiten)
  für das Frontend.
- Backup-Export, CORS und HTTP-Antwort-Header sind gehärtet — siehe
  `backend/internal/handlers/export.go` (Bereinigung von Pfadsegmenten),
  `backend/internal/server/router.go` (CORS verweigert standardmäßig
  Cross-Origin-Zugriff, sofern `CORS_ALLOWED_ORIGINS` nicht gesetzt ist)
  und `backend/internal/middleware/security_headers.go`.
- **Secrets nach einem Leak rotieren**: `backend/cmd/reencrypt-ssh-keys`
  und [`docs/rotate-secrets.md`](docs/rotate-secrets.md) führen durch die
  Rotation von `JWT_SECRET`/`ENCRYPTION_KEY` und die Neuverschlüsselung
  aller gespeicherten SSH-Schlüssel ohne Datenverlust.

---

## Projektstruktur

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Einstiegspunkt
│   │   └── reencrypt-ssh-keys/         # Secret-Rotationstool (siehe Sicherheit)
│   └── internal/
│       ├── config/                     # Config-Loader + Secret-Validierung beim Start
│       ├── database/                   # GORM-Verbindung
│       ├── models/                     # Alle GORM-Modelle
│       ├── middleware/                 # JWT-Auth, RBAC, Rate-Limiting, Security-Header
│       ├── handlers/                   # Gin-HTTP-Handler
│       │   ├── auth.go                 # Login / Logout / Refresh / Passwort ändern
│       │   ├── hosts.go                # Host-CRUD, SSH-Validierung, Deploy-Key, Reset-Host-Key
│       │   ├── instances.go            # Instanz-Discovery/Scan
│       │   ├── rules.go                # pg_hba-Regel-CRUD + Enable/Disable/Reorder/Clone
│       │   ├── temp_access.go          # Temporären Zugriff gewähren/widerrufen
│       │   ├── backup.go, backup_schedules.go   # Backup erstellen/wiederherstellen/herunterladen + Planung
│       │   ├── export.go               # Config-/Backup-Export (lokal + S3)
│       │   ├── healthcheck.go          # Host-Healthchecks + Benachrichtigungen
│       │   ├── users.go                # Benutzer-CRUD (nur SuperAdmin)
│       │   ├── audit.go                # Audit-Log-Liste + Diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH-Client, Host-Key-Pinning (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # pg_hba-Inhaltserstellung + atomares Schreiben + Ablauf
│       │   ├── crypto.go               # AES-256-GCM für SSH-Schlüssel
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — temporäre Regeln alle N Minuten ablaufen lassen
│       └── server/router.go            # Gin-Router, CORS, alle Routen
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios-API-Client
│       ├── types/index.ts              # TypeScript-Interfaces (spiegelt Go-Modelle)
│       ├── stores/auth.ts              # Pinia-Auth-Store
│       ├── router/index.ts             # Vue Router mit Auth- + Passwort-Änderungs-Guards
│       ├── components/layout/AppLayout.vue
│       └── views/                      # Eine View pro Funktionsbereich
│
├── database/migrations/
│   ├── 001_init.sql                    # Vollständiges Schema mit Enums, Indizes, Triggern
│   └── 002–007_*.sql                   # RBAC-Rollen, Backup-Zeitpläne, Passwort-/Host-Key-Härtung
│
├── docs/rotate-secrets.md              # Runbook zur Secret-Rotation
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-Stage: Vue-Build + Go-Build → alpine
├── docker-compose.yml                  # Produktion
└── docker-compose.dev.yml              # Entwicklung (nur DB)
```

---

## API-Referenz

Die vollständige Routenliste befindet sich in
`backend/internal/server/router.go`. Kern-Endpunkte:

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
POST   /api/hosts/validate              # SSH-Vorab-Check vor dem Speichern
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # Gepinnten SSH-Host-Key-Fingerabdruck löschen

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

## RBAC-Berechtigungen

| Funktion               | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Hosts verwalten        | ✓          | ✓   |          |        |               |                    |
| Instanzen verwalten     | ✓          | ✓   |          |        |               |                    |
| Regel hinzufügen/bearbeiten/löschen | ✓ | ✓   |          |        | ✓             |                    |
| Regel aktivieren/deaktivieren | ✓    | ✓   | ✓        |        | ✓             |                    |
| Regeln anzeigen        | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Temp. Zugriff gewähren/widerrufen | ✓ | ✓  | ✓        |        | ✓             | ✓                  |
| Sicherung & Wiederherstellung | ✓    | ✓   |          |        |               |                    |
| Benutzerverwaltung     | ✓          |     |          |        |               |                    |
| Audit-Log anzeigen     | ✓          | ✓   |          |        |               |                    |

`AccessManager` und `TempAccessManager` sind eingeschränkte Rollen für
Teams, die nur Zugriffsregeln / temporären Zugriff verwalten sollen,
ohne volle DBA-Rechte über Hosts, Instanzen, Backups oder Benutzer.

---

## Produktions-Checkliste

- [ ] Echte `JWT_SECRET`/`ENCRYPTION_KEY` setzen — die App startet sonst nicht
- [ ] Einmal anmelden und das Standard-Admin-Passwort ändern (beim ersten Login erzwungen)
- [ ] Ein starkes `DB_PASSWORD` setzen
- [ ] TLS aktivieren (`DB_SSL_MODE=require`)
- [ ] Die App hinter NGINX/Traefik mit HTTPS betreiben
- [ ] Den Postgres-Port nur für das interne Netzwerk zugänglich machen
- [ ] `CORS_ALLOWED_ORIGINS` nur setzen, wenn das Frontend auf einer anderen Origin als die API bereitgestellt wird
