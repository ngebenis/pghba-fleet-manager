# PGHBA Fleet Manager

🇮🇩 [ID](README.md) · 🇬🇧 [EN](README.en.md) · 🇨🇳 [ZH](README.zh.md) · 🇪🇸 [ES](README.es.md) · 🇸🇦 [AR](README.ar.md) · 🇫🇷 [FR](README.fr.md) · 🇵🇹 [PT](README.pt.md) · 🇧🇩 [BN](README.bn.md) · 🇷🇺 [RU](README.ru.md) · 🇮🇳 [HI](README.hi.md) · 🇹🇭 [TH](README.th.md) · 🇰🇭 [KM](README.km.md) · 🇵🇭 [TL](README.tl.md) · 🇻🇳 [VI](README.vi.md) · 🇯🇵 [JA](README.ja.md) · 🇰🇷 [KO](README.ko.md) · 🇵🇱 **PL** · 🇮🇹 [IT](README.it.md) · 🇹🇷 [TR](README.tr.md) · 🇮🇳 [TA](README.ta.md) · 🇮🇷 [FA](README.fa.md) · 🇺🇦 [UK](README.uk.md) · 🇩🇪 [DE](README.de.md) · 🇳🇱 [NL](README.nl.md)

Scentralizowany system zarządzania `pg_hba.conf` PostgreSQL — zarządzaj
regułami dostępu w całej flocie instancji PostgreSQL z jednego interfejsu
webowego, zamiast ręcznie edytować `pg_hba.conf` na każdym hoście.

