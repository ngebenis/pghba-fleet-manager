# PGHBA Fleet Manager

<!--LANGSWITCH-->

Централизованная система управления `pg_hba.conf` для PostgreSQL —
управляйте правилами доступа для целого парка экземпляров PostgreSQL из
единого веб-интерфейса, вместо ручного редактирования `pg_hba.conf` на
каждом хосте.

Этот репозиторий — защищённый, готовый к публикации снимок оригинального
проекта PGHBA Manager: принудительная смена пароля по умолчанию,
проверка секретов при запуске, закрепление (pinning) SSH host key,
ограничение частоты попыток входа и CI-пайплайн, сканирующий зависимости
при каждом push. Подробнее см. раздел [Безопасность](#безопасность) ниже.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Угостите меня чашкой кофе.

---

## Скриншоты

### Панель управления (Dashboard)

![Dashboard](docs/screenshots/dashboard.png)

> Скриншоты ниже (Database Hosts, PG Instances, Access Rules, Temporary
> Access, Backup & Restore, User Management, Audit Log, Settings) —
> **иллюстративные макеты**, созданные на основе реальных компонентов
> фронтенда и цветовой схемы этого репозитория, заполненные
> тестовыми/примерными данными — это не реальное развёрнутое
> приложение. Настоящие скриншоты заменят их по мере появления.

<details>
<summary>Database Hosts (Хосты баз данных)</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>PG Instances (Экземпляры PG)</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>Access Rules (Правила доступа)</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>Temporary Access (Временный доступ)</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>Backup & Restore (Резервное копирование)</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>User Management (Управление пользователями)</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>Audit Log (Журнал аудита)</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>Settings (Настройки)</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## Быстрый старт (Docker)

```bash
# 1. Клонируйте и настройте
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. Создайте .env из шаблона (используется docker-compose.yml)
cp backend/.env.example .env

# 3. Сгенерируйте настоящие секреты — без них приложение не запустится
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. Отредактируйте .env своими значениями, затем:
docker compose up -d

# Приложение доступно по адресу: http://localhost:8080
```

**Первый вход**: `admin` / `Admin@1234`. Эта учётная запись принудительно
помечена для смены пароля при первом входе — сервер блокирует любые
другие действия, пока пароль не будет изменён, поэтому невозможно
«забыть» сменить пароль по умолчанию.

---

## Настройка окружения разработки

```bash
# Запустите только базу данных
docker compose -f docker-compose.dev.yml up -d

# Бэкенд (Go)
cd backend
cp .env.example .env       # отредактируйте при необходимости
go run ./cmd/server

# Фронтенд (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (проксирует /api на :8080)
```

---

## Безопасность

- **Небезопасные значения по умолчанию не могут запуститься.** `main.go`
  вызывает `config.Validate()` при старте, который отказывает в запуске,
  если `JWT_SECRET` или `ENCRYPTION_KEY` пусты или всё ещё содержат
  опубликованное значение-заглушку — неправильно настроенное
  развёртывание явно падает, вместо того чтобы работать с подделываемым
  JWT-секретом или заведомо известным AES-ключом.
- **SSH host key закрепляются (pinning)**, а не игнорируются.
  `services.ConnectToHost` доверяет SSH-ключу управляемого хоста при
  первом подключении и сохраняет его отпечаток
  (`db_hosts.ssh_host_key_fingerprint`); любое последующее подключение
  с другим ключом отклоняется, а не принимается молча. Сброс отпечатка
  после легитимной пересборки хоста выполняется через выделенный,
  аудируемый endpoint (`POST /api/hosts/:id/reset-host-key`), а не через
  обычную форму редактирования хоста.
- **Вход ограничен по частоте**: 10 неудачных попыток с одного IP в
  течение 15 минут блокируют этот IP на 15 минут.
- **Деактивация или понижение роли пользователя вступает в силу
  немедленно** — каждый аутентифицированный запрос заново проверяет
  статус/роль в базе данных, а не доверяет тому, что JWT утверждал до
  8 часов назад.
- **CI сканирует известные уязвимые зависимости** при каждом push/PR и
  еженедельно (`.github/workflows/security.yml`): `govulncheck` для
  бэкенда на Go, `npm audit` (production-зависимости) для фронтенда.
- Экспорт резервных копий, CORS и HTTP-заголовки ответов защищены — см.
  `backend/internal/handlers/export.go` (санитизация сегментов пути),
  `backend/internal/server/router.go` (CORS по умолчанию запрещает
  кросс-доменный доступ, если не задан `CORS_ALLOWED_ORIGINS`), и
  `backend/internal/middleware/security_headers.go`.
- **Ротация секретов после утечки**: `backend/cmd/reencrypt-ssh-keys` и
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) описывают процесс
  ротации `JWT_SECRET`/`ENCRYPTION_KEY` и повторного шифрования всех
  сохранённых SSH-ключей без потери данных.

