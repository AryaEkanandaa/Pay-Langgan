# Email setelah checkout

## Perilaku

- Pembayaran Xendit/Midtrans yang sudah diverifikasi backend membuat email **Pembayaran berhasil** beserta invoice lunas dan tautan melihat invoice.
- Checkout Rp0 membuat konfirmasi tanpa menyatakan ada dana yang ditagihkan. Invoice terbit otomatis tidak dikirim lagi untuk checkout Rp0.
- Trial yang berhasil didaftarkan membuat email **Trial dimulai**, tanggal akhir trial, paket, add-on, diskon, serta estimasi biaya setelah trial. Email menjelaskan belum ada pembayaran/debit otomatis.
- Konfirmasi ini otomatis untuk kejadian baru, terpisah dari opsi `invoice.auto_send` yang mengatur pengiriman saat invoice terbit. Tidak ada pengiriman massal untuk transaksi/trial lama.
- Logo publik, warna, alamat, footer, dan Reply-To mengikuti kustomisasi invoice bisnis. Judul dan pesan transaksi mengikuti jenis konfirmasi supaya tidak mengirim instruksi pembayaran pada invoice lunas.
- Tanggal pada email trial dan tanggal pembayaran ditampilkan dalam UTC secara eksplisit. Harga trial adalah estimasi saat pendaftaran, bukan harga yang dikunci untuk penagihan berikutnya.

## Teknis

Migrasi 43 menambahkan `checkout_email_deliveries`. Pekerjaan dibuat dalam transaksi yang sama dengan pembayaran/pendaftaran trial. Constraint unik membatasi satu konfirmasi per invoice lunas atau subscription trial. Replay webhook dan retry order tidak menambah pekerjaan baru.

Worker yang sama, `go run ./cmd/email-worker`, menjalankan antrean reset password, invoice, dan konfirmasi checkout setiap 5 detik. Mode `EMAIL_DELIVERY_MODE` (`smtp`, `preview`, `disabled`) tetap berlaku. SMTP Gmail dan pengirim PayLanggan menggunakan konfigurasi yang sudah ada.

Pekerjaan diklaim dengan lease 2 menit dan `FOR UPDATE SKIP LOCKED`, maksimal 5 percobaan dengan jeda bertambah. `Message-ID` stabil `checkout-ID` membantu identifikasi pesan. SMTP tidak menjamin exactly-once: jika proses mati sesudah SMTP menerima pesan tetapi sebelum status tersimpan, retry mungkin mengirim ulang. Constraint mencegah pekerjaan ganda dari webhook, bukan menjamin deduplikasi Gmail.

Tidak ada endpoint baru. Tautan invoice lunas memakai endpoint publik invoice yang sudah ada. Panel Email invoice dan tombol Kirim email tetap mengelola email invoice biasa; status konfirmasi checkout disimpan terpisah dan belum memiliki panel riwayat khusus di dashboard.

## Cara mencoba

1. Pastikan API, frontend, dan email-worker aktif; SMTP sudah dikonfigurasi.
2. Buat checkout baru memakai email milik sendiri, lalu selesaikan pembayaran Sandbox.
3. Setelah backend memverifikasi pembayaran, periksa inbox/spam untuk konfirmasi lunas. Biasanya antrean diambil dalam 5 detik, di luar waktu webhook/SMTP.
4. Buat plan trial dan daftar menggunakan email yang belum memakai trial tersebut. Daftarkan kartu melalui gateway yang dipilih, lalu periksa konfirmasi trial. Email baru dijadwalkan setelah pendaftaran kartu selesai.
5. Add-on/kupon harus tercantum dalam rincian email. Logo hanya tampil jika URL aset publik permanen valid dan sudah dikonfigurasi; pengiriman tidak memperbaiki storage yang belum tersedia.

Pengujian otomatis memakai schema PostgreSQL terisolasi dan mode preview, tanpa email sungguhan. Mencakup Xendit/Midtrans, duplikasi event, retry trial, checkout Rp0, rollback, tautan invoice, template escaping, dan worker replay.

Ketika tunnel backend berubah, perbarui `API_PUBLIC_URL` dan webhook gateway serta restart proses terkait. Tautan email yang sebelumnya dikirim menggunakan tunnel lama tetap sementara; produksi memerlukan domain stabil.
