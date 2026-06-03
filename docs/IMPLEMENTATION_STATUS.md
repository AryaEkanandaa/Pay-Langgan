# Implementation Status

## Phase 1 - Foundation (SELESAI)

| Feature | Status | File Terkait | Catatan |
|---|---|---|---|
| Project Structure | DONE | - | Semua folder project sudah dibuat |
| Go Module Init | DONE | go.mod | Dependencies sudah terdaftar |
| Config Loader | DONE | internal/config/config.go | Membaca semua env vars |
| Database Connection | DONE | internal/database/db.go | sqlx + PostgreSQL |
| Migration: businesses | DONE | migrations/001_create_businesses.sql | - |
| Migration: users | DONE | migrations/002_create_users.sql | - |
| Migration: services | DONE | migrations/003_create_services.sql | - |
| Migration: products | DONE | migrations/004_create_products.sql | - |
| Migration: plans | DONE | migrations/005_create_plans.sql | - |
| Migration: add_ons | DONE | migrations/006_create_add_ons.sql | - |
| Migration: coupons | DONE | migrations/007_create_coupons.sql | - |
| Migration: customers | DONE | migrations/008_create_customers.sql | - |
| Migration: subscriptions | DONE | migrations/009_create_subscriptions.sql | - |
| Migration: subscription_add_ons | DONE | migrations/010_create_subscription_add_ons.sql | - |
| Migration: payment_methods | DONE | migrations/011_create_payment_methods.sql | - |
| Migration: invoices | DONE | migrations/012_create_invoices.sql | - |
| Migration: payments | DONE | migrations/013_create_payments.sql | - |
| Migration: billing_queue | DONE | migrations/014_create_billing_queue.sql | - |
| Migration: referral_urls | DONE | migrations/015_create_referral_urls.sql | - |
| Migration: indexes | DONE | migrations/016_create_indexes.sql | - |
| Model: Business | DONE | internal/models/business.go | - |
| Model: User | DONE | internal/models/user.go | - |
| Model: Service | DONE | internal/models/service.go | - |
| Model: Product | DONE | internal/models/product.go | - |
| Model: Plan | DONE | internal/models/plan.go | - |
| Model: AddOn | DONE | internal/models/addon.go | - |
| Model: Coupon | DONE | internal/models/coupon.go | - |
| Model: Customer | DONE | internal/models/customer.go | - |
| Model: Subscription | DONE | internal/models/subscription.go | - |
| Model: Invoice | DONE | internal/models/invoice.go | - |
| Model: Payment | DONE | internal/models/payment.go | - |
| Model: BillingQueue | DONE | internal/models/billing_queue.go | - |
| Model: ReferralURL | DONE | internal/models/referral_url.go | - |
| Model: Request/Response DTO | DONE | internal/models/request_response.go | - |
| Health Endpoint | DONE | cmd/api/main.go | GET /health |
| Docker Compose | DONE | docker-compose.yml | PostgreSQL 16 |
| Makefile | DONE | Makefile | Common commands |
| .env.example | DONE | .env.example | All config keys |
| .gitignore | DONE | .gitignore | - |
| Utils: Response Helper | DONE | internal/utils/response.go | Standard JSON response |
| Utils: Error Helper | DONE | internal/utils/errors.go | Custom error types |
| Utils: JWT Helper | DONE | internal/utils/jwt.go | Token generation & parsing |
| Utils: Password Helper | DONE | internal/utils/password.go | bcrypt hash & verify |
| Utils: ID Generator | DONE | internal/utils/id_generator.go | biz_xxx, INV-xxx |
| JWT Middleware | DONE | internal/middlewares/jwt_middleware.go | Auth middleware with skipper |
| Payment Provider Interface | DONE | pkg/payment/provider.go | PaymentProvider interface |
| Midtrans Provider | DONE | pkg/payment/midtrans.go | Mock + real implementation |
| Worker Skeleton | DONE | cmd/worker/main.go | Billing worker skeleton |
| API Server | DONE | cmd/api/main.go | Echo server with health endpoint |
| Migration: invoice_items | DONE | migrations/017_create_invoice_items.sql | - |
| Migration: subscription_coupons | DONE | migrations/018_create_subscription_coupons.sql | - |
| Migration: payment_attempts | DONE | migrations/019_create_payment_attempts.sql | - |
| Migration: webhook_events | DONE | migrations/020_create_webhook_events.sql | - |
| Migration: audit_logs | DONE | migrations/021_create_audit_logs.sql | - |
| Migration: additional indexes | DONE | migrations/022_create_additional_indexes.sql | - |
| Model: InvoiceItem | DONE | internal/models/invoice_item.go | - |
| Model: SubscriptionCoupon | DONE | internal/models/subscription_coupon.go | - |
| Model: PaymentAttempt | DONE | internal/models/payment_attempt.go | - |
| Model: WebhookEvent | DONE | internal/models/webhook_event.go | - |
| Model: AuditLog | DONE | internal/models/audit_log.go | - |

