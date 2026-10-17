# Status pelanggan dan checkout belum dibayar

Halaman **Pelanggan** menyediakan filter Semua pelanggan, Pelanggan aktif,
Trial, Calon pelanggan, dan Tidak aktif. Status dihitung dari langganan yang
dimiliki pelanggan, dengan prioritas Aktif > Trial > Calon pelanggan > Tidak aktif.
Pelanggan tanpa langganan atau yang hanya menunggu pembayaran tergolong calon
pelanggan. Pelanggan dengan langganan dijeda/dibatalkan tanpa langganan lain
yang aktif, trial, atau pending tergolong tidak aktif. Tidak ada migrasi baru.

Data pelanggan dan invoice tetap dibuat saat checkout dimulai. Langganan berbayar
tetap pending sampai pembayaran terverifikasi. Trial memakai status tersendiri.
KPI pelanggan aktif yang sudah ada mencakup aktif dan trial; pending dikecualikan.
MRR mencakup langganan aktif saja. Pendapatan diterima mengikuti pembayaran sukses.

Daftar invoice/langganan menampilkan **Checkout kedaluwarsa** jika tautan terakhir
sudah kedaluwarsa. Invoice tetap belum lunas dan dapat dilanjutkan melalui portal
atau tautan pembayaran baru untuk invoice yang sama. Kedaluwarsa tautan tidak
menghapus pelanggan, membatalkan invoice/langganan, atau mengaktifkan akses.
Sesi provider yang sedang diproses tidak dianggap kedaluwarsa hanya karena waktu
lokal habis; hasilnya harus dikonfirmasi provider melalui rekonsiliasi/webhook.

Checkout memakai ulang pelanggan berdasarkan email yang dinormalisasi per bisnis
(huruf kecil, tanpa spasi di tepi). Advisory lock mencegah pembelian paralel
membuat pelanggan ganda. Tambah/edit pelanggan manual menolak email yang sudah
dipakai pelanggan lain dalam bisnis yang sama. Pelanggan lama yang sudah duplikat
tidak digabung atau dihapus otomatis. Email yang sama di bisnis berbeda diperbolehkan.
Percobaan ulang pesanan dengan request_id/external_order_id yang sama tetap memakai
pesanan dan invoice yang sama. Pembelian baru dengan ID pesanan berbeda tetap
diperbolehkan; satu pelanggan dapat memiliki beberapa langganan.

## Pengujian manual

1. Mulai checkout dengan email baru, jangan bayar: pelanggan Calon pelanggan,
   langganan Menunggu pembayaran, invoice belum lunas; MRR tidak bertambah.
2. Ulangi checkout menggunakan email sama dengan variasi kapitalisasi/spasi:
   tidak membuat pelanggan baru. Mengulang ID pesanan yang sama tidak membuat invoice baru.
3. Selesaikan pembayaran sandbox: invoice Lunas, langganan Aktif, filter Pelanggan
   aktif menampilkan pelanggan tersebut setelah halaman dimuat ulang.
4. Mulai trial: badge Trial, MRR tidak bertambah. Pelanggan yang sudah mempunyai
   langganan aktif tetap berstatus Aktif meskipun memiliki trial/pending lain.
5. Setelah tautan checkout kedaluwarsa: pesan kedaluwarsa tampil pada invoice dan
   langganan; buka portal untuk melanjutkan pembayaran invoice yang sama.
6. Cari nama/email bersamaan dengan filter dan pagination. Coba tambah/edit email
   duplikat: ditolak dengan pesan yang menjelaskan pelanggan sudah terdaftar.

Tes otomatis memakai schema PostgreSQL sementara dan data sintetis:
`PAYMENT_INTEGRATION=1 go test ./internal/services/payment -run 'TestCustomerLifecyclePostgres|TestMerchantOrderIsScopedAndIdempotent' -count=1`
(sesuaikan sintaks environment dengan terminal). Tidak mengirim email/pembayaran sungguhan.
