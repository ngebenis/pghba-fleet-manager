# PGHBA Fleet Manager

🇮🇩 [ID](README.md) · 🇬🇧 [EN](README.en.md) · 🇨🇳 [ZH](README.zh.md) · 🇪🇸 [ES](README.es.md) · 🇸🇦 [AR](README.ar.md) · 🇫🇷 [FR](README.fr.md) · 🇵🇹 [PT](README.pt.md) · 🇧🇩 **BN** · 🇷🇺 [RU](README.ru.md) · 🇮🇳 [HI](README.hi.md) · 🇹🇭 [TH](README.th.md) · 🇰🇭 [KM](README.km.md) · 🇵🇭 [TL](README.tl.md) · 🇻🇳 [VI](README.vi.md) · 🇯🇵 [JA](README.ja.md) · 🇰🇷 [KO](README.ko.md) · 🇵🇱 [PL](README.pl.md) · 🇮🇹 [IT](README.it.md) · 🇹🇷 [TR](README.tr.md) · 🇮🇳 [TA](README.ta.md) · 🇮🇷 [FA](README.fa.md) · 🇺🇦 [UK](README.uk.md) · 🇩🇪 [DE](README.de.md) · 🇳🇱 [NL](README.nl.md)

PostgreSQL `pg_hba.conf` কেন্দ্রীভূত ব্যবস্থাপনা সিস্টেম — প্রতিটি হোস্টে ম্যানুয়ালি
`pg_hba.conf` এডিট করার বদলে, একটি মাত্র ওয়েব UI থেকে পুরো PostgreSQL
ইনস্ট্যান্স ফ্লিট জুড়ে অ্যাক্সেস নিয়ম পরিচালনা করুন।

