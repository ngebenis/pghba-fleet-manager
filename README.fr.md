# PGHBA Fleet Manager

🇮🇩 [ID](README.md) · 🇬🇧 [EN](README.en.md) · 🇨🇳 [ZH](README.zh.md) · 🇪🇸 [ES](README.es.md) · 🇸🇦 [AR](README.ar.md) · 🇫🇷 **FR** · 🇵🇹 [PT](README.pt.md) · 🇧🇩 [BN](README.bn.md) · 🇷🇺 [RU](README.ru.md) · 🇮🇳 [HI](README.hi.md) · 🇹🇭 [TH](README.th.md) · 🇰🇭 [KM](README.km.md) · 🇵🇭 [TL](README.tl.md) · 🇻🇳 [VI](README.vi.md) · 🇯🇵 [JA](README.ja.md) · 🇰🇷 [KO](README.ko.md) · 🇵🇱 [PL](README.pl.md) · 🇮🇹 [IT](README.it.md) · 🇹🇷 [TR](README.tr.md) · 🇮🇳 [TA](README.ta.md) · 🇮🇷 [FA](README.fa.md) · 🇺🇦 [UK](README.uk.md) · 🇩🇪 [DE](README.de.md) · 🇳🇱 [NL](README.nl.md)

Système de gestion centralisée de `pg_hba.conf` pour PostgreSQL — gérez
les règles d'accès sur toute une flotte d'instances PostgreSQL depuis une
seule interface web, au lieu d'éditer `pg_hba.conf` à la main sur chaque
hôte.

