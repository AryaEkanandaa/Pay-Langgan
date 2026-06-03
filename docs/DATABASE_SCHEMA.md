# Database Schema

Dokumen ini menjelaskan struktur database PayLanggan dalam format yang mirip query `CREATE TABLE`.
Komentar `-- ...` di dalam query dipakai untuk menjelaskan fungsi tabel, fungsi kolom, relasi, dan constraint.

## Gambaran Relasi

```text
businesses
  -> users
  -> services
    -> products
      -> plans
      -> add_ons
  -> customers
    -> subscriptions
      -> subscription_add_ons
      -> subscription_coupons
      -> invoices
        -> invoice_items
        -> payments
          -> payment_attempts
  -> referral_urls
  -> audit_logs

subscriptions -> billing_queue
payment provider -> webhook_events
```

## businesses

```sql
-- Menyimpan data bisnis atau tenant utama.
-- Hampir semua data operasional akan terhubung ke tabel ini.
CREATE TABLE businesses (
    id VARCHAR(50) PRIMARY KEY, -- ID unik bisnis. Dibuat string agar bisa memakai kode/slug/custom ID.
    name VARCHAR(100) NOT NULL, -- Nama bisnis.
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- Status bisnis, default aktif.
    deleted BOOLEAN NOT NULL DEFAULT FALSE, -- Soft delete. TRUE berarti bisnis dianggap dihapus tanpa menghapus row.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu bisnis dibuat.
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu bisnis terakhir diperbarui.
    meta JSONB NOT NULL DEFAULT '{}'::jsonb -- Data tambahan bisnis seperti setting, branding, alamat, atau konfigurasi lain.
);
```

## users

```sql
-- Menyimpan user internal milik bisnis, misalnya owner, admin, atau staff.
CREATE TABLE users (
    id SERIAL PRIMARY KEY, -- ID unik user.
    business_id VARCHAR(50) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE, -- Bisnis pemilik user.
    name VARCHAR(100) NOT NULL, -- Nama user.
    email VARCHAR(100) NOT NULL UNIQUE, -- Email login user. Harus unik.
    password VARCHAR(255) NOT NULL, -- Password user. Seharusnya berisi hash, bukan password asli.
    role VARCHAR(50) NOT NULL DEFAULT 'admin', -- Role user, default admin.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu user dibuat.
    updated_at TIMESTAMP NOT NULL DEFAULT NOW() -- Waktu user terakhir diperbarui.
);
```

## services

```sql
-- Menyimpan layanan besar milik bisnis.
-- Contoh: Hosting, SaaS Tools, Membership, Course.
CREATE TABLE services (
    id SERIAL PRIMARY KEY, -- ID unik service.
    business_id VARCHAR(50) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE, -- Bisnis pemilik service.
    name VARCHAR(100) NOT NULL, -- Nama service.
    description TEXT, -- Deskripsi service.
    meta JSONB NOT NULL DEFAULT '{}'::jsonb, -- Data tambahan service.
    created_at TIMESTAMP NOT NULL DEFAULT NOW() -- Waktu service dibuat.
);
```

## products

```sql
-- Menyimpan produk di dalam service.
-- Satu service bisa punya banyak product.
CREATE TABLE products (
    id SERIAL PRIMARY KEY, -- ID unik product.
    service_id INT NOT NULL REFERENCES services(id) ON DELETE CASCADE, -- Service induk dari product.
    name VARCHAR(100) NOT NULL, -- Nama product.
    description TEXT, -- Deskripsi product.
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- Status product.
    meta JSONB NOT NULL DEFAULT '{}'::jsonb, -- Data tambahan product.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu product dibuat.
    CONSTRAINT chk_products_status CHECK (status IN ('active', 'inactive', 'archived')) -- Membatasi status product.
);
```

## plans

```sql
-- Menyimpan paket harga dari sebuah product.
-- Subscription customer akan memilih salah satu plan.
CREATE TABLE plans (
    id SERIAL PRIMARY KEY, -- ID unik plan.
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE, -- Product pemilik plan.
    name VARCHAR(100) NOT NULL, -- Nama plan, misalnya Basic Monthly atau Pro Yearly.
    price NUMERIC(12,2) NOT NULL, -- Harga plan.
    billing_cycle VARCHAR(20) NOT NULL, -- Siklus tagihan.
    trial_days INT NOT NULL DEFAULT 0, -- Jumlah hari trial. 0 berarti tanpa trial.
    meta JSONB NOT NULL DEFAULT '{}'::jsonb, -- Data tambahan plan seperti limit, fitur, atau konfigurasi.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu plan dibuat.
    CONSTRAINT chk_plans_billing_cycle CHECK (billing_cycle IN ('monthly', 'yearly')) -- Siklus yang valid.
);
```

