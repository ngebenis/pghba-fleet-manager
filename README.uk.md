# PGHBA Fleet Manager

<!--LANGSWITCH-->

Централізована система керування `pg_hba.conf` для PostgreSQL — керуйте
правилами доступу для цілого парку екземплярів PostgreSQL з єдиного
вебінтерфейсу, замість ручного редагування `pg_hba.conf` на кожному
хості.

Цей репозиторій — це захищений, готовий до публічного доступу знімок
оригінального проєкту PGHBA Manager: примусова зміна пароля за
замовчуванням, перевірка секретів під час запуску, прив'язка SSH-ключа
хоста, обмеження частоти входу та CI-конвеєр, який сканує залежності
при кожному push. Дивіться розділ [Безпека](#безпека) нижче.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Пригостіть мене чашкою кави.

---

## Скріншоти

### Дашборд

![Dashboard](docs/screenshots/dashboard.png)

> Скріншоти нижче (Хости баз даних, Екземпляри PG, Правила доступу,
> Тимчасовий доступ, Резервне копіювання та відновлення, Керування
> користувачами, Журнал аудиту, Налаштування) є **ілюстративними
> макетами**, створеними на основі реальних компонентів фронтенду та
> кольорової схеми цього репозиторію, заповненими демонстраційними
> даними — це не знімки реального розгортання. Справжні скріншоти з
> часом замінять їх.

<details>
<summary>Хости баз даних</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>Екземпляри PG</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>Правила доступу</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>Тимчасовий доступ</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>Резервне копіювання та відновлення</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>Керування користувачами</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>Журнал аудиту</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>Налаштування</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## Швидкий старт (Docker)

```bash
# 1. Клонуйте та налаштуйте
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. Створіть .env із шаблону (використовується docker-compose.yml)
cp backend/.env.example .env

# 3. Згенеруйте реальні секрети — без них застосунок не запуститься
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. Відредагуйте .env реальними значеннями, потім:
docker compose up -d

# Застосунок доступний за адресою: http://localhost:8080
```

**Перший вхід**: `admin` / `Admin@1234`. Цей обліковий запис примусово
позначений для зміни пароля під час першого входу — сервер блокує будь-
яку іншу дію, поки пароль не змінено, тож "забув змінити пароль за
замовчуванням" тут неможливе.

---

## Налаштування для розробки

```bash
# Запустіть лише базу даних
docker compose -f docker-compose.dev.yml up -d

# Бекенд (Go)
cd backend
cp .env.example .env       # відредагуйте за потреби
go run ./cmd/server

# Фронтенд (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (проксує /api на :8080)
```

---

## Безпека

- **Жодні небезпечні значення за замовчуванням не можуть запуститись.**
  `main.go` викликає `config.Validate()` під час запуску, що відмовляє
  запускатися, якщо `JWT_SECRET` або `ENCRYPTION_KEY` порожні або досі
  мають опубліковане значення-заглушку — неправильно налаштоване
  розгортання гучно падає, замість того щоб працювати з підробним
  JWT-секретом або відомим AES-ключем.
- **Ключі хостів SSH прив'язуються**, а не ігноруються.
  `services.ConnectToHost` довіряє SSH-ключу керованого хоста під час
  першого з'єднання і зберігає його відбиток
  (`db_hosts.ssh_host_key_fingerprint`); будь-яке наступне з'єднання з
  іншим ключем відхиляється, а не приймається мовчки. Очищення
  відбитка після легітимної пересборки хоста відбувається через
  спеціальну, аудитовану кінцеву точку
  (`POST /api/hosts/:id/reset-host-key`), а не через звичайну форму
  редагування хоста.
- **Вхід обмежений за частотою**: 10 невдалих спроб з однієї IP-адреси
  протягом 15 хвилин блокують цю IP-адресу на 15 хвилин.
- **Деактивація або пониження користувача діє негайно** — кожен
  автентифікований запит повторно перевіряє статус/роль у базі даних,
  замість того щоб довіряти заявленому в JWT, якому може бути до 8
  годин.
- **CI сканує відомі вразливі залежності** при кожному push/PR і щотижня
  (`.github/workflows/security.yml`): `govulncheck` для бекенду на Go,
  `npm audit` (виробничі залежності) для фронтенду.
- Експорт резервних копій, CORS та HTTP-заголовки відповіді захищені —
  дивіться `backend/internal/handlers/export.go` (санітизація сегментів
  шляху), `backend/internal/server/router.go` (CORS за замовчуванням
  забороняє крос-доменний доступ, якщо не встановлено
  `CORS_ALLOWED_ORIGINS`), та
  `backend/internal/middleware/security_headers.go`.
- **Ротація секретів після витоку**: `backend/cmd/reencrypt-ssh-keys` і
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) описують, як
  виконати ротацію `JWT_SECRET`/`ENCRYPTION_KEY` і повторно зашифрувати
  всі збережені SSH-ключі без втрати даних.