---

## Структура проекта

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Точка входа
│   │   └── reencrypt-ssh-keys/         # Инструмент ротации секретов (см. Безопасность)
│   └── internal/
│       ├── config/                     # Загрузчик конфигурации + проверка секретов при старте
│       ├── database/                   # Подключение GORM
│       ├── models/                     # Все модели GORM
│       ├── middleware/                 # JWT-аутентификация, RBAC, ограничение частоты, security headers
│       ├── handlers/                   # HTTP-обработчики Gin
│       │   ├── auth.go                 # Login / logout / refresh / change-password
│       │   ├── hosts.go                # CRUD хостов, проверка SSH, deploy-key, reset-host-key
│       │   ├── instances.go            # Обнаружение/сканирование экземпляров
│       │   ├── rules.go                # CRUD правил pg_hba + enable/disable/reorder/clone
│       │   ├── temp_access.go          # Выдача/отзыв временного доступа
│       │   ├── backup.go, backup_schedules.go   # Создание/восстановление/скачивание резервных копий + расписание
│       │   ├── export.go               # Экспорт конфигурации/резервных копий (локально + S3)
│       │   ├── healthcheck.go          # Проверка состояния хостов + уведомления
│       │   ├── users.go                # CRUD пользователей (только SuperAdmin)
│       │   ├── audit.go                # Список журнала аудита + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH-клиент, закрепление host key (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # Рендеринг содержимого pg_hba + атомарная запись + истечение срока
│       │   ├── crypto.go               # AES-256-GCM для SSH-ключей
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — истечение временных правил каждые N минут
│       └── server/router.go            # Роутер Gin, CORS, все маршруты
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API-клиент
│       ├── types/index.ts              # TypeScript-интерфейсы (зеркало моделей Go)
│       ├── stores/auth.ts              # Pinia auth store
│       ├── router/index.ts             # Vue Router с guard'ами auth + force-password-change
│       ├── components/layout/AppLayout.vue
│       └── views/                      # По одному view на функциональную область
│
├── database/migrations/
│   ├── 001_init.sql                    # Полная схема с enum'ами, индексами, триггерами
│   └── 002–007_*.sql                   # Роли RBAC, расписания резервного копирования, усиление безопасности пароля/host-key
│
├── docs/rotate-secrets.md              # Руководство по ротации секретов
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-stage: сборка Vue + сборка Go → alpine
├── docker-compose.yml                  # Production
└── docker-compose.dev.yml              # Разработка (только БД)
```

---

## Справочник API

Полный список маршрутов находится в `backend/internal/server/router.go`.
Основные endpoints:

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
POST   /api/hosts/validate              # Предварительная проверка SSH перед сохранением
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # Сброс закреплённого отпечатка SSH host key

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

## Права доступа RBAC

| Функция                | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Управление хостами      | ✓          | ✓   |          |        |               |                    |
| Управление экземплярами | ✓          | ✓   |          |        |               |                    |
| Добавление/изменение/удаление правила | ✓ | ✓  |          |        | ✓             |                    |
| Включение/отключение правила | ✓    | ✓   | ✓        |        | ✓             |                    |
| Просмотр правил         | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Выдача/отзыв временного доступа | ✓  | ✓   | ✓        |        | ✓             | ✓                  |
| Резервное копирование и восстановление | ✓ | ✓ |         |        |               |                    |
| Управление пользователями | ✓        |     |          |        |               |                    |
| Просмотр журнала аудита | ✓          | ✓   |          |        |               |                    |

`AccessManager` и `TempAccessManager` — это ограниченные роли для
команд, которым нужно управлять только правилами доступа / временным
доступом, без полных прав DBA на хосты, экземпляры, резервные копии или
пользователей.

---

## Чек-лист перед продакшном

- [ ] Задайте настоящие `JWT_SECRET`/`ENCRYPTION_KEY` — иначе приложение не запустится
- [ ] Войдите один раз и смените пароль администратора по умолчанию (обязательно при первом входе)
- [ ] Задайте надёжный `DB_PASSWORD`
- [ ] Включите TLS (`DB_SSL_MODE=require`)
- [ ] Разместите приложение за NGINX/Traefik с HTTPS
- [ ] Ограничьте доступ к порту Postgres только внутренней сетью
- [ ] Задайте `CORS_ALLOWED_ORIGINS` только если фронтенд развёрнут на другом origin, чем API