এই রিপোজিটরিটি মূল PGHBA Manager প্রকল্পের একটি কঠোরভাবে সুরক্ষিত এবং
পাবলিকের জন্য প্রস্তুত স্ন্যাপশট: ডিফল্ট পাসওয়ার্ড পরিবর্তন বাধ্যতামূলককরণ,
স্টার্টআপে সিক্রেট যাচাই, SSH হোস্ট কী পিনিং, লগইন রেট লিমিটিং, এবং প্রতিটি
পুশে ডিপেন্ডেন্সি স্ক্যান করা একটি CI পাইপলাইন। নিচে [নিরাপত্তা](#নিরাপত্তা)
অংশ দেখুন।

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ আমাকে এক কাপ কফি খাওয়ান।

---

## স্ক্রিনশট

### ড্যাশবোর্ড

![ড্যাশবোর্ড](docs/screenshots/dashboard.png)

> নিচের স্ক্রিনশটগুলো (ডেটাবেস হোস্ট, PG ইনস্ট্যান্স, অ্যাক্সেস নিয়ম, সাময়িক
> অ্যাক্সেস, ব্যাকআপ ও পুনরুদ্ধার, ব্যবহারকারী ব্যবস্থাপনা, অডিট লগ, সেটিংস)
> এই রিপোজিটরির প্রকৃত ফ্রন্টএন্ড কম্পোনেন্ট ও রঙের স্কিম থেকে তৈরি
> **উদাহরণমূলক মকআপ**, যেগুলোতে ডামি/নমুনা ডেটা ভরা হয়েছে — এটি কোনো
> লাইভ ডিপ্লয়মেন্ট নয়। প্রকৃত স্ক্রিনশট উপলব্ধ হলে এগুলো প্রতিস্থাপন করা হবে।

<details>
<summary>ডেটাবেস হোস্ট</summary>

![ডেটাবেস হোস্ট](docs/screenshots/hosts.png)
</details>

<details>
<summary>PG ইনস্ট্যান্স</summary>

![PG ইনস্ট্যান্স](docs/screenshots/instances.png)
</details>

<details>
<summary>অ্যাক্সেস নিয়ম</summary>

![অ্যাক্সেস নিয়ম](docs/screenshots/rules.png)
</details>

<details>
<summary>সাময়িক অ্যাক্সেস</summary>

![সাময়িক অ্যাক্সেস](docs/screenshots/temp-access.png)
</details>

<details>
<summary>ব্যাকআপ ও পুনরুদ্ধার</summary>

![ব্যাকআপ ও পুনরুদ্ধার](docs/screenshots/backup.png)
</details>

<details>
<summary>ব্যবহারকারী ব্যবস্থাপনা</summary>

![ব্যবহারকারী ব্যবস্থাপনা](docs/screenshots/users.png)
</details>

<details>
<summary>অডিট লগ</summary>

![অডিট লগ](docs/screenshots/audit.png)
</details>

<details>
<summary>সেটিংস</summary>

![সেটিংস](docs/screenshots/settings.png)
</details>

---

## দ্রুত শুরু (Docker)

```bash
# ১. ক্লোন করুন এবং কনফিগার করুন
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# ২. টেমপ্লেট থেকে .env তৈরি করুন (docker-compose.yml দ্বারা ব্যবহৃত)
cp backend/.env.example .env

# ৩. প্রকৃত সিক্রেট তৈরি করুন — এগুলো ছাড়া অ্যাপ চালু হবে না
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# ৪. আপনার প্রকৃত মান দিয়ে .env এডিট করুন, তারপর:
docker compose up -d

# অ্যাপ উপলব্ধ: http://localhost:8080
```

**প্রথম লগইন**: `admin` / `Admin@1234`। এই অ্যাকাউন্টটি প্রথম লগইনে পাসওয়ার্ড
পরিবর্তন করতে বাধ্য করা হয়েছে — পরিবর্তন না করা পর্যন্ত সার্ভার অন্য সব
কার্যক্রম ব্লক করে, তাই "ডিফল্ট পাসওয়ার্ড পরিবর্তন করতে ভুলে গেছি" এমন
পরিস্থিতির কোনো সুযোগ নেই।

---

## ডেভেলপমেন্ট সেটআপ

```bash
# শুধু ডেটাবেস চালু করুন
docker compose -f docker-compose.dev.yml up -d

# ব্যাকএন্ড (Go)
cd backend
cp .env.example .env       # প্রয়োজন অনুযায়ী এডিট করুন
go run ./cmd/server

# ফ্রন্টএন্ড (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (/api কে :8080-এ প্রক্সি করে)
```

---

## নিরাপত্তা

- **কোনো অনিরাপদ ডিফল্ট চলতে পারে না।** `main.go` স্টার্টআপে
  `config.Validate()` কল করে, যা `JWT_SECRET` বা `ENCRYPTION_KEY` খালি
  থাকলে বা এখনও প্রকাশিত প্লেসহোল্ডার হলে চালু হতে অস্বীকার করে —
  ভুল কনফিগার করা ডিপ্লয়মেন্ট জালিয়াতিযোগ্য JWT সিক্রেট বা পরিচিত AES
  কী নিয়ে চলার বদলে স্পষ্টভাবে ব্যর্থ হয়।
- **SSH হোস্ট কী পিন করা হয়**, উপেক্ষা করা হয় না। `services.ConnectToHost`
  প্রথম সংযোগে একটি পরিচালিত হোস্টের SSH কী বিশ্বাস করে এবং এর ফিঙ্গারপ্রিন্ট
  সংরক্ষণ করে (`db_hosts.ssh_host_key_fingerprint`); ভিন্ন কী দিয়ে পরবর্তী
  যেকোনো সংযোগ নীরবে বিশ্বাস করার বদলে প্রত্যাখ্যান করা হয়। বৈধ হোস্ট
  পুনর্নির্মাণের পর ফিঙ্গারপ্রিন্ট মোছা একটি নিবেদিত, নিরীক্ষিত এন্ডপয়েন্ট
  (`POST /api/hosts/:id/reset-host-key`) দিয়ে করা হয়, সাধারণ হোস্ট-এডিট
  ফর্ম দিয়ে কখনো নয়।
- **লগইন রেট-লিমিটেড**: একই IP থেকে ১৫ মিনিটে ১০টি ব্যর্থ চেষ্টা সেই IP-কে
  ১৫ মিনিটের জন্য লক করে দেয়।
- **একজন ব্যবহারকারীকে নিষ্ক্রিয় বা ডিমোট করা তাৎক্ষণিক কার্যকর হয়** — প্রতিটি
  প্রমাণীকৃত রিকোয়েস্ট ৮ ঘণ্টা আগে JWT-তে দাবি করা তথ্যের বদলে ডেটাবেসে
  status/role পুনরায় যাচাই করে।
- **CI পরিচিত দুর্বলতাযুক্ত ডিপেন্ডেন্সি স্ক্যান করে** প্রতিটি push/PR-এ এবং
  সাপ্তাহিকভাবে (`.github/workflows/security.yml`): Go ব্যাকএন্ডের জন্য
  `govulncheck`, ফ্রন্টএন্ডের জন্য `npm audit` (প্রোডাকশন ডিপেন্ডেন্সি)।
- ব্যাকআপ এক্সপোর্ট, CORS এবং HTTP রেসপন্স হেডার সুরক্ষিত করা হয়েছে —
  দেখুন `backend/internal/handlers/export.go` (পাথ-সেগমেন্ট স্যানিটাইজেশন),
  `backend/internal/server/router.go` (`CORS_ALLOWED_ORIGINS` সেট না করলে
  CORS ডিফল্টভাবে কোনো ক্রস-অরিজিন অ্যাক্সেস দেয় না), এবং
  `backend/internal/middleware/security_headers.go`।
- **ফাঁস হওয়ার পর সিক্রেট রোটেশন**: `backend/cmd/reencrypt-ssh-keys` এবং
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) ডেটা হারানো ছাড়াই
  `JWT_SECRET`/`ENCRYPTION_KEY` রোটেট করা এবং সংরক্ষিত প্রতিটি SSH কী পুনরায়
  এনক্রিপ্ট করার পদ্ধতি দেখায়।