---

## Структура проєкту

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Точка входу
│   │   └── reencrypt-ssh-keys/         # Інструмент ротації секретів (див. Безпека)
│   └── internal/
│       ├── config/                     # Завантажувач конфігурації + перевірка секретів під час запуску
│       ├── database/                   # Підключення GORM
│       ├── models/                     # Усі моделі GORM
│       ├── middleware/                 # JWT-автентифікація, RBAC, обмеження частоти, заголовки безпеки
│       ├── handlers/                   # HTTP-обробники Gin
│       │   ├── auth.go                 # Вхід / вихід / оновлення токена / зміна пароля
│       │   ├── hosts.go                # CRUD хостів, перевірка SSH, розгортання ключа, скидання ключа хоста
│       │   ├── instances.go            # Виявлення/сканування екземплярів
│       │   ├── rules.go                # CRUD правил pg_hba + увімкнення/вимкнення/зміна порядку/клонування
│       │   ├── temp_access.go          # Надання/відкликання тимчасового доступу
│       │   ├── backup.go, backup_schedules.go   # Створення/відновлення/завантаження резервних копій + розклад
│       │   ├── export.go               # Експорт конфігурації/резервних копій (локально + S3)
│       │   ├── healthcheck.go          # Перевірка стану хостів + сповіщення
│       │   ├── users.go                # CRUD користувачів (лише SuperAdmin)
│       │   ├── audit.go                # Список журналу аудиту + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH-клієнт, прив'язка ключа хоста (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # Рендеринг вмісту pg_hba + атомарний запис + закінчення терміну дії
│       │   ├── crypto.go               # AES-256-GCM для SSH-ключів
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — закінчення терміну дії тимчасових правил кожні N хвилин
│       └── server/router.go            # Маршрутизатор Gin, CORS, усі маршрути
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API-клієнт
│       ├── types/index.ts              # TypeScript-інтерфейси (віддзеркалюють моделі Go)
│       ├── stores/auth.ts              # Pinia-сховище автентифікації
│       ├── router/index.ts             # Vue Router із захистами автентифікації + примусової зміни пароля
│       ├── components/layout/AppLayout.vue
│       └── views/                      # Одне подання для кожної функціональної області
│
├── database/migrations/
│   ├── 001_init.sql                    # Повна схема з переліками, індексами, тригерами
│   └── 002–007_*.sql                   # Ролі RBAC, розклади резервного копіювання, захист пароля/ключа хоста
│
├── docs/rotate-secrets.md              # Посібник із ротації секретів
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Багатоетапний: збірка Vue + збірка Go → alpine
├── docker-compose.yml                  # Продакшн
└── docker-compose.dev.yml              # Розробка (лише БД)
```

---

## Довідник API

Повний список маршрутів знаходиться в `backend/internal/server/router.go`.
Основні кінцеві точки:

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
POST   /api/hosts/validate              # Попередня перевірка SSH перед збереженням
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # Очистити прив'язаний відбиток ключа хоста

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

## Дозволи RBAC

| Функція                | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Керування хостами       | ✓          | ✓   |          |        |               |                    |
| Керування екземплярами  | ✓          | ✓   |          |        |               |                    |
| Додати/редагувати/видалити правило | ✓ | ✓ |         |        | ✓             |                    |
| Увімкнути/вимкнути правило | ✓       | ✓   | ✓        |        | ✓             |                    |
| Перегляд правил         | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Надати/відкликати тимчасовий доступ | ✓ | ✓ | ✓       |        | ✓             | ✓                  |
| Резервне копіювання та відновлення | ✓ | ✓ |         |        |               |                    |
| Керування користувачами | ✓          |     |          |        |               |                    |
| Перегляд журналу аудиту | ✓          | ✓   |          |        |               |                    |

`AccessManager` і `TempAccessManager` — це обмежені ролі для команд,
яким потрібно керувати лише правилами доступу / тимчасовим доступом без
повних прав DBA щодо хостів, екземплярів, резервних копій чи
користувачів.

---

## Чекліст для продакшну

- [ ] Встановіть реальні `JWT_SECRET`/`ENCRYPTION_KEY` — інакше застосунок не запуститься
- [ ] Увійдіть один раз і змініть пароль адміністратора за замовчуванням (обов'язково при першому вході)
- [ ] Встановіть надійний `DB_PASSWORD`
- [ ] Увімкніть TLS (`DB_SSL_MODE=require`)
- [ ] Розмістіть застосунок за NGINX/Traefik з HTTPS
- [ ] Обмежте порт Postgres лише внутрішньою мережею
- [ ] Встановіть `CORS_ALLOWED_ORIGINS` лише якщо фронтенд розгорнуто на іншому домені, ніж API
