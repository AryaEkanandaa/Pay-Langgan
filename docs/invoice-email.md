# Email invoice pelanggan

## Yang tersedia

- **Pengaturan Bisnis → Tampilan Invoice** (`/dashboard/pengaturan/invoice`): logo raster PNG/JPG atau URL HTTPS permanen, warna, alamat, catatan, subjek email, pesan, dan alamat balasan bisnis.
- **Invoice → Detail → Preview email**: melihat HTML email memakai invoice asli dan pengaturan yang sudah disimpan.
- **Buat preview lokal / Kirim email**: satu antrean per invoice, dengan status queued, sending, previewed, sent, atau failed. Klik berulang tidak membuat antrean ganda. Preview/failed dapat disiapkan ulang dengan pengaturan dan alamat pelanggan terbaru. Invoice yang sudah terkirim tidak dikirim ulang oleh tombol ini.
- Pengiriman otomatis adalah opsi bisnis untuk invoice **baru**, termasuk invoice pembaruan langganan. Mengaktifkannya tidak mengirim invoice lama secara massal. Alamat pelanggan yang tidak valid dilewati pada penerbitan otomatis; pengiriman manual menunjukkan error.
- Email berisi rincian plan, add-on, diskon, total, dan jatuh tempo dari invoice. Isi bisnis diperlakukan sebagai teks, bukan HTML bebas. Nominal tidak dapat diganti lewat template.
- Email mempunyai tautan privat untuk melihat status invoice terbaru dan melanjutkan pembayaran. Tautan berlaku 90 hari; perlakukan tautan tersebut sebagai informasi privat.
- Pengirim berasal dari SMTP **PayLanggan**, bukan SMTP bisnis. `reply_to` mengarahkan balasan pelanggan ke bisnis.

## Mode preview (opsional)

Di `Backend/.env`:

```dotenv
EMAIL_DELIVERY_MODE=preview
EMAIL_PREVIEW_DIR=../.tmp/email-preview
```

Mode ini tidak membuat koneksi SMTP dan tidak mengirim email. Worker membuat file `D:\PayLanggan\.tmp\email-preview\invoice-ID.html`; status menjadi **Preview dibuat — belum dikirim**. `sent_at` tetap kosong. Preview di dashboard dapat dibuka tanpa worker.

Jalankan dari terminal terpisah:

```powershell
cd D:\PayLanggan\Backend
go run ./cmd/api
```

```powershell
cd D:\PayLanggan\Backend
go run ./cmd/email-worker
```

Untuk penerbitan invoice pembaruan otomatis, jalankan worker billing yang sudah ada:

```powershell
cd D:\PayLanggan\Backend
go run ./cmd/billing
```

Jangan jalankan dua API pada port yang sama. Setelah mengubah env, hentikan hanya API dan worker email dengan Ctrl+C di terminal masing-masing, kemudian jalankan lagi. Frontend harus mengarah ke backend yang aktif.

## Mengaktifkan pengiriman sungguhan nanti

Untuk logo email sungguhan, gunakan URL gambar HTTPS publik; beberapa aplikasi email memblokir logo yang disimpan sebagai data gambar.

1. Siapkan layanan SMTP untuk email transaksi dan alamat pengirim milik PayLanggan. Ambil host, port, username, dan password SMTP dari layanan tersebut. Jangan memasukkan password SMTP ke frontend atau membagikannya di chat.
2. Verifikasi alamat/domain pengirim dan pasang SPF/DKIM sesuai instruksi layanan SMTP. Pastikan alamat pengirim diizinkan layanan tersebut.
3. Isi `Backend/.env`:

```dotenv
EMAIL_DELIVERY_MODE=smtp
MAIL_FROM="PayLanggan <invoice@DOMAIN_ANDA>"
SMTP_HOST=HOST_DARI_LAYANAN_SMTP
SMTP_PORT=587
SMTP_USERNAME=USERNAME_SMTP
SMTP_PASSWORD=PASSWORD_SMTP
SMTP_IMPLICIT_TLS=false
API_PUBLIC_URL=https://DOMAIN_BACKEND_AKTIF
APP_PUBLIC_URL=https://DOMAIN_FRONTEND_AKTIF
```

Port 587 memakai STARTTLS. Jika layanan meminta TLS langsung pada port 465, gunakan `SMTP_PORT=465` dan `SMTP_IMPLICIT_TLS=true`. Koneksi selalu memverifikasi sertifikat TLS; SMTP tanpa TLS tidak didukung.

4. Gunakan URL backend/frontend publik yang aktif. Domain `trycloudflare.com` dapat berganti ketika tunnel dimulai ulang. Perbarui env terkait, skrip demo, dan webhook gateway sebagaimana alur tunnel sebelumnya. Untuk penggunaan tetap, gunakan domain tetap.
5. Restart API dan worker email. Coba **satu invoice dengan alamat email Anda sendiri** melalui tombol Kirim email sebelum mengaktifkan pengiriman otomatis bisnis.
6. Periksa status **Terkirim**, inbox, dan folder spam. Status Terkirim berarti server SMTP menerima email, bukan jaminan email masuk inbox. SMTP Gmail sudah dikonfigurasi pada lingkungan lokal dan satu email uji telah diterima server SMTP.
7. Preview yang dibuat sebelumnya tetap berstatus previewed sampai Anda meminta pengiriman lagi. Beralih ke SMTP tidak otomatis mengirim semua preview terdahulu. Namun pekerjaan yang masih queued akan diproses worker; periksa antrean sebelum berpindah mode.

## Cara kerja antrean

Invoice dan pekerjaan otomatis disimpan dalam transaksi yang sama. Worker mengambil pekerjaan dengan row lock dan lease dua menit, mencatat upaya, dan mencoba kegagalan maksimal lima kali. Kegagalan diberi kode umum tanpa merekam password SMTP. Proses yang berhenti dapat dilanjutkan worker berikutnya. Pengiriman SMTP tidak bisa dijamin tepat satu kali apabila proses berhenti setelah SMTP menerima email tetapi sebelum status database tersimpan; Message-ID dibuat tetap untuk invoice tersebut.

Kustomisasi disalin saat pekerjaan disiapkan. Preview dashboard memakai snapshot pekerjaan jika sudah tersedia; sebelum ada pekerjaan, pengaturan bisnis terkini digunakan. Mengubah tampilan invoice tidak mengganti isi email yang sudah terkirim. Email berupa HTML di badan pesan; tahap ini belum membuat lampiran PDF atau laporan bounce dari penyedia SMTP.

## API

Semua endpoint dashboard memeriksa kepemilikan invoice terhadap bisnis pada token login:

- `GET /api/v1/invoices/:id/email` — status dan mode.
- `GET /api/v1/invoices/:id/email-preview` — subjek dan HTML, tanpa pengiriman.
- `POST /api/v1/invoices/:id/email` — antrean manual, hanya admin/finance.
- `GET /api/v1/invoice-emails/:token` — invoice privat melalui token acak.
- `POST /api/v1/invoice-emails/:token/checkout` — membuka checkout milik gateway bisnis. Pembayaran yang sedang berlangsung tidak dibuat ulang.

Migrasi: `Migrations/2026-10-05-38_invoice_email_deliveries.sql`. Terapkan migrasi sebelum menjalankan versi API/billing worker ini.

## Status tombol dan logo publik

Lihat [panduan status dan storage logo](invoice-email-status-logo.md) untuk konfigurasi PUBLIC_ASSET_URL, disk persisten, dan pengujian Gmail.
