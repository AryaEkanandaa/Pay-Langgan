# Koleksi Postman backend PAY-LANGGAN

File:

- `PAY-LANGGAN.postman_collection.json`: semua endpoint HTTP yang didaftarkan di `internal/routes` serta `/health`.
- `PAY-LANGGAN.local.postman_environment.json`: contoh variabel lokal tanpa kredensial asli.
- `generate.py`: generator kedua JSON tersebut. Jalankan lagi bila rute atau contoh payload berubah.

## Persiapan

1. Jalankan PostgreSQL, migrasi, dan API sesuai README proyek. Default API: `http://localhost:8080`.
2. Import kedua file JSON ke Postman dan pilih environment **PAY-LANGGAN Local**.
3. Isi `login_email` dan `login_password` untuk akun tenant yang sudah ada, lalu jalankan **Login**. Respons sukses mengisi `token` otomatis.
4. Untuk tenant baru, isi `signup_email` unik dan `signup_password` (minimal 6 karakter), set `run_signup=true`, lalu jalankan **Signup bisnis dan admin**. Setelah itu set `run_signup=false`. Signup juga mengisi token.
5. Gunakan folder sesuai urutan nomor. Untuk membuat data uji, set `run_mutations=true`. Request create menyimpan ID ke environment: `service_id`, `product_id`, `plan_id`, `add_on_id`, `coupon_id`, `customer_id`, `subscription_id`. Jika data sudah ada, isi ID manual dari respons daftar.
6. Invoice tidak mempunyai endpoint create. Isi `invoice_id` dari **Daftar invoice** setelah invoice dibuat oleh alur subscription atau worker billing. Jalankan `go run ./cmd/worker` dari `Backend` bila perlu menghasilkan invoice berkala.

Semua respons API umumnya memakai `{ "success": true, "message": "...", "data": ... }`. Request daftar juga mengembalikan `meta` untuk pagination. Setiap request mempunyai pemeriksaan status HTTP dan `success`; request create/login menyimpan variabel bila berhasil. Folder dan request berisi penjelasan tujuan, role, dan prasyarat. Request dengan ID yang belum diisi dilewati oleh Collection Runner.

Request daftar mengirim `page` dan `limit` secara default. Untuk filter opsional, tambahkan sendiri `search` atau `status` di tab **Params** saat diperlukan. Nilai filter kosong sengaja tidak disertakan agar Postman tidak mengirim variabel kosong sebagai teks `{{nama_variabel}}`.

## Flag keamanan data uji

Semua flag default `false`:

| Variabel | Mengizinkan |
| --- | --- |
| `run_signup` | Membuat tenant dan admin baru |
| `run_mutations` | Membuat dan mengubah bisnis, staf, katalog, pelanggan, dan subscription |
| `run_checkout` | Membuat checkout link dan memulai sesi pembayaran |
| `run_webhook` | Mengirim payload callback uji ke endpoint Xendit |
| `run_autopay` | Mengaktifkan autopay, hanya setelah ada persetujuan pelanggan yang sah |
| `run_platform_checkout` | Membuat checkout tagihan bisnis PayLanggan sesudah admin menyetujui debit Rp300.000/bulan |
| `run_irreversible` | Hapus data, ubah status subscription, tandai invoice lunas, atau nonaktifkan autopay |

Aktifkan hanya flag yang diperlukan, lalu kembalikan ke `false`. Menjalankan seluruh koleksi dengan `run_irreversible=true` dapat menghapus data yang baru dibuat serta mengubah status invoice/subscription. Gunakan database pengujian.

**Jika panel respons Postman kosong setelah menekan Send**, periksa flag request pada environment yang sedang terpilih di pojok kanan atas. Request yang dijaga flag memakai `pm.execution.skipRequest()` sehingga tidak mengirim HTTP sama sekali ketika flag masih `false`. Misalnya, untuk **Buat Layanan**, ubah `run_mutations` menjadi `true` di environment aktif dan kirim ulang. Jika ada beberapa salinan environment (misalnya `PAY-LANGGAN Local` dan `PAY-LANGGAN Local 2.0`), token login serta flag harus ada pada salinan yang sedang aktif.

