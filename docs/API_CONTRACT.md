# API Contract

Dokumen ini adalah kontrak API backend PayLanggan saat ini.
Kontrak ini dibuat untuk menjadi acuan frontend, Postman, atau integrasi client lain.

## Base

```text
Base URL local: http://localhost:8080
API Prefix: /api/v1
Content-Type: application/json
```

## Authentication

Endpoint public:

```text
GET  /health
POST /api/v1/signup
POST /api/v1/login
```

Selain endpoint di atas, semua endpoint wajib memakai Bearer token:

```http
Authorization: Bearer <jwt_token>
```

## Standard Response

### Success

```json
{
  "success": true,
  "message": "Success message",
  "data": {}
}
```

### Success with Pagination

```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": [],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 100
  }
}
```

### Error

```json
{
  "success": false,
  "message": "Error message"
}
```

## Common Query Params

Endpoint list umumnya menerima query berikut:

```text
page=1
limit=10
search=keyword
```

Khusus subscription list:

```text
status=trial|active|paused|cancelled
```

## Common Status Codes

```text
200 OK                  Request berhasil.
201 Created             Data berhasil dibuat.
400 Bad Request          Request body atau parameter tidak valid.
401 Unauthorized         Token tidak ada atau tidak valid.
403 Forbidden            Tidak punya akses ke resource.
404 Not Found            Resource tidak ditemukan.
409 Conflict             Data konflik, misalnya email/coupon duplicate.
500 Internal Server Error Error server.
```

## Data Types

```text
ID integer              Auto increment ID.
business_id string      ID bisnis, contoh: biz_xxxxx.
timestamp string        Format RFC3339, contoh: 2026-01-01T00:00:00Z.
money number            Angka desimal, contoh: 100000 atau 100000.50.
meta object             JSON object fleksibel.
nullable                Bisa bernilai null.
```

## Resource Schemas

### UserDTO

```json
{
  "id": 1,
  "business_id": "biz_xxxxx",
  "name": "Admin Demo",
  "email": "admin@example.com",
  "role": "admin"
}
```

### Business

```json
{
  "id": "biz_xxxxx",
  "name": "PT Demo SaaS",
  "status": "active",
  "deleted": false,
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z",
  "meta": {}
}
```

### Service

```json
{
  "id": 1,
  "business_id": "biz_xxxxx",
  "name": "Meeting App",
  "description": "Aplikasi meeting online",
  "meta": {},
  "created_at": "2026-01-01T00:00:00Z"
}
```

### Product

```json
{
  "id": 1,
  "service_id": 1,
  "name": "Meeting Basic Product",
  "description": "Produk meeting basic",
  "status": "active",
  "meta": {},
  "created_at": "2026-01-01T00:00:00Z"
}
```

Valid `product.status`:

```text
active
inactive
archived
```

### Plan

```json
{
  "id": 1,
  "product_id": 1,
  "name": "Basic Monthly",
  "price": 100000,
  "billing_cycle": "monthly",
  "trial_days": 14,
  "meta": {},
  "created_at": "2026-01-01T00:00:00Z"
}
```

Valid `billing_cycle`:

```text
monthly
yearly
```

### AddOn

```json
{
  "id": 1,
  "product_id": 1,
  "name": "Extra Storage",
  "price": 25000,
  "billing_cycle": "monthly",
  "meta": {},
  "created_at": "2026-01-01T00:00:00Z"
}
```

### Coupon

```json
{
  "id": 1,
  "code": "DISC10",
  "discount_type": "percentage",
  "discount_value": 10,
  "max_usage": 100,
  "used_count": 0,
  "expires_at": "2026-12-31T23:59:59Z",
  "created_at": "2026-01-01T00:00:00Z"
}
```

Valid `discount_type`:

```text
percentage
fixed
```

### Customer

```json
{
  "id": 1,
  "business_id": "biz_xxxxx",
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "contact": "08123456789",
  "meta": {},
  "created_at": "2026-01-01T00:00:00Z"
}
```

### SubscriptionDetail

```json
{
  "id": 1,
  "customer_id": 1,
  "plan_id": 1,
  "status": "trial",
  "start_date": "2026-01-01T00:00:00Z",
  "next_billing_date": "2026-01-15T00:00:00Z",
  "end_date": null,
  "trial_ends_at": "2026-01-15T00:00:00Z",
  "meta": {},
  "created_at": "2026-01-01T00:00:00Z",
  "customer": {
    "id": 1,
    "name": "Budi Santoso",
    "email": "budi@example.com"
  },
  "plan": {
    "id": 1,
    "name": "Basic Monthly",
    "price": 100000,
    "billing_cycle": "monthly",
    "trial_days": 14
  },
  "product": {
    "id": 1,
    "name": "Meeting Basic Product"
  },
  "service": {
    "id": 1,
    "name": "Meeting App"
  },
  "add_ons": [
    {
      "id": 1,
      "add_on_id": 1,
      "name": "Extra Storage",
      "price": 25000,
      "quantity": 2
    }
  ],
  "coupons": [
    {
      "id": 1,
      "coupon_id": 1,
      "code": "DISC10",
      "discount_type": "percentage",
      "discount_value": 10
    }
  ]
}
```

