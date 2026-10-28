# PGHBA Fleet Manager

🇮🇩 [ID](README.md) · 🇬🇧 [EN](README.en.md) · 🇨🇳 [ZH](README.zh.md) · 🇪🇸 [ES](README.es.md) · 🇸🇦 **AR** · 🇫🇷 [FR](README.fr.md) · 🇵🇹 [PT](README.pt.md) · 🇧🇩 [BN](README.bn.md) · 🇷🇺 [RU](README.ru.md) · 🇮🇳 [HI](README.hi.md) · 🇹🇭 [TH](README.th.md) · 🇰🇭 [KM](README.km.md) · 🇵🇭 [TL](README.tl.md) · 🇻🇳 [VI](README.vi.md) · 🇯🇵 [JA](README.ja.md) · 🇰🇷 [KO](README.ko.md) · 🇵🇱 [PL](README.pl.md) · 🇮🇹 [IT](README.it.md) · 🇹🇷 [TR](README.tr.md) · 🇮🇳 [TA](README.ta.md) · 🇮🇷 [FA](README.fa.md) · 🇺🇦 [UK](README.uk.md) · 🇩🇪 [DE](README.de.md) · 🇳🇱 [NL](README.nl.md)

نظام مركزي لإدارة ملف `pg_hba.conf` الخاص بـ PostgreSQL — يتيح لك إدارة
قواعد الوصول عبر أسطول كامل من نسخ PostgreSQL من واجهة ويب واحدة، بدلاً
من تحرير `pg_hba.conf` يدويًا على كل مضيف.

