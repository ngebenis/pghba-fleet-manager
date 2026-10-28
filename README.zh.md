# PGHBA Fleet Manager

<!--LANGSWITCH-->

PostgreSQL `pg_hba.conf` 集中管理系统 —— 通过单一 Web 界面管理整个
PostgreSQL 实例集群的访问规则，无需在每台主机上手动编辑 `pg_hba.conf`。

本仓库是原始 PGHBA Manager 项目经过安全加固、可公开发布的快照版本：
强制修改默认密码、启动时验证密钥、SSH 主机密钥固定（pinning）、登录
频率限制，以及在每次推送时扫描依赖项的 CI 流水线。详见下方的
[安全性](#安全性) 部分。

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ 请我喝杯咖啡吧。

---

## 截图

### 仪表盘

![Dashboard](docs/screenshots/dashboard.png)

> 以下截图（数据库主机、PG 实例、访问规则、临时访问、备份与恢复、
> 用户管理、审计日志、设置）均为**示意性模拟图**，基于本仓库实际的
> 前端组件和配色方案制作，并填充了示例/虚拟数据 —— 并非真实部署的
> 截图。真实截图将在后续逐步替换。

<details>
<summary>数据库主机</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>PG 实例</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>访问规则</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>临时访问</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>备份与恢复</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>用户管理</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>审计日志</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>设置</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## 快速开始（Docker）

```bash
# 1. 克隆并配置
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. 从模板创建 .env（供 docker-compose.yml 使用）
cp backend/.env.example .env

# 3. 生成真实的密钥 —— 没有密钥应用将拒绝启动
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. 用真实值编辑 .env，然后执行：
docker compose up -d

# 应用访问地址：http://localhost:8080
```

**首次登录**：`admin` / `Admin@1234`。该账号被强制标记为首次登录时
必须修改密码 —— 在密码修改完成之前，服务器会阻止其他所有操作，因此
不存在"忘记修改默认密码"的风险窗口。

---

## 开发环境搭建

```bash
# 仅启动数据库
docker compose -f docker-compose.dev.yml up -d

# 后端（Go）
cd backend
cp .env.example .env       # 按需修改
go run ./cmd/server

# 前端（Vue 3）
cd frontend
npm install
npm run dev
# → http://localhost:5173（将 /api 代理到 :8080）
```

---

## 安全性

- **不允许在不安全的默认配置下运行。** `main.go` 在启动时调用
  `config.Validate()`，如果 `JWT_SECRET` 或 `ENCRYPTION_KEY` 为空，
  或仍是公开发布的占位符，将直接拒绝启动 —— 配置错误的部署会
  明确报错失败，而不是带着可伪造的 JWT 密钥或已知的 AES 密钥运行。
- **SSH 主机密钥会被固定（pinned）**，而非被忽略。
  `services.ConnectToHost` 在首次连接时信任被管理主机的 SSH 密钥，
  并持久化其指纹（`db_hosts.ssh_host_key_fingerprint`）；之后若使用
  不同的密钥连接将被拒绝，而不是被静默信任。在主机合法重建后清除
  指纹，需通过专用且有审计记录的接口
  （`POST /api/hosts/:id/reset-host-key`），而非普通的主机编辑表单。
- **登录有频率限制**：同一 IP 在 15 分钟内失败 10 次登录，将被锁定
  15 分钟。
- **停用或降级用户权限会立即生效** —— 每个已认证的请求都会重新
  校验数据库中的状态/角色，而不是信任长达 8 小时前签发的 JWT 中
  的声明。
- **CI 会在每次推送/PR 以及每周扫描已知存在漏洞的依赖**
  （`.github/workflows/security.yml`）：后端 Go 使用 `govulncheck`，
  前端使用 `npm audit`（仅生产依赖）。
- 备份导出、CORS 以及 HTTP 响应头均已加固 —— 详见
  `backend/internal/handlers/export.go`（路径片段净化处理）、
  `backend/internal/server/router.go`（默认不允许任何跨域访问，
  除非设置了 `CORS_ALLOWED_ORIGINS`），以及
  `backend/internal/middleware/security_headers.go`。
- **密钥泄露后的轮换**：`backend/cmd/reencrypt-ssh-keys` 以及
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) 提供了完整指引，
  指导您在不丢失数据的情况下轮换 `JWT_SECRET`/`ENCRYPTION_KEY` 并
  重新加密所有已存储的 SSH 密钥。

---

