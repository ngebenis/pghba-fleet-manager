# PGHBA Fleet Manager

<!--LANGSWITCH-->

Sistema de gerenciamento centralizado do `pg_hba.conf` do PostgreSQL —
gerencie regras de acesso em toda uma frota de instâncias PostgreSQL a
partir de uma única interface web, em vez de editar manualmente o
`pg_hba.conf` em cada host.

Este repositório é um snapshot reforçado e pronto para o público do
projeto original PGHBA Manager: aplicação obrigatória de troca de senha
padrão, validação de segredos na inicialização, fixação de chave de host
SSH, limitação de taxa no login e um pipeline de CI que verifica
dependências a cada push. Veja [Segurança](#segurança) abaixo.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Me pague um cafezinho.

---

## Capturas de Tela

### Painel

![Painel](docs/screenshots/dashboard.png)

> As capturas de tela abaixo (Hosts de Banco de Dados, Instâncias PG,
> Regras de Acesso, Acesso Temporário, Backup e Restauração,
> Gerenciamento de Usuários, Log de Auditoria, Configurações) são
> **mockups ilustrativos** construídos a partir dos componentes de
> frontend e do esquema de cores reais deste repositório, preenchidos
> com dados fictícios/de exemplo — não é uma implantação ao vivo.
> Capturas de tela reais vão substituí-las conforme ficarem disponíveis.

<details>
<summary>Hosts de Banco de Dados</summary>

![Hosts de Banco de Dados](docs/screenshots/hosts.png)
</details>

<details>
<summary>Instâncias PG</summary>

![Instâncias PG](docs/screenshots/instances.png)
</details>

<details>
<summary>Regras de Acesso</summary>

![Regras de Acesso](docs/screenshots/rules.png)
</details>

<details>
<summary>Acesso Temporário</summary>

![Acesso Temporário](docs/screenshots/temp-access.png)
</details>

<details>
<summary>Backup e Restauração</summary>

![Backup e Restauração](docs/screenshots/backup.png)
</details>

<details>
<summary>Gerenciamento de Usuários</summary>

![Gerenciamento de Usuários](docs/screenshots/users.png)
</details>

<details>
<summary>Log de Auditoria</summary>

![Log de Auditoria](docs/screenshots/audit.png)
</details>

<details>
<summary>Configurações</summary>

![Configurações](docs/screenshots/settings.png)
</details>

---

## Início Rápido (Docker)

```bash
# 1. Clone e configure
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. Crie o .env a partir do modelo (usado pelo docker-compose.yml)
cp backend/.env.example .env

# 3. Gere segredos reais — o aplicativo recusa iniciar sem eles
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. Edite o .env com seus valores reais e então:
docker compose up -d

# Aplicativo disponível em: http://localhost:8080
```

**Primeiro login**: `admin` / `Admin@1234`. Esta conta é forçada a
trocar a senha no primeiro login — o servidor bloqueia qualquer outra
ação até que ela seja alterada, então não há janela de "esqueci de
trocar a senha padrão".

---

## Configuração de Desenvolvimento

```bash
# Inicie apenas o banco de dados
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # edite conforme necessário
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (encaminha /api para :8080)
```

---

## Segurança

- **Nenhum padrão inseguro pode rodar.** O `main.go` chama
  `config.Validate()` na inicialização, que recusa iniciar se
  `JWT_SECRET` ou `ENCRYPTION_KEY` estiverem vazios ou ainda forem o
  valor de exemplo publicado — uma implantação mal configurada falha de
  forma visível em vez de rodar com um segredo JWT falsificável ou uma
  chave AES conhecida.
- **Chaves de host SSH são fixadas**, não ignoradas. `services.ConnectToHost`
  confia na chave SSH de um host gerenciado na primeira conexão e
  persiste sua impressão digital (`db_hosts.ssh_host_key_fingerprint`);
  qualquer conexão posterior com uma chave diferente é recusada em vez
  de confiada silenciosamente. Limpar uma impressão digital após uma
  reconstrução legítima do host passa por um endpoint dedicado e
  auditado (`POST /api/hosts/:id/reset-host-key`), nunca pelo formulário
  genérico de edição de host.
- **O login tem limite de taxa**: 10 tentativas falhas de um mesmo IP em
  15 minutos bloqueiam esse IP por 15 minutos.
- **Desativar ou rebaixar um usuário tem efeito imediato** — cada
  requisição autenticada reverifica status/função no banco de dados em
  vez de confiar no que um JWT declarou até 8 horas atrás.
- **O CI verifica dependências com vulnerabilidades conhecidas** a cada
  push/PR e semanalmente (`.github/workflows/security.yml`):
  `govulncheck` para o backend Go, `npm audit` (dependências de
  produção) para o frontend.
- Exportação de backup, CORS e cabeçalhos de resposta HTTP são
  reforçados — veja `backend/internal/handlers/export.go` (sanitização
  de segmentos de caminho), `backend/internal/server/router.go` (CORS
  bloqueia acesso cross-origin por padrão, a menos que
  `CORS_ALLOWED_ORIGINS` esteja definido) e
  `backend/internal/middleware/security_headers.go`.
- **Rotação de segredos após um vazamento**:
  `backend/cmd/reencrypt-ssh-keys` e
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md) orientam a rotação
  de `JWT_SECRET`/`ENCRYPTION_KEY` e a re-criptografia de todas as
  chaves SSH armazenadas sem perda de dados.

