# PAY-LANGGAN.COM

Platform backend untuk mengelola subscription SaaS. Dibangun dengan Golang (Echo Framework) dan PostgreSQL.

## Tech Stack

- **Go** 1.21+
- **Echo Framework** v4 - HTTP router & middleware
- **PostgreSQL** 16 - Database utama
- **sqlx** - Database driver dengan query builder
- **goose** - Database migration
- **JWT** - Authentication
- **bcrypt** - Password hashing
  - **Xendit** - Payment gateway (sandbox/mock mode)

## Struktur Project

```
├── Backend/
│   ├── cmd/
│   │   ├── api/         # HTTP API server
│   │   └── worker/      # Billing worker
│   ├── internal/
│   │   ├── config/      # Konfigurasi dari .env
│   │   ├── database/    # Koneksi PostgreSQL
│   │   ├── handlers/    # HTTP handlers, dikelompokkan per domain
│   │   │   ├── identity/     # auth, business
│   │   │   ├── catalog/      # service, product, plan, add-on
│   │   │   ├── coupon/
│   │   │   ├── customer/
│   │   │   ├── subscription/
│   │   │   ├── billing/      # placeholder - invoice & recurring billing
│   │   │   ├── payment/      # placeholder - payment gateway webhook
│   │   │   ├── revenue/      # placeholder - revenue dashboard
│   │   │   └── audit/        # placeholder - audit log
│   │   ├── middlewares/  # JWT, error, logging
│   │   ├── models/      # Database structs & per-domain DTO files
│   │   ├── repositories/ # Database queries, dikelompokkan per domain (sama seperti handlers)
│   │   ├── routes/      # Route registrations, satu file per domain
│   │   ├── services/    # Business logic, dikelompokkan per domain (sama seperti handlers)
│   │   └── utils/       # Helpers (response, pagination, jwt, password)
│   ├── pkg/
│   │   └── payment/     # Payment provider interface
│   ├── go.mod / go.sum
│   ├── Makefile
│   └── docker-compose.yml
├── Frontend/         # Aplikasi frontend (Vite)
├── Migrations/       # Goose SQL migrations
└── docs/             # Dokumentasi
```

## Cara Setup

### 1. Clone & Masuk Folder

```bash
git clone <repo-url>
cd pay-langgan/Backend
```

### 2. Copy .env

```bash
cp .env.example .env
# Edit .env sesuai kebutuhan
```

### 3. Jalankan PostgreSQL dengan Docker

```bash
docker compose up -d
```

### 4. Download Dependencies

```bash
go mod tidy
```

### 5. Jalankan Migration

```bash
make migrate-up
```

Cek status migration:

```bash
make migrate-status
```

### 6. Jalankan Server

```bash
make run
```

Server akan berjalan di `http://localhost:8080`.

Cek health endpoint:

```bash
curl http://localhost:8080/health
```



## API Endpoints

### Public

| Method | Endpoint | Deskripsi |
|---|---|---|
| GET | /health | Server health check |
| POST | /api/v1/signup | Registrasi business + admin |
| POST | /api/v1/login | Login user |

### Protected (Bearer Token)

| Method | Endpoint | Deskripsi |
|---|---|---|
| GET | /api/v1/me | Profil user saat ini |
| GET | /api/v1/businesses/me | Detail business |
| PUT | /api/v1/businesses/me | Update business |
| GET | /api/v1/users | List user dalam business (Admin) |
| POST | /api/v1/users | Buat user Sales/Finance (Admin) |
| GET | /api/v1/dashboard/summary | Statistik dashboard (Admin/Finance) |
| GET | /api/v1/services | List services |
| GET | /api/v1/services/:id | Detail service |
| POST | /api/v1/services | Create service |
| PUT | /api/v1/services/:id | Update service |
| DELETE | /api/v1/services/:id | Delete service |
| GET | /api/v1/products | List products |
| GET | /api/v1/products/:id | Detail product |
| POST | /api/v1/products | Create product |
| PUT | /api/v1/products/:id | Update product |
| DELETE | /api/v1/products/:id | Delete product |
| GET | /api/v1/plans | List plans |
| GET | /api/v1/plans/:id | Detail plan |
| POST | /api/v1/plans | Create plan |
| PUT | /api/v1/plans/:id | Update plan |
| DELETE | /api/v1/plans/:id | Delete plan |
| GET | /api/v1/add-ons | List add-ons |
| GET | /api/v1/add-ons/:id | Detail add-on |
| POST | /api/v1/add-ons | Create add-on |
| PUT | /api/v1/add-ons/:id | Update add-on |
| DELETE | /api/v1/add-ons/:id | Delete add-on |
| GET | /api/v1/coupons | List coupons |
| GET | /api/v1/coupons/:id | Detail coupon |
| POST | /api/v1/coupons | Create coupon |
| PUT | /api/v1/coupons/:id | Update coupon |
| DELETE | /api/v1/coupons/:id | Delete coupon |
| GET | /api/v1/customers | List customers |
| GET | /api/v1/customers/:id | Detail customer |
| POST | /api/v1/customers | Create customer |
| PUT | /api/v1/customers/:id | Update customer |
| DELETE | /api/v1/customers/:id | Delete customer |
| POST | /api/v1/subscriptions/preview | Preview subscription price |
| GET | /api/v1/subscriptions | List subscriptions |
| GET | /api/v1/subscriptions/:id | Detail subscription |
| POST | /api/v1/subscriptions | Create subscription |
| POST | /api/v1/subscriptions/:id/cancel | Cancel subscription |
| POST | /api/v1/subscriptions/:id/pause | Pause subscription |
| POST | /api/v1/subscriptions/:id/resume | Resume subscription |
| POST | /api/v1/subscriptions/:id/add-ons | Add add-on to subscription |
| DELETE | /api/v1/subscriptions/:id/add-ons/:add_on_id | Remove add-on from subscription |
| POST | /api/v1/subscriptions/:id/coupons | Apply coupon to subscription |
| DELETE | /api/v1/subscriptions/:id/coupons/:coupon_id | Remove coupon from subscription |

