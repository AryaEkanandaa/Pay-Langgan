# Architecture

Dokumentasi ini menjelaskan arsitektur clean architecture sederhana yang digunakan di PAY-LANGGAN.COM.

## Layer

### 1. Handler (HTTP Layer)

Berada di `internal/handlers/`. Bertanggung jawab:
- Menerima request HTTP
- Validasi basic request
- Memanggil service
- Mengembalikan response JSON standar

### 2. Service (Business Logic)

Berada di `internal/services/`. Bertanggung jawab:
- Berisi business logic aplikasi
- Orchestrasi multiple repository jika diperlukan
- Contoh: signup flow, billing logic, payment logic

### 3. Repository (Data Access)

Berada di `internal/repositories/`. Bertanggung jawab:
- Query ke database menggunakan sqlx
- Tidak berisi business logic berat
- Tenant-aware dengan business_id

### 4. Model

Berada di `internal/models/`. Berisi:
- Struct mapping ke tabel database (tag: db, json)
- Request/Response DTO
- Reusable type definitions

## Alur Request

```
Client -> HTTP Request
  -> Route (internal/routes/routes.go)
    -> Middleware (JWT, Logging, Error)
      -> Handler (validasi input)
        -> Service (business logic)
          -> Repository (database query)
            -> Database (PostgreSQL)
          <- Repository
        <- Service
      <- Handler (response JSON)
    <- Middleware
  <- Route
<- Client
```

## Alasan Pemisahan Layer

1. **Separation of Concerns** - Setiap layer punya tanggung jawab spesifik
2. **Testability** - Service bisa di-test tanpa HTTP layer
3. **Maintainability** - Perubahan database hanya di repository
4. **Scalability** - Tim bisa kerja paralel di layer berbeda