To repozytorium to zabezpieczona, gotowa do publikacji wersja oryginalnego
projektu PGHBA Manager: wymuszanie zmiany domyślnego hasła, walidacja
sekretów przy starcie, przypinanie kluczy hosta SSH, ograniczanie liczby
prób logowania oraz pipeline CI skanujący zależności przy każdym pushu.
Zobacz sekcję [Bezpieczeństwo](#bezpieczeństwo) poniżej.

🌐 **24 języki od razu dostępne** — zmień w dowolnym momencie z rozwijanej
listy w pasku górnym: English, Bahasa Indonesia, 中文, Español, العربية,
Français, Português, বাংলা, Русский, हिन्दी, ไทย, ខ្មែរ, Tagalog, Tiếng Việt,
日本語, 한국어, Polski, Italiano, Türkçe, தமிழ், فارسی, Українська, Deutsch,
Nederlands. Arabski i perski są automatycznie renderowane od prawej do lewej.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Postaw mi kawę.

---

## Zrzuty ekranu

### Panel główny

![Dashboard](docs/screenshots/dashboard.png)

> Poniższe zrzuty ekranu (Hosty bazy danych, Instancje PG, Reguły dostępu,
> Dostęp tymczasowy, Kopie zapasowe i przywracanie, Zarządzanie
> użytkownikami, Dziennik audytu, Ustawienia) to **ilustracyjne makiety**
> zbudowane na podstawie rzeczywistych komponentów frontendu i schematu
> kolorów tego repozytorium, wypełnione przykładowymi danymi — to nie jest
> aktywne wdrożenie. Prawdziwe zrzuty ekranu zastąpią je w miarę
> dostępności.

<details>
<summary>Hosty bazy danych</summary>

![Hosty bazy danych](docs/screenshots/hosts.png)
</details>

<details>
<summary>Instancje PG</summary>

![Instancje PG](docs/screenshots/instances.png)
</details>

<details>
<summary>Reguły dostępu</summary>

![Reguły dostępu](docs/screenshots/rules.png)
</details>

<details>
<summary>Dostęp tymczasowy</summary>

![Dostęp tymczasowy](docs/screenshots/temp-access.png)
</details>

<details>
<summary>Kopie zapasowe i przywracanie</summary>

![Kopie zapasowe i przywracanie](docs/screenshots/backup.png)
</details>

<details>
<summary>Zarządzanie użytkownikami</summary>

![Zarządzanie użytkownikami](docs/screenshots/users.png)
</details>

<details>
<summary>Dziennik audytu</summary>

![Dziennik audytu](docs/screenshots/audit.png)
</details>

<details>
<summary>Ustawienia</summary>

![Ustawienia](docs/screenshots/settings.png)
</details>

---

## Szybki start (Docker)

```bash
# 1. Sklonuj i skonfiguruj
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. Utwórz .env z szablonu (używany przez docker-compose.yml)
cp backend/.env.example .env

# 3. Wygeneruj prawdziwe sekrety — aplikacja nie uruchomi się bez nich
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. Edytuj .env swoimi rzeczywistymi wartościami, następnie:
docker compose up -d

# Aplikacja dostępna pod adresem: http://localhost:8080
```

**Pierwsze logowanie**: `admin` / `Admin@1234`. To konto jest oznaczone jako
wymagające zmiany hasła przy pierwszym logowaniu — serwer blokuje każdą
inną akcję, dopóki hasło nie zostanie zmienione, więc nie ma okna czasowego
na "zapomniałem zmienić domyślne hasło".

---

## Konfiguracja środowiska deweloperskiego

```bash
# Uruchom tylko bazę danych
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # dostosuj w razie potrzeby
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (proxy /api do :8080)
```

---

## Bezpieczeństwo

- **Żadne niebezpieczne wartości domyślne nie mogą działać.** `main.go`
  wywołuje `config.Validate()` przy starcie, co odmawia uruchomienia, jeśli
  `JWT_SECRET` lub `ENCRYPTION_KEY` są puste lub nadal mają opublikowaną
  wartość placeholder — błędnie skonfigurowane wdrożenie zawodzi głośno
  zamiast działać z możliwym do sfałszowania sekretem JWT lub znanym
  kluczem AES.
- **Klucze hosta SSH są przypinane**, a nie ignorowane. `services.ConnectToHost`
  ufa kluczowi SSH zarządzanego hosta przy pierwszym połączeniu i zapisuje
  jego odcisk palca (`db_hosts.ssh_host_key_fingerprint`); każde kolejne
  połączenie z innym kluczem jest odrzucane zamiast cicho zaufane.
  Usunięcie odcisku palca po legalnej przebudowie hosta odbywa się przez
  dedykowany, audytowany endpoint (`POST /api/hosts/:id/reset-host-key`),
  nigdy przez standardowy formularz edycji hosta.
- **Logowanie jest ograniczane**: 10 nieudanych prób z jednego adresu IP w
  ciągu 15 minut blokuje ten adres IP na 15 minut.
- **Dezaktywacja lub obniżenie uprawnień użytkownika działa natychmiast** —
  każde uwierzytelnione żądanie ponownie sprawdza status/rolę w bazie
  danych zamiast ufać temu, co token JWT twierdził nawet do 8 godzin
  wcześniej.
- **CI skanuje znane podatne zależności** przy każdym push/PR oraz co
  tydzień (`.github/workflows/security.yml`): `govulncheck` dla backendu
  Go, `npm audit` (zależności produkcyjne) dla frontendu.
- **Migracje bazy danych są osadzone i samoczynnie stosowane.** Każdy plik
  SQL w `backend/internal/database/migrations/` jest wbudowany w binarkę
  (`go:embed`) i stosowany automatycznie — śledzony w tabeli
  `schema_migrations` — przy każdym starcie, więc wdrożenie nigdy nie
  pominie cicho zmiany schematu tak, jak mogłoby to zrobić ręcznie
  uruchamiane skrypt migracyjny.
- Eksport kopii zapasowych, CORS i nagłówki odpowiedzi HTTP są
  zabezpieczone — zobacz `backend/internal/handlers/export.go` (sanityzacja
  segmentów ścieżki), `backend/internal/server/router.go` (CORS domyślnie
  blokuje dostęp cross-origin, chyba że ustawiono
  `CORS_ALLOWED_ORIGINS`) oraz `backend/internal/middleware/security_headers.go`.
- **Rotacja sekretów po wycieku**: `backend/cmd/reencrypt-ssh-keys` oraz
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) prowadzą przez rotację
  `JWT_SECRET`/`ENCRYPTION_KEY` i ponowne szyfrowanie każdego zapisanego
  klucza SSH bez utraty danych.

---