Valid `subscription.status`:

```text
trial
active
paused
cancelled
```

## Health

### GET /health

Auth: public.

Response `200`:

```json
{
  "success": true,
  "message": "PAY-LANGGAN.COM is running",
  "data": {
    "app_name": "PAY-LANGGAN.COM",
    "env": "development"
  }
}
```

## Auth

### POST /api/v1/signup

Auth: public.

Request:

```json
{
  "business_name": "PT Demo SaaS",
  "name": "Admin Demo",
  "email": "admin@example.com",
  "password": "password123"
}
```

Rules:

```text
business_name required
name required
email required, valid email, unique
password required, minimum 6 characters
```

Response `201`:

```json
{
  "success": true,
  "message": "Signup success",
  "data": {
    "token": "<jwt_token>",
    "user": {
      "id": 1,
      "business_id": "biz_xxxxx",
      "name": "Admin Demo",
      "email": "admin@example.com",
      "role": "admin"
    },
    "business": {
      "id": "biz_xxxxx",
      "name": "PT Demo SaaS",
      "status": "active"
    }
  }
}
```

Possible errors:

```text
400 invalid request body
409 email already registered
500 internal server error
```

### POST /api/v1/login

Auth: public.

Request:

```json
{
  "email": "admin@example.com",
  "password": "password123"
}
```

Response `200`:

```json
{
  "success": true,
  "message": "Login success",
  "data": {
    "token": "<jwt_token>",
    "user": {
      "id": 1,
      "business_id": "biz_xxxxx",
      "name": "Admin Demo",
      "email": "admin@example.com",
      "role": "admin"
    }
  }
}
```

Possible errors:

```text
400 invalid request body
401 invalid email or password
500 internal server error
```

### GET /api/v1/me

Auth: bearer token.

Response `200`:

```json
{
  "success": true,
  "message": "User profile retrieved successfully",
  "data": {
    "user_id": 1,
    "business_id": "biz_xxxxx",
    "email": "admin@example.com",
    "role": "admin"
  }
}
```

## Businesses

### GET /api/v1/businesses/me

Auth: bearer token.

Response `200`:

```json
{
  "success": true,
  "message": "Business retrieved successfully",
  "data": {
    "id": "biz_xxxxx",
    "name": "PT Demo SaaS",
    "status": "active",
    "deleted": false,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "meta": {}
  }
}
```

### PUT /api/v1/businesses/me

Auth: bearer token.

Request:

```json
{
  "name": "PT Demo SaaS Updated",
  "meta": {
    "brand_color": "#0f172a"
  }
}
```

Response `200`:

```json
{
  "success": true,
  "message": "Business updated successfully"
}
```

## Services

### GET /api/v1/services

Auth: bearer token.

Query:

```text
page optional, default 1
limit optional, default 10
search optional
```

Response `200`:

```json
{
  "success": true,
  "message": "Services retrieved successfully",
  "data": [
    {
      "id": 1,
      "business_id": "biz_xxxxx",
      "name": "Meeting App",
      "description": "Aplikasi meeting online",
      "meta": {},
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1
  }
}
```

### GET /api/v1/services/:id

Auth: bearer token.

Path params:

```text
id integer, required
```

Response `200`: `Service`

Possible errors:

```text
404 service not found
```

### POST /api/v1/services

Auth: bearer token.

Request:

```json
{
  "name": "Meeting App",
  "description": "Aplikasi meeting online",
  "meta": {}
}
```

Response `201`: `Service`

### PUT /api/v1/services/:id

Auth: bearer token.

Request:

```json
{
  "name": "Meeting App Pro",
  "description": "Updated description",
  "meta": {}
}
```

Response `200`: `Service`

### DELETE /api/v1/services/:id

Auth: bearer token.

Response `200`:

```json
{
  "success": true,
  "message": "Service deleted successfully"
}
```

## Products

### GET /api/v1/products

Auth: bearer token.

Query:

```text
page optional, default 1
limit optional, default 10
search optional
```

Response `200`: paginated array of `Product`.

### GET /api/v1/products/:id

Auth: bearer token.

Response `200`: `Product`.

### POST /api/v1/products

Auth: bearer token.

Request:

```json
{
  "service_id": 1,
  "name": "Meeting Basic Product",
  "description": "Produk meeting basic",
  "status": "active",
  "meta": {}
}
```

Rules:

```text
service_id must belong to current business
status must be active, inactive, or archived
```

