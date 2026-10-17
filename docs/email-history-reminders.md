# Riwayat email dan pengingat pelanggan

## Menggunakan fitur

Buka **Email Pelanggan** (`/dashboard/email-pelanggan`). Halaman memuat email invoice, konfirmasi pembayaran, trial dimulai, pengingat trial, dan pengingat jatuh tempo milik bisnis yang sedang login. Filter tersedia untuk status, jenis email, nama pelanggan, dan alamat email. Riwayat diperbarui setiap 5 detik saat tab terlihat.

Admin dan finance dapat mencoba ulang email berstatus **Gagal**, dengan konfirmasi. Sales dapat melihat riwayat. Status berasal dari antrean backend: Dalam antrean, Mengirim, Terkirim, Preview lokal, Gagal, atau Dilewati. Mode preview tidak mengirim email sungguhan.

**Detail** menampilkan subjek dan HTML yang disimpan ketika worker memproses email. Email lama sebelum migrasi 45 belum memiliki salinan HTML historis; preview untuk email tersebut tidak tersedia. Preview memakai iframe sandbox tanpa skrip. Riwayat dan detail selalu dibatasi business_id dari sesi pengguna.

## Pengingat otomatis

Pengingat **nonaktif secara default**. Aktifkan melalui switch **Pengingat otomatis** pada halaman Email Pelanggan:

- Trial: masuk antrean ketika sisa trial paling lama 72 jam dan belum berakhir.
- Invoice: masuk antrean ketika jatuh tempo paling lama 24 jam lagi, belum lunas, dan jumlahnya positif.
- Hanya trial/invoice yang dibuat sejak aktivasi pengingat yang diproses. Mengaktifkan kembali menetapkan batas baru.
- Invoice yang langsung jatuh tempo saat checkout tidak mendapat pengingat sebelumnya. Fitur ini belum mencakup pengingat lewat jatuh tempo.
- Worker memeriksa kembali preferensi, status bisnis, trial, dan invoice sebelum mengirim. Pengingat yang tidak diperlukan ditandai **Dilewati**.
- Batas unik per jenis email/invoice atau trial mencegah penjadwalan ganda. Antrean menggunakan retry worker yang sudah ada, maksimum lima percobaan sebelum Gagal. Retry manual mengatur ulang percobaan.
- Tanggal akhir trial dan jatuh tempo mengikuti penyimpanan UTC proyek. Batas aktivasi untuk `created_at` mengikuti zona waktu sesi PostgreSQL karena kolom tersebut diisi `NOW()` tanpa zona waktu.

Syarat menjalankan: migrasi **45**, API terbaru, dan `go run ./cmd/email-worker`. Pengiriman sungguhan memerlukan `EMAIL_DELIVERY_MODE=smtp` beserta konfigurasi SMTP yang sudah digunakan. Jangan memasukkan kata sandi SMTP ke frontend. Logo dan tautan email tetap memerlukan alamat publik yang aktif; domain tunnel sementara dapat berubah.

SMTP tidak menyediakan jaminan pengiriman tepat sekali: bila proses berhenti setelah SMTP menerima email tetapi sebelum status tersimpan, retry dapat mengirim ulang. Status Terkirim berarti server SMTP menerima pesan, bukan jaminan pesan masuk inbox atau dibaca.

## Pengujian

`PAYMENT_INTEGRATION=1 go test ./internal/services/payment -run TestEmailHistoryRemindersPostgres -count=1` (atur environment dengan sintaks terminal yang digunakan).

Tes menggunakan schema PostgreSQL sementara dan mode preview: default nonaktif, batas data baru, jadwal, deduplikasi paralel, filter, akses antar bisnis, salinan HTML, tautan invoice, retry, serta pembatalan pengingat setelah opt-out/pelunasan/pembatalan trial. Tidak mengirim email sungguhan.
