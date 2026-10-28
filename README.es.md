# PGHBA Fleet Manager

<!--LANGSWITCH-->

Sistema de gestión centralizada de `pg_hba.conf` de PostgreSQL — gestione
las reglas de acceso de toda una flota de instancias de PostgreSQL desde
una única interfaz web, sin necesidad de editar manualmente
`pg_hba.conf` en cada host.

Este repositorio es una instantánea (snapshot) reforzada y lista para
uso público del proyecto original PGHBA Manager: aplicación obligatoria
del cambio de contraseña por defecto, validación de secretos al inicio,
fijación (pinning) de claves de host SSH, limitación de intentos de
inicio de sesión, y un pipeline de CI que escanea dependencias en cada
push. Vea la sección [Seguridad](#seguridad) más abajo.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Invíteme un café.

---

## Capturas de pantalla

### Panel (Dashboard)

![Dashboard](docs/screenshots/dashboard.png)

> Las capturas de pantalla a continuación (Hosts de base de datos,
> Instancias PG, Reglas de acceso, Acceso temporal, Copias de
> seguridad, Gestión de usuarios, Registro de auditoría, Configuración)
> son **maquetas ilustrativas** construidas a partir de los componentes
> reales del frontend de este repositorio y su esquema de colores,
> rellenadas con datos de ejemplo/ficticios — no corresponden a un
> despliegue real. Las capturas reales las irán reemplazando con el
> tiempo.

<details>
<summary>Hosts de base de datos</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>Instancias PG</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>Reglas de acceso</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>Acceso temporal</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>Copias de seguridad</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>Gestión de usuarios</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>Registro de auditoría</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>Configuración</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## Inicio rápido (Docker)

```bash
# 1. Clonar y configurar
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. Crear el .env a partir de la plantilla (usado por docker-compose.yml)
cp backend/.env.example .env

# 3. Generar secretos reales — la aplicación se niega a arrancar sin ellos
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. Edite .env con sus valores reales y luego:
docker compose up -d

# Aplicación disponible en: http://localhost:8080
```

**Primer inicio de sesión**: `admin` / `Admin@1234`. Esta cuenta está
marcada para forzar el cambio de contraseña en el primer inicio de
sesión — el servidor bloquea cualquier otra acción hasta que se cambie,
de modo que no existe la posibilidad de "olvidar cambiar la contraseña
por defecto".

---

## Configuración de desarrollo

```bash
# Iniciar solo la base de datos
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # editar según sea necesario
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (redirige /api a :8080)
```

---

## Seguridad

- **No se permite ejecutar con valores por defecto inseguros.**
  `main.go` llama a `config.Validate()` al iniciar, lo que impide el
  arranque si `JWT_SECRET` o `ENCRYPTION_KEY` están vacíos o siguen
  siendo el valor de ejemplo publicado — un despliegue mal configurado
  falla de forma explícita, en lugar de ejecutarse con un secreto JWT
  falsificable o una clave AES conocida públicamente.
- **Las claves de host SSH se fijan (pinning)**, no se ignoran.
  `services.ConnectToHost` confía en la clave SSH de un host gestionado
  en la primera conexión y persiste su huella digital
  (`db_hosts.ssh_host_key_fingerprint`); cualquier conexión posterior
  con una clave diferente será rechazada en lugar de aceptarse
  silenciosamente. Borrar una huella tras una reconstrucción legítima
  del host se realiza mediante un endpoint dedicado y auditado
  (`POST /api/hosts/:id/reset-host-key`), nunca mediante el formulario
  genérico de edición de hosts.
- **El inicio de sesión tiene límite de intentos**: 10 intentos
  fallidos desde una misma IP en 15 minutos bloquean esa IP durante
  15 minutos.
- **Desactivar o degradar a un usuario surte efecto de inmediato** —
  cada solicitud autenticada vuelve a verificar el estado/rol contra la
  base de datos, en lugar de confiar en lo que un JWT declaró hasta
  8 horas atrás.
- **El CI escanea dependencias con vulnerabilidades conocidas** en
  cada push/PR y semanalmente (`.github/workflows/security.yml`):
  `govulncheck` para el backend en Go, `npm audit` (dependencias de
  producción) para el frontend.
- La exportación de copias de seguridad, CORS y las cabeceras de
  respuesta HTTP están reforzadas — vea
  `backend/internal/handlers/export.go` (sanitización de segmentos de
  ruta), `backend/internal/server/router.go` (CORS no permite acceso
  entre orígenes por defecto, a menos que se configure
  `CORS_ALLOWED_ORIGINS`), y
  `backend/internal/middleware/security_headers.go`.
- **Rotación de secretos tras una filtración**:
  `backend/cmd/reencrypt-ssh-keys` y
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) explican cómo
  rotar `JWT_SECRET`/`ENCRYPTION_KEY` y volver a cifrar todas las
  claves SSH almacenadas sin pérdida de datos.

