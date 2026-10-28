# PGHBA Fleet Manager

<!--LANGSWITCH-->

سامانه مدیریت متمرکز `pg_hba.conf` در PostgreSQL — مدیریت قوانین دسترسی
در کل ناوگانی از نمونه‌های PostgreSQL از طریق یک رابط وب واحد، بدون نیاز
به ویرایش دستی `pg_hba.conf` روی هر هاست.

این مخزن یک نسخهٔ سخت‌شده (hardened) و آمادهٔ انتشار عمومی از پروژهٔ
اصلی PGHBA Manager است: الزام تغییر رمز پیش‌فرض، اعتبارسنجی اسرار هنگام
راه‌اندازی، پین‌کردن کلید میزبان SSH، محدودسازی نرخ ورود، و یک خط لولهٔ
CI که در هر push وابستگی‌ها را اسکن می‌کند. بخش [امنیت](#امنیت) در ادامه
را ببینید.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ یک فنجان قهوه برایم مهمان کنید.

---

## تصاویر برنامه

### داشبورد

![Dashboard](docs/screenshots/dashboard.png)

> تصاویر زیر (هاست‌های پایگاه‌داده، نمونه‌های PG، قوانین دسترسی، دسترسی
> موقت، پشتیبان‌گیری و بازیابی، مدیریت کاربران، لاگ حسابرسی، تنظیمات)
> **نمونه‌های تصویری ساختگی (mockup)** هستند که از مؤلفه‌های واقعی
> فرانت‌اند و طرح رنگی همین مخزن ساخته شده‌اند و با داده‌های نمایشی پر
> شده‌اند — نه تصویر یک استقرار واقعی. تصاویر واقعی به‌مرور جایگزین
> آن‌ها خواهند شد.

<details>
<summary>هاست‌های پایگاه‌داده</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>نمونه‌های PG</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>قوانین دسترسی</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>دسترسی موقت</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>پشتیبان‌گیری و بازیابی</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>مدیریت کاربران</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>لاگ حسابرسی</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>تنظیمات</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## شروع سریع (Docker)

```bash
# ۱. کلون و پیکربندی
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# ۲. ساخت .env از روی الگو (مورد استفادهٔ docker-compose.yml)
cp backend/.env.example .env

# ۳. تولید اسرار واقعی — برنامه بدون آن‌ها اجرا نمی‌شود
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# ۴. مقادیر واقعی را در .env ویرایش کنید، سپس:
docker compose up -d

# برنامه در دسترس است در: http://localhost:8080
```

**ورود اولیه**: `admin` / `Admin@1234`. این حساب در اولین ورود مجبور به
تغییر رمز عبور می‌شود — سرور تا زمان تغییر رمز هر اقدام دیگری را مسدود
می‌کند، بنابراین هیچ بازهٔ «فراموش‌کردن تغییر رمز پیش‌فرض» وجود ندارد.

---

## راه‌اندازی محیط توسعه

```bash
# فقط پایگاه‌داده را اجرا کنید
docker compose -f docker-compose.dev.yml up -d

# بک‌اند (Go)
cd backend
cp .env.example .env       # در صورت نیاز ویرایش کنید
go run ./cmd/server

# فرانت‌اند (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (پراکسی /api به :8080)
```

---

## امنیت

- **هیچ مقدار پیش‌فرض ناامنی اجرا نمی‌شود.** `main.go` هنگام
  راه‌اندازی `config.Validate()` را فرا می‌خواند که اگر `JWT_SECRET` یا
  `ENCRYPTION_KEY` خالی باشند یا هنوز همان مقدار نمونهٔ منتشرشده باشند،
  از اجرا خودداری می‌کند — یک استقرار پیکربندی‌نشده به‌جای اجرا با یک
  JWT secret قابل‌جعل یا یک کلید AES شناخته‌شده، با خطا متوقف می‌شود.
- **کلیدهای میزبان SSH پین می‌شوند**، نه نادیده گرفته می‌شوند.
  `services.ConnectToHost` در اولین اتصال به کلید SSH یک میزبان مدیریت‌شده
  اعتماد می‌کند و اثر انگشت آن را ذخیره می‌کند
  (`db_hosts.ssh_host_key_fingerprint`)؛ هر اتصال بعدی با کلیدی متفاوت
  رد می‌شود، نه اینکه به‌طور خاموش پذیرفته شود. پاک‌کردن یک اثر انگشت پس
  از بازسازی مشروع یک میزبان، از طریق یک نقطهٔ پایان اختصاصی و
  ممیزی‌شده (`POST /api/hosts/:id/reset-host-key`) انجام می‌شود، نه از
  طریق فرم عمومی ویرایش هاست.
- **ورود نرخ‌محدود است**: ۱۰ تلاش ناموفق از یک IP ظرف ۱۵ دقیقه، آن IP
  را به‌مدت ۱۵ دقیقه قفل می‌کند.
- **غیرفعال‌کردن یا کاهش نقش یک کاربر بلافاصله اعمال می‌شود** — هر
  درخواست احراز‌هویت‌شده وضعیت/نقش را دوباره در برابر پایگاه‌داده
  بررسی می‌کند، نه اینکه به ادعای یک JWT که ممکن است تا ۸ ساعت قدیمی
  باشد اعتماد کند.
- **CI در هر push/PR و هفتگی وابستگی‌های دارای آسیب‌پذیری شناخته‌شده را
  اسکن می‌کند** (`.github/workflows/security.yml`): `govulncheck` برای
  بک‌اند Go، `npm audit` (وابستگی‌های تولید) برای فرانت‌اند.
- صدور پشتیبان، CORS و هدرهای پاسخ HTTP سخت‌شده‌اند — ببینید
  `backend/internal/handlers/export.go` (پاک‌سازی بخش مسیر)،
  `backend/internal/server/router.go` (CORS به‌طور پیش‌فرض اجازهٔ
  دسترسی cross-origin نمی‌دهد مگر اینکه `CORS_ALLOWED_ORIGINS` تنظیم
  شده باشد)، و `backend/internal/middleware/security_headers.go`.
- **چرخش اسرار پس از نشت**: `backend/cmd/reencrypt-ssh-keys` و
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) مراحل چرخش
  `JWT_SECRET`/`ENCRYPTION_KEY` و رمزگذاری مجدد همهٔ کلیدهای SSH
  ذخیره‌شده را بدون از‌دست‌رفتن داده توضیح می‌دهند.

