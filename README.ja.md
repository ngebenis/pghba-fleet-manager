# PGHBA Fleet Manager

🇮🇩 [ID](README.md) · 🇬🇧 [EN](README.en.md) · 🇨🇳 [ZH](README.zh.md) · 🇪🇸 [ES](README.es.md) · 🇸🇦 [AR](README.ar.md) · 🇫🇷 [FR](README.fr.md) · 🇵🇹 [PT](README.pt.md) · 🇧🇩 [BN](README.bn.md) · 🇷🇺 [RU](README.ru.md) · 🇮🇳 [HI](README.hi.md) · 🇹🇭 [TH](README.th.md) · 🇰🇭 [KM](README.km.md) · 🇵🇭 [TL](README.tl.md) · 🇻🇳 [VI](README.vi.md) · 🇯🇵 **JA** · 🇰🇷 [KO](README.ko.md) · 🇵🇱 [PL](README.pl.md) · 🇮🇹 [IT](README.it.md) · 🇹🇷 [TR](README.tr.md) · 🇮🇳 [TA](README.ta.md) · 🇮🇷 [FA](README.fa.md) · 🇺🇦 [UK](README.uk.md) · 🇩🇪 [DE](README.de.md) · 🇳🇱 [NL](README.nl.md)

PostgreSQLの`pg_hba.conf`を一元管理するシステムです。各ホストで
`pg_hba.conf`を手作業で編集する代わりに、1つのWeb UIから
PostgreSQLインスタンス群全体のアクセスルールを管理できます。

