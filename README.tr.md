# PGHBA Fleet Manager

🇮🇩 [ID](README.md) · 🇬🇧 [EN](README.en.md) · 🇨🇳 [ZH](README.zh.md) · 🇪🇸 [ES](README.es.md) · 🇸🇦 [AR](README.ar.md) · 🇫🇷 [FR](README.fr.md) · 🇵🇹 [PT](README.pt.md) · 🇧🇩 [BN](README.bn.md) · 🇷🇺 [RU](README.ru.md) · 🇮🇳 [HI](README.hi.md) · 🇹🇭 [TH](README.th.md) · 🇰🇭 [KM](README.km.md) · 🇵🇭 [TL](README.tl.md) · 🇻🇳 [VI](README.vi.md) · 🇯🇵 [JA](README.ja.md) · 🇰🇷 [KO](README.ko.md) · 🇵🇱 [PL](README.pl.md) · 🇮🇹 [IT](README.it.md) · 🇹🇷 **TR** · 🇮🇳 [TA](README.ta.md) · 🇮🇷 [FA](README.fa.md) · 🇺🇦 [UK](README.uk.md) · 🇩🇪 [DE](README.de.md) · 🇳🇱 [NL](README.nl.md)

PostgreSQL `pg_hba.conf` merkezi yönetim sistemi — bir PostgreSQL sunucu
filosundaki erişim kurallarını tek bir web arayüzünden yönetin, her
sunucuda `pg_hba.conf` dosyasını tek tek elle düzenlemek yerine.