### Contoh Signup

```bash
curl -X POST http://localhost:8080/api/v1/signup \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "PT Demo SaaS",
    "name": "Admin Demo",
    "email": "admin@example.com",
    "password": "password123"
  }'
```

### Contoh Login

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'
```

### Contoh Me (dengan token)

```bash
curl http://localhost:8080/api/v1/me \
  -H "Authorization: Bearer <token_dari_login>"
```

### Contoh Create Service

```bash
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "Meeting App",
    "description": "Aplikasi meeting online"
  }'
```

### Contoh Create Customer

```bash
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "contact": "08123456789"
  }'
```

### Contoh Create Plan

```bash
curl -X POST http://localhost:8080/api/v1/plans \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "product_id": 1,
    "name": "Basic Monthly",
    "price": 100000,
    "billing_cycle": "monthly",
    "trial_days": 14
  }'
```

### Contoh Create Subscription

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "customer_id": 1,
    "plan_id": 1,
    "add_ons": [
      {"add_on_id": 1, "quantity": 2}
    ],
    "coupon_code": "DISC10"
  }'
```

### Contoh Subscription Preview

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions/preview \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "plan_id": 1,
    "add_ons": [
      {"add_on_id": 1, "quantity": 2}
    ],
    "coupon_code": "DISC10"
  }'
```

### Contoh Cancel Subscription

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions/1/cancel \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"reason": "Customer requested cancellation"}'
```

### Contoh List dengan Pagination

```bash
curl "http://localhost:8080/api/v1/customers?page=1&limit=10&search=budi" \
  -H "Authorization: Bearer <token>"
```

## Postman Testing

### Install Newman

```bash
npm install -g newman
```

### Run API Test

```bash
make test-api
```

Atau manual:

```bash
newman run postman/PAY-LANGGAN.postman_collection.json \
  -e postman/PAY-LANGGAN.local.postman_environment.json
```

## Makefile Commands

| Command | Deskripsi |
|---|---|
| `make run` | Jalankan HTTP API server |
| `make migrate-up` | Jalankan semua migration |
| `make migrate-down` | Rollback migration terakhir |
| `make migrate-status` | Cek status migration |
| `make test` | Jalankan semua test |
| `make test-api` | Jalankan Postman API test via Newman |
| `make tidy` | Rapikan go.mod & go.sum |
| `make build` | Build binary |
| `make docker-up` | Start PostgreSQL container |
| `make docker-down` | Stop PostgreSQL container |

## Role dan Alur Akses

- **Admin** dibuat otomatis sebagai user pertama ketika signup dan dapat mengelola business serta membuat user Sales/Finance.
- **Sales** dibuat oleh Admin dan dapat mengelola pelanggan, katalog, kupon, dan subscription.
- **Finance** dibuat oleh Admin dan dapat melihat dashboard, pelanggan, kupon, serta subscription.
- **Super Admin** berada di level platform dan tidak dibuat melalui signup publik. Akunnya harus diprovision secara internal.

Semua user tenant membawa `business_id` di JWT. Endpoint tenant melakukan pembatasan role dan business agar data antar-business tetap terisolasi.

Untuk membuat Super Admin setelah migration role dijalankan:

```bash
go run ./cmd/create-super-admin -name "Platform Admin" -email "superadmin@example.com" -password "password123"
```

## Environment Variables

Semua konfigurasi melalui file `.env`. Lihat `.env.example` untuk daftar lengkap.