Response `201`: `Product`.

### PUT /api/v1/products/:id

Auth: bearer token.

Request:

```json
{
  "name": "Meeting Basic Product Updated",
  "description": "Updated description",
  "status": "active",
  "meta": {}
}
```

Response `200`: `Product`.

### DELETE /api/v1/products/:id

Auth: bearer token.

Response `200`:

```json
{
  "success": true,
  "message": "Product deleted successfully"
}
```

## Plans

### GET /api/v1/plans

Auth: bearer token.

Query:

```text
page optional, default 1
limit optional, default 10
search optional
```

Response `200`: paginated array of `Plan`.

### GET /api/v1/plans/:id

Auth: bearer token.

Response `200`: `Plan`.

### POST /api/v1/plans

Auth: bearer token.

Request:

```json
{
  "product_id": 1,
  "name": "Basic Monthly",
  "price": 100000,
  "billing_cycle": "monthly",
  "trial_days": 14,
  "meta": {}
}
```

Rules:

```text
product_id must belong to current business
price must be greater than or equal to 0
billing_cycle must be monthly or yearly
trial_days defaults to 0 if not used
```

Response `201`: `Plan`.

### PUT /api/v1/plans/:id

Auth: bearer token.

Request:

```json
{
  "name": "Basic Monthly Updated",
  "price": 120000,
  "billing_cycle": "monthly",
  "trial_days": 7,
  "meta": {}
}
```

Response `200`: `Plan`.

### DELETE /api/v1/plans/:id

Auth: bearer token.

Response `200`:

```json
{
  "success": true,
  "message": "Plan deleted successfully"
}
```

## Add-ons

### GET /api/v1/add-ons

Auth: bearer token.

Query:

```text
page optional, default 1
limit optional, default 10
search optional
```

Response `200`: paginated array of `AddOn`.

### GET /api/v1/add-ons/:id

Auth: bearer token.

Response `200`: `AddOn`.

### POST /api/v1/add-ons

Auth: bearer token.

Request:

```json
{
  "product_id": 1,
  "name": "Extra Storage",
  "price": 25000,
  "billing_cycle": "monthly",
  "meta": {}
}
```

Rules:

```text
product_id must belong to current business
price must be greater than or equal to 0
billing_cycle must be monthly or yearly
```

Response `201`: `AddOn`.

### PUT /api/v1/add-ons/:id

Auth: bearer token.

Request:

```json
{
  "name": "Extra Storage Updated",
  "price": 30000,
  "billing_cycle": "monthly",
  "meta": {}
}
```

Response `200`: `AddOn`.

### DELETE /api/v1/add-ons/:id

Auth: bearer token.

Response `200`:

```json
{
  "success": true,
  "message": "Add-on deleted successfully"
}
```

## Coupons

### GET /api/v1/coupons

Auth: bearer token.

Query:

```text
page optional, default 1
limit optional, default 10
search optional
```

Response `200`: paginated array of `Coupon`.

Important:

```text
Coupons are currently global and not filtered by business_id.
```

### GET /api/v1/coupons/:id

Auth: bearer token.

Response `200`: `Coupon`.

### POST /api/v1/coupons

Auth: bearer token.

Request:

```json
{
  "code": "DISC10",
  "discount_type": "percentage",
  "discount_value": 10,
  "max_usage": 100,
  "expires_at": "2026-12-31T23:59:59Z"
}
```

Rules:

```text
code must be unique
discount_type must be percentage or fixed
if discount_type is percentage, discount_value should be between 0 and 100
max_usage nullable
expires_at nullable
```

Response `201`: `Coupon`.

### PUT /api/v1/coupons/:id

Auth: bearer token.

Request:

```json
{
  "code": "DISC20",
  "discount_type": "percentage",
  "discount_value": 20,
  "max_usage": 100,
  "expires_at": "2026-12-31T23:59:59Z"
}
```

Response `200`: `Coupon`.

### DELETE /api/v1/coupons/:id

Auth: bearer token.

Response `200`:

```json
{
  "success": true,
  "message": "Coupon deleted successfully"
}
```

## Customers

### GET /api/v1/customers

Auth: bearer token.

Query:

```text
page optional, default 1
limit optional, default 10
search optional
```

Response `200`: paginated array of `Customer`.

### GET /api/v1/customers/:id

Auth: bearer token.

Response `200`: `Customer`.

### POST /api/v1/customers

Auth: bearer token.

Request:

```json
{
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "contact": "08123456789",
  "meta": {}
}
```

Response `201`: `Customer`.

### PUT /api/v1/customers/:id

Auth: bearer token.

Request:

```json
{
  "name": "Budi Santoso Updated",
  "email": "budi.updated@example.com",
  "contact": "08123456789",
  "meta": {}
}
```

Response `200`: `Customer`.

