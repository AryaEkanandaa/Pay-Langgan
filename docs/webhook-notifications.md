# Notifikasi Webhook Logs

## Navigasi dan perilaku

- Menu utama: Ringkasan, Pelanggan, Langganan, Invoice, Katalog, Kupon, Webhook Logs, Pengaturan. Akses sesuai peran tetap berlaku; Webhook Logs khusus admin bisnis.
- Submenu Pengaturan hanya Bisnis, Payment Gateway dan Integrasi Website.
- URL utama `/dashboard/webhook-logs`. URL lama `/dashboard/pengaturan/webhook-logs` dialihkan dengan replace; tautan Payment Gateway diperbarui.
- Badge menunjukkan jumlah log dengan `received_at` setelah waktu terakhir pengguna melihat halaman. Nol disembunyikan; lebih dari 99 ditampilkan `99+`. Animasi pendek mendukung reduced motion.
- Penanda terakhir dilihat disimpan di database per pengguna, bukan localStorage. Query memakai user ID dan business ID dari token login, bukan dari input browser.
- Polling COUNT setiap 30 detik, tanpa mengambil seluruh daftar. Tidak ada request saat tab tersembunyi; saat tab aktif kembali count langsung diperiksa. Gangguan jaringan mempertahankan nilai terakhir.
- Halaman menyimpan waktu terakhir dilihat yang lama di memori untuk penanda baris selama kunjungan. Setelah daftar tampil, backend diperbarui sampai waktu server yang ditangkap **sebelum** mengambil daftar. Log lebih baru dari batas itu tidak ikut dianggap dibaca secara tidak sengaja.
- Log baru yang masuk selama halaman terbuka ditawarkan melalui tombol “N log baru, muat ulang”. Baris tidak disisipkan otomatis. Perbarui, pagination, filter dan detail tetap tersedia.
- Membuka halaman dianggap membaca semua log sampai batas waktu tersebut, termasuk log di halaman lain atau di luar filter. Ini mengikuti definisi terakhir membuka halaman, bukan status baca tiap baris.
- Perangkat lain tidak boleh memundurkan penanda yang sudah lebih baru. Waktu masa depan dibatasi ke waktu server.

## Database dan API

Migrasi `Migrations/2026-10-07-39_webhook_logs_last_seen.sql` menambah:

