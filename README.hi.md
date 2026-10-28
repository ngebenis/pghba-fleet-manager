# PGHBA Fleet Manager

<!--LANGSWITCH-->

PostgreSQL `pg_hba.conf` केंद्रीकृत प्रबंधन सिस्टम — एक ही वेब UI से
PostgreSQL इंस्टेंस के पूरे फ्लीट में एक्सेस नियम प्रबंधित करें, बजाय
इसके कि हर होस्ट पर मैन्युअली `pg_hba.conf` संपादित किया जाए।

यह रिपॉजिटरी मूल PGHBA Manager प्रोजेक्ट का एक हार्डेन्ड, सार्वजनिक
प्रकाशन के लिए तैयार स्नैपशॉट है: डिफ़ॉल्ट पासवर्ड बदलने की अनिवार्यता,
स्टार्टअप पर सीक्रेट सत्यापन, SSH host key पिनिंग, लॉगिन रेट लिमिटिंग,
और हर push पर dependency स्कैन करने वाला CI पाइपलाइन। नीचे
[सुरक्षा](#सुरक्षा) अनुभाग देखें।

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ कृपया मुझे एक कप कॉफी पिलाएं।

---

## स्क्रीनशॉट

### डैशबोर्ड

![Dashboard](docs/screenshots/dashboard.png)

> नीचे दिए गए स्क्रीनशॉट (Database Hosts, PG Instances, Access Rules,
> Temporary Access, Backup & Restore, User Management, Audit Log,
> Settings) इस रिपॉजिटरी के वास्तविक फ्रंटएंड कंपोनेंट्स और रंग योजना
> से बनाए गए **उदाहरण मॉकअप** हैं, जिनमें डमी/नमूना डेटा भरा गया है —
> यह किसी लाइव डिप्लॉयमेंट का स्क्रीनशॉट नहीं है। जैसे-जैसे असली
> स्क्रीनशॉट उपलब्ध होंगे, इन्हें बदल दिया जाएगा।

<details>
<summary>Database Hosts (डेटाबेस होस्ट)</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>PG Instances (PG इंस्टेंस)</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>Access Rules (एक्सेस नियम)</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>Temporary Access (अस्थायी एक्सेस)</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>Backup & Restore (बैकअप और पुनर्स्थापना)</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>User Management (उपयोगकर्ता प्रबंधन)</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>Audit Log (ऑडिट लॉग)</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>Settings (सेटिंग्स)</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## क्विक स्टार्ट (Docker)

```bash
# 1. क्लोन करें और कॉन्फ़िगर करें
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. टेम्पलेट से .env बनाएं (docker-compose.yml द्वारा उपयोग किया जाता है)
cp backend/.env.example .env

# 3. असली सीक्रेट जनरेट करें — इनके बिना ऐप शुरू नहीं होगा
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. .env को अपने असली मूल्यों के साथ संपादित करें, फिर:
docker compose up -d

# ऐप यहां उपलब्ध है: http://localhost:8080
```

**पहला लॉगिन**: `admin` / `Admin@1234`। इस अकाउंट को पहली बार लॉगिन
करते समय पासवर्ड बदलने के लिए फोर्स-फ्लैग किया गया है — जब तक पासवर्ड
नहीं बदला जाता, सर्वर अन्य सभी क्रियाओं को ब्लॉक कर देता है, इसलिए
"डिफ़ॉल्ट पासवर्ड बदलना भूल गए" जैसी कोई स्थिति नहीं बनती।

---

## डेवलपमेंट सेटअप

```bash
# केवल डेटाबेस शुरू करें
docker compose -f docker-compose.dev.yml up -d

# बैकएंड (Go)
cd backend
cp .env.example .env       # आवश्यकतानुसार संपादित करें
go run ./cmd/server

# फ्रंटएंड (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (/api को :8080 पर प्रॉक्सी करता है)
```

---

## सुरक्षा

- **कोई असुरक्षित डिफ़ॉल्ट मान नहीं चल सकता।** स्टार्टअप पर `main.go`
  `config.Validate()` को कॉल करता है, जो यदि `JWT_SECRET` या
  `ENCRYPTION_KEY` खाली हैं या अभी भी प्रकाशित प्लेसहोल्डर मान हैं तो
  शुरू होने से इनकार कर देता है — गलत कॉन्फ़िगर की गई डिप्लॉयमेंट
  जाली बनाए जा सकने वाले JWT सीक्रेट या ज्ञात AES की के साथ चलने के
  बजाय स्पष्ट रूप से विफल हो जाती है।
- **SSH host key पिन किए जाते हैं**, अनदेखा नहीं किए जाते।
  `services.ConnectToHost` पहली बार कनेक्शन पर किसी प्रबंधित होस्ट की
  SSH key पर भरोसा करता है और उसका फिंगरप्रिंट सहेजता है
  (`db_hosts.ssh_host_key_fingerprint`); अलग key के साथ कोई भी बाद का
  कनेक्शन चुपचाप स्वीकार करने के बजाय अस्वीकार कर दिया जाता है। किसी
  वैध होस्ट रीबिल्ड के बाद फिंगरप्रिंट साफ़ करना एक समर्पित, ऑडिट किए
  गए endpoint (`POST /api/hosts/:id/reset-host-key`) से होता है, न कि
  सामान्य होस्ट-एडिट फॉर्म से।
- **लॉगिन रेट-लिमिटेड है**: एक IP से 15 मिनट के भीतर 10 असफल प्रयास
  उस IP को 15 मिनट के लिए लॉक कर देते हैं।
- **किसी उपयोगकर्ता को निष्क्रिय करना या उसकी भूमिका घटाना तुरंत
  प्रभावी होता है** — हर प्रमाणित अनुरोध डेटाबेस के विरुद्ध
  स्थिति/भूमिका को फिर से जांचता है, न कि यह मानता है कि JWT ने 8 घंटे
  पहले क्या दावा किया था।
- **CI हर push/PR पर और साप्ताहिक रूप से ज्ञात कमजोर dependency के लिए
  स्कैन करता है** (`.github/workflows/security.yml`): Go बैकएंड के लिए
  `govulncheck`, फ्रंटएंड के लिए `npm audit` (production dependency)।
- बैकअप एक्सपोर्ट, CORS और HTTP रिस्पॉन्स हेडर को हार्डन किया गया है —
  देखें `backend/internal/handlers/export.go` (path-segment सैनिटाइजेशन),
  `backend/internal/server/router.go` (जब तक `CORS_ALLOWED_ORIGINS` सेट
  न हो, CORS डिफ़ॉल्ट रूप से कोई क्रॉस-ओरिजिन एक्सेस नहीं देता), और
  `backend/internal/middleware/security_headers.go`।
- **लीक होने के बाद सीक्रेट रोटेट करना**: `backend/cmd/reencrypt-ssh-keys`
  और [`docs/rotate-secrets.md`](docs/rotate-secrets.md) बिना डेटा खोए
  `JWT_SECRET`/`ENCRYPTION_KEY` को रोटेट करने और सभी सहेजी गई SSH key
  को फिर से एन्क्रिप्ट करने की प्रक्रिया बताते हैं।

---

## प्रोजेक्ट संरचना

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # एंट्री पॉइंट
│   │   └── reencrypt-ssh-keys/         # सीक्रेट-रोटेशन टूल (सुरक्षा देखें)
│   └── internal/
│       ├── config/                     # कॉन्फ़िग लोडर + स्टार्टअप सीक्रेट सत्यापन
│       ├── database/                   # GORM कनेक्शन
│       ├── models/                     # सभी GORM मॉडल
│       ├── middleware/                 # JWT auth, RBAC, रेट लिमिटिंग, security headers
│       ├── handlers/                   # Gin HTTP हैंडलर
│       │   ├── auth.go                 # Login / logout / refresh / change-password
│       │   ├── hosts.go                # होस्ट CRUD, SSH validate, deploy-key, reset-host-key
│       │   ├── instances.go            # इंस्टेंस डिस्कवरी/स्कैन
│       │   ├── rules.go                # pg_hba रूल CRUD + enable/disable/reorder/clone
│       │   ├── temp_access.go          # अस्थायी एक्सेस grant/revoke
│       │   ├── backup.go, backup_schedules.go   # बैकअप create/restore/download + शेड्यूलिंग
│       │   ├── export.go               # कॉन्फ़िग/बैकअप एक्सपोर्ट (local + S3)
│       │   ├── healthcheck.go          # होस्ट हेल्थ चेक + नोटिफिकेशन
│       │   ├── users.go                # उपयोगकर्ता CRUD (केवल SuperAdmin)
│       │   ├── audit.go                # ऑडिट लॉग लिस्ट + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH क्लाइंट, host key पिनिंग (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # pg_hba कंटेंट रेंडरिंग + एटॉमिक राइट + एक्सपायरी
│       │   ├── crypto.go               # SSH key के लिए AES-256-GCM
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — हर N मिनट में अस्थायी नियम समाप्त करता है
│       └── server/router.go            # Gin राउटर, CORS, सभी रूट
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API क्लाइंट
│       ├── types/index.ts              # TypeScript इंटरफेस (Go मॉडल का प्रतिबिंब)
│       ├── stores/auth.ts              # Pinia auth store
│       ├── router/index.ts             # auth + force-password-change गार्ड के साथ Vue Router
│       ├── components/layout/AppLayout.vue
│       └── views/                      # हर फीचर एरिया के लिए एक view
│
├── database/migrations/
│   ├── 001_init.sql                    # enum, index, trigger के साथ पूरा स्कीमा
│   └── 002–007_*.sql                   # RBAC भूमिकाएं, बैकअप शेड्यूल, पासवर्ड/host-key हार्डनिंग
│
├── docs/rotate-secrets.md              # सीक्रेट रोटेशन रनबुक
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-stage: Vue build + Go build → alpine
├── docker-compose.yml                  # Production
└── docker-compose.dev.yml              # Development (केवल DB)
```

---

## API संदर्भ

पूरी रूट सूची `backend/internal/server/router.go` में है। मुख्य
endpoints:

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
POST   /api/hosts/validate              # सहेजने से पहले SSH पूर्व-जांच
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # पिन किए गए SSH host key फिंगरप्रिंट को साफ़ करें

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

## RBAC अनुमतियां

| फीचर                   | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| होस्ट प्रबंधन            | ✓          | ✓   |          |        |               |                    |
| इंस्टेंस प्रबंधन          | ✓          | ✓   |          |        |               |                    |
| नियम जोड़ें/संपादित करें/हटाएं | ✓    | ✓   |          |        | ✓             |                    |
| नियम सक्षम/अक्षम करें    | ✓          | ✓   | ✓        |        | ✓             |                    |
| नियम देखें               | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| अस्थायी एक्सेस दें/रद्द करें | ✓        | ✓   | ✓        |        | ✓             | ✓                  |
| बैकअप और पुनर्स्थापना    | ✓          | ✓   |          |        |               |                    |
| उपयोगकर्ता प्रबंधन        | ✓          |     |          |        |               |                    |
| ऑडिट लॉग देखें           | ✓          | ✓   |          |        |               |                    |

`AccessManager` और `TempAccessManager` उन टीमों के लिए सीमित भूमिकाएं
हैं जिन्हें होस्ट, इंस्टेंस, बैकअप, या उपयोगकर्ताओं पर पूर्ण DBA
अधिकारों के बिना केवल एक्सेस नियम / अस्थायी एक्सेस प्रबंधित करना है।

---

## प्रोडक्शन चेकलिस्ट

- [ ] असली `JWT_SECRET`/`ENCRYPTION_KEY` सेट करें — अन्यथा ऐप शुरू नहीं होगा
- [ ] एक बार लॉगिन करें और डिफ़ॉल्ट एडमिन पासवर्ड बदलें (पहले लॉगिन पर अनिवार्य)
- [ ] एक मजबूत `DB_PASSWORD` सेट करें
- [ ] TLS सक्षम करें (`DB_SSL_MODE=require`)
- [ ] ऐप को HTTPS के साथ NGINX/Traefik के पीछे रखें
- [ ] Postgres के पोर्ट को केवल इंटरनल नेटवर्क तक सीमित रखें
- [ ] `CORS_ALLOWED_ORIGINS` केवल तभी सेट करें जब फ्रंटएंड API से अलग origin पर डिप्लॉय किया गया हो
