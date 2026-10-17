# Aktivasi langganan setelah pembayaran

## Alur

| Kejadian | Status langganan |
| --- | --- |
| Daftar paket berbayar tanpa trial | `pending` — Menunggu pembayaran |
| Checkout gagal, ditutup, atau kedaluwarsa tanpa pelunasan | Tetap `pending` |
| Pembayaran Xendit/Midtrans terverifikasi | `active` — Aktif |
| Invoice Rp0 | Langsung lunas dan `active` |
| Trial dimulai | `trial` — Trial |
| Trial berakhir dan invoice pertama belum lunas | `pending` |
| Invoice pertama trial dilunasi | `active` |
| Pembayaran terlambat setelah dibatalkan/dijeda | Invoice lunas; langganan tetap dibatalkan/dijeda |

Aktivasi, pelunasan invoice, dan antrean konfirmasi email berada dalam transaksi yang sama. Replay webhook tidak membuat aktivasi/email ganda. Penandaan lunas manual oleh admin dan hasil debit otomatis terverifikasi menggunakan aktivasi yang sama.

Status pending tidak diproses worker renewal, sehingga invoice pertama yang belum dibayar tidak menghasilkan invoice periode berikutnya terus-menerus. Status pending tidak dapat dijeda/dilanjutkan; bisnis dapat membatalkannya. Pelanggan tetap tercatat di daftar pelanggan, tetapi tidak termasuk pelanggan aktif bila tidak punya langganan aktif/trial lain.

Langganan aktif yang sudah memasuki perpanjangan tetap mengikuti kebijakan renewal sebelumnya; perubahan ini tidak menambahkan penangguhan otomatis untuk tunggakan renewal. Tanggal mulai, anchor, dan periode invoice tetap mengikuti saat pendaftaran/periode trial yang disepakati; pembayaran terlambat tidak otomatis memperpanjang tanggal akhir. Perubahan periode/ganti plan/pro-rata tidak termasuk pekerjaan ini.

## Data lama

Migrasi 44 menambah constraint status pending dan mengoreksi hanya langganan aktif yang terhubung ke merchant order dengan invoice berbayar berstatus pending, tanpa satu pun invoice paid. Langganan pernah lunas, cancelled/paused, dan data admin/legacy tanpa bukti order tersebut tidak diubah. Rollback migrasi memindahkan pending menjadi paused agar tidak mengaktifkan akses yang belum dibayar.

## Pengujian

Pengujian schema PostgreSQL terisolasi: checkout/order pending, Xendit/Midtrans mengaktifkan setelah pembayaran, replay, trial expiry, pending dikecualikan renewal, invoice Rp0, dan pembayaran tidak mengaktifkan cancelled/paused/trial yang belum waktunya. UI diuji menggunakan API mock untuk label, filter, tombol Batalkan, tanpa Jeda/Lanjutkan, mode terang/gelap pada 360/768/1440px. Tidak membuat pembayaran atau mengirim email uji sungguhan.