---

## ساختار پروژه

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # نقطهٔ ورود
│   │   └── reencrypt-ssh-keys/         # ابزار چرخش اسرار (به بخش امنیت مراجعه کنید)
│   └── internal/
│       ├── config/                     # بارگذار پیکربندی + اعتبارسنجی اسرار هنگام راه‌اندازی
│       ├── database/                   # اتصال GORM
│       ├── models/                     # همهٔ مدل‌های GORM
│       ├── middleware/                 # احراز هویت JWT، RBAC، محدودسازی نرخ، هدرهای امنیتی
│       ├── handlers/                   # هندلرهای HTTP در Gin
│       │   ├── auth.go                 # ورود / خروج / تمدید / تغییر رمز
│       │   ├── hosts.go                # CRUD هاست، اعتبارسنجی SSH، استقرار کلید، بازنشانی کلید میزبان
│       │   ├── instances.go            # کشف/اسکن نمونه
│       │   ├── rules.go                # CRUD قوانین pg_hba + فعال/غیرفعال/ترتیب‌دهی/شبیه‌سازی
│       │   ├── temp_access.go          # اعطا/لغو دسترسی موقت
│       │   ├── backup.go, backup_schedules.go   # ایجاد/بازیابی/دانلود پشتیبان + زمان‌بندی
│       │   ├── export.go               # صدور پیکربندی/پشتیبان (محلی + S3)
│       │   ├── healthcheck.go          # بررسی سلامت هاست + اعلان‌ها
│       │   ├── users.go                # CRUD کاربر (فقط SuperAdmin)
│       │   ├── audit.go                # فهرست و تفاوت لاگ حسابرسی
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # کلاینت SSH، پین‌کردن کلید میزبان (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # رندر محتوای pg_hba + نوشتن اتمیک + انقضا
│       │   ├── crypto.go               # AES-256-GCM برای کلیدهای SSH
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — انقضای قوانین موقت هر N دقیقه
│       └── server/router.go            # روتر Gin، CORS، همهٔ مسیرها
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # کلاینت API مبتنی بر Axios
│       ├── types/index.ts              # اینترفیس‌های TypeScript (منعکس‌کنندهٔ مدل‌های Go)
│       ├── stores/auth.ts              # استور احراز هویت Pinia
│       ├── router/index.ts             # Vue Router با نگهبان‌های احراز هویت + تغییر اجباری رمز
│       ├── components/layout/AppLayout.vue
│       └── views/                      # یک ویو برای هر حوزهٔ ویژگی
│
├── database/migrations/
│   ├── 001_init.sql                    # طرحوارهٔ کامل با enum، ایندکس، تریگر
│   └── 002–007_*.sql                   # نقش‌های RBAC، زمان‌بندی پشتیبان، سخت‌سازی رمز/کلید میزبان
│
├── docs/rotate-secrets.md              # راهنمای چرخش اسرار
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # چندمرحله‌ای: ساخت Vue + ساخت Go → alpine
├── docker-compose.yml                  # تولید
└── docker-compose.dev.yml              # توسعه (فقط پایگاه‌داده)
```

---

## مرجع API

فهرست کامل مسیرها در `backend/internal/server/router.go` قرار دارد.
نقاط پایانی اصلی:

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
POST   /api/hosts/validate              # بررسی پیش‌از‌اجرای SSH پیش از ذخیره
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # پاک‌کردن اثر انگشت کلید میزبان پین‌شده

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

## مجوزهای RBAC

| ویژگی                 | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| مدیریت هاست‌ها          | ✓          | ✓   |          |        |               |                    |
| مدیریت نمونه‌ها         | ✓          | ✓   |          |        |               |                    |
| افزودن/ویرایش/حذف قانون | ✓          | ✓   |          |        | ✓             |                    |
| فعال/غیرفعال‌کردن قانون | ✓          | ✓   | ✓        |        | ✓             |                    |
| مشاهدهٔ قوانین          | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| اعطا/لغو دسترسی موقت    | ✓          | ✓   | ✓        |        | ✓             | ✓                  |
| پشتیبان‌گیری و بازیابی  | ✓          | ✓   |          |        |               |                    |
| مدیریت کاربران          | ✓          |     |          |        |               |                    |
| مشاهدهٔ لاگ حسابرسی     | ✓          | ✓   |          |        |               |                    |

`AccessManager` و `TempAccessManager` نقش‌های محدودی هستند برای
تیم‌هایی که فقط باید قوانین دسترسی/دسترسی موقت را مدیریت کنند، بدون
امتیازات کامل DBA روی هاست‌ها، نمونه‌ها، پشتیبان‌ها یا کاربران.

---

## چک‌لیست تولید

- [ ] مقادیر واقعی `JWT_SECRET`/`ENCRYPTION_KEY` را تنظیم کنید — در غیر این صورت برنامه اجرا نمی‌شود
- [ ] یک‌بار وارد شوید و رمز پیش‌فرض ادمین را تغییر دهید (در اولین ورود الزامی است)
- [ ] یک `DB_PASSWORD` قوی تنظیم کنید
- [ ] TLS را فعال کنید (`DB_SSL_MODE=require`)
- [ ] برنامه را پشت NGINX/Traefik با HTTPS قرار دهید
- [ ] پورت Postgres را فقط به شبکهٔ داخلی محدود کنید
- [ ] `CORS_ALLOWED_ORIGINS` را فقط در صورتی تنظیم کنید که فرانت‌اند روی خاستگاهی متفاوت از API مستقر شده باشد
