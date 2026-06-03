# Backend Progress

Dokumen ini merangkum progress backend PayLanggan saat ini berdasarkan source code yang ada di repository.

## Ringkasan Status

```text
Foundation/API Server        DONE
Authentication & JWT         DONE
Business Profile API         DONE
Master Data CRUD             DONE
Tenant Isolation             DONE
Customer CRUD                DONE
Coupon CRUD                  DONE
Subscription Lifecycle       DONE
Subscription Pricing Preview DONE
Audit Log for Subscription   DONE
Database Migrations          DONE
Payment Provider Interface   PARTIAL
Midtrans Integration         PARTIAL
Billing Worker               TODO
Invoice Runtime/API          TODO
Payment Runtime/API          TODO
Webhook Handler              TODO
Automated Tests              TODO
```

## Yang Sudah Jadi

### 1. API Server

Backend sudah punya HTTP API server berbasis Echo.

File utama:

```text
cmd/api/main.go
```

Yang sudah tersedia:

```text
Config loader dari .env
Database connection ke PostgreSQL
Echo server
Logger middleware
Recover middleware
CORS middleware
Health endpoint
Dependency injection manual: repository -> service -> handler -> route
```

Endpoint:

```http
GET /health
```

### 2. Authentication

Authentication sudah berjalan dengan JWT.

Endpoint:

```http
POST /api/v1/signup
POST /api/v1/login
GET  /api/v1/me
```

Yang sudah jadi:

```text
Signup business + admin user
Login user
Generate JWT
Parse JWT
JWT middleware untuk protected routes
Password hashing dengan bcrypt
Password verification
Response standard untuk auth
```

File terkait:

```text
internal/handlers/auth_handler.go
internal/services/auth_service.go
internal/repositories/auth_repository.go
internal/middlewares/jwt_middleware.go
internal/utils/jwt.go
internal/utils/password.go
```

### 3. Business Profile

API untuk business milik user login sudah tersedia.

Endpoint:

```http
GET /api/v1/businesses/me
PUT /api/v1/businesses/me
```

Yang sudah jadi:

```text
Ambil detail business dari business_id JWT
Update nama dan meta business
```

File terkait:

```text
internal/handlers/business_handler.go
internal/services/business_service.go
internal/repositories/business_repository.go
```

### 4. Master Data CRUD

CRUD untuk data katalog produk sudah dibuat.

Endpoint service:

```http
GET    /api/v1/services
GET    /api/v1/services/:id
POST   /api/v1/services
PUT    /api/v1/services/:id
DELETE /api/v1/services/:id
```

Endpoint product:

```http
GET    /api/v1/products
GET    /api/v1/products/:id
POST   /api/v1/products
PUT    /api/v1/products/:id
DELETE /api/v1/products/:id
```

Endpoint plan:

```http
GET    /api/v1/plans
GET    /api/v1/plans/:id
POST   /api/v1/plans
PUT    /api/v1/plans/:id
DELETE /api/v1/plans/:id
```

Endpoint add-on:

```http
GET    /api/v1/add-ons
GET    /api/v1/add-ons/:id
POST   /api/v1/add-ons
PUT    /api/v1/add-ons/:id
DELETE /api/v1/add-ons/:id
```

Yang sudah jadi:

```text
List data dengan pagination
Search data
Detail data
Create data
Update data
Delete data
Validasi parent ownership
Validasi billing_cycle untuk plan dan add-on
Validasi status product
```

File terkait:

```text
internal/handlers/service_handler.go
internal/handlers/product_handler.go
internal/handlers/plan_handler.go
internal/handlers/addon_handler.go
internal/services/service_service.go
internal/services/product_service.go
internal/services/plan_service.go
internal/services/addon_service.go
internal/repositories/service_repository.go
internal/repositories/product_repository.go
internal/repositories/plan_repository.go
internal/repositories/addon_repository.go
```

### 5. Tenant Isolation

Backend sudah menerapkan isolasi data antar business.

Yang sudah jadi:

```text
business_id diambil dari JWT
Service difilter berdasarkan business_id
Product dicek lewat relasi product -> service -> business
Plan dicek lewat relasi plan -> product -> service -> business
Add-on dicek lewat relasi add_on -> product -> service -> business
Customer difilter berdasarkan business_id
Subscription difilter berdasarkan customer -> business
```

Catatan:

```text
Coupons saat ini masih global dan belum tenant-specific.
Ini juga sudah tercatat sebagai improvement di docs/TODO.md.
```

### 6. Customer CRUD

Customer management sudah tersedia.

Endpoint:

```http
GET    /api/v1/customers
GET    /api/v1/customers/:id
POST   /api/v1/customers
PUT    /api/v1/customers/:id
DELETE /api/v1/customers/:id
```

Yang sudah jadi:

```text
List customer per business
Search customer
Detail customer
Create customer
Update customer
Delete customer
```

File terkait:

```text
internal/handlers/customer_handler.go
internal/services/customer_service.go
internal/repositories/customer_repository.go
```

### 7. Coupon CRUD

Coupon management sudah tersedia.

Endpoint:

```http
GET    /api/v1/coupons
GET    /api/v1/coupons/:id
POST   /api/v1/coupons
PUT    /api/v1/coupons/:id
DELETE /api/v1/coupons/:id
```

Yang sudah jadi:

```text
List coupon
Search coupon
Detail coupon
Create coupon
Update coupon
Delete coupon
Validasi duplicate code
Validasi discount_type percentage/fixed
Validasi percentage agar tidak lebih dari 100
Validasi max_usage dan expires_at digunakan saat apply coupon
```

Catatan:

```text
Coupon belum memiliki business_id, jadi masih global.
```

File terkait:

```text
internal/handlers/coupon_handler.go
internal/services/coupon_service.go
internal/repositories/coupon_repository.go
```

### 8. Subscription Lifecycle

Subscription adalah fitur backend yang sudah cukup lengkap.

Endpoint:

```http
POST   /api/v1/subscriptions/preview
GET    /api/v1/subscriptions
GET    /api/v1/subscriptions/:id
POST   /api/v1/subscriptions
POST   /api/v1/subscriptions/:id/cancel
POST   /api/v1/subscriptions/:id/pause
POST   /api/v1/subscriptions/:id/resume
POST   /api/v1/subscriptions/:id/add-ons
DELETE /api/v1/subscriptions/:id/add-ons/:add_on_id
POST   /api/v1/subscriptions/:id/coupons
DELETE /api/v1/subscriptions/:id/coupons/:coupon_id
```

Yang sudah jadi:

```text
Preview harga subscription
Create subscription secara transactional
Set status trial jika plan punya trial_days
Set status active jika plan tidak punya trial
Set start_date
Set next_billing_date
List subscription dengan pagination
Filter subscription berdasarkan status
Search subscription berdasarkan customer
Detail subscription lengkap dengan customer, plan, product, service, add-ons, dan coupons
Cancel subscription
Pause subscription
Resume subscription
Add add-on ke subscription
Remove add-on dari subscription
Apply coupon ke subscription
Remove coupon dari subscription
Increment coupon used_count saat apply
Decrement coupon used_count saat remove
Audit log untuk aksi subscription
```

File terkait:

```text
internal/handlers/subscription_handler.go
internal/services/subscription_service.go
internal/services/subscription_pricing_service.go
internal/repositories/subscription_repository.go
internal/repositories/subscription_addon_repository.go
internal/repositories/subscription_coupon_repository.go
internal/repositories/audit_log_repository.go
```

### 9. Database Migrations

Migration database sudah tersedia untuk domain utama.

Tabel yang sudah dibuat:

```text
businesses
users
services
products
plans
add_ons
coupons
customers
subscriptions
subscription_add_ons
payment_methods
invoices
payments
billing_queue
referral_urls
invoice_items
subscription_coupons
payment_attempts
webhook_events
audit_logs
```

Index juga sudah dibuat untuk kolom-kolom penting seperti:

```text
users.email
users.business_id
services.business_id
customers.business_id
subscriptions.customer_id
subscriptions.plan_id
subscriptions.status
subscriptions.next_billing_date
invoices.subscription_id
invoices.status
payments.invoice_id
payments.status
billing_queue.processed
billing_queue.scheduled_at
webhook_events.processed
audit_logs.business_id
audit_logs.created_at
```

Dokumentasi database:

```text
docs/DATABASE_SCHEMA.md
```

### 10. API Documentation

Dokumentasi API sudah tersedia.

File:

```text
docs/API_SPEC.md
docs/API_CONTRACT.md
```

`API_CONTRACT.md` berisi kontrak API yang lebih formal untuk frontend/integrasi:

```text
Base URL
Auth rule
Standard response
Request body
Response schema
Endpoint summary
Validation rules
Endpoint yang belum diimplementasikan
```

### 11. Payment Provider Interface

Package payment sudah mulai dibuat.

File:

```text
pkg/payment/provider.go
pkg/payment/midtrans.go
```

Yang sudah jadi:

```text
Interface Provider
Struct ChargeRequest
Struct ChargeResponse
MidtransProvider
Mock charge
Validasi amount di mock charge
Generate mock transaction ID
```

Status:

```text
PARTIAL
```

Catatan:

```text
Real Midtrans belum diimplementasikan.
Payment provider belum terhubung ke invoice/payment API.
```

## Yang Sebagian Jadi

### 1. Billing & Payment Foundation

Database untuk billing dan payment sudah tersedia, tetapi runtime flow belum dibuat.

Yang sudah ada di database:

```text
invoices
invoice_items
payments
payment_attempts
billing_queue
webhook_events
payment_methods
```

Yang belum ada di API/service:

```text
Generate invoice dari subscription
Create invoice_items dari plan/add-ons/coupon
Charge payment ke provider
Simpan payment_attempts saat charge
Update invoice status
Update payment status
Handle webhook provider
Process billing_queue
```

### 2. Worker

Folder worker ada:

```text
cmd/worker
```

Tetapi source code worker saat ini belum tersedia di folder tersebut.

Status:

```text
TODO dari sisi source code
```

