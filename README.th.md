# PGHBA Fleet Manager

<!--LANGSWITCH-->

ระบบจัดการส่วนกลางสำหรับ `pg_hba.conf` ของ PostgreSQL — จัดการกฎการเข้าถึง
ทั่วทั้งกลุ่มอินสแตนซ์ PostgreSQL จากเว็บ UI เดียว แทนที่จะแก้ไข
`pg_hba.conf` ด้วยมือในแต่ละโฮสต์

รีพอซิทอรีนี้เป็นสแนปช็อตที่ผ่านการเสริมความแข็งแกร่งและพร้อมเผยแพร่สู่สาธารณะ
ของโปรเจกต์ PGHBA Manager ดั้งเดิม: การบังคับใช้รหัสผ่านที่ไม่ใช่ค่าเริ่มต้น
การตรวจสอบความถูกต้องของ secret ตอนเริ่มทำงาน การปักหมุดคีย์โฮสต์ SSH
การจำกัดอัตราการเข้าสู่ระบบ และไปป์ไลน์ CI ที่สแกนดีเพนเดนซีทุกครั้งที่ push
ดูรายละเอียดที่ [ความปลอดภัย](#ความปลอดภัย) ด้านล่าง

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ เลี้ยงกาแฟฉันสักแก้วได้นะ

---

## ภาพหน้าจอ

### แดชบอร์ด

![Dashboard](docs/screenshots/dashboard.png)

> ภาพหน้าจอด้านล่าง (โฮสต์ฐานข้อมูล, อินสแตนซ์ PG, กฎการเข้าถึง, การเข้าถึงชั่วคราว,
> สำรองและกู้คืน, จัดการผู้ใช้, บันทึกการตรวจสอบ, ตั้งค่า) เป็น
> **ภาพจำลองประกอบ** ที่สร้างจากคอมโพเนนต์ frontend และชุดสีจริงของรีพอซิทอรีนี้
> โดยใช้ข้อมูลตัวอย่าง/ข้อมูลจำลอง ไม่ใช่ภาพจากการใช้งานจริง
> ภาพหน้าจอจริงจะมาแทนที่เมื่อพร้อมใช้งาน

<details>
<summary>โฮสต์ฐานข้อมูล</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>อินสแตนซ์ PG</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>กฎการเข้าถึง</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>การเข้าถึงชั่วคราว</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>สำรองและกู้คืน</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>จัดการผู้ใช้</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>บันทึกการตรวจสอบ</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>ตั้งค่า</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## เริ่มต้นใช้งานอย่างรวดเร็ว (Docker)

```bash
# 1. โคลนและตั้งค่า
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. สร้าง .env จากเทมเพลต (ใช้โดย docker-compose.yml)
cp backend/.env.example .env

# 3. สร้าง secret จริง — แอปจะปฏิเสธการเริ่มทำงานหากไม่มี
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. แก้ไข .env ด้วยค่าจริงของคุณ จากนั้น:
docker compose up -d

# แอปพร้อมใช้งานที่: http://localhost:8080
```

**เข้าสู่ระบบครั้งแรก**: `admin` / `Admin@1234` บัญชีนี้ถูกบังคับให้
เปลี่ยนรหัสผ่านในการเข้าสู่ระบบครั้งแรก — เซิร์ฟเวอร์จะบล็อกการดำเนินการอื่นๆ
ทั้งหมดจนกว่าจะเปลี่ยนรหัสผ่าน จึงไม่มีช่องโหว่ "ลืมเปลี่ยนรหัสผ่านเริ่มต้น"

---

## การตั้งค่าสำหรับพัฒนา

```bash
# เริ่มเฉพาะฐานข้อมูล
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # แก้ไขตามต้องการ
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (พร็อกซี /api ไปยัง :8080)
```

---

## ความปลอดภัย

- **ไม่สามารถรันด้วยค่าเริ่มต้นที่ไม่ปลอดภัยได้** `main.go` เรียก `config.Validate()`
  ตอนเริ่มทำงาน ซึ่งจะปฏิเสธการเริ่มทำงานหาก `JWT_SECRET` หรือ `ENCRYPTION_KEY`
  ว่างเปล่าหรือยังเป็นค่า placeholder ที่เผยแพร่ไว้ — การตั้งค่าที่ผิดพลาด
  จะล้มเหลวอย่างชัดเจนแทนที่จะทำงานด้วย JWT secret ที่ปลอมแปลงได้
  หรือ AES key ที่รู้จักกันทั่วไป
- **คีย์โฮสต์ SSH ถูกปักหมุด** ไม่ใช่ถูกละเลย `services.ConnectToHost`
  เชื่อถือคีย์ SSH ของโฮสต์ที่จัดการในการเชื่อมต่อครั้งแรก และบันทึก
  fingerprint ไว้ (`db_hosts.ssh_host_key_fingerprint`) การเชื่อมต่อครั้งถัดไป
  ด้วยคีย์ที่แตกต่างกันจะถูกปฏิเสธแทนที่จะเชื่อถือโดยเงียบๆ การล้าง
  fingerprint หลังจากสร้างโฮสต์ใหม่โดยถูกต้องตามกฎหมายต้องผ่าน endpoint
  เฉพาะที่มีการตรวจสอบ (`POST /api/hosts/:id/reset-host-key`) ไม่ใช่ผ่าน
  ฟอร์มแก้ไขโฮสต์ทั่วไป
- **การเข้าสู่ระบบถูกจำกัดอัตรา**: ความพยายามล้มเหลว 10 ครั้งจาก IP เดียว
  ภายใน 15 นาที จะล็อก IP นั้นเป็นเวลา 15 นาที
- **การปิดใช้งานหรือลดระดับผู้ใช้มีผลทันที** — ทุกคำขอที่ผ่านการยืนยันตัวตน
  จะตรวจสอบสถานะ/บทบาทกับฐานข้อมูลใหม่อีกครั้ง แทนที่จะเชื่อถือสิ่งที่ JWT
  อ้างไว้นานถึง 8 ชั่วโมง
- **CI สแกนหาดีเพนเดนซีที่มีช่องโหว่ที่รู้จัก** ทุกครั้งที่ push/PR และทุกสัปดาห์
  (`.github/workflows/security.yml`): `govulncheck` สำหรับ backend Go
  `npm audit` (ดีเพนเดนซีสำหรับการใช้งานจริง) สำหรับ frontend
- การส่งออกข้อมูลสำรอง, CORS, และ HTTP response headers ได้รับการเสริมความแข็งแกร่ง — ดู
  `backend/internal/handlers/export.go` (การทำความสะอาด path segment),
  `backend/internal/server/router.go` (CORS จะไม่อนุญาต cross-origin
  โดยค่าเริ่มต้น เว้นแต่จะตั้งค่า `CORS_ALLOWED_ORIGINS`), และ
  `backend/internal/middleware/security_headers.go`
- **การหมุนเวียน secret หลังจากรั่วไหล**: `backend/cmd/reencrypt-ssh-keys` และ
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) แนะนำขั้นตอนการหมุนเวียน
  `JWT_SECRET`/`ENCRYPTION_KEY` และเข้ารหัสคีย์ SSH ที่จัดเก็บไว้ทั้งหมดใหม่
  โดยไม่สูญเสียข้อมูล

