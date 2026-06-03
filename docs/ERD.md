# Entity Relationship Diagram

## Relasi Antar Tabel

```
businesses 1-m users
businesses 1-m services
businesses 1-m customers
businesses 1-m referral_urls

services 1-m products

products 1-m plans
products 1-m add_ons

customers 1-m subscriptions
customers 1-m payment_methods

subscriptions 1-m subscription_add_ons
subscriptions 1-m invoices
subscriptions 1-m billing_queue

subscription_add_ons m-1 add_ons

invoices 1-m payments
invoices 1-m invoice_items

subscriptions m-m coupons (via subscription_coupons)

payments 1-m payment_attempts

webhook_events (independent, mencatat callback dari payment gateway)

businesses 1-m audit_logs
users 1-m audit_logs
```

## Detail Relasi

### businesses -> users
- Satu business memiliki banyak user (admin/staff)
- User tidak bisa ada tanpa business

### businesses -> services
- Satu business memiliki banyak services (produk SaaS)
- Service terisolasi per business

### businesses -> audit_logs
- Satu business memiliki banyak audit logs

### services -> products
- Satu service bisa memiliki banyak products (varian pricing)

### products -> plans
- Satu product bisa memiliki banyak plans (monthly/yearly)

### products -> add_ons
- Satu product bisa memiliki banyak add-ons

### customers -> subscriptions
- Satu customer bisa memiliki banyak subscriptions
- Subscription mengacu ke plan (bukan ke product langsung)

### subscriptions -> invoices
- Subscription periodik menghasilkan invoice

### subscriptions -> billing_queue
- Queue untuk recurring billing job

### subscriptions -> coupons (via subscription_coupons)
- Banyak subscription bisa memakai banyak coupons
- Dicatat di tabel penghubung subscription_coupons

### invoices -> payments
- Satu invoice bisa memiliki banyak payments (percobaan pembayaran ulang)

### invoices -> invoice_items
- Satu invoice memiliki banyak item (plan, add-on, discount, adjustment)
- Setiap item adalah baris rincian tagihan

### payments -> payment_attempts
- Satu payment bisa memiliki banyak attempt (retry)
- Mencatat setiap percobaan charge ke payment gateway

### webhook_events
- Mencatat callback dari payment gateway (Midtrans, Xendit, DOKU)
- Tidak memiliki relasi FK langsung, reference_id menghubungkan ke transaksi asli
