# Catatan bimbingan PayLanggan — 24 September 2026

## Ringkasan proyek

PayLanggan adalah aplikasi untuk bisnis yang mengelola pelanggan, katalog layanan dan paket, langganan, invoice, serta pembayaran berulang. Backend menggunakan Go, Echo, dan PostgreSQL; frontend menggunakan React. Integrasi pembayaran diuji memakai Xendit Test Mode.

Ada **dua alur pembayaran yang berbeda**:

1. **Pelanggan membayar bisnis:** pelanggan membayar invoice langganan layanan milik bisnis. Invoice dan metode pembayaran pelanggan dicatat dalam data tenant bisnis tersebut.
2. **Bisnis membayar PayLanggan:** admin bisnis membayar biaya penggunaan platform. Saat ini hanya ada satu paket tetap, **PayLanggan Bulanan Rp300.000/bulan**. Tagihan dan metode pembayarannya disimpan terpisah dari invoice pelanggan bisnis.

## Pekerjaan yang sudah dibuat

| Area | Hasil |
| --- | --- |
| API dan Postman | Collection dan environment lokal mencakup autentikasi, katalog, pelanggan, langganan, invoice, checkout, riwayat pembayaran, autopay pelanggan, dan tagihan bisnis ke PayLanggan. Request mutasi dijaga flag agar tidak terkirim tanpa sengaja. |
| Checkout pelanggan | Link checkout publik menghubungkan invoice dengan halaman pembayaran kartu Xendit. Webhook dan worker rekonsiliasi memperbarui status pembayaran. |
| Tagihan berkala pelanggan | Billing worker membuat invoice periode berikutnya berdasarkan jadwal langganan. |
| Autopay pelanggan | Metode pembayaran tersimpan dapat dipakai untuk debit periode berikutnya setelah pengaturan persetujuan autopay diaktifkan. |
| Langganan bisnis PayLanggan | Paket tetap Rp300.000/bulan, checkout pertama dengan persetujuan debit otomatis, penyimpanan token kartu, daftar tagihan bisnis, pembatalan perpanjangan, worker invoice dan debit berkala, serta halaman admin **Tagihan PayLanggan**. |
| Database | Migrasi versi **30** diterapkan pada PostgreSQL lokal. Tabel baru: `platform_subscriptions`, `platform_invoices`, `platform_checkout_sessions`, dan `platform_autopay_charges`. |

## Bukti pengujian yang sudah diperoleh

### Pembayaran pelanggan bisnis — diuji langsung di Xendit Test Mode

1. Akun bisnis dibuat dan login melalui Postman; data katalog, pelanggan, dan langganan bulanan disiapkan.
2. Invoice awal Rp100.000 dibayar memakai kartu uji Xendit. Respons invoice menunjukkan status **`paid`**.
3. Billing worker menghasilkan invoice periode kedua dengan status **`pending`** untuk langganan yang sama.
4. Metode pembayaran Xendit tersimpan dan autopay diaktifkan dengan persetujuan; respons pengaturan menunjukkan **`enabled: true`**.
5. Pada simulasi jatuh tempo, worker pembayaran memulai debit otomatis. Hasil pemeriksaan database menunjukkan **charge `succeeded`**, pembayaran **`success`**, dan invoice **`paid`**.

### Langganan bisnis ke PayLanggan — diuji pada database sementara

- Alur checkout pertama, penolakan tanpa persetujuan, pemisahan tenant, validasi token webhook, pengulangan webhook, pembuatan invoice perpanjangan, satu debit per invoice, dan pembatalan telah diuji pada skema PostgreSQL sementara dengan provider Xendit tiruan.
- `go test ./...` pada backend lulus; build frontend (`npm run build`) lulus.
- Migrasi lokal berhasil sampai **Goose versi 30**, dan keempat tabel tagihan bisnis sudah diverifikasi ada.
- **Pembayaran bisnis Rp300.000 melalui Xendit Test Mode yang sesungguhnya belum dijalankan.** Karena itu, hasil debit bisnis di provider nyata belum dapat diklaim berhasil.

## Alur demo yang bisa ditunjukkan

1. Tampilkan Postman: login, daftar langganan pelanggan, invoice awal `paid`, dan invoice perpanjangan.
2. Jelaskan hasil autopay pelanggan: persetujuan aktif, charge `succeeded`, dan invoice berubah menjadi `paid`.
3. Buka halaman admin `/dashboard/tagihan-platform`: satu paket Rp300.000/bulan, persetujuan debit otomatis, checkout, status, dan riwayat tagihan.
4. Tunjukkan migrasi versi 30 dan tabel baru sebagai bukti pemisahan tagihan bisnis dari invoice pelanggan.
5. Jelaskan pengujian otomatis yang lulus dan batasnya: provider tiruan menguji logika aplikasi, sedangkan uji Xendit Test Mode untuk alur bisnis masih menjadi langkah berikutnya.

## Pekerjaan berikutnya

1. Jalankan API, frontend, dan `go run ./cmd/platform-worker` dengan konfigurasi Xendit Test Mode dan URL HTTPS yang aktif.
2. Bayar checkout bisnis Rp300.000 memakai kartu uji, lalu pastikan status langganan `active`, invoice `paid`, dan `auto_renew=true`.
3. Simulasikan periode berikutnya pada database pengujian, jalankan platform worker, dan cocokkan status charge lokal dengan transaksi Xendit.
4. Uji skenario gagal bayar, webhook terlambat, dan pembatalan pada Xendit Test Mode sebelum menyatakan alur bisnis siap produksi.

## Kalimat ringkas untuk disampaikan

> “Saya sudah menyelesaikan alur langganan pelanggan sampai pembayaran awal, pembuatan invoice periode berikutnya, dan debit otomatis dengan Xendit Test Mode. Saya juga menambahkan alur terpisah agar bisnis membayar biaya platform PayLanggan Rp300.000 per bulan setelah memberi persetujuan debit otomatis. Migrasi database dan pengujian integrasi aplikasi sudah lulus. Langkah berikutnya adalah membuktikan pembayaran dan perpanjangan biaya platform itu langsung di Xendit Test Mode.”

Rincian langkah teknis pengujian tersedia di [`Backend/postman/README.md`](../Backend/postman/README.md).