---

## প্রজেক্ট স্ট্রাকচার

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # এন্ট্রি পয়েন্ট
│   │   └── reencrypt-ssh-keys/         # সিক্রেট-রোটেশন টুল (নিরাপত্তা দেখুন)
│   └── internal/
│       ├── config/                     # কনফিগ লোডার + স্টার্টআপ সিক্রেট যাচাই
│       ├── database/                   # GORM সংযোগ
│       ├── models/                     # সব GORM মডেল
│       ├── middleware/                 # JWT auth, RBAC, রেট লিমিটিং, সিকিউরিটি হেডার
│       ├── handlers/                   # Gin HTTP হ্যান্ডলার
│       │   ├── auth.go                 # Login / logout / refresh / change-password
│       │   ├── hosts.go                # হোস্ট CRUD, SSH validate, deploy-key, reset-host-key
│       │   ├── instances.go            # ইনস্ট্যান্স discovery/scan
│       │   ├── rules.go                # pg_hba নিয়ম CRUD + enable/disable/reorder/clone
│       │   ├── temp_access.go          # সাময়িক অ্যাক্সেস grant/revoke
│       │   ├── backup.go, backup_schedules.go   # ব্যাকআপ create/restore/download + শিডিউলিং
│       │   ├── export.go               # কনফিগ/ব্যাকআপ এক্সপোর্ট (local + S3)
│       │   ├── healthcheck.go          # হোস্ট হেলথ চেক + নোটিফিকেশন
│       │   ├── users.go                # ব্যবহারকারী CRUD (শুধু SuperAdmin)
│       │   ├── audit.go                # অডিট লগ তালিকা + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH client, হোস্ট কী পিনিং (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # pg_hba কন্টেন্ট রেন্ডারিং + atomic write + expiry
│       │   ├── crypto.go               # SSH কীর জন্য AES-256-GCM
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — প্রতি N মিনিটে সাময়িক নিয়মের মেয়াদ শেষ করে
│       └── server/router.go            # Gin router, CORS, সব রুট
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API ক্লায়েন্ট
│       ├── types/index.ts              # TypeScript ইন্টারফেস (Go মডেলের প্রতিফলন)
│       ├── stores/auth.ts              # Pinia auth store
│       ├── router/index.ts             # auth + force-password-change গার্ডসহ Vue Router
│       ├── components/layout/AppLayout.vue
│       └── views/                      # প্রতিটি ফিচার এলাকার জন্য একটি view
│
├── database/migrations/
│   ├── 001_init.sql                    # enum, index, trigger সহ সম্পূর্ণ স্কিমা
│   └── 002–007_*.sql                   # RBAC রোল, ব্যাকআপ শিডিউল, পাসওয়ার্ড/হোস্ট-কী হার্ডেনিং
│
├── docs/rotate-secrets.md              # সিক্রেট রোটেশন রানবুক
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-stage: Vue build + Go build → alpine
├── docker-compose.yml                  # প্রোডাকশন
└── docker-compose.dev.yml              # ডেভেলপমেন্ট (শুধু DB)
```

---

## API রেফারেন্স

সম্পূর্ণ রুট তালিকা `backend/internal/server/router.go`-তে আছে। মূল
এন্ডপয়েন্টসমূহ:

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

## RBAC অনুমতি

| ফিচার                  | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| হোস্ট পরিচালনা          | ✓          | ✓   |          |        |               |                    |
| ইনস্ট্যান্স পরিচালনা     | ✓          | ✓   |          |        |               |                    |
| নিয়ম যোগ/সম্পাদনা/মোছা   | ✓          | ✓   |          |        | ✓             |                    |
| নিয়ম সক্রিয়/নিষ্ক্রিয়    | ✓          | ✓   | ✓        |        | ✓             |                    |
| নিয়ম দেখা               | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| সাময়িক অ্যাক্সেস প্রদান/বাতিল | ✓    | ✓   | ✓        |        | ✓             | ✓                  |
| ব্যাকআপ ও পুনরুদ্ধার      | ✓          | ✓   |          |        |               |                    |
| ব্যবহারকারী ব্যবস্থাপনা   | ✓          |     |          |        |               |                    |
| অডিট লগ দেখা            | ✓          | ✓   |          |        |               |                    |

`AccessManager` এবং `TempAccessManager` হলো সীমিত পরিসরের রোল, যেসব দলকে
হোস্ট, ইনস্ট্যান্স, ব্যাকআপ বা ব্যবহারকারীর ওপর সম্পূর্ণ DBA সুবিধা ছাড়াই শুধু
অ্যাক্সেস নিয়ম / সাময়িক অ্যাক্সেস পরিচালনা করতে হয়, তাদের জন্য।

---

## প্রোডাকশন চেকলিস্ট

- [ ] প্রকৃত `JWT_SECRET`/`ENCRYPTION_KEY` সেট করুন — এগুলো ছাড়া অ্যাপ চালু হবে না
- [ ] একবার লগইন করে ডিফল্ট অ্যাডমিন পাসওয়ার্ড পরিবর্তন করুন (প্রথম লগইনে বাধ্যতামূলক)
- [ ] শক্তিশালী `DB_PASSWORD` সেট করুন
- [ ] TLS সক্রিয় করুন (`DB_SSL_MODE=require`)
- [ ] HTTPS সহ NGINX/Traefik-এর পেছনে অ্যাপটি রাখুন
- [ ] Postgres-এর পোর্ট শুধু ইন্টারনাল নেটওয়ার্কে সীমাবদ্ধ রাখুন
- [ ] ফ্রন্টএন্ড API-এর থেকে ভিন্ন origin-এ ডিপ্লয় করা হলেই কেবল `CORS_ALLOWED_ORIGINS` সেট করুন