## add_ons

```sql
-- Menyimpan fitur tambahan yang bisa ditambahkan ke subscription.
CREATE TABLE add_ons (
    id SERIAL PRIMARY KEY, -- ID unik add-on.
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE, -- Product pemilik add-on.
    name VARCHAR(100) NOT NULL, -- Nama add-on.
    price NUMERIC(12,2) NOT NULL, -- Harga add-on.
    billing_cycle VARCHAR(20) NOT NULL, -- Siklus tagihan add-on.
    meta JSONB NOT NULL DEFAULT '{}'::jsonb, -- Data tambahan add-on.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu add-on dibuat.
    CONSTRAINT chk_add_ons_billing_cycle CHECK (billing_cycle IN ('monthly', 'yearly')) -- Siklus yang valid.
);
```

## coupons

```sql
-- Menyimpan kode diskon.
CREATE TABLE coupons (
    id SERIAL PRIMARY KEY, -- ID unik coupon.
    code VARCHAR(50) NOT NULL UNIQUE, -- Kode promo. Harus unik.
    discount_type VARCHAR(20) NOT NULL, -- Jenis diskon.
    discount_value NUMERIC(12,2) NOT NULL, -- Nilai diskon. Bisa persen atau nominal, tergantung discount_type.
    max_usage INT, -- Batas maksimal pemakaian coupon. NULL berarti tidak dibatasi oleh database.
    used_count INT NOT NULL DEFAULT 0, -- Jumlah pemakaian coupon saat ini.
    expires_at TIMESTAMP, -- Waktu coupon kedaluwarsa. NULL berarti tidak punya expiry di database.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu coupon dibuat.
    CONSTRAINT chk_coupons_discount_type CHECK (discount_type IN ('percentage', 'fixed')) -- Jenis diskon valid.
);
```

## customers

```sql
-- Menyimpan customer milik bisnis.
CREATE TABLE customers (
    id SERIAL PRIMARY KEY, -- ID unik customer.
    business_id VARCHAR(50) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE, -- Bisnis pemilik customer.
    name VARCHAR(100) NOT NULL, -- Nama customer.
    email VARCHAR(100), -- Email customer.
    contact VARCHAR(20), -- Nomor kontak customer.
    meta JSONB NOT NULL DEFAULT '{}'::jsonb, -- Data tambahan customer seperti alamat atau tax info.
    created_at TIMESTAMP NOT NULL DEFAULT NOW() -- Waktu customer dibuat.
);
```

## subscriptions

```sql
-- Menyimpan langganan customer terhadap plan tertentu.
-- Tabel ini adalah pusat proses billing berulang.
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY, -- ID unik subscription.
    customer_id INT NOT NULL REFERENCES customers(id) ON DELETE CASCADE, -- Customer pemilik subscription.
    plan_id INT NOT NULL REFERENCES plans(id), -- Plan yang dipilih customer.
    status VARCHAR(20) NOT NULL, -- Status subscription.
    start_date TIMESTAMP NOT NULL DEFAULT NOW(), -- Tanggal mulai subscription.
    next_billing_date TIMESTAMP, -- Jadwal tagihan berikutnya.
    end_date TIMESTAMP, -- Tanggal subscription berakhir.
    trial_ends_at TIMESTAMP, -- Tanggal trial berakhir.
    meta JSONB NOT NULL DEFAULT '{}'::jsonb, -- Data tambahan subscription.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu subscription dibuat.
    CONSTRAINT chk_subscriptions_status CHECK (status IN ('trial', 'active', 'cancelled', 'paused')) -- Status valid.
);
```

## subscription_add_ons

```sql
-- Menghubungkan subscription dengan add-on.
-- Satu subscription bisa punya banyak add-on.
CREATE TABLE subscription_add_ons (
    id SERIAL PRIMARY KEY, -- ID unik record.
    subscription_id INT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE, -- Subscription yang memakai add-on.
    add_on_id INT NOT NULL REFERENCES add_ons(id), -- Add-on yang dipakai.
    quantity INT NOT NULL DEFAULT 1, -- Jumlah add-on yang dipakai.
    created_at TIMESTAMP NOT NULL DEFAULT NOW() -- Waktu add-on dipasang ke subscription.
);
```

