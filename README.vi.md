# PGHBA Fleet Manager

<!--LANGSWITCH-->

Hệ thống quản lý tập trung `pg_hba.conf` của PostgreSQL — quản lý các quy
tắc truy cập trên toàn bộ fleet instance PostgreSQL từ một giao diện web
duy nhất, thay vì phải chỉnh sửa `pg_hba.conf` thủ công trên từng host.

Repository này là bản snapshot đã được tăng cường bảo mật, sẵn sàng công
khai của dự án PGHBA Manager gốc: bắt buộc đổi mật khẩu mặc định, xác
thực secret khi khởi động, SSH host key pinning, giới hạn tốc độ đăng
nhập, và pipeline CI quét dependency ở mỗi lần push. Xem phần
[Bảo mật](#bảo-mật) bên dưới.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Mời tôi một tách cà phê nhé.

---

## Ảnh chụp màn hình

### Dashboard

![Dashboard](docs/screenshots/dashboard.png)

> Các ảnh chụp màn hình bên dưới (Database Hosts, PG Instances, Access
> Rules, Temporary Access, Backup & Restore, User Management, Audit Log,
> Settings) là **mockup minh họa**, được dựng từ các component frontend
> và bảng màu thực tế của repo này, sử dụng dữ liệu mẫu/giả — không phải
> từ bản triển khai thực tế. Ảnh chụp màn hình thật sẽ thay thế dần theo
> thời gian.

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

## Bắt đầu nhanh (Docker)

```bash
# 1. Clone và cấu hình
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. Tạo .env từ template (được dùng bởi docker-compose.yml)
cp backend/.env.example .env

# 3. Tạo secret thật — ứng dụng sẽ từ chối khởi động nếu không có
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. Chỉnh sửa .env với giá trị thật của bạn, sau đó:
docker compose up -d

# Ứng dụng khả dụng tại: http://localhost:8080
```

**Đăng nhập lần đầu**: `admin` / `Admin@1234`. Tài khoản này bị bắt buộc
đổi mật khẩu ngay lần đăng nhập đầu tiên — server sẽ chặn mọi hành động
khác cho đến khi mật khẩu được đổi, nên không có chuyện "quên đổi mật
khẩu mặc định".

---

## Thiết lập môi trường phát triển

```bash
# Chỉ khởi chạy database
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # chỉnh sửa nếu cần
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (proxy /api đến :8080)
```

---

## Bảo mật

- **Không có cấu hình mặc định không an toàn nào có thể chạy được.**
  `main.go` gọi `config.Validate()` khi khởi động, từ chối khởi động nếu
  `JWT_SECRET` hoặc `ENCRYPTION_KEY` trống hoặc vẫn là giá trị placeholder
  đã công khai — một triển khai cấu hình sai sẽ thất bại rõ ràng thay vì
  chạy với một JWT secret có thể bị giả mạo hoặc một AES key đã bị lộ.
- **SSH host key được ghim (pinned)**, không bị bỏ qua.
  `services.ConnectToHost` tin tưởng SSH key của một host được quản lý ở
  lần kết nối đầu tiên và lưu lại fingerprint của nó
  (`db_hosts.ssh_host_key_fingerprint`); mọi kết nối sau đó với một key
  khác sẽ bị từ chối thay vì âm thầm được tin tưởng. Việc xóa fingerprint
  sau khi rebuild host hợp lệ phải đi qua một endpoint riêng, có ghi audit
  (`POST /api/hosts/:id/reset-host-key`), không phải qua form chỉnh sửa
  host thông thường.
- **Đăng nhập bị giới hạn tốc độ**: 10 lần đăng nhập thất bại từ một IP
  trong vòng 15 phút sẽ khóa IP đó trong 15 phút.
- **Vô hiệu hóa hoặc hạ quyền người dùng có hiệu lực ngay lập tức** — mọi
  request đã xác thực đều kiểm tra lại status/role với database thay vì
  tin vào claim của JWT có thể đã cũ tới 8 giờ.
- **CI quét các dependency có lỗ hổng đã biết** ở mỗi lần push/PR và hàng
  tuần (`.github/workflows/security.yml`): `govulncheck` cho backend Go,
  `npm audit` (dependency production) cho frontend.
- Backup export, CORS và HTTP response headers đã được tăng cường bảo mật
  — xem `backend/internal/handlers/export.go` (sanitize path-segment),
  `backend/internal/server/router.go` (CORS mặc định không cho truy cập
  cross-origin trừ khi `CORS_ALLOWED_ORIGINS` được thiết lập), và
  `backend/internal/middleware/security_headers.go`.
- **Xoay vòng secret sau khi bị rò rỉ**: `backend/cmd/reencrypt-ssh-keys`
  và [`docs/rotate-secrets.md`](docs/rotate-secrets.md) hướng dẫn xoay
  vòng `JWT_SECRET`/`ENCRYPTION_KEY` và mã hóa lại toàn bộ SSH key đã lưu
  mà không mất dữ liệu.

---

## Cấu trúc dự án

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Entry point
│   │   └── reencrypt-ssh-keys/         # Công cụ xoay vòng secret (xem Bảo mật)
│   └── internal/
│       ├── config/                     # Config loader + xác thực secret khi khởi động
│       ├── database/                   # Kết nối GORM
│       ├── models/                     # Toàn bộ model GORM
│       ├── middleware/                 # JWT auth, RBAC, rate limiting, security headers
│       ├── handlers/                   # Gin HTTP handlers
│       │   ├── auth.go                 # Login / logout / refresh / change-password
│       │   ├── hosts.go                # Host CRUD, SSH validate, deploy-key, reset-host-key
│       │   ├── instances.go            # Discovery/scan instance
│       │   ├── rules.go                # pg_hba rule CRUD + enable/disable/reorder/clone
│       │   ├── temp_access.go          # Grant/revoke temporary access
│       │   ├── backup.go, backup_schedules.go   # Create/restore/download backup + lập lịch
│       │   ├── export.go               # Export config/backup (local + S3)
│       │   ├── healthcheck.go          # Health check host + thông báo
│       │   ├── users.go                # User CRUD (chỉ SuperAdmin)
│       │   ├── audit.go                # Danh sách audit log + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH client, host key pinning (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # Render nội dung pg_hba + atomic write + expiry
│       │   ├── crypto.go               # AES-256-GCM cho SSH key
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — hết hạn temp rules mỗi N phút
│       └── server/router.go            # Gin router, CORS, toàn bộ route
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API client
│       ├── types/index.ts              # TypeScript interface (phản chiếu Go model)
│       ├── stores/auth.ts              # Pinia auth store
│       ├── router/index.ts             # Vue Router với auth + force-password-change guard
│       ├── components/layout/AppLayout.vue
│       └── views/                      # Mỗi view cho một khu vực tính năng
│
├── database/migrations/
│   ├── 001_init.sql                    # Schema đầy đủ với enum, index, trigger
│   └── 002–007_*.sql                   # RBAC roles, backup schedule, hardening password/host-key
│
├── docs/rotate-secrets.md              # Runbook xoay vòng secret
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-stage: build Vue + build Go → alpine
├── docker-compose.yml                  # Production
└── docker-compose.dev.yml              # Development (chỉ DB)
```

---

## Tham khảo API

Danh sách route đầy đủ nằm trong
`backend/internal/server/router.go`. Các endpoint chính:

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
POST   /api/hosts/validate              # Kiểm tra SSH trước khi lưu
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # Xóa fingerprint SSH host key đã ghim

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

## Quyền hạn RBAC

| Tính năng               | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|--------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Quản lý Host             | ✓          | ✓   |          |        |               |                    |
| Quản lý Instance         | ✓          | ✓   |          |        |               |                    |
| Thêm/Sửa/Xóa Rule        | ✓          | ✓   |          |        | ✓             |                    |
| Bật/Tắt Rule             | ✓          | ✓   | ✓        |        | ✓             |                    |
| Xem Rule                 | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Cấp/Thu hồi Temp Access  | ✓          | ✓   | ✓        |        | ✓             | ✓                  |
| Backup & Restore         | ✓          | ✓   |          |        |               |                    |
| Quản lý người dùng       | ✓          |     |          |        |               |                    |
| Xem Audit Log            | ✓          | ✓   |          |        |               |                    |

`AccessManager` và `TempAccessManager` là các role giới hạn dành cho các
nhóm chỉ cần quản lý access rule / temporary access mà không cần đầy đủ
quyền DBA đối với host, instance, backup hoặc người dùng.

---

## Checklist cho Production

- [ ] Thiết lập `JWT_SECRET`/`ENCRYPTION_KEY` thật — ứng dụng sẽ không khởi động nếu thiếu
- [ ] Đăng nhập một lần và đổi mật khẩu admin mặc định (bắt buộc ở lần đăng nhập đầu)
- [ ] Đặt `DB_PASSWORD` mạnh
- [ ] Bật TLS (`DB_SSL_MODE=require`)
- [ ] Đặt ứng dụng sau NGINX/Traefik với HTTPS
- [ ] Giới hạn cổng Postgres chỉ trong mạng nội bộ
- [ ] Chỉ thiết lập `CORS_ALLOWED_ORIGINS` nếu frontend được triển khai trên một origin khác với API