Ce dépôt est une version durcie et prête pour le public du projet PGHBA
Manager original : application du changement de mot de passe par défaut,
validation des secrets au démarrage, épinglage de la clé d'hôte SSH,
limitation du taux de connexion, et un pipeline CI qui scanne les
dépendances à chaque push. Voir la section [Sécurité](#security)
ci-dessous.

🌐 **24 langues prêtes à l'emploi** — changez à tout moment depuis le
menu déroulant de la barre supérieure : English, Bahasa Indonesia, 中文,
Español, العربية, Français, Português, বাংলা, Русский, हिन्दी, ไทย,
ខ្មែរ, Tagalog, Tiếng Việt, 日本語, 한국어, Polski, Italiano, Türkçe,
தமிழ், فارسی, Українська, Deutsch, Nederlands. L'arabe et le persan
s'affichent automatiquement de droite à gauche.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Offrez-moi un café.

---

## Captures d'écran

### Tableau de bord

![Dashboard](docs/screenshots/dashboard.png)

> Les captures d'écran ci-dessous (Hôtes de base de données, Instances PG,
> Règles d'accès, Accès temporaire, Sauvegarde et restauration, Gestion
> des utilisateurs, Journal d'audit, Paramètres) sont des **maquettes
> illustratives** construites à partir des composants frontend réels de
> ce dépôt et de son schéma de couleurs, remplies avec des données
> fictives/d'exemple — pas un déploiement en production. Elles seront
> remplacées par de vraies captures d'écran dès qu'elles seront
> disponibles.

<details>
<summary>Hôtes de base de données</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>Instances PG</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>Règles d'accès</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>Accès temporaire</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>Sauvegarde et restauration</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>Gestion des utilisateurs</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>Journal d'audit</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>Paramètres</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## Démarrage rapide (Docker)

```bash
# 1. Cloner et configurer
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. Créer .env à partir du modèle (utilisé par docker-compose.yml)
cp backend/.env.example .env

# 3. Générer de vrais secrets — l'application refuse de démarrer sans eux
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. Modifiez .env avec vos vraies valeurs, puis :
docker compose up -d

# Application disponible sur : http://localhost:8080
```

**Première connexion** : `admin` / `Admin@1234`. Ce compte est marqué
pour forcer le changement de mot de passe à la première connexion — le
serveur bloque toute autre action tant qu'il n'est pas changé, donc il
n'y a pas de fenêtre où l'on « oublierait » de changer le mot de passe
par défaut.

---

## Configuration de développement

```bash
# Démarrer uniquement la base de données
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # ajuster si nécessaire
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (proxy /api vers :8080)
```

---

## Sécurité

- **Aucun paramètre par défaut non sécurisé ne peut s'exécuter.**
  `main.go` appelle `config.Validate()` au démarrage, qui refuse de
  démarrer si `JWT_SECRET` ou `ENCRYPTION_KEY` est vide ou toujours la
  valeur par défaut publiée — un déploiement mal configuré échoue
  bruyamment plutôt que de s'exécuter avec un secret JWT falsifiable ou
  une clé AES connue.
- **Les clés d'hôte SSH sont épinglées**, pas ignorées.
  `services.ConnectToHost` fait confiance à la clé SSH d'un hôte géré
  lors de la première connexion et conserve son empreinte
  (`db_hosts.ssh_host_key_fingerprint`) ; toute connexion ultérieure avec
  une clé différente est refusée au lieu d'être acceptée silencieusement.
  Effacer une empreinte après une reconstruction légitime de l'hôte passe
  par un endpoint dédié et audité (`POST /api/hosts/:id/reset-host-key`),
  jamais par le formulaire générique d'édition d'hôte.
- **La connexion est limitée en débit** : 10 tentatives échouées depuis
  une même IP en 15 minutes bloquent cette IP pendant 15 minutes.
- **Désactiver ou rétrograder un utilisateur prend effet immédiatement**
  — chaque requête authentifiée revérifie le statut/rôle dans la base de
  données au lieu de faire confiance à ce qu'un JWT prétendait, qui peut
  dater de jusqu'à 8 heures.
- **La CI scanne les dépendances vulnérables connues** à chaque push/PR
  et chaque semaine (`.github/workflows/security.yml`) : `govulncheck`
  pour le backend Go, `npm audit` (dépendances de production) pour le
  frontend.
- **Les migrations de base de données sont embarquées et s'appliquent
  automatiquement.** Chaque fichier SQL dans
  `backend/internal/database/migrations/` est intégré au binaire
  (`go:embed`) et appliqué automatiquement — suivi dans une table
  `schema_migrations` — à chaque démarrage, si bien qu'un déploiement ne
  peut jamais sauter silencieusement un changement de schéma comme le
  ferait un script de migration lancé à la main.
- L'export des sauvegardes, CORS, et les en-têtes de réponse HTTP sont
  durcis — voir `backend/internal/handlers/export.go` (assainissement
  des segments de chemin), `backend/internal/server/router.go` (CORS
  refuse par défaut l'accès cross-origin sauf si
  `CORS_ALLOWED_ORIGINS` est défini), et
  `backend/internal/middleware/security_headers.go`.
- **Rotation des secrets après une fuite** : `backend/cmd/reencrypt-ssh-keys`
  et [`docs/rotate-secrets.md`](docs/rotate-secrets.md) détaillent la
  rotation de `JWT_SECRET`/`ENCRYPTION_KEY` et le rechiffrement de
  toutes les clés SSH stockées sans perte de données.

---

## Structure du projet

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Point d'entrée
│   │   └── reencrypt-ssh-keys/         # Outil de rotation des secrets (voir Sécurité)
│   └── internal/
│       ├── config/                     # Chargeur de config + validation des secrets au démarrage
│       ├── database/                   # Connexion GORM + exécuteur de migrations SQL embarqué
│       │   └── migrations/             # 001 (base, via initdb) + 002–008, appliquées automatiquement au démarrage
│       ├── models/                     # Tous les modèles GORM
│       ├── middleware/                 # Auth JWT, RBAC, limitation de débit, en-têtes de sécurité
│       ├── handlers/                   # Handlers HTTP Gin
│       │   ├── auth.go                 # Login / logout / refresh / changement de mot de passe
│       │   ├── hosts.go                # CRUD hôtes, validation SSH, déploiement de clé, reset-host-key
│       │   ├── instances.go            # Découverte/scan d'instances
│       │   ├── rules.go                # CRUD règles pg_hba + activer/désactiver/réordonner/cloner
│       │   ├── temp_access.go          # Octroi/révocation d'accès temporaire
│       │   ├── backup.go, backup_schedules.go   # Création/restauration/téléchargement de sauvegarde + planification
│       │   ├── export.go               # Export config/sauvegarde (local + S3)
│       │   ├── healthcheck.go          # Vérifications de santé des hôtes + notifications
│       │   ├── users.go                # CRUD utilisateurs (SuperAdmin uniquement)
│       │   ├── audit.go                # Liste du journal d'audit + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # Client SSH, épinglage de clé d'hôte (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # Rendu du contenu pg_hba + écriture atomique + expiration
│       │   ├── crypto.go               # AES-256-GCM pour les clés SSH
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — expire les règles temporaires toutes les N minutes
│       └── server/router.go            # Routeur Gin, CORS, toutes les routes
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Client API Axios
│       ├── types/index.ts              # Interfaces TypeScript (reflètent les modèles Go)
│       ├── i18n/                       # 24 fichiers de langue + registre AVAILABLE_LOCALES (compatible RTL)
│       ├── stores/auth.ts              # Store d'authentification Pinia
│       ├── router/index.ts             # Vue Router avec gardes d'auth + changement de mot de passe forcé
│       ├── components/layout/AppLayout.vue
│       └── views/                      # Une vue par domaine fonctionnel
│
├── docs/rotate-secrets.md              # Guide de rotation des secrets
├── .github/workflows/security.yml      # CI : build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-stage : build Vue + build Go → alpine
├── docker-compose.yml                  # Production
└── docker-compose.dev.yml              # Développement (DB uniquement)
```

---

## Référence API

La liste complète des routes se trouve dans
`backend/internal/server/router.go`. Endpoints principaux :

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
POST   /api/hosts/validate              # Vérification SSH préalable avant sauvegarde
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # Efface l'empreinte de clé d'hôte SSH épinglée

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

## Permissions RBAC

| Fonctionnalité          | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Gérer les hôtes         | ✓          | ✓   |          |        |               |                    |
| Gérer les instances     | ✓          | ✓   |          |        |               |                    |
| Ajouter/Modifier/Supprimer une règle | ✓          | ✓   |          |        | ✓             |                    |
| Activer/Désactiver une règle | ✓          | ✓   | ✓        |        | ✓             |                    |
| Voir les règles         | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Accorder/Révoquer un accès temporaire | ✓        | ✓   | ✓        |        | ✓             | ✓                  |
| Sauvegarde et restauration | ✓          | ✓   |          |        |               |                    |
| Gestion des utilisateurs | ✓          |     |          |        |               |                    |
| Voir le journal d'audit | ✓          | ✓   |          |        |               |                    |

`AccessManager` et `TempAccessManager` sont des rôles à portée limitée
pour les équipes qui ne doivent gérer que les règles d'accès / l'accès
temporaire sans les privilèges DBA complets sur les hôtes, instances,
sauvegardes ou utilisateurs.

---

## Checklist de production

- [ ] Définir de vrais `JWT_SECRET`/`ENCRYPTION_KEY` — l'application ne démarrera pas sans eux
- [ ] Se connecter une fois et changer le mot de passe admin par défaut (forcé à la première connexion)
- [ ] Définir un `DB_PASSWORD` fort
- [ ] Activer TLS (`DB_SSL_MODE=require`)
- [ ] Placer l'application derrière NGINX/Traefik avec HTTPS
- [ ] Restreindre le port de Postgres au réseau interne uniquement
- [ ] Définir `CORS_ALLOWED_ORIGINS` uniquement si le frontend est déployé sur une origine différente de l'API
