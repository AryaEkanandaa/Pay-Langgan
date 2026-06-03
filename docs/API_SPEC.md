# API Specification

## Base URL

```
http://localhost:8080
```

## Authentication

Semua endpoint `/api/v1/*` kecuali `/signup` dan `/login` memerlukan JWT Bearer Token.

```
Authorization: Bearer <token>
```

## Response Format

### Success

```json
{
  "success": true,
  "message": "Data retrieved successfully",
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
    "total": 20
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

## Endpoints

### Health

#### GET /health

Cek status server.

Response:

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

### Auth

#### POST /api/v1/signup

Registrasi business baru + user admin.

Request:

```json
{
  "business_name": "PT Demo SaaS",
  "name": "Admin Demo",
  "email": "admin@example.com",
  "password": "password123"
}
```

Response 201:

```json
{
  "success": true,
  "message": "Signup success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
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

Response 409:

```json
{
  "success": false,
  "message": "email already registered"
}
```

#### POST /api/v1/login

Login user.

Request:

```json
{
  "email": "admin@example.com",
  "password": "password123"
}
```

Response 200:

```json
{
  "success": true,
  "message": "Login success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
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

Response 401:

```json
{
  "success": false,
  "message": "invalid email or password"
}
```

#### GET /api/v1/me

Lihat profil user yang sedang login. (Protected)

Headers: `Authorization: Bearer <token>`

Response 200:

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

Response 401:

```json
{
  "success": false,
  "message": "missing authorization header"
}
```

### Businesses

#### GET /api/v1/businesses/me

Ambil detail business milik user login. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Business retrieved successfully",
  "data": {
    "id": "biz_xxxxx",
    "name": "PT Demo SaaS",
    "status": "active",
    "meta": {}
  }
}
```

#### PUT /api/v1/businesses/me

Update business milik user login. (Protected)

Request:

```json
{
  "name": "PT Demo SaaS Updated"
}
```

Response 200:

```json
{
  "success": true,
  "message": "Business updated successfully"
}
```

### Services

#### GET /api/v1/services

List services milik business user login. (Protected)

Query params: `?page=1&limit=10&search=keyword`

Response 200:

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

#### GET /awl/v1/services/:id

Detail service. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Service retrieved successfully",
  "data": {
    "id": 1,
    "business_id": "biz_xxxxx",
    "name": "Meeting App",
    "description": "Aplikasi meeting online",
    "meta": {},
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

Response 404:

```json
{
  "success": false,
  "message": "service not found"
}
```

#### POST /api/v1/services

Buat service baru. (Protected)

Request:

```json
{
  "name": "Meeting App",
  "description": "Aplikasi meeting online"
}
```

Response 201:

```json
{
  "success": true,
  "message": "Service created successfully",
  "data": {
    "id": 1,
    "business_id": "biz_xxxxx",
    "name": "Meeting App",
    "description": "Aplikasi meeting online",
    "meta": {},
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

#### PUT /api/v1/services/:id

Update service. (Protected)

Request:

```json
{
  "name": "Meeting App Pro",
  "description": "Updated description"
}
```

Response 200:

```json
{
  "success": true,
  "message": "Service updated successfully",
  "data": { ... }
}
```

#### DELETE /api/v1/services/:id

Hapus service. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Service deleted successfully"
}
```

### Products

#### GET /api/v1/products

List products milik business user login (via service relationship). (Protected)

Query params: `?page=1&limit=10&search=keyword`

Response 200:

```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": 1,
      "service_id": 1,
      "name": "Meeting Basic Product",
      "description": "Produk meeting basic",
      "status": "active",
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

#### GET /api/v1/products/:id

Detail product. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Product retrieved successfully",
  "data": { ... }
}
```

Response 404:

```json
{
  "success": false,
  "message": "product not found"
}
```

#### POST /api/v1/products

Buat product baru. (Protected) Service_id harus milik business user.

Request:

```json
{
  "service_id": 1,
  "name": "Meeting Basic Product",
  "description": "Produk meeting basic",
  "status": "active"
}
```

Response 201:

```json
{
  "success": true,
  "message": "Product created successfully",
  "data": { ... }
}
```

#### PUT /api/v1/products/:id

Update product. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Product updated successfully",
  "data": { ... }
}
```

#### DELETE /api/v1/products/:id

Hapus product. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Product deleted successfully"
}
```

### Plans

#### GET /api/v1/plans

List plans milik business user login (via product -> service relationship). (Protected)

Query params: `?page=1&limit=10&search=keyword`

Response 200:

```json
{
  "success": true,
  "message": "Plans retrieved successfully",
  "data": [
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
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1
  }
}
```

#### GET /api/v1/plans/:id

Detail plan. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Plan retrieved successfully",
  "data": { ... }
}
```

#### POST /api/v1/plans

Buat plan baru. (Protected) Product_id harus milik business user.

Request:

```json
{
  "product_id": 1,
  "name": "Basic Monthly",
  "price": 100000,
  "billing_cycle": "monthly",
  "trial_days": 14
}
```

Valid billing_cycle: `monthly` atau `yearly`.

Response 201:

```json
{
  "success": true,
  "message": "Plan created successfully",
  "data": { ... }
}
```

#### PUT /api/v1/plans/:id

Update plan. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Plan updated successfully",
  "data": { ... }
}
```