このリポジトリは、オリジナルのPGHBA Managerプロジェクトを
ハードニングし、公開用に整備したスナップショットです。
デフォルトパスワードの強制変更、起動時のシークレット検証、
SSHホストキーのピン留め、ログインのレート制限、そして
push毎に依存関係をスキャンするCIパイプラインを備えています。
詳細は下記の[セキュリティ](#セキュリティ)をご覧ください。

🌐 **24言語に標準対応** — トップバーのドロップダウンからいつでも
切り替え可能です: English、Bahasa Indonesia、中文、Español、العربية、
Français、Português、বাংলা、Русский、हिन्दी、ไทย、ខ្មែរ、Tagalog、
Tiếng Việt、日本語、한국어、Polski、Italiano、Türkçe、தமிழ்、فارسی、
Українська、Deutsch、Nederlands。アラビア語とペルシア語は自動的に
右から左へ表示されます。

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ コーヒー一杯分のご支援をいただけると嬉しいです。

---

## スクリーンショット

### ダッシュボード

![Dashboard](docs/screenshots/dashboard.png)

> 以下のスクリーンショット(データベースホスト、PGインスタンス、
> アクセスルール、一時アクセス、バックアップと復元、ユーザー管理、
> 監査ログ、設定)は、本リポジトリの実際のフロントエンドコンポーネントと
> カラースキームを元に作成した**イメージ用のモックアップ**であり、
> サンプルデータを使用しています。実稼働環境のスクリーンショットでは
> ありません。実際のスクリーンショットは今後差し替えられます。

<details>
<summary>データベースホスト</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>PGインスタンス</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>アクセスルール</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>一時アクセス</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>バックアップと復元</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>ユーザー管理</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>監査ログ</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>設定</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## クイックスタート(Docker)

```bash
# 1. クローンして設定
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. .envをテンプレートから作成(docker-compose.ymlが使用)
cp backend/.env.example .env

# 3. 実際のシークレットを生成 — これがないとアプリは起動しません
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. .envを実際の値で編集してから:
docker compose up -d

# アプリは次のURLで利用可能: http://localhost:8080
```

**初回ログイン**: `admin` / `Admin@1234`。このアカウントは初回ログイン時に
パスワード変更が強制されます。変更が完了するまでサーバーは他の操作を
すべてブロックするため、「デフォルトパスワードを変更し忘れる」という
隙間はありません。

---

## 開発環境のセットアップ

```bash
# データベースのみ起動
docker compose -f docker-compose.dev.yml up -d

# バックエンド(Go)
cd backend
cp .env.example .env       # 必要に応じて編集
go run ./cmd/server

# フロントエンド(Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (/apiを:8080へプロキシ)
```

---

## セキュリティ

- **安全でないデフォルト設定では起動できません。** `main.go`は起動時に
  `config.Validate()`を呼び出し、`JWT_SECRET`や`ENCRYPTION_KEY`が
  空または公開されたプレースホルダーのままの場合は起動を拒否します。
  設定ミスのあるデプロイは、偽造可能なJWTシークレットや既知のAESキーで
  動作するのではなく、はっきりと失敗します。
- **SSHホストキーはピン留め**され、無視されません。
  `services.ConnectToHost`は管理対象ホストのSSHキーを初回接続時に
  信頼し、そのフィンガープリント(`db_hosts.ssh_host_key_fingerprint`)を
  保存します。以降、異なるキーでの接続は黙って信頼されるのではなく
  拒否されます。正当なホスト再構築後にフィンガープリントをクリアする
  場合は、一般的なホスト編集フォームではなく、専用の監査対象エンドポイント
  (`POST /api/hosts/:id/reset-host-key`)を経由します。
- **ログインにはレート制限**があります。15分以内に同一IPから
  10回ログインに失敗すると、そのIPは15分間ロックアウトされます。
- **ユーザーの無効化や降格は即座に反映されます** — 認証済みのすべての
  リクエストは、最大8時間前のJWTの主張を信頼するのではなく、
  データベースに対してステータス/ロールを再チェックします。
- **CIは既知の脆弱性のある依存関係をスキャンします**。push/PRごと、
  および毎週(`.github/workflows/security.yml`)、Goバックエンドには
  `govulncheck`、フロントエンドには`npm audit`(本番依存関係)を
  使用します。
- **データベースマイグレーションは組み込まれ、自動的に適用されます。**
  `backend/internal/database/migrations/`内のすべてのSQLファイルは
  バイナリに埋め込まれ(`go:embed`)、起動のたびに自動的に適用されます
  — `schema_migrations`テーブルで追跡されるため、手動実行の
  マイグレーションスクリプトのようにデプロイがスキーマ変更を
  黙って見逃すことはありません。
- バックアップのエクスポート、CORS、HTTPレスポンスヘッダーも
  ハードニングされています。詳細は`backend/internal/handlers/export.go`
  (パスセグメントのサニタイズ)、`backend/internal/server/router.go`
  (`CORS_ALLOWED_ORIGINS`が設定されない限りクロスオリジンアクセスを
  デフォルトで拒否)、`backend/internal/middleware/security_headers.go`
  を参照してください。
- **漏洩後のシークレットローテーション**: `backend/cmd/reencrypt-ssh-keys`
  と[`docs/rotate-secrets.md`](docs/rotate-secrets.md)では、データを
  失うことなく`JWT_SECRET`/`ENCRYPTION_KEY`をローテーションし、
  保存されているすべてのSSHキーを再暗号化する手順を説明しています。

---

## プロジェクト構成

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # エントリーポイント
│   │   └── reencrypt-ssh-keys/         # シークレットローテーションツール(セキュリティ参照)
│   └── internal/
│       ├── config/                     # 設定ローダー + 起動時シークレット検証
│       ├── database/                   # GORM接続 + 埋め込みSQLマイグレーションランナー
│       │   └── migrations/             # 001(initdb経由のベースライン)+ 002–008、起動時に自動適用
│       ├── models/                     # すべてのGORMモデル
│       ├── middleware/                 # JWT認証、RBAC、レート制限、セキュリティヘッダー
│       ├── handlers/                   # Gin HTTPハンドラー
│       │   ├── auth.go                 # ログイン / ログアウト / リフレッシュ / パスワード変更
│       │   ├── hosts.go                # ホストCRUD、SSH検証、デプロイキー、ホストキーリセット
│       │   ├── instances.go            # インスタンスの検出/スキャン
│       │   ├── rules.go                # pg_hbaルールCRUD + 有効/無効/並べ替え/クローン
│       │   ├── temp_access.go          # 一時アクセスの付与/取り消し
│       │   ├── backup.go, backup_schedules.go   # バックアップ作成/復元/ダウンロード + スケジューリング
│       │   ├── export.go               # 設定/バックアップのエクスポート(ローカル + S3)
│       │   ├── healthcheck.go          # ホストのヘルスチェック + 通知
│       │   ├── users.go                # ユーザーCRUD(SuperAdminのみ)
│       │   ├── audit.go                # 監査ログの一覧 + 差分
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSHクライアント、ホストキーピン留め(ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # pg_hba内容のレンダリング + アトミック書き込み + 有効期限
│       │   ├── crypto.go               # SSHキー用のAES-256-GCM
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — N分ごとに一時ルールを失効
│       └── server/router.go            # Ginルーター、CORS、全ルート
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios APIクライアント
│       ├── types/index.ts              # TypeScriptインターフェース(Goモデルをミラー)
│       ├── i18n/                       # 24のロケールファイル + AVAILABLE_LOCALESレジストリ(RTL対応)
│       ├── stores/auth.ts              # Pinia認証ストア
│       ├── router/index.ts             # 認証 + パスワード強制変更ガード付きVue Router
│       ├── components/layout/AppLayout.vue
│       └── views/                      # 機能領域ごとのビュー
│
├── docs/rotate-secrets.md              # シークレットローテーションのランブック
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # マルチステージ: Vueビルド + Goビルド → alpine
├── docker-compose.yml                  # 本番用
└── docker-compose.dev.yml              # 開発用(DBのみ)
```

---

## APIリファレンス

完全なルート一覧は`backend/internal/server/router.go`にあります。
主なエンドポイント:

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
POST   /api/hosts/validate              # 保存前のSSH事前チェック
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # ピン留めされたSSHホストキーフィンガープリントをクリア

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

## RBAC権限

| 機能                   | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| ホスト管理              | ✓          | ✓   |          |        |               |                    |
| インスタンス管理         | ✓          | ✓   |          |        |               |                    |
| ルールの追加/編集/削除   | ✓          | ✓   |          |        | ✓             |                    |
| ルールの有効/無効化      | ✓          | ✓   | ✓        |        | ✓             |                    |
| ルールの閲覧            | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| 一時アクセスの付与/取り消し | ✓        | ✓   | ✓        |        | ✓             | ✓                  |
| バックアップと復元       | ✓          | ✓   |          |        |               |                    |
| ユーザー管理            | ✓          |     |          |        |               |                    |
| 監査ログの閲覧          | ✓          | ✓   |          |        |               |                    |

`AccessManager`と`TempAccessManager`は、ホスト、インスタンス、
バックアップ、ユーザーに対する完全なDBA権限なしに、アクセスルール /
一時アクセスのみを管理するチーム向けのスコープ限定ロールです。

---

## 本番環境チェックリスト

- [ ] 実際の`JWT_SECRET`/`ENCRYPTION_KEY`を設定する — 設定しないとアプリは起動しません
- [ ] 一度ログインしてデフォルトの管理者パスワードを変更する(初回ログイン時に強制)
- [ ] 強力な`DB_PASSWORD`を設定する
- [ ] TLSを有効化する(`DB_SSL_MODE=require`)
- [ ] HTTPS付きのNGINX/Traefikの背後にアプリを配置する
- [ ] Postgresのポートを内部ネットワークのみに制限する
- [ ] フロントエンドがAPIと異なるオリジンにデプロイされている場合のみ`CORS_ALLOWED_ORIGINS`を設定する
