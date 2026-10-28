# PGHBA Fleet Manager

🇮🇩 **ID** · 🇬🇧 [EN](README.en.md) · 🇨🇳 [ZH](README.zh.md) · 🇪🇸 [ES](README.es.md) · 🇸🇦 [AR](README.ar.md) · 🇫🇷 [FR](README.fr.md) · 🇵🇹 [PT](README.pt.md) · 🇧🇩 [BN](README.bn.md) · 🇷🇺 [RU](README.ru.md) · 🇮🇳 [HI](README.hi.md) · 🇹🇭 [TH](README.th.md) · 🇰🇭 [KM](README.km.md) · 🇵🇭 [TL](README.tl.md) · 🇻🇳 [VI](README.vi.md) · 🇯🇵 [JA](README.ja.md) · 🇰🇷 [KO](README.ko.md) · 🇵🇱 [PL](README.pl.md) · 🇮🇹 [IT](README.it.md) · 🇹🇷 [TR](README.tr.md) · 🇮🇳 [TA](README.ta.md) · 🇮🇷 [FA](README.fa.md) · 🇺🇦 [UK](README.uk.md) · 🇩🇪 [DE](README.de.md) · 🇳🇱 [NL](README.nl.md)

Sistem manajemen terpusat untuk `pg_hba.conf` PostgreSQL — kelola aturan
akses di seluruh fleet instance PostgreSQL dari satu web UI, tanpa perlu
mengedit `pg_hba.conf` secara manual di setiap host.