#### DELETE /api/v1/plans/:id

Hapus plan. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Plan deleted successfully"
}
```

### Add-ons

#### GET /api/v1/add-ons

List add-ons milik business user login (via product -> service relationship). (Protected)

Query params: `?page=1&limit=10&search=keyword`

Response 200:

```json
{
  "success": true,
  "message": "Add-ons retrieved successfully",
  "data": [
    {
      "id": 1,
      "product_id": 1,
      "name": "Extra Storage",
      "price": 25000,
      "billing_cycle": "monthly",
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

#### GET /api/v1/add-ons/:id

Detail add-on. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Add-on retrieved successfully",
  "data": { ... }
}
```

#### POST /api/v1/add-ons

Buat add-on baru. (Protected) Product_id harus milik business user.

Request:

```json
{
  "product_id": 1,
  "name": "Extra Storage",
  "price": 25000,
  "billing_cycle": "monthly"
}
```

Valid billing_cycle: `monthly` atau `yearly`.

Response 201:

```json
{
  "success": true,
  "message": "Add-on created successfully",
  "data": { ... }
}
```

#### PUT /api/v1/add-ons/:id

Update add-on. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Add-on updated successfully",
  "data": { ... }
}
```

#### DELETE /api/v1/add-ons/:id

Hapus add-on. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Add-on deleted successfully"
}
```

### Coupons

#### GET /api/v1/coupons

List semua coupons (global, tidak di-filter per business). (Protected)

Query params: `?page=1&limit=10&search=keyword`

Response 200:

```json
{
  "success": true,
  "message": "Coupons retrieved successfully",
  "data": [
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
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1
  }
}
```

#### GET /api/v1/coupons/:id

Detail coupon. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Coupon retrieved successfully",
  "data": { ... }
}
```

#### POST /api/v1/coupons

Buat coupon baru. (Protected)

Request (percentage):

```json
{
  "code": "DISC10",
  "discount_type": "percentage",
  "discount_value": 10,
  "max_usage": 100,
  "expires_at": "2026-12-31T23:59:59Z"
}
```

Request (fixed):

```json
{
  "code": "HEMAT50000",
  "discount_type": "fixed",
  "discount_value": 50000,
  "max_usage": 50,
  "expires_at": "2026-12-31T23:59:59Z"
}
```

Valid discount_type: `percentage` atau `fixed`.
Jika percentage, discount_value maksimal 100.

Response 201:

```json
{
  "success": true,
  "message": "Coupon created successfully",
  "data": { ... }
}
```

Response 409 (duplicate code):

```json
{
  "success": false,
  "message": "coupon code already exists"
}
```

#### PUT /api/v1/coupons/:id

Update coupon. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Coupon updated successfully",
  "data": { ... }
}
```

#### DELETE /api/v1/coupons/:id

Hapus coupon. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Coupon deleted successfully"
}
```

### Customers

#### GET /api/v1/customers

List customers milik business user login. (Protected)

Query params: `?page=1&limit=10&search=budi`

Response 200:

```json
{
  "success": true,
  "message": "Customers retrieved successfully",
  "data": [
    {
      "id": 1,
      "business_id": "biz_xxxxx",
      "name": "Budi Santoso",
      "email": "budi@example.com",
      "contact": "08123456789",
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

#### GET /api/v1/customers/:id

Detail customer. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Customer retrieved successfully",
  "data": { ... }
}
```

#### POST /api/v1/customers

Buat customer baru. (Protected) Business_id otomatis dari JWT.

Request:

```json
{
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "contact": "08123456789"
}
```

Response 201:

```json
{
  "success": true,
  "message": "Customer created successfully",
  "data": { ... }
}
```

#### PUT /api/v1/customers/:id

Update customer. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Customer updated successfully",
  "data": { ... }
}
```

#### DELETE /api/v1/customers/:id

Hapus customer. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Customer deleted successfully"
}
```

### Subscriptions

#### POST /api/v1/subscriptions/preview

Hitung estimasi biaya subscription sebelum dibuat. (Protected)

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

Response 200:

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

#### POST /api/v1/subscriptions

Buat subscription baru. (Protected)

Customer_id dan plan_id wajib milik business user login.

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

Response 201:

```json
{
  "success": true,
  "message": "Subscription created successfully",
  "data": {
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
    "customer": { "id": 1, "name": "Budi Santoso", "email": "budi@example.com" },
    "plan": { "id": 1, "name": "Basic Monthly", "price": 100000, "billing_cycle": "monthly", "trial_days": 14 },
    "product": { "id": 1, "name": "Meeting Basic Product" },
    "service": { "id": 1, "name": "Meeting App" },
    "add_ons": [
      { "id": 1, "add_on_id": 1, "name": "Extra Storage", "price": 25000, "quantity": 2 }
    ],
    "coupons": [
      { "id": 1, "coupon_id": 1, "code": "DISC10", "discount_type": "percentage", "discount_value": 10 }
    ]
  }
}
```

#### GET /api/v1/subscriptions

List subscriptions milik business user login. (Protected)

Query params: `?page=1&limit=10&status=active&search=budi`

Filter status: `trial`, `active`, `paused`, `cancelled`.

Response 200:

```json
{
  "success": true,
  "message": "Subscriptions retrieved successfully",
  "data": [ ... ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1
  }
}
```

#### GET /api/v1/subscriptions/:id

Detail subscription dengan relasi customer, plan, product, service, add-ons, coupons. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Subscription retrieved successfully",
  "data": {
    "id": 1,
    "customer_id": 1,
    "plan_id": 1,
    "status": "trial",
    "start_date": "...",
    "next_billing_date": "...",
    "end_date": null,
    "trial_ends_at": "...",
    "meta": {},
    "created_at": "...",
    "customer": {},
    "plan": {},
    "product": {},
    "service": {},
    "add_ons": [],
    "coupons": []
  }
}
```

#### POST /api/v1/subscriptions/:id/cancel

Batalkan subscription. (Protected)

Hanya subscription dengan status `trial`, `active`, atau `paused` yang boleh dibatalkan.

Request:

```json
{
  "reason": "Customer requested cancellation"
}
```

Response 200:

```json
{
  "success": true,
  "message": "Subscription cancelled successfully",
  "data": { ... }
}
```

#### POST /api/v1/subscriptions/:id/pause

Jeda subscription. (Protected) Hanya subscription `active` yang boleh di-pause.

Request:

```json
{
  "reason": "Temporary pause"
}
```

Response 200:

```json
{
  "success": true,
  "message": "Subscription paused successfully",
  "data": { ... }
}
```

#### POST /api/v1/subscriptions/:id/resume

Lanjutkan subscription yang di-pause. (Protected) Hanya subscription `paused` yang boleh di-resume.

Response 200:

```json
{
  "success": true,
  "message": "Subscription resumed successfully",
  "data": { ... }
}
```

#### POST /api/v1/subscriptions/:id/add-ons

Tambah add-on ke subscription. (Protected)

Request:

```json
{
  "add_on_id": 1,
  "quantity": 3
}
```

Response 200:

```json
{
  "success": true,
  "message": "Add-on added to subscription successfully",
  "data": { ... }
}
```

#### DELETE /api/v1/subscriptions/:id/add-ons/:add_on_id

Hapus add-on dari subscription. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Add-on removed from subscription successfully",
  "data": { ... }
}
```

#### POST /api/v1/subscriptions/:id/coupons

Terapkan kupon ke subscription. (Protected)

Request:

```json
{
  "coupon_code": "DISC10"
}
```

Response 200:

```json
{
  "success": true,
  "message": "Coupon applied to subscription successfully",
  "data": { ... }
}
```

Response 409:

```json
{
  "success": false,
  "message": "coupon already applied to this subscription"
}
```

#### DELETE /api/v1/subscriptions/:id/coupons/:coupon_id

Hapus kupon dari subscription. (Protected)

Response 200:

```json
{
  "success": true,
  "message": "Coupon removed from subscription successfully",
  "data": { ... }
}
## Status Codes

| Code | Description |
|---|---|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - input tidak valid |
| 401 | Unauthorized - token salah / tidak ada |
| 403 | Forbidden - tidak punya akses |
| 404 | Not Found - data tidak ditemukan |
| 409 | Conflict - email/coupon sudah ada |
| 500 | Internal Server Error |

## Postman Testing

### Collection

File: `postman/PAY-LANGGAN.postman_collection.json`

### Environment

File: `postman/PAY-LANGGAN.local.postman_environment.json`

Variable:
- `base_url` = `http://localhost:8080`
- `token` - terisi otomatis setelah Signup/Login
- `business_id`, `user_id` - terisi otomatis setelah Signup/Login
- `service_id`, `product_id`, `plan_id`, `addon_id`, `coupon_id`, `customer_id` - terisi otomatis dari response create

### Cara Import ke Postman

1. Buka Postman
2. File -> Import -> Pilih file `postman/PAY-LANGGAN.postman_collection.json`
3. File -> Import -> Pilih file `postman/PAY-LANGGAN.local.postman_environment.json`
4. Pilih environment "PAY-LANGGAN Local" di pojok kanan atas

### Cara Run dengan Newman

Install Newman:

```bash
npm install -g newman
```

Run:

```bash
newman run postman/PAY-LANGGAN.postman_collection.json \
  -e postman/PAY-LANGGAN.local.postman_environment.json
```