### DELETE /api/v1/customers/:id

Auth: bearer token.

Response `200`:

```json
{
  "success": true,
  "message": "Customer deleted successfully"
}
```

## Subscriptions

### POST /api/v1/subscriptions/preview

Auth: bearer token.

Request:

```json
{
  "plan_id": 1,
  "add_ons": [
    {
      "add_on_id": 1,
      "quantity": 2
    }
  ],
  "coupon_code": "DISC10"
}
```

Response `200`:

```json
{
  "success": true,
  "message": "Subscription preview calculated successfully",
  "data": {
    "plan_amount": 100000,
    "add_on_amount": 50000,
    "discount_amount": 15000,
    "total_amount": 135000,
    "items": [
      {
        "type": "plan",
        "name": "Basic Monthly",
        "quantity": 1,
        "unit_price": 100000,
        "subtotal": 100000
      },
      {
        "type": "add_on",
        "name": "Extra Storage",
        "quantity": 2,
        "unit_price": 25000,
        "subtotal": 50000
      },
      {
        "type": "discount",
        "name": "Coupon DISC10 (percentage)",
        "quantity": 1,
        "unit_price": -15000,
        "subtotal": -15000
      }
    ]
  }
}
```

### GET /api/v1/subscriptions

Auth: bearer token.

Query:

```text
page optional, default 1
limit optional, default 10
status optional: trial|active|paused|cancelled
search optional: customer search
```

Response `200`: paginated array of `Subscription`.

### GET /api/v1/subscriptions/:id

Auth: bearer token.

Response `200`: `SubscriptionDetail`.

### POST /api/v1/subscriptions

Auth: bearer token.

Request:

```json
{
  "customer_id": 1,
  "plan_id": 1,
  "add_ons": [
    {
      "add_on_id": 1,
      "quantity": 2
    }
  ],
  "coupon_code": "DISC10",
  "meta": {}
}
```

Rules:

```text
customer_id must belong to current business
plan_id must belong to current business
add_on_id must belong to current business
coupon_code optional
if plan has trial_days > 0, status becomes trial
if plan has no trial, status becomes active
```

Response `201`: `SubscriptionDetail`.

### POST /api/v1/subscriptions/:id/cancel

Auth: bearer token.

Request:

```json
{
  "reason": "Customer requested cancellation"
}
```

Rules:

```text
Allowed current status: trial, active, paused
New status: cancelled
end_date will be set to current time
reason will be stored in subscription meta if provided
audit log will be created
```

Response `200`: `SubscriptionDetail`.

### POST /api/v1/subscriptions/:id/pause

Auth: bearer token.

Request:

```json
{
  "reason": "Temporary pause"
}
```

Rules:

```text
Allowed current status: active
New status: paused
reason will be stored in subscription meta if provided
audit log will be created
```

Response `200`: `SubscriptionDetail`.

### POST /api/v1/subscriptions/:id/resume

Auth: bearer token.

Request body: none.

Rules:

```text
Allowed current status: paused
New status: active
next_billing_date will be reset to current time if missing or already in the past
audit log will be created
```

Response `200`: `SubscriptionDetail`.

### POST /api/v1/subscriptions/:id/add-ons

Auth: bearer token.

Request:

```json
{
  "add_on_id": 1,
  "quantity": 3
}
```

Rules:

```text
quantity must be greater than 0
add_on_id must belong to current business
if add-on already exists on subscription, quantity will be updated
audit log will be created
```

Response `200`: `SubscriptionDetail`.

### DELETE /api/v1/subscriptions/:id/add-ons/:add_on_id

Auth: bearer token.

Rules:

```text
Deletes add-on relation from subscription
audit log will be created
```

Response `200`: `SubscriptionDetail`.

### POST /api/v1/subscriptions/:id/coupons

Auth: bearer token.

Request:

```json
{
  "coupon_code": "DISC10"
}
```

Rules:

```text
coupon must exist
coupon must not be expired
coupon used_count must be below max_usage if max_usage is set
same coupon cannot be applied twice to one subscription
coupon used_count will be incremented
audit log will be created
```

Response `200`: `SubscriptionDetail`.

Possible errors:

```text
404 coupon not found
409 coupon already applied to this subscription
```

### DELETE /api/v1/subscriptions/:id/coupons/:coupon_id

Auth: bearer token.

Rules:

```text
Deletes coupon relation from subscription
coupon used_count will be decremented if greater than 0
audit log will be created
```

Response `200`: `SubscriptionDetail`.

## Endpoint Summary

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

## Not Implemented Yet

Endpoint berikut belum ada di route backend saat ini, walaupun beberapa tabel database sudah tersedia:

```text
Invoice read API
Invoice generation API
Payment charge API
Payment attempts API
Billing queue API
Midtrans webhook API
Audit log read API
```