```sql
users.webhook_logs_last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

Pengguna lama diinisialisasi pada waktu migrasi agar riwayat lama tidak memunculkan ratusan notifikasi. Migrasi sudah diterapkan di database lokal dan dicatat sebagai versi 39. Untuk lingkungan lain, jalankan migrasi melalui alur proyek sebelum memuat ulang API. Tidak ada perubahan pada data pembayaran, kredensial gateway, atau isi webhook.

Index yang sudah ada `idx_business_webhook_logs_recent (business_id, received_at DESC, id DESC)` digunakan; tidak dibuat index duplikat.

Endpoint baru, dengan middleware autentikasi, tenant dan role admin yang sama dengan daftar log:

- `GET /api/v1/settings/webhook-logs/new-count`: `{ count, last_seen_at, checked_at }`.
- `POST /api/v1/settings/webhook-logs/seen`: body `{ "seen_through": "waktu checked_at dari server" }`; mengembalikan count terbaru beserta penanda waktu.
- Respons menggunakan envelope API proyek dan `Cache-Control: no-store`.

## File yang berubah

| File | Alasan |
| --- | --- |
| `Frontend/src/data/dashboardNav.ts` | Urutan menu dan ikon Webhook Logs sebagai menu utama. |
| `Frontend/src/components/layout/DashboardLayout.tsx` | Badge desktop/mobile, aria-label, state aktif submenu dan penyedia state notifikasi. |
| `Frontend/src/components/ui/NavBadge.tsx` | Komponen badge reusable dengan 99+ dan animasi kenaikan. |
| `Frontend/src/hooks/useNewWebhookCount.ts` | Polling, visibility, abort, state bersama dan penanganan kegagalan jaringan. |
| `Frontend/src/pages/dashboard/WebhookLogsPage.tsx` | Baseline per kunjungan, mark-seen setelah data tampil, penanda baris dan tombol muat ulang. |
| `Frontend/src/lib/api.ts` | Tipe respons count, fetch count/seen dan AbortSignal daftar log. |
| `Frontend/src/App.tsx` | Route utama dan redirect URL lama. |
| `Frontend/src/pages/dashboard/PaymentGatewayPage.tsx` | Link internal menuju URL utama. |
| `Frontend/src/index.css` | Badge, aksen baris, mode gelap dan reduced motion. |
| `Backend/internal/services/payment/webhook_logs.go` | COUNT terbatas tenant/pengguna dan update penanda yang monoton. |
| `Backend/internal/handlers/payment/checkout_handler.go` | Handler count dan mark-seen dengan validasi input. |
| `Backend/internal/routes/routes.go` | Registrasi endpoint di grup admin yang sudah ada. |
| `Backend/internal/services/payment/webhook_unread_integration_test.go` | Pengujian PostgreSQL terisolasi untuk tenant, admin, race timestamp, antar perangkat dan pengguna lama. |
| `Migrations/2026-10-07-39_webhook_logs_last_seen.sql` | Kolom waktu terakhir dilihat dan rollback. |
| `docs/webhook-logs.md`, `docs/webhook-notifications.md` | Panduan navigasi, implementasi dan pengujian. |

Tidak ada dependency baru.

## Cara menguji dengan gateway sandbox

1. Login sebagai admin bisnis, buka halaman selain Webhook Logs. Riwayat lama tidak memunculkan badge setelah migrasi.
2. Kirim test webhook melalui Xendit/Midtrans ke URL webhook bisnis yang aktif. Pastikan request mencapai PayLanggan; request yang gagal di DNS/tunnel tidak menghasilkan log lokal.
3. Tunggu maksimal sekitar 30 detik atau pindah tab lalu kembali. Badge harus muncul. Beberapa pengiriman menghasilkan hitungan lebih tinggi, termasuk duplikat/gagal yang dicatat.
4. Buka Webhook Logs: log setelah penanda lama mendapat aksen kiri, latar tipis dan label Baru. Setelah data tampil badge kembali nol, kecuali ada log tambahan yang masuk selama pemuatan.
5. Kirim test lagi saat halaman terbuka. Tunggu polling: tombol “N log baru, muat ulang” muncul; daftar tetap. Klik tombol atau Perbarui untuk mengambil data.
6. Buka Detail dan coba filter gateway/status. Keluar lalu kunjungi halaman kembali: penanda Baru untuk log yang sudah dilihat menghilang.
7. Login dengan admin yang sama di browser/perangkat lain: penanda baca mengikuti database. Admin lain dalam bisnis yang sama mempunyai penanda terpisah.
8. Uji 360px, 768px, 1440px, mode terang/gelap dan reduced motion. Buka URL lama untuk memeriksa redirect. Saat tab tersembunyi, endpoint count tidak boleh terus dipanggil; jaringan gagal tidak mengosongkan badge.

## Verifikasi selama implementasi

- TypeScript strict, lint dan build frontend lulus.
- Uji PostgreSQL di schema sementara lulus: pemisahan tenant/pengguna, log yang datang selama pemuatan, penanda tidak mundur, penolakan user tenant lain, serta migrasi riwayat lama.
- Uji browser dengan API tiruan lulus: navigasi, redirect, 0/3/99+, polling 30 detik, pause tab tersembunyi, network failure, penanda per kunjungan, tombol muat ulang, filter/detail, responsif dan reduced motion.
- Pengiriman webhook sandbox sungguhan tidak dipicu selama implementasi; langkah di atas adalah pemeriksaan end-to-end melalui akun gateway Anda.
