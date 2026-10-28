# PGHBA Fleet Manager

🇮🇩 [ID](README.md) · 🇬🇧 [EN](README.en.md) · 🇨🇳 [ZH](README.zh.md) · 🇪🇸 [ES](README.es.md) · 🇸🇦 [AR](README.ar.md) · 🇫🇷 [FR](README.fr.md) · 🇵🇹 [PT](README.pt.md) · 🇧🇩 [BN](README.bn.md) · 🇷🇺 [RU](README.ru.md) · 🇮🇳 [HI](README.hi.md) · 🇹🇭 [TH](README.th.md) · 🇰🇭 [KM](README.km.md) · 🇵🇭 [TL](README.tl.md) · 🇻🇳 [VI](README.vi.md) · 🇯🇵 [JA](README.ja.md) · 🇰🇷 **KO** · 🇵🇱 [PL](README.pl.md) · 🇮🇹 [IT](README.it.md) · 🇹🇷 [TR](README.tr.md) · 🇮🇳 [TA](README.ta.md) · 🇮🇷 [FA](README.fa.md) · 🇺🇦 [UK](README.uk.md) · 🇩🇪 [DE](README.de.md) · 🇳🇱 [NL](README.nl.md)

PostgreSQL `pg_hba.conf` 중앙 관리 시스템 — 각 호스트에서
`pg_hba.conf`를 직접 편집하는 대신, 하나의 웹 UI에서 전체
PostgreSQL 인스턴스 집합의 접근 규칙을 관리할 수 있습니다.

이 저장소는 원본 PGHBA Manager 프로젝트를 하드닝하여 공개용으로
준비한 스냅샷입니다: 기본 비밀번호 강제 변경, 시작 시 시크릿 검증,
SSH 호스트 키 고정(pinning), 로그인 속도 제한, 그리고 매 push마다
의존성을 스캔하는 CI 파이프라인을 포함합니다. 자세한 내용은
아래 [보안](#보안) 섹션을 참고하세요.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ 커피 한 잔 후원해 주세요.

---

## 스크린샷

### 대시보드

![Dashboard](docs/screenshots/dashboard.png)

> 아래 스크린샷(데이터베이스 호스트, PG 인스턴스, 접근 규칙, 임시
> 접근, 백업 및 복원, 사용자 관리, 감사 로그, 설정)은 이 저장소의
> 실제 프론트엔드 컴포넌트와 색상 체계를 바탕으로 만든 **예시용
> 목업(mockup)**이며, 더미/샘플 데이터로 채워져 있습니다 — 실제
> 운영 환경의 화면이 아닙니다. 실제 스크린샷은 준비되는 대로
> 교체될 예정입니다.

<details>
<summary>데이터베이스 호스트</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>PG 인스턴스</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>접근 규칙</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>임시 접근</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>백업 및 복원</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>사용자 관리</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>감사 로그</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>설정</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## 빠른 시작 (Docker)

```bash
# 1. 클론 및 설정
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. docker-compose.yml이 사용할 .env를 템플릿에서 생성
cp backend/.env.example .env

# 3. 실제 시크릿 생성 — 없으면 앱이 시작되지 않습니다
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. .env를 실제 값으로 편집한 후:
docker compose up -d

# 앱 접속 주소: http://localhost:8080
```

**최초 로그인**: `admin` / `Admin@1234`. 이 계정은 최초 로그인 시
비밀번호 변경이 강제되며, 변경이 완료될 때까지 서버는 다른 모든
작업을 차단하므로 "기본 비밀번호 변경을 깜빡하는" 상황이 발생하지
않습니다.

---

## 개발 환경 설정

```bash
# 데이터베이스만 시작
docker compose -f docker-compose.dev.yml up -d

# 백엔드 (Go)
cd backend
cp .env.example .env       # 필요에 따라 편집
go run ./cmd/server

# 프론트엔드 (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (/api를 :8080으로 프록시)
```

---

## 보안

- **안전하지 않은 기본값으로는 실행되지 않습니다.** `main.go`는 시작
  시 `config.Validate()`를 호출하여 `JWT_SECRET` 또는
  `ENCRYPTION_KEY`가 비어 있거나 공개된 플레이스홀더 값 그대로인
  경우 시작을 거부합니다 — 잘못 구성된 배포는 위조 가능한 JWT
  시크릿이나 알려진 AES 키로 조용히 실행되는 대신 명확하게
  실패합니다.
- **SSH 호스트 키는 고정(pinning)**되며 무시되지 않습니다.
  `services.ConnectToHost`는 관리 대상 호스트의 SSH 키를 최초
  연결 시 신뢰하고 해당 지문(`db_hosts.ssh_host_key_fingerprint`)을
  저장합니다. 이후 다른 키로의 연결은 조용히 신뢰되는 대신
  거부됩니다. 정상적인 호스트 재구축 후 지문을 지워야 할 경우,
  일반적인 호스트 편집 폼이 아닌 전용 감사 대상 엔드포인트
  (`POST /api/hosts/:id/reset-host-key`)를 통해서만 가능합니다.
- **로그인에는 속도 제한이 적용**됩니다: 15분 이내 동일 IP에서
  10회 로그인 실패 시 해당 IP는 15분간 잠깁니다.
- **사용자 비활성화 또는 권한 강등은 즉시 적용됩니다** — 인증된
  모든 요청은 최대 8시간 전 JWT에 담긴 값을 신뢰하는 대신
  데이터베이스에서 상태/역할을 다시 확인합니다.
- **CI는 알려진 취약점이 있는 의존성을 스캔**합니다. 매 push/PR
  및 매주(`.github/workflows/security.yml`), Go 백엔드는
  `govulncheck`, 프론트엔드는 `npm audit`(프로덕션 의존성)을
  사용합니다.
- 백업 내보내기, CORS, HTTP 응답 헤더도 하드닝되어 있습니다 —
  `backend/internal/handlers/export.go`(경로 세그먼트 살균 처리),
  `backend/internal/server/router.go`(`CORS_ALLOWED_ORIGINS`가
  설정되지 않는 한 기본적으로 교차 출처 접근을 차단), 그리고
  `backend/internal/middleware/security_headers.go`를 참고하세요.
- **유출 후 시크릿 교체**: `backend/cmd/reencrypt-ssh-keys`와
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md)에서 데이터
  손실 없이 `JWT_SECRET`/`ENCRYPTION_KEY`를 교체하고 저장된 모든
  SSH 키를 재암호화하는 방법을 안내합니다.

