# Drop-in script PayLanggan

Drop-in script menampilkan checkout PayLanggan sebagai popup di website bisnis. Form kartu di dalam checkout tetap menggunakan Xendit Components. Pembayaran sukses hanya dinyatakan setelah webhook Xendit mengubah status invoice menjadi `paid` di backend PayLanggan.

## Yang sudah tersedia

- `Frontend/public/paylanggan.js`: skrip popup tanpa dependensi framework.
- `POST /api/v1/integrations/orders`: membuat pelanggan, subscription, invoice pertama, dan tautan checkout untuk paket milik bisnis.
- `GET /api/v1/integrations/orders/:reference`: membaca status invoice untuk pesanan bisnis.
- `POST /api/v1/integrations/keys`, `GET /api/v1/integrations/keys`, `DELETE /api/v1/integrations/keys/:id`: membuat, melihat daftar, dan mencabut kunci integrasi. Endpoint ini memerlukan JWT admin bisnis. Secret key hanya muncul sekali saat dibuat.
- `examples/dropin-demo`: contoh website bisnis dengan server sederhana. Secret key tidak dikirim ke browser.

## Persiapan

1. Jalankan migrasi terbaru, termasuk `2026-09-29-31_merchant_dropin.sql`.
2. Jalankan backend dan frontend. `APP_PUBLIC_URL` di backend harus mengarah ke origin frontend. Untuk Xendit sungguhan, `XENDIT_RETURN_BASE_URL` harus HTTPS dan mengarah ke frontend yang bisa diakses Xendit.
3. Pastikan langganan bisnis ke PayLanggan berstatus `active` dan masa bayarnya belum habis. Buat paket di katalog bisnis. Versi ini menerima paket aktif dengan siklus `monthly` atau `yearly`, tanpa masa trial, harga kartu Rp5.000–Rp200.000.000 dalam rupiah bulat. Harga ditentukan dari database PayLanggan.
4. Login sebagai admin bisnis, lalu buat kunci integrasi:

```http
POST /api/v1/integrations/keys
Authorization: Bearer <JWT_ADMIN_BISNIS>
Content-Type: application/json

{"name":"Website toko"}
```

Simpan `data.secret_key` hanya di server bisnis (misalnya environment variable `PAYLANGGAN_SECRET_KEY`). Jangan taruh di HTML, JavaScript, Git, atau Postman environment yang dibagikan. Kunci yang dicabut tidak dapat dipakai lagi.

## Alur pembelian

Server bisnis mengirim request berikut ke backend PayLanggan:

```http
POST /api/v1/integrations/orders
Authorization: Bearer pl_sk_<SECRET>
Content-Type: application/json

{
  "external_order_id": "order-12345",
  "plan_id": 10,
  "customer": {
    "name": "Budi",
    "email": "budi@example.com",
    "mobile": "+6281234567890"
  }
}
```

Respons memuat `checkout_url`, `invoice_id`, `invoice_status`, dan `external_order_id`. `external_order_id` harus unik per bisnis. Mengirim ulang referensi yang sama tidak membuat invoice baru; untuk checkout yang belum dimulai, PayLanggan membuat tautan pengganti. Jika checkout sudah aktif, API bisa memberi `409` sampai pembayaran selesai atau sesi berakhir. Server bisnis dapat membaca status pesanan dengan `GET /api/v1/integrations/orders/order-12345` menggunakan secret key yang sama.

Di website bisnis, pasang skrip dari **origin frontend PayLanggan yang sama dengan `checkout_url`**:

```html
<script src="https://checkout.paylanggan.example/paylanggan.js"></script>
<script>
  // checkoutUrl berasal dari server bisnis setelah POST /integrations/orders.
  PayLanggan.open({
    checkoutUrl,
    onPaid() {
      // Tampilkan hasil. Server bisnis tetap harus memeriksa status order ke API.
    },
    onClose() {
      // Pelanggan menutup popup; jangan langsung tandai order gagal.
    }
  })
</script>
```

`onPaid` adalah sinyal untuk memperbarui tampilan. **Jangan aktifkan akses pelanggan hanya karena callback browser terpanggil.** Server bisnis perlu memeriksa `invoice_status === "paid"` melalui API. Untuk operasi produksi, buat mekanisme notifikasi server ke server atau polling terjadwal agar status tetap tersinkron walau pelanggan menutup browser.

## Menjalankan website contoh

Di PowerShell, dari folder proyek:

```powershell
$env:PAYLANGGAN_API_URL = 'http://localhost:8080'
$env:PAYLANGGAN_CHECKOUT_ORIGIN = 'http://localhost:5173'
$env:PAYLANGGAN_SECRET_KEY = '<kunci-dari-endpoint-admin>'
$env:PAYLANGGAN_PLAN_ID = '10'
node .\examples\dropin-demo\server.mjs
```

Buka `http://localhost:4174`, isi kontak, lalu klik **Beli langganan**. Untuk Test Mode Xendit sungguhan, gunakan URL HTTPS untuk frontend dan sesuaikan `APP_PUBLIC_URL`, `XENDIT_RETURN_BASE_URL`, serta `PAYLANGGAN_CHECKOUT_ORIGIN` dengan URL frontend tersebut. Skrip dan checkout harus berasal dari origin frontend yang sama.

## Batas versi awal

- Checkout popup memakai iframe. Alur autentikasi kartu 3DS harus diuji pada browser dan metode kartu yang dipakai, karena penyedia autentikasi dapat membatasi tampilan dalam iframe. Jika iframe ditolak, pelanggan dapat membuka `checkout_url` PayLanggan sebagai halaman penuh.
- Pemrosesan pesanan mendukung paket dasar tanpa trial, kupon, atau add-on. Ini tidak mengubah alur invoice dan autopay yang sudah ada.
- API status order tersedia; pengiriman webhook khusus dari PayLanggan ke server bisnis belum dibuat.
- Drop-in script tidak mengatur tujuan dana. Sebelum transaksi nyata untuk banyak bisnis, routing pembayaran ke sub-account bisnis melalui xenPlatform atau rancangan settlement yang disetujui Xendit perlu diselesaikan.
