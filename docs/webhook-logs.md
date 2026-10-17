# Webhook Logs

Buka menu utama **Webhook Logs** dengan akun admin bisnis, di antara Kupon dan Pengaturan.

- Filter berdasarkan Xendit/Midtrans dan status penerimaan.
- Lihat waktu, event, referensi transaksi, respons HTTP, status pemrosesan, dan detail kegagalan.
- Klik **Perbarui** untuk mengambil log terbaru.
- **Diterima** berarti endpoint mengakui notifikasi; ini tidak selalu berarti pembayaran berhasil. Notifikasi pending atau pesanan di luar PayLanggan juga dapat diterima.
- **Duplikat** berarti event sudah pernah diterima. Pembayaran tidak dicatat dua kali.
- **Gagal** berarti notifikasi mencapai endpoint tetapi ditolak atau gagal diproses.

Migrasi `2026-10-05-37_business_webhook_logs.sql` menambahkan tabel metadata penerimaan. Riwayat event yang sudah terhubung ke checkout bisnis disalin ke tabel tersebut; kegagalan lama yang belum disimpan tidak dapat direkonstruksi. Kegagalan baru dicatat melalui endpoint webhook bisnis.

Payload mentah, nomor kartu, API key, signature, dan token verifikasi tidak dimasukkan dalam log ini. Setiap query dibatasi ke bisnis pengguna yang login. Webhook yang tidak mencapai server, misalnya tunnel mati atau DNS gagal, hanya dapat diperiksa di dashboard gateway.