## payment_methods

```sql
-- Menyimpan metode pembayaran customer.
CREATE TABLE payment_methods (
    id SERIAL PRIMARY KEY, -- ID unik payment method.
    customer_id INT NOT NULL REFERENCES customers(id) ON DELETE CASCADE, -- Customer pemilik payment method.
    provider VARCHAR(50) NOT NULL, -- Provider payment.
    token VARCHAR(255) NOT NULL, -- Token aman dari provider, bukan data kartu asli.
    is_default BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE jika menjadi metode pembayaran default customer.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu payment method dibuat.
    CONSTRAINT chk_payment_methods_provider CHECK (provider IN ('midtrans', 'xendit', 'doku')) -- Provider valid.
);
```

## invoices

```sql
-- Menyimpan invoice atau tagihan dari subscription.
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY, -- ID unik invoice.
    subscription_id INT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE, -- Subscription sumber invoice.
    invoice_number VARCHAR(100) NOT NULL UNIQUE, -- Nomor invoice yang tampil ke customer.
    amount NUMERIC(12,2) NOT NULL, -- Total tagihan invoice.
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- Status invoice.
    due_date TIMESTAMP, -- Batas waktu pembayaran.
    paid_at TIMESTAMP, -- Waktu invoice dibayar.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu invoice dibuat.
    CONSTRAINT chk_invoices_status CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')) -- Status valid.
);
```

## payments

```sql
-- Menyimpan transaksi pembayaran untuk invoice.
CREATE TABLE payments (
    id SERIAL PRIMARY KEY, -- ID unik payment.
    invoice_id INT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE, -- Invoice yang dibayar.
    provider VARCHAR(50) NOT NULL, -- Provider pembayaran.
    provider_transaction_id VARCHAR(255), -- ID transaksi dari provider.
    amount NUMERIC(12,2) NOT NULL, -- Nominal pembayaran.
    status VARCHAR(20) NOT NULL, -- Status payment.
    paid_at TIMESTAMP, -- Waktu pembayaran sukses.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu payment dibuat.
    CONSTRAINT chk_payments_status CHECK (status IN ('pending', 'success', 'failed', 'expired')) -- Status valid.
);
```

## billing_queue

```sql
-- Menyimpan antrean proses billing.
-- Dipakai worker/job scheduler untuk memproses tagihan subscription.
CREATE TABLE billing_queue (
    id SERIAL PRIMARY KEY, -- ID unik job billing.
    subscription_id INT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE, -- Subscription yang akan diproses.
    scheduled_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Jadwal proses billing.
    processed BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE jika job sudah diproses.
    retry_count INT NOT NULL DEFAULT 0, -- Jumlah percobaan ulang jika gagal.
    last_error TEXT, -- Error terakhir saat proses billing gagal.
    processed_at TIMESTAMP, -- Waktu job selesai diproses.
    created_at TIMESTAMP NOT NULL DEFAULT NOW() -- Waktu job dibuat.
);
```

## referral_urls

```sql
-- Menyimpan URL referral atau tracking link milik bisnis.
CREATE TABLE referral_urls (
    id SERIAL PRIMARY KEY, -- ID unik referral URL.
    business_id VARCHAR(50) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE, -- Bisnis pemilik referral URL.
    code VARCHAR(100) NOT NULL UNIQUE, -- Kode referral unik.
    target_url TEXT NOT NULL, -- URL tujuan ketika referral dibuka.
    created_at TIMESTAMP NOT NULL DEFAULT NOW() -- Waktu referral URL dibuat.
);
```

## invoice_items

```sql
-- Menyimpan rincian item pada invoice.
-- Total invoice bisa dihitung dari kumpulan item ini.
CREATE TABLE invoice_items (
    id SERIAL PRIMARY KEY, -- ID unik invoice item.
    invoice_id INT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE, -- Invoice pemilik item.
    item_type VARCHAR(50) NOT NULL, -- Jenis item invoice.
    description TEXT NOT NULL, -- Deskripsi item yang tampil di invoice.
    quantity INT NOT NULL DEFAULT 1, -- Jumlah item.
    unit_price NUMERIC(12,2) NOT NULL, -- Harga per unit.
    subtotal NUMERIC(12,2) NOT NULL, -- Total item. Biasanya quantity * unit_price.
    meta JSONB NOT NULL DEFAULT '{}'::jsonb, -- Data tambahan item, seperti referensi plan/add-on/coupon.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu item dibuat.
    CONSTRAINT chk_invoice_items_item_type CHECK (item_type IN ('plan', 'add_on', 'discount', 'manual_adjustment')) -- Jenis item valid.
);
```