Bu depo, orijinal PGHBA Manager projesinin güçlendirilmiş, herkese açık
yayına hazır bir anlık görüntüsüdür: varsayılan parola zorunluluğu,
başlangıçta gizli anahtar doğrulaması, SSH host key pinleme, giriş hız
sınırlaması ve her push'ta bağımlılıkları tarayan bir CI pipeline'ı
içerir. Aşağıdaki [Güvenlik](#güvenlik) bölümüne bakın.

🌐 **Kullanıma hazır 24 dil** — üst çubuktaki açılır menüden istediğiniz
zaman değiştirin: English, Bahasa Indonesia, 中文, Español, العربية,
Français, Português, বাংলা, Русский, हिन्दी, ไทย, ខ្មែរ, Tagalog,
Tiếng Việt, 日本語, 한국어, Polski, Italiano, Türkçe, தமிழ், فارسی,
Українська, Deutsch, Nederlands. Arapça ve Farsça otomatik olarak
sağdan sola görüntülenir.

[![Donate via PayPal](https://img.shields.io/badge/Donate-PayPal-00457C?style=flat&logo=paypal&logoColor=white)](https://www.paypal.com/donate/?business=b3n1_id@yahoo.com)
<br>☕ Bana bir fincan kahve ısmarlayabilirsiniz.

---

## Ekran Görüntüleri

### Pano

![Dashboard](docs/screenshots/dashboard.png)

> Aşağıdaki ekran görüntüleri (Veritabanı Sunucuları, PG Örnekleri, Erişim
> Kuralları, Geçici Erişim, Yedekleme ve Geri Yükleme, Kullanıcı Yönetimi,
> Denetim Günlüğü, Ayarlar) bu deponun gerçek ön yüz bileşenleri ve renk
> şeması kullanılarak oluşturulmuş, örnek/sahte verilerle doldurulmuş
> **gösterim amaçlı maketlerdir** — canlı bir dağıtımdan alınmamıştır.
> Gerçek ekran görüntüleri hazır olduğunda bunların yerini alacaktır.

<details>
<summary>Veritabanı Sunucuları</summary>

![Database Hosts](docs/screenshots/hosts.png)
</details>

<details>
<summary>PG Örnekleri</summary>

![PG Instances](docs/screenshots/instances.png)
</details>

<details>
<summary>Erişim Kuralları</summary>

![Access Rules](docs/screenshots/rules.png)
</details>

<details>
<summary>Geçici Erişim</summary>

![Temporary Access](docs/screenshots/temp-access.png)
</details>

<details>
<summary>Yedekleme ve Geri Yükleme</summary>

![Backup & Restore](docs/screenshots/backup.png)
</details>

<details>
<summary>Kullanıcı Yönetimi</summary>

![User Management](docs/screenshots/users.png)
</details>

<details>
<summary>Denetim Günlüğü</summary>

![Audit Log](docs/screenshots/audit.png)
</details>

<details>
<summary>Ayarlar</summary>

![Settings](docs/screenshots/settings.png)
</details>

---

## Hızlı Başlangıç (Docker)

```bash
# 1. Klonlayın ve yapılandırın
git clone https://github.com/ngebenis/pghba-fleet-manager.git
cd pghba-fleet-manager

# 2. Şablondan .env oluşturun (docker-compose.yml tarafından kullanılır)
cp backend/.env.example .env

# 3. Gerçek gizli anahtarlar üretin — uygulama bunlar olmadan başlamaz
openssl rand -base64 48   # → JWT_SECRET
openssl rand -hex 32      # → ENCRYPTION_KEY

# 4. .env dosyasını gerçek değerlerinizle düzenleyin, ardından:
docker compose up -d

# Uygulama şu adreste kullanılabilir: http://localhost:8080
```

**İlk giriş**: `admin` / `Admin@1234`. Bu hesap, ilk girişte parola
değiştirmeye zorlanacak şekilde işaretlenmiştir — parola değiştirilene
kadar sunucu diğer tüm işlemleri engeller, böylece "varsayılan parolayı
değiştirmeyi unutma" riski ortadan kalkar.

---

## Geliştirme Kurulumu

```bash
# Sadece veritabanını başlatın
docker compose -f docker-compose.dev.yml up -d

# Backend (Go)
cd backend
cp .env.example .env       # gerektiği gibi düzenleyin
go run ./cmd/server

# Frontend (Vue 3)
cd frontend
npm install
npm run dev
# → http://localhost:5173 (/api isteklerini :8080'e yönlendirir)
```

---

## Güvenlik

- **Güvensiz varsayılanlarla çalışamaz.** `main.go`, başlangıçta
  `config.Validate()` fonksiyonunu çağırır; bu fonksiyon `JWT_SECRET`
  veya `ENCRYPTION_KEY` boşsa ya da hâlâ yayınlanan yer tutucu değerse
  başlamayı reddeder — yanlış yapılandırılmış bir dağıtım, sahtelenebilir
  bir JWT anahtarı veya bilinen bir AES anahtarıyla çalışmak yerine
  açıkça başarısız olur.
- **SSH host key'leri yoksayılmaz, pinlenir.** `services.ConnectToHost`,
  yönetilen bir sunucunun SSH anahtarına ilk bağlantıda güvenir ve
  parmak izini (`db_hosts.ssh_host_key_fingerprint`) kalıcı hale getirir;
  farklı bir anahtarla yapılan sonraki bağlantılar sessizce kabul
  edilmek yerine reddedilir. Meşru bir sunucu yeniden kurulumundan sonra
  parmak izini temizlemek, genel sunucu düzenleme formu yerine özel,
  denetlenen bir uç nokta (`POST /api/hosts/:id/reset-host-key`)
  üzerinden yapılır.
- **Giriş hız sınırlıdır**: 15 dakika içinde bir IP'den gelen 10 başarısız
  deneme, o IP'yi 15 dakikalığına kilitler.
- **Bir kullanıcının devre dışı bırakılması veya rolünün düşürülmesi
  hemen etkili olur** — her kimliği doğrulanmış istek, JWT'nin 8 saate
  kadar eski olabilecek iddiasına güvenmek yerine durumu/rolü
  veritabanına karşı yeniden kontrol eder.
- **CI, her push/PR'da ve haftalık olarak** (`.github/workflows/security.yml`)
  bilinen güvenlik açığı bulunan bağımlılıkları tarar: Go backend için
  `govulncheck`, frontend için `npm audit` (üretim bağımlılıkları).
- **Veritabanı migration'ları gömülüdür ve kendiliğinden uygulanır.**
  `backend/internal/database/migrations/` içindeki her SQL dosyası
  ikili dosyanın (binary) içine gömülür (`go:embed`) ve her başlangıçta
  otomatik olarak uygulanır — bir `schema_migrations` tablosunda takip
  edilerek — böylece bir dağıtım, elle çalıştırılan bir migration
  betiğinde olabileceği gibi bir şema değişikliğini sessizce asla
  atlayamaz.
- Yedek dışa aktarma, CORS ve HTTP yanıt başlıkları güçlendirilmiştir —
  bkz. `backend/internal/handlers/export.go` (yol segmenti temizleme),
  `backend/internal/server/router.go` (`CORS_ALLOWED_ORIGINS`
  ayarlanmadıkça CORS varsayılan olarak çapraz kaynak erişimine izin
  vermez) ve `backend/internal/middleware/security_headers.go`.
- **Bir sızıntıdan sonra gizli anahtarları döndürme**:
  `backend/cmd/reencrypt-ssh-keys` ve
  [`docs/rotate-secrets.md`](docs/rotate-secrets.md), `JWT_SECRET`/
  `ENCRYPTION_KEY` döndürmeyi ve saklanan tüm SSH anahtarlarını veri
  kaybı olmadan yeniden şifrelemeyi adım adım anlatır.

---

## Proje Yapısı

```
pghba-fleet-manager/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go              # Giriş noktası
│   │   └── reencrypt-ssh-keys/         # Gizli anahtar döndürme aracı (bkz. Güvenlik)
│   └── internal/
│       ├── config/                     # Yapılandırma yükleyici + başlangıç gizli anahtar doğrulaması
│       ├── database/                   # GORM bağlantısı + gömülü SQL migration motoru
│       │   └── migrations/             # 001 (temel, initdb ile) + 002–008, başlangıçta otomatik uygulanır
│       ├── models/                     # Tüm GORM modelleri
│       ├── middleware/                 # JWT auth, RBAC, hız sınırlama, güvenlik başlıkları
│       ├── handlers/                   # Gin HTTP işleyicileri
│       │   ├── auth.go                 # Giriş / çıkış / yenileme / parola değiştirme
│       │   ├── hosts.go                # Sunucu CRUD, SSH doğrulama, deploy-key, reset-host-key
│       │   ├── instances.go            # Örnek keşfi/taraması
│       │   ├── rules.go                # pg_hba kural CRUD + etkinleştir/devre dışı bırak/yeniden sırala/klonla
│       │   ├── temp_access.go          # Geçici erişim verme/iptal etme
│       │   ├── backup.go, backup_schedules.go   # Yedek oluşturma/geri yükleme/indirme + zamanlama
│       │   ├── export.go               # Yapılandırma/yedek dışa aktarma (yerel + S3)
│       │   ├── healthcheck.go          # Sunucu sağlık kontrolleri + bildirimler
│       │   ├── users.go                # Kullanıcı CRUD (yalnızca SuperAdmin)
│       │   ├── audit.go                # Denetim günlüğü listesi + fark
│       │   └── dashboard.go, settings.go, conflicts.go, notifications.go
│       ├── services/
│       │   ├── ssh.go                  # SSH istemcisi, host key pinleme (ConnectToHost)
│       │   ├── hba.go, hba_writer.go    # pg_hba içerik oluşturma + atomik yazma + süre sonu
│       │   ├── crypto.go               # SSH anahtarları için AES-256-GCM
│       │   ├── audit.go, notifier.go
│       ├── scheduler/                  # Cron — her N dakikada geçici kuralların süresini doldur
│       └── server/router.go            # Gin router, CORS, tüm rotalar
│
├── frontend/
│   └── src/
│       ├── api/index.ts                # Axios API istemcisi
│       ├── types/index.ts              # TypeScript arayüzleri (Go modellerini yansıtır)
│       ├── i18n/                       # 24 dil dosyası + AVAILABLE_LOCALES kaydı (RTL destekli)
│       ├── stores/auth.ts              # Pinia auth store
│       ├── router/index.ts             # Auth + zorunlu parola değiştirme korumalı Vue Router
│       ├── components/layout/AppLayout.vue
│       └── views/                      # Her özellik alanı için bir görünüm
│
├── docs/rotate-secrets.md              # Gizli anahtar döndürme kılavuzu
├── .github/workflows/security.yml      # CI: build/vet/govulncheck + npm audit
├── Dockerfile                          # Çok aşamalı: Vue build + Go build → alpine
├── docker-compose.yml                  # Üretim
└── docker-compose.dev.yml              # Geliştirme (yalnızca DB)
```

---

## API Referansı

Tam rota listesi `backend/internal/server/router.go` içinde yer alır.
Temel uç noktalar:

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
POST   /api/hosts/validate              # Kaydetmeden önce SSH ön kontrolü
POST   /api/hosts/:id/deploy-key
POST   /api/hosts/:id/reset-host-key    # Pinlenmiş SSH host key parmak izini temizle

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

## RBAC İzinleri

| Özellik                 | SuperAdmin | DBA | Operator | Viewer | AccessManager | TempAccessManager |
|--------------------------|:----------:|:---:|:--------:|:------:|:-------------:|:-----------------:|
| Sunucuları Yönet          | ✓          | ✓   |          |        |               |                    |
| Örnekleri Yönet           | ✓          | ✓   |          |        |               |                    |
| Kural Ekle/Düzenle/Sil    | ✓          | ✓   |          |        | ✓             |                    |
| Kuralı Etkinleştir/Devre Dışı Bırak | ✓ | ✓   | ✓        |        | ✓             |                    |
| Kuralları Görüntüle       | ✓          | ✓   | ✓        | ✓      | ✓             |                    |
| Geçici Erişim Ver/İptal Et | ✓        | ✓   | ✓        |        | ✓             | ✓                  |
| Yedekleme ve Geri Yükleme | ✓          | ✓   |          |        |               |                    |
| Kullanıcı Yönetimi        | ✓          |     |          |        |               |                    |
| Denetim Günlüğünü Görüntüle | ✓        | ✓   |          |        |               |                    |

`AccessManager` ve `TempAccessManager`, sunucular, örnekler, yedekler
veya kullanıcılar üzerinde tam DBA yetkisi olmadan yalnızca erişim
kurallarını/geçici erişimi yönetmesi gereken ekipler için kapsamı
sınırlı rollerdir.

---

## Üretim Kontrol Listesi

- [ ] Gerçek `JWT_SECRET`/`ENCRYPTION_KEY` ayarlayın — aksi takdirde uygulama başlamaz
- [ ] Bir kez giriş yapıp varsayılan admin parolasını değiştirin (ilk girişte zorunludur)
- [ ] Güçlü bir `DB_PASSWORD` belirleyin
- [ ] TLS'i etkinleştirin (`DB_SSL_MODE=require`)
- [ ] Uygulamayı HTTPS ile NGINX/Traefik arkasına yerleştirin
- [ ] Postgres'in portunu yalnızca iç ağla sınırlayın
- [ ] Frontend, API'den farklı bir kaynakta dağıtılıyorsa `CORS_ALLOWED_ORIGINS` ayarlayın