## Struktura projektu

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Punkt wejścia
│   │   └── reencrypt-ssh-keys/         # Narzędzie do rotacji sekretów (patrz Bezpieczeństwo)
│   └── internal/
│       ├── config/                     # Loader konfiguracji + walidacja sekretów przy starcie
│       ├── database/                   # Połączenie GORM + osadzony mechanizm migracji SQL
│       │   └── migrations/             # 001 (bazowa, przez initdb) + 002–008, stosowane automatycznie przy starcie
│       ├── models/                     # Wszystkie modele GORM
│       ├── middleware/                 # Uwierzytelnianie JWT, RBAC, limitowanie, nagłówki bezpieczeństwa
│       ├── handlers/                   # Handlery HTTP Gin
│       │   ├── auth.go                 # Logowanie / wylogowanie / odświeżanie / zmiana hasła
│       │   ├── hosts.go                # CRUD hostów, walidacja SSH, wdrożenie klucza, reset klucza hosta
│       │   ├── instances.go            # Wykrywanie/skanowanie instancji
│       │   ├── rules.go                # CRUD reguł pg_hba + włącz/wyłącz/zmień kolejność/sklonuj
│       │   ├── temp_access.go          # Przyznawanie/odwoływanie dostępu tymczasowego
│       │   ├── backup.go, backup_schedules.go   # Tworzenie/przywracanie/pobieranie kopii zapasowych + harmonogram
│       │   ├── export.go               # Eksport konfiguracji/kopii zapasowych (lokalnie + S3)
│       │   ├── healthcheck.go          # Sprawdzanie stanu hostów + powiadomienia
│       │   ├── users.go                # CRUD użytkowników (tylko SuperAdmin)
│       │   ├── audit.go                # Lista dziennika audytu + różnice
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # Klient SSH, przypinanie kluczy hosta (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # Renderowanie treści pg_hba + atomowy zapis + wygasanie
│       │   ├── crypto.go               # AES-256-GCM dla kluczy SSH
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — wygaszanie reguł tymczasowych co N minut
│       └── server/router.go            # Router Gin, CORS, wszystkie trasy
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Klient API Axios
│       ├── types/index.ts              # Interfejsy TypeScript (odzwierciedlają modele Go)
│       ├── i18n/                       # 24 pliki lokalizacji + rejestr AVAILABLE_LOCALES (obsługa RTL)
│       ├── stores/auth.ts              # Magazyn uwierzytelniania Pinia
│       ├── router/index.ts             # Vue Router z zabezpieczeniami auth + zmiana hasła
│       ├── components/layout/AppLayout.vue
│       └── views/                      # Jeden widok na obszar funkcjonalny
│
├── docs/rotate-secrets.md              # Przewodnik rotacji sekretów
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Wieloetapowy: build Vue + build Go → alpine
├── docker-compose.yml                  # Produkcja
└── docker-compose.dev.yml              # Deweloperski (tylko DB)
```

---

## Dokumentacja API

Pełna lista tras znajduje się w `backend/internal/server/router.go`. Główne
endpointy:

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

## Uprawnienia RBAC

| Funkcja                | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Zarządzanie hostami     | ✓          | ✓   |          |        |               |                    |
| Zarządzanie instancjami | ✓          | ✓   |          |        |               |                    |
| Dodaj/edytuj/usuń regułę | ✓        | ✓   |          |        | ✓             |                    |
| Włącz/wyłącz regułę     | ✓          | ✓   | ✓        |        | ✓             |                    |
| Przeglądaj reguły       | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Przyznaj/odwołaj dostęp tymczasowy | ✓ | ✓   | ✓        |        | ✓             | ✓                  |
| Kopie zapasowe i przywracanie | ✓    | ✓   |          |        |               |                    |
| Zarządzanie użytkownikami | ✓        |     |          |        |               |                    |
| Przeglądaj dziennik audytu | ✓       | ✓   |          |        |               |                    |

`AccessManager` i `TempAccessManager` to ograniczone role dla zespołów,
które powinny zarządzać tylko regułami dostępu / dostępem tymczasowym, bez
pełnych uprawnień DBA nad hostami, instancjami, kopiami zapasowymi czy
użytkownikami.

---

## Lista kontrolna wdrożenia produkcyjnego

- [ ] Ustaw prawdziwe `JWT_SECRET`/`ENCRYPTION_KEY` — aplikacja inaczej się nie uruchomi
- [ ] Zaloguj się raz i zmień domyślne hasło administratora (wymuszane przy pierwszym logowaniu)
- [ ] Ustaw silne `DB_PASSWORD`
- [ ] Włącz TLS (`DB_SSL_MODE=require`)
- [ ] Umieść aplikację za NGINX/Traefik z HTTPS
- [ ] Ogranicz port Postgresa tylko do sieci wewnętrznej
- [ ] Ustaw `CORS_ALLOWED_ORIGINS` tylko jeśli frontend jest wdrożony na innej domenie niż API