## subscription_coupons

```sql
-- Menghubungkan subscription dengan coupon.
CREATE TABLE subscription_coupons (
    id SERIAL PRIMARY KEY, -- ID unik record.
    subscription_id INT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE, -- Subscription yang memakai coupon.
    coupon_id INT NOT NULL REFERENCES coupons(id) ON DELETE CASCADE, -- Coupon yang dipakai.
    applied_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu coupon mulai diterapkan.
    expired_at TIMESTAMP, -- Waktu coupon berhenti berlaku untuk subscription ini.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu record dibuat.
    UNIQUE(subscription_id, coupon_id) -- Mencegah coupon yang sama dipasang dua kali di subscription yang sama.
);
```

## payment_attempts

```sql
-- Menyimpan riwayat percobaan pembayaran.
-- Berguna untuk retry payment dan debugging pembayaran gagal.
CREATE TABLE payment_attempts (
    id SERIAL PRIMARY KEY, -- ID unik payment attempt.
    payment_id INT NOT NULL REFERENCES payments(id) ON DELETE CASCADE, -- Payment utama yang dicoba.
    invoice_id INT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE, -- Invoice terkait attempt.
    provider VARCHAR(50) NOT NULL, -- Provider yang dipakai saat attempt.
    attempt_number INT NOT NULL DEFAULT 1, -- Nomor urutan percobaan.
    status VARCHAR(20) NOT NULL, -- Status attempt.
    error_message TEXT, -- Pesan error jika attempt gagal.
    attempted_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu percobaan dilakukan.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu record dibuat.
    CONSTRAINT chk_payment_attempts_status CHECK (status IN ('pending', 'success', 'failed')) -- Status valid.
);
```

## webhook_events

```sql
-- Menyimpan webhook yang diterima dari provider pembayaran.
CREATE TABLE webhook_events (
    id SERIAL PRIMARY KEY, -- ID unik webhook event.
    provider VARCHAR(50) NOT NULL, -- Provider pengirim webhook.
    event_type VARCHAR(100) NOT NULL, -- Jenis event webhook.
    reference_id VARCHAR(255), -- ID referensi transaksi atau object dari provider.
    payload JSONB NOT NULL DEFAULT '{}'::jsonb, -- Payload webhook mentah atau hasil normalisasi.
    processed BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE jika webhook sudah diproses aplikasi.
    received_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu webhook diterima.
    processed_at TIMESTAMP, -- Waktu webhook selesai diproses.
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), -- Waktu record dibuat.
    CONSTRAINT chk_webhook_events_provider CHECK (provider IN ('midtrans', 'xendit', 'doku')) -- Provider valid.
);
```

## audit_logs

```sql
-- Menyimpan riwayat aksi dan perubahan data penting.
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY, -- ID unik audit log.
    business_id VARCHAR(50) NOT NULL REFERENCES businesses(id) ON DELETE CASCADE, -- Bisnis tempat aksi terjadi.
    user_id INT REFERENCES users(id) ON DELETE SET NULL, -- User pelaku aksi. Menjadi NULL jika user dihapus.
    action VARCHAR(100) NOT NULL, -- Nama aksi, misalnya create, update, delete, login, atau payment_retry.
    entity_type VARCHAR(100) NOT NULL, -- Jenis entity yang terdampak, misalnya subscription, invoice, customer.
    entity_id VARCHAR(100), -- ID entity yang terdampak.
    old_value JSONB NOT NULL DEFAULT '{}'::jsonb, -- Data sebelum perubahan.
    new_value JSONB NOT NULL DEFAULT '{}'::jsonb, -- Data setelah perubahan.
    created_at TIMESTAMP NOT NULL DEFAULT NOW() -- Waktu aksi dicatat.
);
```

## Indexes

