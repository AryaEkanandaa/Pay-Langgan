# Portal pelanggan

## Akses dan penggunaan

Portal berada di `/portal/ID_INTEGRASI_PUBLIK`, menggunakan ID integrasi aktif yang sama dengan checkout. Di halaman **Integrasi Website**, bagian **Portal pelanggan** menyediakan tautan untuk membuka portal dan kode tombol:

```html
<button data-pl-portal>Kelola langganan saya</button>
```

Pasang dengan skrip header `paylanggan.js` yang sudah berisi `data-pl-site` dan `data-pl-api-url`. Tombol membuka portal PayLanggan. Portal menggunakan alamat API dari konfigurasi frontend PayLanggan.

1. Pelanggan mengisi email yang digunakan saat checkout.
2. PayLanggan mengantrekan email berisi tautan masuk, berlaku **15 menit** dan **sekali pakai**. Respons tetap sama untuk email yang tidak terdaftar.
3. Pelanggan membuka tautan lalu menekan **Masuk ke portal**. Konfirmasi ini menghindari konsumsi token oleh pemindai tautan email yang hanya membuka halaman.
4. Portal menampilkan langganan dan invoice untuk email terverifikasi pada bisnis tersebut. Data pelanggan duplikat dengan email sama dalam bisnis yang sama digabung dalam daftar; bisnis lain tetap terpisah.
5. Pelanggan dapat membuka rincian invoice, termasuk add-on dan diskon. Harga di kartu langganan adalah harga dasar plan saat ini; jumlah yang harus dibayar mengikuti invoice.
6. **Lanjutkan pembayaran** membuka checkout yang sudah ada. Invoice lunas tidak bisa dibayar ulang. Checkout yang masih berjalan dilanjutkan dengan token akses tambahan, tanpa mengganti token awal atau membuat transaksi provider baru. Checkout yang sedang dibuat atau sesi aktif yang kedaluwarsa memerlukan penyelesaian/reconciliation terlebih dahulu.
7. Setelah pembayaran berhasil, tombol kembali ke portal tersedia. Status portal diperbarui saat tab kembali fokus, setiap 30 detik ketika terlihat, atau melalui **Perbarui status**.

Sesi berlaku **12 jam** dan dapat dicabut melalui **Keluar**. Token sesi disimpan di sessionStorage khusus portal per bisnis/tab, terpisah dari akun admin. Bila storage browser tidak tersedia, sesi hanya tersimpan dalam memori dan hilang setelah reload. Token tetap perlu dilindungi seperti sesi login lain; jangan membagikan tautan masuk.

## Backend dan konfigurasi

Memerlukan migrasi **46**, API terbaru, dan worker email terbaru. Gunakan konfigurasi SMTP yang sudah ada. `APP_PUBLIC_URL` harus HTTPS publik aktif untuk email sungguhan; mode preview mengizinkan localhost HTTP. Setelah tunnel berubah, alamat frontend dan backend perlu diperbarui seperti alur checkout sebelumnya.

Endpoint baru di `/api/v1/portal/:site`:

- GET root: nama bisnis publik.
- POST `/request-link`: minta email masuk, cooldown 1 menit dan maksimum 5 permintaan per email/site per jam.
- POST `/exchange`: tukar token email sekali pakai menjadi sesi portal.
- GET `/overview`: daftar langganan/invoice milik email terverifikasi, paginasi terpisah 20 item.
- GET `/invoices/:id`: detail invoice milik pelanggan.
- POST `/invoices/:id/checkout`: buat atau lanjutkan checkout.
- POST `/logout`: cabut sesi.

Endpoint privat menerima token portal melalui header Authorization Bearer. JWT admin tidak diterima sebagai sesi portal. Semua akses memeriksa ID integrasi masih aktif, bisnis aktif, dan email masih terdaftar. Pencabutan ID integrasi juga menutup akses portal melalui ID itu. Token email berada di fragment URL, dihapus dari alamat browser sebelum interaksi lanjutan, dan hanya hash disimpan untuk verifikasi. Worker merekonstruksi token email dari nonce + HMAC `JWT_SECRET`; nonce dibersihkan setelah pengiriman/konsumsi/kedaluwarsa.

Email login memakai antrean tersendiri dan worker email yang sama, maksimal 3 percobaan. Email ini tidak ditampilkan pada riwayat email transaksi bisnis agar tautan autentikasi tidak terlihat di dashboard. Mode preview membuat `portal-ID.html` lokal dan tidak mengirim email. SMTP menerima pesan bukan jaminan pesan sudah masuk inbox. Belum ada ubah plan, pembatalan langganan oleh pelanggan, atau perubahan data profil melalui portal ini.

## Pengujian

Atur `PAYMENT_INTEGRATION=1`, lalu dari Backend jalankan:

```text
go test ./internal/services/payment -run TestCustomerPortalPostgres -count=1
```

Tes menggunakan schema PostgreSQL sementara, email preview, dan provider mock. Mencakup email tidak terdaftar, cooldown, pelanggan duplikat, token lintas bisnis, token replay, expiry, logout, pencabutan integrasi, ownership invoice/detail/checkout, invoice lunas, serta resume pembayaran tanpa transaksi ganda. UI diuji dengan API tiruan pada 360/768/1440px terang/gelap, rincian diskon, empty state, permintaan email, login/logout, penghapusan fragment, dan navigasi checkout. Tidak mengirim email atau pembayaran sungguhan.