Catatan:

```text
Ada file worker.exe di root project, tetapi source cmd/worker saat ini kosong.
```

## Yang Belum Jadi

### 1. Invoice API

Belum ada route untuk invoice.

Belum tersedia:

```http
GET  /api/v1/invoices
GET  /api/v1/invoices/:id
POST /api/v1/invoices/generate
```

Kebutuhan berikutnya:

```text
Repository invoice
Repository invoice_items
Service invoice generation
Handler invoice read/detail
Route invoice
```

### 2. Payment API

Belum ada route payment.

Belum tersedia:

```http
GET  /api/v1/payments
GET  /api/v1/payments/:id
POST /api/v1/payments/charge
```

Kebutuhan berikutnya:

```text
Repository payment
Repository payment_attempts
Payment service
Payment handler
Integrasi dengan pkg/payment
```

### 3. Webhook API

Belum ada endpoint webhook provider.

Belum tersedia:

```http
POST /api/v1/webhooks/midtrans
```

Kebutuhan berikutnya:

```text
Webhook handler
Validasi signature provider
Simpan payload ke webhook_events
Update payment status
Update invoice status
Idempotency handling
```

### 4. Billing Worker

Belum ada worker untuk proses billing otomatis.

Kebutuhan berikutnya:

```text
Cari subscription yang next_billing_date sudah due
Buat billing_queue
Process billing_queue
Generate invoice
Charge payment
Retry jika gagal
Update retry_count dan last_error
Set processed_at
```

### 5. Audit Log Read API

Audit log sudah dibuat saat aksi subscription, tetapi belum ada endpoint untuk membaca audit log.

Belum tersedia:

```http
GET /api/v1/audit-logs
GET /api/v1/audit-logs/:id
```

### 6. Automated Testing

Belum terlihat test otomatis untuk repository, service, handler, atau integration test.

Belum tersedia:

```text
Unit test repositories
Unit test services
Unit test handlers
Integration test API
Load test
```

## Progress per Layer

```text
cmd/api                  DONE
internal/config          DONE
internal/database        DONE
internal/models          DONE
internal/routes          DONE
internal/middlewares     DONE
internal/utils           DONE
internal/handlers        DONE for active endpoints
internal/services        DONE for active endpoints
internal/repositories    DONE for active endpoints
pkg/payment              PARTIAL
cmd/worker               TODO
tests                    TODO
```

## Endpoint Aktif Saat Ini

```text
GET    /health

POST   /api/v1/signup
POST   /api/v1/login
GET    /api/v1/me

GET    /api/v1/businesses/me
PUT    /api/v1/businesses/me

GET    /api/v1/services
GET    /api/v1/services/:id
POST   /api/v1/services
PUT    /api/v1/services/:id
DELETE /api/v1/services/:id

GET    /api/v1/products
GET    /api/v1/products/:id
POST   /api/v1/products
PUT    /api/v1/products/:id
DELETE /api/v1/products/:id

GET    /api/v1/plans
GET    /api/v1/plans/:id
POST   /api/v1/plans
PUT    /api/v1/plans/:id
DELETE /api/v1/plans/:id

GET    /api/v1/add-ons
GET    /api/v1/add-ons/:id
POST   /api/v1/add-ons
PUT    /api/v1/add-ons/:id
DELETE /api/v1/add-ons/:id

GET    /api/v1/coupons
GET    /api/v1/coupons/:id
POST   /api/v1/coupons
PUT    /api/v1/coupons/:id
DELETE /api/v1/coupons/:id

GET    /api/v1/customers
GET    /api/v1/customers/:id
POST   /api/v1/customers
PUT    /api/v1/customers/:id
DELETE /api/v1/customers/:id

POST   /api/v1/subscriptions/preview
GET    /api/v1/subscriptions
GET    /api/v1/subscriptions/:id
POST   /api/v1/subscriptions
POST   /api/v1/subscriptions/:id/cancel
POST   /api/v1/subscriptions/:id/pause
POST   /api/v1/subscriptions/:id/resume
POST   /api/v1/subscriptions/:id/add-ons
DELETE /api/v1/subscriptions/:id/add-ons/:add_on_id
POST   /api/v1/subscriptions/:id/coupons
DELETE /api/v1/subscriptions/:id/coupons/:coupon_id
```

## Prioritas Berikutnya

Urutan paling masuk akal untuk melanjutkan backend:

```text
1. Invoice repository + invoice_items repository
2. Invoice generation service dari subscription
3. Invoice read/detail endpoint
4. Payment service yang memakai pkg/payment
5. Payment attempt tracking
6. Midtrans webhook handler
7. Billing queue worker
8. Test untuk flow subscription -> invoice -> payment
```

## Kesimpulan

Backend PayLanggan saat ini sudah kuat di bagian foundation, authentication, tenant isolation, CRUD master data, customer, coupon, dan subscription lifecycle.

Bagian yang belum selesai adalah fase billing dan payment: invoice generation, payment charge, payment attempts, webhook, dan worker. Database untuk bagian tersebut sudah siap, tetapi API/service/worker runtime belum dibuat.