هذا المستودع هو نسخة معزَّزة أمنيًا وجاهزة للنشر العام من مشروع PGHBA
Manager الأصلي: فرض تغيير كلمة المرور الافتراضية، التحقق من الأسرار عند
بدء التشغيل، تثبيت بصمة مفتاح مضيف SSH، تحديد معدل تسجيل الدخول، وخط
أنابيب CI يفحص الاعتماديات في كل عملية دفع. راجع قسم [الأمان](#security)
أدناه.

🌐 **24 لغة جاهزة للاستخدام** — يمكنك التبديل بينها في أي وقت من القائمة
المنسدلة في الشريط العلوي: English، Bahasa Indonesia، 中文، Español،
العربية، Français، Português، বাংলা، Русский، हिन्दी، ไทย، ខ្មែរ،
Tagalog، Tiếng Việt، 日本語، 한국어، Polski، Italiano، Türkçe، தமிழ்،
فارسی، Українська، Deutsch، Nederlands. تُعرَض العربية والفارسية
تلقائيًا من اليمين إلى اليسار.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ تفضل بدعوتي لفنجان قهوة.

---

## لقطات الشاشة

### لوحة التحكم

![Dashboard](docs/screenshots/dashboard.png)

> لقطات الشاشة أدناه (مضيفو قواعد البيانات، نسخ PG، قواعد الوصول، الوصول
> المؤقت، النسخ الاحتياطي والاستعادة، إدارة المستخدمين، سجل التدقيق،
> الإعدادات) هي **نماذج توضيحية** مبنية على مكونات الواجهة الأمامية
> الفعلية لهذا المستودع ونظام الألوان الخاص به، مع بيانات تجريبية/وهمية —
> وليست نشرًا حيًا. سيتم استبدالها بلقطات شاشة حقيقية عند توفرها.

<details>
<summary>مضيفو قواعد البيانات</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>نسخ PG</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>قواعد الوصول</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>الوصول المؤقت</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>النسخ الاحتياطي والاستعادة</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>إدارة المستخدمين</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>سجل التدقيق</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>الإعدادات</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## البدء السريع (Docker)

```bash
# 1. استنساخ المستودع وإعداده
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. إنشاء ملف .env من القالب (يُستخدم بواسطة docker-compose.yml)
cp backend/.env.example .env

# 3. إنشاء أسرار حقيقية — لن يبدأ التطبيق بدونها
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. عدّل .env بقيمك الحقيقية، ثم:
docker compose up -d

# التطبيق متاح على: http://localhost:8080
```

**أول تسجيل دخول**: `admin` / `Admin@1234`. هذا الحساب مُعلَّم إجباريًا
لتغيير كلمة المرور عند أول تسجيل دخول — يمنع الخادم أي إجراء آخر حتى
يتم تغييرها، لذا لا توجد فرصة لنسيان تغيير كلمة المرور الافتراضية.

---

## إعداد بيئة التطوير

```bash
# تشغيل قاعدة البيانات فقط
docker compose -f docker-compose.dev.yml up -d

# الواجهة الخلفية (Go)
cd backend
cp .env.example .env       # عدّل حسب الحاجة
go run ./cmd/server

# الواجهة الأمامية (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (يوجّه /api إلى :8080)
```

---

## الأمان

- **لا يمكن تشغيل أي إعدادات افتراضية غير آمنة.** يستدعي `main.go` دالة
  `config.Validate()` عند بدء التشغيل، والتي ترفض البدء إذا كان
  `JWT_SECRET` أو `ENCRYPTION_KEY` فارغًا أو ما زال القيمة الافتراضية
  المنشورة — فالنشر الخاطئ يفشل بوضوح بدلاً من العمل بسرّ JWT قابل
  للتزوير أو مفتاح AES معروف.
- **بصمات مفاتيح مضيفي SSH مثبّتة**، لا يتم تجاهلها. تقوم
  `services.ConnectToHost` بالثقة ببصمة مفتاح SSH الخاص بالمضيف المُدار
  عند أول اتصال وتخزينها (`db_hosts.ssh_host_key_fingerprint`)؛ أي اتصال
  لاحق بمفتاح مختلف يُرفض بدلاً من قبوله بصمت. حذف بصمة بعد إعادة بناء
  مضيف شرعي يتم عبر نقطة نهاية مخصّصة وموثَّقة في سجل التدقيق
  (`POST /api/hosts/:id/reset-host-key`)، وليس عبر نموذج تعديل المضيف
  العام.
- **تسجيل الدخول مقيّد بمعدل**: 10 محاولات فاشلة من نفس عنوان IP خلال 15
  دقيقة تؤدي إلى حظر ذلك العنوان لمدة 15 دقيقة.
- **تعطيل أو خفض صلاحية مستخدم يسري فورًا** — يُعيد كل طلب مصادَق عليه
  التحقق من الحالة/الدور في قاعدة البيانات بدلاً من الوثوق بما ادّعاه
  رمز JWT قد يصل عمره إلى 8 ساعات.
- **يفحص CI الاعتماديات المعروفة بالثغرات** في كل دفع/طلب سحب وأسبوعيًا
  (`.github/workflows/security.yml`): `govulncheck` للواجهة الخلفية Go،
  و`npm audit` (اعتماديات الإنتاج) للواجهة الأمامية.
- **ترحيلات قاعدة البيانات مضمَّنة في البرنامج وتُطبَّق تلقائيًا.** كل
  ملف SQL في `backend/internal/database/migrations/` يُضمَّن داخل الملف
  التنفيذي (`go:embed`) ويُطبَّق تلقائيًا — مع تتبعه عبر جدول
  `schema_migrations` — في كل تشغيل، بحيث لا يمكن لأي نشر أن يتخطى
  بصمت تغييرًا في المخطط كما قد يحدث مع سكربت ترحيل يُشغَّل يدويًا.
- تصدير النسخ الاحتياطية، وCORS، ورؤوس استجابة HTTP معزَّزة أمنيًا — راجع
  `backend/internal/handlers/export.go` (تعقيم مقاطع المسار)،
  `backend/internal/server/router.go` (CORS يرفض الوصول عبر النطاقات
  افتراضيًا ما لم يُحدَّد `CORS_ALLOWED_ORIGINS`)، و
  `backend/internal/middleware/security_headers.go`.
- **تدوير الأسرار بعد تسرّبها**: يرشدك `backend/cmd/reencrypt-ssh-keys`
  و[`docs/rotate-secrets.md`](docs/rotate-secrets.md) خلال عملية تدوير
  `JWT_SECRET`/`ENCRYPTION_KEY` وإعادة تشفير جميع مفاتيح SSH المخزَّنة
  دون فقدان بيانات.

---

## هيكل المشروع

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # نقطة الدخول
│   │   └── reencrypt-ssh-keys/         # أداة تدوير الأسرار (راجع الأمان)
│   └── internal/
│       ├── config/                     # محمّل الإعدادات + التحقق من الأسرار عند البدء
│       ├── database/                   # اتصال GORM + منفّذ ترحيلات SQL مضمَّن
│       │   └── migrations/             # 001 (أساسي، عبر initdb) + 002–008، تُطبَّق تلقائيًا عند التشغيل
│       ├── models/                     # جميع نماذج GORM
│       ├── middleware/                 # مصادقة JWT، RBAC، تحديد المعدل، رؤوس الأمان
│       ├── handlers/                   # معالجات HTTP لـ Gin
│       │   ├── auth.go                 # تسجيل الدخول / الخروج / التحديث / تغيير كلمة المرور
│       │   ├── hosts.go                # CRUD للمضيفين، التحقق من SSH، نشر المفتاح، إعادة تعيين البصمة
│       │   ├── instances.go            # اكتشاف/فحص النسخ
│       │   ├── rules.go                # CRUD لقواعد pg_hba + تفعيل/تعطيل/إعادة ترتيب/استنساخ
│       │   ├── temp_access.go          # منح/إلغاء الوصول المؤقت
│       │   ├── backup.go, backup_schedules.go   # إنشاء/استعادة/تنزيل النسخ الاحتياطية + الجدولة
│       │   ├── export.go               # تصدير الإعدادات/النسخ الاحتياطية (محلي + S3)
│       │   ├── healthcheck.go          # فحوصات صحة المضيف + الإشعارات
│       │   ├── users.go                # CRUD للمستخدمين (SuperAdmin فقط)
│       │   ├── audit.go                # قائمة سجل التدقيق + المقارنة
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # عميل SSH، تثبيت بصمة المضيف (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # توليد محتوى pg_hba + كتابة ذرية + انتهاء الصلاحية
│       │   ├── crypto.go               # AES-256-GCM لمفاتيح SSH
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # مجدول — إنهاء القواعد المؤقتة كل N دقيقة
│       └── server/router.go            # موجّه Gin، CORS، جميع المسارات
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # عميل Axios للـ API
│       ├── types/index.ts              # واجهات TypeScript (تعكس نماذج Go)
│       ├── i18n/                       # 24 ملف لغة + سجل AVAILABLE_LOCALES (يدعم RTL)
│       ├── stores/auth.ts              # متجر مصادقة Pinia
│       ├── router/index.ts             # موجّه Vue مع حراس المصادقة وتغيير كلمة المرور الإجباري
│       ├── components/layout/AppLayout.vue
│       └── views/                      # واجهة واحدة لكل مجال وظيفي
│
├── docs/rotate-secrets.md              # دليل تدوير الأسرار
├── .github/workflows/security.yml      # CI: بناء/فحص/govulncheck + npm audit
├── Dockerfile                          # متعدد المراحل: بناء Vue + بناء Go → alpine
├── docker-compose.yml                  # الإنتاج
└── docker-compose.dev.yml              # التطوير (قاعدة البيانات فقط)
```

---

## مرجع API

تعيش قائمة المسارات الكاملة في `backend/internal/server/router.go`.
نقاط النهاية الأساسية:

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
POST   /api/hosts/validate              # فحص مسبق عبر SSH قبل الحفظ
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # مسح بصمة مفتاح المضيف المثبَّتة

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

## صلاحيات RBAC

| الميزة                  | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| إدارة المضيفين           | ✓          | ✓   |          |        |               |                    |
| إدارة النسخ              | ✓          | ✓   |          |        |               |                    |
| إضافة/تعديل/حذف قاعدة    | ✓          | ✓   |          |        | ✓             |                    |
| تفعيل/تعطيل قاعدة        | ✓          | ✓   | ✓        |        | ✓             |                    |
| عرض القواعد              | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| منح/إلغاء الوصول المؤقت  | ✓        | ✓   | ✓        |        | ✓             | ✓                  |
| النسخ الاحتياطي والاستعادة | ✓          | ✓   |          |        |               |                    |
| إدارة المستخدمين         | ✓          |     |          |        |               |                    |
| عرض سجل التدقيق          | ✓          | ✓   |          |        |               |                    |

`AccessManager` و`TempAccessManager` دوران محدودان النطاق للفرق التي
تحتاج فقط إلى إدارة قواعد الوصول / الوصول المؤقت دون صلاحيات DBA
الكاملة على المضيفين أو النسخ أو النسخ الاحتياطية أو المستخدمين.

---

## قائمة التحقق للإنتاج

- [ ] عيّن `JWT_SECRET`/`ENCRYPTION_KEY` حقيقيين — لن يبدأ التطبيق بدونهما
- [ ] سجّل الدخول مرة واحدة وغيّر كلمة مرور المسؤول الافتراضية (إجباري عند أول تسجيل دخول)
- [ ] عيّن `DB_PASSWORD` قويًا
- [ ] فعّل TLS (`DB_SSL_MODE=require`)
- [ ] ضع التطبيق خلف NGINX/Traefik مع HTTPS
- [ ] قصر منفذ Postgres على الشبكة الداخلية فقط
- [ ] عيّن `CORS_ALLOWED_ORIGINS` فقط إذا كانت الواجهة الأمامية منشورة على نطاق مختلف عن الـ API