## Phase 2 - Authentication & API Foundation (SELESAI)

| Feature | Status | File Terkait | Catatan |
|---|---|---|---|
| Auth Signup | DONE | internal/handlers/auth_handler.go | Transactional, bcrypt, JWT |
| Auth Login | DONE | internal/handlers/auth_handler.go | Validasi email & password |
| Auth Repository | DONE | internal/repositories/auth_repository.go | FindUserByEmail, CreateBusinessAndUserTx |
| Auth Service | DONE | internal/services/auth_service.go | Signup, Login, GetMe |
| JWT Utility | DONE | internal/utils/jwt.go | GenerateToken, ParseToken |
| Password Utility | DONE | internal/utils/password.go | HashPassword, CheckPassword |
| ID Generator | DONE | internal/utils/id_generator.go | GenerateBusinessID, GenerateInvoiceNumber |
| Standard Response | DONE | internal/utils/response.go | Success, BadRequest, Unauthorized, dll |
| Error Handling | DONE | internal/utils/errors.go | ErrNotFound, ErrConflict, dll |
| JWT Middleware | DONE | internal/middlewares/jwt_middleware.go | Auth with skipper |
| Protected Route /me | DONE | internal/routes/routes.go | GET /api/v1/me |
| Route Registration | DONE | internal/routes/routes.go | Public + Protected groups |
| DI Wiring | DONE | cmd/api/main.go | Repo -> Service -> Handler -> Route |
| Postman Collection | DONE | postman/ | 7 request + environment |
| Newman Test Command | DONE | Makefile | make test-api |
| API Documentation | DONE | docs/API_SPEC.md | Endpoint detail + Postman guide |

## Phase 3 - CRUD Master Data + Tenant Isolation (SELESAI)