---

## โครงสร้างโปรเจกต์

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # จุดเริ่มต้น
│   │   └── reencrypt-ssh-keys/         # เครื่องมือหมุนเวียน secret (ดูความปลอดภัย)
│   └── internal/
│       ├── config/                     # ตัวโหลด config + ตรวจสอบ secret ตอนเริ่มทำงาน
│       ├── database/                   # การเชื่อมต่อ GORM
│       ├── models/                     # โมเดล GORM ทั้งหมด
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

## เอกสารอ้างอิง API

รายการ route ทั้งหมดอยู่ใน `backend/internal/server/router.go` endpoint หลัก:

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

## สิทธิ์ RBAC

| ฟีเจอร์                | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| จัดการโฮสต์             | ✓          | ✓   |          |        |               |                    |
| จัดการอินสแตนซ์          | ✓          | ✓   |          |        |               |                    |
| เพิ่ม/แก้ไข/ลบกฎ         | ✓          | ✓   |          |        | ✓             |                    |
| เปิด/ปิดใช้งานกฎ         | ✓          | ✓   | ✓        |        | ✓             |                    |
| ดูกฎ                    | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| ให้/เพิกถอนการเข้าถึงชั่วคราว | ✓      | ✓   | ✓        |        | ✓             | ✓                  |
| สำรองและกู้คืน           | ✓          | ✓   |          |        |               |                    |
| จัดการผู้ใช้             | ✓          |     |          |        |               |                    |
| ดูบันทึกการตรวจสอบ       | ✓          | ✓   |          |        |               |                    |

`AccessManager` และ `TempAccessManager` เป็นบทบาทที่จำกัดขอบเขตสำหรับทีมที่
ต้องการจัดการเฉพาะกฎการเข้าถึง / การเข้าถึงชั่วคราว โดยไม่ต้องมีสิทธิ์ DBA
เต็มรูปแบบเหนือโฮสต์, อินสแตนซ์, การสำรองข้อมูล, หรือผู้ใช้

---

## รายการตรวจสอบสำหรับการใช้งานจริง

- [ ] ตั้งค่า `JWT_SECRET`/`ENCRYPTION_KEY` จริง — แอปจะไม่เริ่มทำงานหากไม่มี
- [ ] เข้าสู่ระบบครั้งหนึ่งและเปลี่ยนรหัสผ่าน admin เริ่มต้น (บังคับใช้ในการเข้าสู่ระบบครั้งแรก)
- [ ] ตั้งค่า `DB_PASSWORD` ที่แข็งแรง
- [ ] เปิดใช้งาน TLS (`DB_SSL_MODE=require`)
- [ ] วางแอปไว้หลัง NGINX/Traefik ด้วย HTTPS
- [ ] จำกัดพอร์ตของ Postgres ให้เฉพาะเครือข่ายภายในเท่านั้น
- [ ] ตั้งค่า `CORS_ALLOWED_ORIGINS` เฉพาะเมื่อ frontend ถูกติดตั้งบน origin ที่แตกต่างจาก API