---

## Estrutura do Projeto

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Ponto de entrada
│   │   └── reencrypt-ssh-keys/         # Ferramenta de rotação de segredos (ver Segurança)
│   └── internal/
│       ├── config/                     # Carregador de config + validação de segredos na inicialização
│       ├── database/                   # Conexão GORM
│       ├── models/                     # Todos os modelos GORM
│       ├── middleware/                 # Autenticação JWT, RBAC, limitação de taxa, cabeçalhos de segurança
│       ├── handlers/                   # Handlers HTTP do Gin
│       │   ├── auth.go                 # Login / logout / refresh / troca de senha
│       │   ├── hosts.go                # CRUD de hosts, validação SSH, deploy-key, reset-host-key
│       │   ├── instances.go            # Descoberta/scan de instâncias
│       │   ├── rules.go                # CRUD de regras pg_hba + enable/disable/reorder/clone
│       │   ├── temp_access.go          # Concessão/revogação de acesso temporário
│       │   ├── backup.go, backup_schedules.go   # Criação/restauração/download de backup + agendamento
│       │   ├── export.go               # Exportação de config/backup (local + S3)
│       │   ├── healthcheck.go          # Verificações de saúde de hosts + notificações
│       │   ├── users.go                # CRUD de usuários (apenas SuperAdmin)
│       │   ├── audit.go                # Lista de log de auditoria + diff
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # Cliente SSH, fixação de chave de host (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # Renderização do conteúdo pg_hba + gravação atômica + expiração
│       │   ├── crypto.go               # AES-256-GCM para chaves SSH
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — expira regras temporárias a cada N minutos
│       └── server/router.go            # Roteador Gin, CORS, todas as rotas
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Cliente de API Axios
│       ├── types/index.ts              # Interfaces TypeScript (espelham os modelos Go)
│       ├── stores/auth.ts              # Store Pinia de autenticação
│       ├── router/index.ts             # Vue Router com guards de autenticação + troca forçada de senha
│       ├── components/layout/AppLayout.vue
│       └── views/                      # Uma view por área de funcionalidade
│
├── database/migrations/
│   ├── 001_init.sql                    # Esquema completo com enums, índices, triggers
│   └── 002–007_*.sql                   # Funções RBAC, agendamento de backup, reforço de senha/chave de host
│
├── docs/rotate-secrets.md              # Runbook de rotação de segredos
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Multi-stage: build do Vue + build do Go → alpine
├── docker-compose.yml                  # Produção
└── docker-compose.dev.yml              # Desenvolvimento (apenas DB)
```

---

## Referência da API

A lista completa de rotas está em `backend/internal/server/router.go`.
Endpoints principais:

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

## Permissões RBAC

| Funcionalidade         | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Gerenciar Hosts        | ✓          | ✓   |          |        |               |                    |
| Gerenciar Instâncias   | ✓          | ✓   |          |        |               |                    |
| Adicionar/Editar/Excluir Regra | ✓  | ✓   |          |        | ✓             |                    |
| Ativar/Desativar Regra | ✓          | ✓   | ✓        |        | ✓             |                    |
| Ver Regras             | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Conceder/Revogar Acesso Temporário | ✓ | ✓ | ✓      |        | ✓             | ✓                  |
| Backup e Restauração   | ✓          | ✓   |          |        |               |                    |
| Gerenciamento de Usuários | ✓       |     |          |        |               |                    |
| Ver Log de Auditoria   | ✓          | ✓   |          |        |               |                    |

`AccessManager` e `TempAccessManager` são funções com escopo limitado
para equipes que devem apenas gerenciar regras de acesso / acesso
temporário, sem os privilégios completos de DBA sobre hosts, instâncias,
backups ou usuários.

---

## Checklist de Produção

- [ ] Defina valores reais para `JWT_SECRET`/`ENCRYPTION_KEY` — o aplicativo não inicia sem eles
- [ ] Faça login uma vez e troque a senha padrão do admin (obrigatório no primeiro login)
- [ ] Defina uma `DB_PASSWORD` forte
- [ ] Ative TLS (`DB_SSL_MODE=require`)
- [ ] Coloque o aplicativo atrás de NGINX/Traefik com HTTPS
- [ ] Restrinja a porta do Postgres apenas à rede interna
- [ ] Defina `CORS_ALLOWED_ORIGINS` apenas se o frontend for implantado em uma origem diferente da API