| Feature | Status | File Terkait | Catatan |
|---|---|---|---|
| Business Me API | DONE | internal/handlers/business_handler.go | GET + PUT /businesses/me |
| Business Repository | DONE | internal/repositories/business_repository.go | FindByID, Update |
| Business Service | DONE | internal/services/business_service.go | GetMyBusiness, UpdateMyBusiness |
| Services CRUD | DONE | internal/handlers/service_handler.go | CRUD dengan tenant isolation |
| Services Repository | DONE | internal/repositories/service_repository.go | FindAllByBusinessID, dll |
| Services Service | DONE | internal/services/service_service.go | CRUD dengan validasi |
| Products CRUD | DONE | internal/handlers/product_handler.go | CRUD dengan tenant isolation via JOIN |
| Products Repository | DONE | internal/repositories/product_repository.go | JOIN services untuk tenant check |
| Products Service | DONE | internal/services/product_service.go | Validasi service_id milik business |
| Plans CRUD | DONE | internal/handlers/plan_handler.go | CRUD dengan tenant isolation via JOIN |
| Plans Repository | DONE | internal/repositories/plan_repository.go | JOIN products->services untuk tenant check |
| Plans Service | DONE | internal/services/plan_service.go | Validasi billing_cycle, price, product_id |
| Add-ons CRUD | DONE | internal/handlers/addon_handler.go | CRUD dengan tenant isolation via JOIN |
| Add-ons Repository | DONE | internal/repositories/addon_repository.go | JOIN products->services untuk tenant check |
| Add-ons Service | DONE | internal/services/addon_service.go | Validasi billing_cycle, price, product_id |
| Coupons CRUD | DONE | internal/handlers/coupon_handler.go | CRUD global (tanpa tenant isolation) |
| Coupons Repository | DONE | internal/repositories/coupon_repository.go | FindByCode, Create dengan unique check |
| Coupons Service | DONE | internal/services/coupon_service.go | Validasi discount_type, duplicate code |
| Customers CRUD | DONE | internal/handlers/customer_handler.go | CRUD dengan tenant isolation |
| Customers Repository | DONE | internal/repositories/customer_repository.go | FindAllByBusinessID, dll |
| Customers Service | DONE | internal/services/customer_service.go | CRUD dengan validasi |
| Tenant Isolation | DONE | All Phase 3 | business_id dari JWT, JOIN chain untuk ownership |
| Pagination | DONE | All Phase 3 | Page, limit, ILIKE search |
| Search | DONE | All Phase 3 | ILIKE untuk name/email |
| JSONMap Type | DONE | internal/models/jsonb.go | SQL Scanner/Valuer untuk JSONB |
| Postman Phase 3 | DONE | postman/ | 41 request + environment update |
| API Documentation | DONE | docs/API_SPEC.md | Semua endpoint Phase 3 |
| DB Model JSONB Fix | DONE | internal/models/*.go | JSONMap replaces map[string]any |

## Phase 4 - Subscription Lifecycle (SELESAI)

| Feature | Status | File Terkait | Catatan |
|---|---|---|---|
| Subscription Create | DONE | internal/handlers/subscription_handler.go | Transactional, trial/active, add-ons, coupons |
| Subscription List | DONE | internal/handlers/subscription_handler.go | Pagination, status filter, search customer |
| Subscription Detail | DONE | internal/handlers/subscription_handler.go | Relasi customer, plan, product, service, add-ons, coupons |
| Subscription Cancel | DONE | internal/handlers/subscription_handler.go | trial/active/paused -> cancelled, audit log |
| Subscription Pause | DONE | internal/handlers/subscription_handler.go | active -> paused, audit log |
| Subscription Resume | DONE | internal/handlers/subscription_handler.go | paused -> active, reset next_billing_date, audit log |
| Subscription Add-ons | DONE | internal/handlers/subscription_handler.go | Add/remove add-ons dengan upsert |
| Subscription Coupons | DONE | internal/handlers/subscription_handler.go | Apply/remove, validasi, increment/decrement used_count |
| Subscription Preview | DONE | internal/services/subscription_pricing_service.go | Kalkulasi biaya, breakdown items |
| Audit Log for Subscription | DONE | internal/repositories/audit_log_repository.go | create/cancel/pause/resume/add-on/coupon actions |
| Subscription Repository | DONE | internal/repositories/subscription_repository.go | CreateTx, FindAllByBusinessID, UpdateStatus |
| Subscription AddOn Repository | DONE | internal/repositories/subscription_addon_repository.go | Upsert, Delete, FindBySubscriptionID |
| Subscription Coupon Repository | DONE | internal/repositories/subscription_coupon_repository.go | Apply, Remove, FindBySubscriptionID, Exists check |
| Postman Phase 4 | DONE | postman/ | 52 request (11 subscription baru) |
| API Documentation | DONE | docs/API_SPEC.md | Semua endpoint subscription |
| DB Model JSONB Fix | DONE | internal/models/subscription.go, audit_log.go | JSONMap for JSONB columns |

## Phase 5 - Coming Soon

| Feature | Status | File Terkait | Catatan |
|---|---|---|---|
| Invoice/Payment Read | TODO | - | Belum dibuat |
| Billing Cron/Queue | TODO | - | Belum dibuat |
| Midtrans Integration | TODO | - | Belum dibuat |
| Invoice Generation | TODO | - | Belum dibuat |
| Invoice Items | TODO | - | Belum dibuat |
| Billing Worker | TODO | - | Belum dibuat |
| Webhook Handler | TODO | - | Belum dibuat |
| Unit Test | TODO | - | Belum dibuat |