Repository ini adalah snapshot yang sudah di-hardening dan siap untuk
publik dari proyek PGHBA Manager: penegakan ganti password default,
validasi secret saat startup, SSH host key pinning, rate limiting login,
dan pipeline CI yang memindai dependency di setiap push. Lihat bagian
[Keamanan](#keamanan) di bawah.

🌐 **24 bahasa siap pakai** — ganti kapan saja lewat dropdown di topbar:
English, Bahasa Indonesia, 中文, Español, العربية, Français, Português,
বাংলা, Русский, हिन्दी, ไทย, ខ្មែរ, Tagalog, Tiếng Việt, 日本語, 한국어,
Polski, Italiano, Türkçe, தமிழ், فارسی, Українська, Deutsch, Nederlands.
Arab dan Persia otomatis dirender right-to-left.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Silakan traktir saya dengan secangkir kopi.

---

## Screenshot

### Dashboard

![Dashboard](docs/screenshots/dashboard.png)

> Screenshot di bawah ini (Database Hosts, PG Instances, Access Rules,
> Temporary Access, Backup & Restore, User Management, Audit Log,
> Settings) adalah **mockup ilustratif** yang dibuat berdasarkan komponen
> frontend dan skema warna asli repo ini, diisi dengan data
> dummy/contoh — bukan hasil tangkapan layar dari deployment yang
> berjalan. Screenshot asli akan menggantikan ini seiring waktu.

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
# 1. Clone dan konfigurasi
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. Buat .env dari template (dipakai oleh docker-compose.yml)
cp backend/.env.example .env

# 3. Generate secret asli — aplikasi akan menolak start tanpa ini
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. Edit .env dengan nilai asli Anda, lalu:
docker compose up -d

# Aplikasi tersedia di: http://localhost:8080
```

**Login pertama**: `admin` / `Admin@1234`. Akun ini diwajibkan untuk
mengganti password pada login pertama — server akan memblokir semua
aksi lain sampai password diganti, sehingga tidak ada celah "lupa
mengganti password default".

---

## Setup Development

```bash
# Jalankan hanya database
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # sesuaikan jika perlu
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (proxy /api ke :8080)
```

---

## Keamanan

- **Tidak ada default yang tidak aman yang bisa jalan.** `main.go`
  memanggil `config.Validate()` saat startup, yang menolak menyala jika
  `JWT_SECRET` atau `ENCRYPTION_KEY` kosong atau masih nilai placeholder
  yang dipublikasikan — deployment yang salah konfigurasi akan gagal
  secara jelas, bukan berjalan dengan JWT secret yang bisa dipalsukan
  atau AES key yang sudah diketahui publik.
- **SSH host key di-pin**, bukan diabaikan. `services.ConnectToHost`
  mempercayai SSH key sebuah host pada koneksi pertama dan menyimpan
  fingerprint-nya (`db_hosts.ssh_host_key_fingerprint`); koneksi
  berikutnya dengan key yang berbeda akan ditolak, bukan dipercaya diam-
  diam. Menghapus fingerprint setelah rebuild host yang sah dilakukan
  lewat endpoint khusus yang tercatat di audit
  (`POST /api/hosts/:id/reset-host-key`), bukan lewat form edit host
  biasa.
- **Login dibatasi rate-nya**: 10 kali gagal login dari satu IP dalam 15
  menit akan mengunci IP tersebut selama 15 menit.
- **Menonaktifkan atau menurunkan role user langsung berlaku** — setiap
  request yang terautentikasi akan mengecek ulang status/role ke
  database, bukan mempercayai klaim JWT yang bisa berumur sampai 8 jam.
- **CI memindai dependency yang diketahui rentan** di setiap push/PR dan
  setiap minggu (`.github/workflows/security.yml`): `govulncheck` untuk
  backend Go, `npm audit` (dependency produksi) untuk frontend.
- **Migration database ter-embed dan jalan otomatis.** Setiap file SQL di
  `backend/internal/database/migrations/` dibakukan ke dalam binary
  (`go:embed`) dan diterapkan otomatis — dilacak lewat tabel
  `schema_migrations` — di setiap startup, jadi deployment tidak akan
  pernah diam-diam melewati perubahan skema seperti yang bisa terjadi
  kalau migration dijalankan manual.
- Export backup, CORS, dan header response HTTP sudah di-hardening —
  lihat `backend/internal/handlers/export.go` (sanitasi path segment),
  `backend/internal/server/router.go` (CORS default-nya menolak akses
  cross-origin kecuali `CORS_ALLOWED_ORIGINS` di-set), dan
  `backend/internal/middleware/security_headers.go`.
- **Rotasi secret setelah ada kebocoran**: `backend/cmd/reencrypt-ssh-keys`
  dan [`docs/rotate-secrets.md`](docs/rotate-secrets.md) memandu Anda
  merotasi `JWT_SECRET`/`ENCRYPTION_KEY` dan meng-enkripsi ulang semua
  SSH key yang tersimpan tanpa kehilangan data.

---

## Struktur Project

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Entry point
│   │   └── reencrypt-ssh-keys/         # Tool rotasi secret (lihat Keamanan)
│   └── internal/
│       ├── config/                     # Loader config + validasi secret saat startup
│       ├── database/                   # Koneksi GORM + embedded SQL migration runner
│       │   └── migrations/             # 001 (baseline, via initdb) + 002–008, otomatis jalan saat startup
│       ├── models/                     # Semua model GORM
│       ├── middleware/                 # JWT auth, RBAC, rate limiting, security headers
│       ├── handlers/                   # Handler HTTP Gin
│       │   ├── auth.go                 # Login / logout / refresh / change-password
│       │   ├── hosts.go                # CRUD host, validasi SSH, deploy-key, reset-host-key
│       │   ├── instances.go            # Discovery/scan instance
│       │   ├── rules.go                # CRUD rule pg_hba + enable/disable/reorder/clone
│       │   ├── temp_access.go          # Grant/revoke temporary access
│       │   ├── backup.go, backup_schedules.go   # Create/restore/download backup + penjadwalan
│       │   ├── export.go               # Export config/backup (lokal + S3)
│       │   ├── healthcheck.go          # Health check host + notifikasi
│       │   ├── users.go                # CRUD user (khusus SuperAdmin)
│       │   ├── audit.go                # List audit log + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH client, host key pinning (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # Rendering konten pg_hba + atomic write + expiry
│       │   ├── crypto.go               # AES-256-GCM untuk SSH key
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — expire rule sementara setiap N menit
│       └── server/router.go            # Router Gin, CORS, semua route
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Klien API Axios
│       ├── types/index.ts              # Interface TypeScript (mencerminkan model Go)
│       ├── i18n/                       # 24 file locale + registry AVAILABLE_LOCALES (RTL-aware)
│       ├── stores/auth.ts              # Pinia auth store
│       ├── router/index.ts             # Vue Router dengan guard auth + force-password-change
│       ├── components/layout/AppLayout.vue
│       └── views/                      # Satu view per area fitur
│
├── docs/rotate-secrets.md              # Runbook rotasi secret
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-stage: build Vue + build Go → alpine
├── docker-compose.yml                  # Produksi
└── docker-compose.dev.yml              # Development (DB saja)
```

---

## Referensi API

Daftar route lengkap ada di `backend/internal/server/router.go`. Endpoint
inti:

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
POST   /api/hosts/validate              # Cek SSH pre-flight sebelum disimpan
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # Hapus fingerprint SSH host key yang di-pin

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

## Hak Akses RBAC

| Fitur                  | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Kelola Host            | ✓          | ✓   |          |        |               |                    |
| Kelola Instance        | ✓          | ✓   |          |        |               |                    |
| Tambah/Edit/Hapus Rule | ✓          | ✓   |          |        | ✓             |                    |
| Enable/Disable Rule    | ✓          | ✓   | ✓        |        | ✓             |                    |
| Lihat Rule             | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Grant/Revoke Temp Access | ✓        | ✓   | ✓        |        | ✓             | ✓                  |
| Backup & Restore       | ✓          | ✓   |          |        |               |                    |
| Manajemen User         | ✓          |     |          |        |               |                    |
| Lihat Audit Log        | ✓          | ✓   |          |        |               |                    |

`AccessManager` dan `TempAccessManager` adalah role terbatas untuk tim
yang hanya perlu mengelola rule akses / temporary access tanpa privilege
DBA penuh atas host, instance, backup, atau user.

---

## Checklist Produksi

- [ ] Set `JWT_SECRET`/`ENCRYPTION_KEY` yang asli — aplikasi tidak akan start tanpa ini
- [ ] Login sekali dan ganti password default admin (diwajibkan saat login pertama)
- [ ] Set `DB_PASSWORD` yang kuat
- [ ] Aktifkan TLS (`DB_SSL_MODE=require`)
- [ ] Letakkan aplikasi di belakang NGINX/Traefik dengan HTTPS
- [ ] Batasi port Postgres hanya untuk internal network
- [ ] Set `CORS_ALLOWED_ORIGINS` hanya jika frontend di-deploy di origin yang berbeda dari API