## 项目结构

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # 入口文件
│   │   └── reencrypt-ssh-keys/         # 密钥轮换工具（详见"安全性"）
│   └── internal/
│       ├── config/                     # 配置加载器 + 启动时密钥校验
│       ├── database/                   # GORM 数据库连接
│       ├── models/                     # 所有 GORM 模型
│       ├── middleware/                 # JWT 认证、RBAC、频率限制、安全响应头
│       ├── handlers/                   # Gin HTTP 处理器
│       │   ├── auth.go                 # 登录 / 登出 / 刷新令牌 / 修改密码
│       │   ├── hosts.go                # 主机 CRUD、SSH 校验、部署密钥、重置主机密钥
│       │   ├── instances.go            # 实例发现/扫描
│       │   ├── rules.go                # pg_hba 规则 CRUD + 启用/禁用/重排序/克隆
│       │   ├── temp_access.go          # 临时访问的授予/撤销
│       │   ├── backup.go, backup_schedules.go   # 备份创建/恢复/下载 + 定时调度
│       │   ├── export.go               # 配置/备份导出（本地 + S3）
│       │   ├── healthcheck.go          # 主机健康检查 + 通知
│       │   ├── users.go                # 用户 CRUD（仅限 SuperAdmin）
│       │   ├── audit.go                # 审计日志列表 + 差异对比
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH 客户端、主机密钥固定（ConnectToHost）
│       │   ├── hba.go, hba_writer.go    # pg_hba 内容渲染 + 原子写入 + 过期处理
│       │   ├── crypto.go               # SSH 密钥的 AES-256-GCM 加密
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # 定时任务 —— 每隔 N 分钟过期临时规则
│       └── server/router.go            # Gin 路由、CORS、所有路由定义
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API 客户端
│       ├── types/index.ts              # TypeScript 接口（对应 Go 模型）
│       ├── stores/auth.ts              # Pinia 认证 store
│       ├── router/index.ts             # Vue Router，含认证 + 强制改密守卫
│       ├── components/layout/AppLayout.vue
│       └── views/                      # 每个功能区域对应一个视图
│
├── database/migrations/
│   ├── 001_init.sql                    # 完整的数据库结构，含枚举、索引、触发器
│   └── 002–007_*.sql                   # RBAC 角色、备份计划、密码/主机密钥加固
│
├── docs/rotate-secrets.md              # 密钥轮换操作手册
├── .github/workflows/security.yml      # CI：build/vet/govulncheck + npm audit
├── Dockerfile                          # 多阶段构建：Vue 构建 + Go 构建 → alpine
├── docker-compose.yml                  # 生产环境
└── docker-compose.dev.yml              # 开发环境（仅数据库）
```

---

## API 参考

完整路由列表位于 `backend/internal/server/router.go`。核心接口：

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
POST   /api/hosts/validate              # 保存前的 SSH 预检
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # 清除已固定的 SSH 主机密钥指纹

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

## RBAC 权限

| 功能                   | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| 管理主机                | ✓          | ✓   |          |        |               |                    |
| 管理实例                | ✓          | ✓   |          |        |               |                    |
| 添加/编辑/删除规则       | ✓          | ✓   |          |        | ✓             |                    |
| 启用/禁用规则            | ✓          | ✓   | ✓        |        | ✓             |                    |
| 查看规则                | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| 授予/撤销临时访问        | ✓        | ✓   | ✓        |        | ✓             | ✓                  |
| 备份与恢复               | ✓          | ✓   |          |        |               |                    |
| 用户管理                | ✓          |     |          |        |               |                    |
| 查看审计日志             | ✓          | ✓   |          |        |               |                    |

`AccessManager` 和 `TempAccessManager` 是面向特定团队的受限角色，
用于仅管理访问规则/临时访问，而无需拥有对主机、实例、备份或用户的
完整 DBA 权限。

---

## 生产环境检查清单

- [ ] 设置真实的 `JWT_SECRET`/`ENCRYPTION_KEY` —— 否则应用无法启动
- [ ] 首次登录后立即修改默认管理员密码（首次登录时强制要求）
- [ ] 设置强度足够的 `DB_PASSWORD`
- [ ] 启用 TLS（`DB_SSL_MODE=require`）
- [ ] 在 NGINX/Traefik 后部署应用并启用 HTTPS
- [ ] 将 Postgres 端口限制为仅内网可访问
- [ ] 仅当前端与 API 部署在不同源（origin）时，才设置 `CORS_ALLOWED_ORIGINS`