## Checkout dan pembayaran

**Buat checkout link** membutuhkan invoice pending dan role Admin/Finance. Respons menyimpan `checkout_token` dan `checkout_url`. **Lihat checkout publik** dan **Mulai sesi kartu** tidak memakai JWT. Isi `customer_mobile` dengan nomor E.164, misalnya `+6281234567890`. `XENDIT_MOCK=true` hanya mensimulasikan sesi; `components_sdk_key` yang dihasilkan tidak berlaku untuk SDK Xendit sesungguhnya.

### Uji pembayaran kartu di Xendit Test Mode

1. Isi `XENDIT_SECRET_KEY` dan `XENDIT_CALLBACK_TOKEN` Test Mode di `Backend/.env`, set `XENDIT_MOCK=false`, dan biarkan `AUTOPAY_ENABLED=false`. Jangan masukkan secret key ke environment Postman.
2. Siapkan **dua URL HTTPS publik** jika menguji lewat tunnel: URL frontend untuk `APP_PUBLIC_URL` dan `XENDIT_RETURN_BASE_URL`, serta URL backend untuk API dan webhook. Kedua variabel frontend itu cukup berisi alamat dasar frontend; backend menambahkan path `/checkout/:token` secara otomatis.
3. Buat `Frontend/.env.local` dari `Frontend/.env.example`, lalu isi `VITE_API_URL` dengan URL HTTPS backend dan `VITE_ALLOWED_HOST` dengan hostname tunnel frontend tanpa `https://`. Restart server Vite setelah mengubahnya. Saat browser dan backend berjalan di mesin yang sama, `http://localhost:8080` dapat dipakai untuk API lokal, tetapi gunakan URL HTTPS backend untuk uji melalui frontend HTTPS publik.
4. Di dashboard Xendit Test Mode, daftarkan webhook **Payment Session Completed** dan **Payment Session Expired** ke `https://URL-BACKEND/api/v1/webhooks/xendit/payment-sessions`. URL webhook harus menuju backend, sedangkan return URL menuju frontend.
5. Jalankan migrasi (termasuk migrasi checkout, recovery, dan controls), API (`go run ./cmd/api`), frontend (`npm run dev`), dan worker pembayaran (`go run ./cmd/payments`) pada proses terpisah. Worker membantu rekonsiliasi bila webhook terlambat atau perlu retry.
6. Login sebagai Admin atau Finance. Dari invoice berstatus `pending`, jalankan **Buat checkout link** di Postman (`invoice_id` terisi, `run_checkout=true`), lalu buka `checkout_url` di browser. Isi kontak E.164, lanjutkan ke formulir kartu Xendit, dan gunakan kartu uji yang tersedia pada akun Xendit Test Mode Anda. **Jangan** jalankan **Mulai sesi kartu** di Postman untuk link yang sama sebelum membukanya di browser: halaman akan mengambil sesi aktif, tetapi untuk alur UI paling sederhana biarkan halaman yang memulai sesi.
7. Setelah pembayaran, halaman menunggu status invoice dari backend. Periksa **Detail invoice** dan **Riwayat pembayaran** di Postman. Invoice `paid` adalah hasil yang dicari; jika tetap `pending`, periksa pengiriman webhook di dashboard Xendit dan log worker.

Webhook membutuhkan `callback_token` yang sesuai konfigurasi backend dan `webhook_payload` yang benar-benar sesuai sesi/provider; jangan memakai contoh payload acak untuk mengubah status invoice. Endpoint webhook ini hanya menerima event Payment Session, bukan Payments API v3. `run_webhook` sengaja mati secara default. Autopay memerlukan metode pembayaran pelanggan yang tersimpan, `payment_method_id`, referensi bukti persetujuan, tanggal ISO 8601 yang sah, dan `AUTOPAY_ENABLED=true` pada worker. Pengiriman request **Aktifkan autopay** saja tidak menggantikan persetujuan pelanggan.