```sql
-- Index user.
CREATE INDEX idx_users_email ON users(email); -- Mempercepat login atau lookup user berdasarkan email.
CREATE INDEX idx_users_business_id ON users(business_id); -- Mempercepat daftar user per bisnis.

-- Index service dan customer.
CREATE INDEX idx_services_business_id ON services(business_id); -- Mempercepat daftar service per bisnis.
CREATE INDEX idx_customers_business_id ON customers(business_id); -- Mempercepat daftar customer per bisnis.

-- Index subscription.
CREATE INDEX idx_subscriptions_customer_id ON subscriptions(customer_id); -- Mempercepat daftar subscription customer.
CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id); -- Mempercepat query subscription berdasarkan plan.
CREATE INDEX idx_subscriptions_status ON subscriptions(status); -- Mempercepat filter status subscription.
CREATE INDEX idx_subscriptions_next_billing_date ON subscriptions(next_billing_date); -- Mempercepat billing job.

-- Index invoice dan payment.
CREATE INDEX idx_invoices_subscription_id ON invoices(subscription_id); -- Mempercepat daftar invoice per subscription.
CREATE INDEX idx_invoices_status ON invoices(status); -- Mempercepat filter status invoice.
CREATE INDEX idx_payments_invoice_id ON payments(invoice_id); -- Mempercepat daftar payment per invoice.
CREATE INDEX idx_payments_status ON payments(status); -- Mempercepat filter status payment.

-- Index billing queue.
CREATE INDEX idx_billing_queue_processed ON billing_queue(processed); -- Mempercepat pencarian job belum diproses.
CREATE INDEX idx_billing_queue_scheduled_at ON billing_queue(scheduled_at); -- Mempercepat pencarian job berdasarkan jadwal.
CREATE INDEX idx_billing_queue_subscription_id ON billing_queue(subscription_id); -- Mempercepat pencarian queue per subscription.

-- Index coupon dan referral.
CREATE INDEX idx_coupons_code ON coupons(code); -- Mempercepat lookup coupon berdasarkan kode.
CREATE INDEX idx_referral_urls_code ON referral_urls(code); -- Mempercepat lookup referral berdasarkan kode.

-- Index invoice items.
CREATE INDEX idx_invoice_items_invoice_id ON invoice_items(invoice_id); -- Mempercepat ambil item invoice.
CREATE INDEX idx_invoice_items_item_type ON invoice_items(item_type); -- Mempercepat filter item berdasarkan tipe.

-- Index subscription coupons.
CREATE INDEX idx_subscription_coupons_subscription_id ON subscription_coupons(subscription_id); -- Mempercepat ambil coupon per subscription.
CREATE INDEX idx_subscription_coupons_coupon_id ON subscription_coupons(coupon_id); -- Mempercepat cari pemakaian coupon.

-- Index payment attempts.
CREATE INDEX idx_payment_attempts_payment_id ON payment_attempts(payment_id); -- Mempercepat ambil attempt per payment.
CREATE INDEX idx_payment_attempts_invoice_id ON payment_attempts(invoice_id); -- Mempercepat ambil attempt per invoice.
CREATE INDEX idx_payment_attempts_status ON payment_attempts(status); -- Mempercepat filter attempt berdasarkan status.

-- Index webhook events.
CREATE INDEX idx_webhook_events_provider ON webhook_events(provider); -- Mempercepat filter webhook per provider.
CREATE INDEX idx_webhook_events_event_type ON webhook_events(event_type); -- Mempercepat filter webhook per event.
CREATE INDEX idx_webhook_events_reference_id ON webhook_events(reference_id); -- Mempercepat pencocokan webhook dengan transaksi.
CREATE INDEX idx_webhook_events_processed ON webhook_events(processed); -- Mempercepat pencarian webhook belum diproses.

-- Index audit logs.
CREATE INDEX idx_audit_logs_business_id ON audit_logs(business_id); -- Mempercepat audit log per bisnis.
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id); -- Mempercepat audit log per user.
CREATE INDEX idx_audit_logs_entity_type ON audit_logs(entity_type); -- Mempercepat audit log per jenis entity.
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at); -- Mempercepat filter/sort audit berdasarkan waktu.
```

## Catatan

```text
1. Kolom meta JSONB dipakai sebagai tempat data tambahan yang fleksibel.
2. Kolom updated_at belum otomatis berubah saat update, jadi perlu logic aplikasi atau trigger.
3. Kolom invoice_items.subtotal perlu dijaga aplikasi agar sesuai dengan quantity * unit_price.
4. coupons.used_count perlu di-update dengan hati-hati agar tidak bentrok saat coupon dipakai bersamaan.
5. Beberapa kolom provider sudah punya CHECK, tetapi payments.provider belum dibatasi oleh CHECK.
```