---

## 프로젝트 구조

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # 진입점
│   │   └── reencrypt-ssh-keys/         # 시크릿 교체 도구(보안 섹션 참고)
│   └── internal/
│       ├── config/                     # 설정 로더 + 시작 시 시크릿 검증
│       ├── database/                   # GORM 연결
│       ├── models/                     # 모든 GORM 모델
│       ├── middleware/                 # JWT 인증, RBAC, 속도 제한, 보안 헤더
│       ├── handlers/                   # Gin HTTP 핸들러
│       │   ├── auth.go                 # 로그인 / 로그아웃 / 갱신 / 비밀번호 변경
│       │   ├── hosts.go                # 호스트 CRUD, SSH 검증, 배포 키, 호스트 키 재설정
│       │   ├── instances.go            # 인스턴스 검색/스캔
│       │   ├── rules.go                # pg_hba 규칙 CRUD + 활성화/비활성화/재정렬/복제
│       │   ├── temp_access.go          # 임시 접근 부여/취소
│       │   ├── backup.go, backup_schedules.go   # 백업 생성/복원/다운로드 + 스케줄링
│       │   ├── export.go               # 설정/백업 내보내기 (로컬 + S3)
│       │   ├── healthcheck.go          # 호스트 상태 점검 + 알림
│       │   ├── users.go                # 사용자 CRUD (SuperAdmin 전용)
│       │   ├── audit.go                # 감사 로그 목록 + 변경 비교
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH 클라이언트, 호스트 키 고정(ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # pg_hba 내용 렌더링 + 원자적 쓰기 + 만료 처리
│       │   ├── crypto.go               # SSH 키용 AES-256-GCM
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — N분마다 임시 규칙 만료 처리
│       └── server/router.go            # Gin 라우터, CORS, 전체 라우트
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API 클라이언트
│       ├── types/index.ts              # TypeScript 인터페이스 (Go 모델 미러링)
│       ├── stores/auth.ts              # Pinia 인증 스토어
│       ├── router/index.ts             # 인증 + 비밀번호 강제 변경 가드가 있는 Vue Router
│       ├── components/layout/AppLayout.vue
│       └── views/                      # 기능 영역별 뷰
│
├── database/migrations/
│   ├── 001_init.sql                    # enum, 인덱스, 트리거를 포함한 전체 스키마
│   └── 002–007_*.sql                   # RBAC 역할, 백업 스케줄, 비밀번호/호스트 키 하드닝
│
├── docs/rotate-secrets.md              # 시크릿 교체 런북
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # 멀티 스테이지: Vue 빌드 + Go 빌드 → alpine
├── docker-compose.yml                  # 프로덕션용
└── docker-compose.dev.yml              # 개발용 (DB만)
```

---

## API 레퍼런스

전체 라우트 목록은 `backend/internal/server/router.go`에 있습니다.
주요 엔드포인트:

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
POST   /api/hosts/validate              # 저장 전 SSH 사전 점검
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # 고정된 SSH 호스트 키 지문 초기화

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

## RBAC 권한

| 기능                    | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| 호스트 관리              | ✓          | ✓   |          |        |               |                    |
| 인스턴스 관리             | ✓          | ✓   |          |        |               |                    |
| 규칙 추가/편집/삭제        | ✓          | ✓   |          |        | ✓             |                    |
| 규칙 활성화/비활성화       | ✓          | ✓   | ✓        |        | ✓             |                    |
| 규칙 조회                | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| 임시 접근 부여/취소        | ✓          | ✓   | ✓        |        | ✓             | ✓                  |
| 백업 및 복원              | ✓          | ✓   |          |        |               |                    |
| 사용자 관리               | ✓          |     |          |        |               |                    |
| 감사 로그 조회            | ✓          | ✓   |          |        |               |                    |

`AccessManager`와 `TempAccessManager`는 호스트, 인스턴스, 백업,
사용자에 대한 전체 DBA 권한 없이 접근 규칙 / 임시 접근만 관리해야
하는 팀을 위한 제한된 역할입니다.

---

## 프로덕션 체크리스트

- [ ] 실제 `JWT_SECRET`/`ENCRYPTION_KEY` 설정 — 설정하지 않으면 앱이 시작되지 않습니다
- [ ] 한 번 로그인하여 기본 관리자 비밀번호 변경(최초 로그인 시 강제됨)
- [ ] 강력한 `DB_PASSWORD` 설정
- [ ] TLS 활성화(`DB_SSL_MODE=require`)
- [ ] HTTPS를 적용한 NGINX/Traefik 뒤에 앱 배치
- [ ] Postgres 포트를 내부 네트워크로만 제한
- [ ] 프론트엔드가 API와 다른 출처에 배포된 경우에만 `CORS_ALLOWED_ORIGINS` 설정