---

## Estructura del proyecto

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Punto de entrada
│   │   └── reencrypt-ssh-keys/         # Herramienta de rotación de secretos (ver Seguridad)
│   └── internal/
│       ├── config/                     # Cargador de configuración + validación de secretos al inicio
│       ├── database/                   # Conexión GORM
│       ├── models/                     # Todos los modelos GORM
│       ├── middleware/                 # Autenticación JWT, RBAC, límite de intentos, cabeceras de seguridad
│       ├── handlers/                   # Manejadores HTTP de Gin
│       │   ├── auth.go                 # Login / logout / refresh / cambio de contraseña
│       │   ├── hosts.go                # CRUD de hosts, validación SSH, deploy-key, reset-host-key
│       │   ├── instances.go            # Descubrimiento/escaneo de instancias
│       │   ├── rules.go                # CRUD de reglas pg_hba + enable/disable/reorder/clone
│       │   ├── temp_access.go          # Concesión/revocación de acceso temporal
│       │   ├── backup.go, backup_schedules.go   # Crear/restaurar/descargar copias + programación
│       │   ├── export.go               # Exportación de configuración/copias (local + S3)
│       │   ├── healthcheck.go          # Comprobaciones de salud de hosts + notificaciones
│       │   ├── users.go                # CRUD de usuarios (solo SuperAdmin)
│       │   ├── audit.go                # Listado de auditoría + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # Cliente SSH, fijación de clave de host (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # Renderizado de pg_hba + escritura atómica + expiración
│       │   ├── crypto.go               # AES-256-GCM para claves SSH
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — expira reglas temporales cada N minutos
│       └── server/router.go            # Router de Gin, CORS, todas las rutas
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Cliente API Axios
│       ├── types/index.ts              # Interfaces TypeScript (reflejan los modelos de Go)
│       ├── stores/auth.ts              # Store de autenticación Pinia
│       ├── router/index.ts             # Vue Router con guardias de auth + cambio forzado de contraseña
│       ├── components/layout/AppLayout.vue
│       └── views/                      # Una vista por área funcional
│
├── database/migrations/
│   ├── 001_init.sql                    # Esquema completo con enums, índices, triggers
│   └── 002–007_*.sql                   # Roles RBAC, programación de copias, refuerzo de contraseña/host-key
│
├── docs/rotate-secrets.md              # Guía de rotación de secretos
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-stage: build de Vue + build de Go → alpine
├── docker-compose.yml                  # Producción
└── docker-compose.dev.yml              # Desarrollo (solo BD)
```

---

## Referencia de la API

El listado completo de rutas se encuentra en
`backend/internal/server/router.go`. Endpoints principales:

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
POST   /api/hosts/validate              # Comprobación SSH previa antes de guardar
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # Borra la huella de clave de host SSH fijada

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

## Permisos RBAC

| Funcionalidad          | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Gestionar hosts        | ✓          | ✓   |          |        |               |                    |
| Gestionar instancias   | ✓          | ✓   |          |        |               |                    |
| Añadir/editar/eliminar regla | ✓    | ✓   |          |        | ✓             |                    |
| Habilitar/deshabilitar regla | ✓    | ✓   | ✓        |        | ✓             |                    |
| Ver reglas              | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Conceder/revocar acceso temporal | ✓ | ✓   | ✓        |        | ✓             | ✓                  |
| Copias de seguridad y restauración | ✓ | ✓ |         |        |               |                    |
| Gestión de usuarios     | ✓          |     |          |        |               |                    |
| Ver registro de auditoría | ✓        | ✓   |          |        |               |                    |

`AccessManager` y `TempAccessManager` son roles acotados para equipos
que solo deben gestionar reglas de acceso / acceso temporal, sin tener
privilegios completos de DBA sobre hosts, instancias, copias de
seguridad o usuarios.

---

## Checklist de producción

- [ ] Configure `JWT_SECRET`/`ENCRYPTION_KEY` reales — la aplicación no arrancará sin ellos
- [ ] Inicie sesión una vez y cambie la contraseña de administrador por defecto (obligatorio en el primer inicio de sesión)
- [ ] Configure una `DB_PASSWORD` robusta
- [ ] Habilite TLS (`DB_SSL_MODE=require`)
- [ ] Coloque la aplicación detrás de NGINX/Traefik con HTTPS
- [ ] Restrinja el puerto de Postgres solo a la red interna
- [ ] Configure `CORS_ALLOWED_ORIGINS` únicamente si el frontend se despliega en un origen distinto al de la API