## Tagihan bisnis ke PayLanggan

Fitur ini terpisah dari invoice dan autopay **pelanggan bisnis**. Hanya admin bisnis yang dapat mengakses folder **10 - Tagihan bisnis ke PayLanggan**. Satu paket tetap **PayLanggan Bulanan, Rp300.000/bulan**.

1. Jalankan migrasi terbaru (`goose ... up` atau `make migrate-up`), API, frontend, dan `go run ./cmd/platform-worker` di terminal terpisah. Worker ini memproses perpanjangan tagihan bisnis; worker `./cmd/payments` tetap menangani pembayaran pelanggan.
2. Di Xendit Test Mode, aktifkan izin API untuk Payment Sessions dan Payments API v3 serta kemampuan penyimpanan kartu/debit berulang. Biarkan **Payment Session Completed/Expired** menuju URL webhook yang sudah dipakai: `https://URL-BACKEND/api/v1/webhooks/xendit/payment-sessions`. Backend mengarahkan event bisnis berdasarkan referensi `platform_`. Gunakan callback token yang sama dengan `XENDIT_CALLBACK_TOKEN`.
3. Set `XENDIT_MOCK=false`, `XENDIT_SECRET_KEY` Test Mode, dan `XENDIT_RETURN_BASE_URL` ke URL HTTPS frontend. Login sebagai admin bisnis. Jalankan **Paket PayLanggan** dan **Status langganan PayLanggan**.
4. Sesudah persetujuan admin atas biaya awal dan debit otomatis bulanan, set `run_platform_checkout=true` dan jalankan **Mulai checkout PayLanggan** sekali. Buka `platform_checkout_url` di browser, bayar dengan kartu uji Xendit, lalu set flag kembali `false`. Jangan memakai kartu sungguhan. Xendit dapat memerlukan aktivasi fitur merchant initiated transactions untuk debit berikutnya.
5. Jalankan **Status langganan PayLanggan** dan **Daftar tagihan PayLanggan**. Setelah webhook atau rekonsiliasi worker, status harus `active`, invoice `paid`, dan `auto_renew=true` bila token kartu berhasil tersimpan. Bila `auto_renew=false`, periksa token kartu pada sesi Xendit dan konfigurasi penyimpanan kartu sebelum menguji perpanjangan.
6. Untuk uji perpanjangan tanpa menunggu sebulan, gunakan database **khusus pengujian** untuk menggeser `paid_through` dan `next_billing_date` ke masa lalu, lalu jalankan `go run ./cmd/platform-worker -once -limit 10`. Verifikasi invoice periode kedua berstatus `paid` dan hanya ada satu charge per invoice. Pengubahan tanggal langsung di database hanya untuk simulasi.
7. **Batalkan langganan PayLanggan** memerlukan `run_irreversible=true`. Pembatalan menghentikan debit baru dan mempertahankan periode yang sudah dibayar. Jangan jalankan saat masih menguji perpanjangan.

Jika provider menolak atau respons debit tidak pasti, worker mencatat status `failed` atau `review_required` pada `platform_autopay_charges`; jangan ulangi debit secara manual tanpa memastikan status terakhir di Xendit.

## Newman

Dari direktori `Backend`, setelah memasang Newman:

```sh
newman run postman/PAY-LANGGAN.postman_collection.json -e postman/PAY-LANGGAN.local.postman_environment.json --env-var login_email=admin@example.com --env-var login_password=PASSWORD_ANDA
```

Atau gunakan `make test-api`; target tersebut memakai kedua file yang sama. Untuk pengujian CI, simpan kredensial di secret runner dan berikan melalui `--env-var`; jangan commit environment hasil export yang memuat token, password, atau callback token. Nilai variabel yang diubah saat Newman berjalan bersifat sementara kecuali diekspor secara eksplisit.
