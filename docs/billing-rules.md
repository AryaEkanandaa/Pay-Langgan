# Aturan billing PayLanggan

## Pembuatan subscription

- Waktu billing menggunakan UTC. Kolom TIMESTAMP disimpan sebagai waktu UTC.
- `next_billing_date` adalah awal periode yang belum dibuatkan invoice, bukan
  batas pembayaran invoice yang sudah terbit.
- Tanpa trial: subscription `active`, invoice pertama jatuh tempo saat dibuat,
  dan tanggal billing berikutnya maju satu periode.
- Dengan trial: subscription `trial`, belum ada invoice; tanggal billing dan
  `billing_anchor` dimulai pada akhir trial.
- `billing_cycle` pada subscription menyimpan siklus saat pendaftaran. Perubahan
  siklus di katalog tidak mengubah jadwal subscription yang sudah ada.
- Status subscription menunjukkan lifecycle, bukan bukti pembayaran.
- Subscription, add-on, penerapan kupon, invoice awal beserta rinciannya, dan
  audit log disimpan dalam satu transaksi. Kegagalan membatalkan semuanya.

## Periode dan rincian invoice

- Setiap invoice baru menyimpan `period_start` dan `period_end`. Periode mencakup
  awal dan tidak mencakup akhir: `[period_start, period_end)`.
- Tanggal acuan tetap dipertahankan: 31 Januari -> 28 Februari -> 31 Maret.
  Subscription tahunan tanggal 29 Februari kembali ke 29 Februari pada tahun
  kabisat, dan menggunakan 28 Februari pada tahun lainnya.
- Rincian plan, jumlah add-on, harga satuan, subtotal, dan potongan kupon disimpan
  pada `invoice_items`. Jumlah rincian harus sama dengan total invoice.
- Perhitungan invoice menggunakan bilangan bulat dalam satuan 0,01; persentase
  diskon dibulatkan ke bawah pada satuan tersebut. Nilai di luar NUMERIC(12,2)
  ditolak sebelum disimpan.
- Harga memakai katalog saat invoice diterbitkan. Edit katalog tidak mengubah
  invoice yang sudah terbit. Riwayat harga untuk menghitung ulang periode lampau
  belum tersedia; backlog memakai katalog saat diproses.
- Add-on harus memiliki siklus yang sama dengan subscription. Prorata dan
  penagihan add-on dengan siklus berbeda belum didukung.

## Renewal manual

`RenewalService.Run(ctx, at, limit)` mengambil subscription `active` atau `trial`
berdasarkan `next_billing_date <= at`. Setiap periode diproses dalam transaksi
tersendiri:

1. Ambil dan kunci subscription dengan `FOR UPDATE SKIP LOCKED`.
2. Hitung batas periode berikutnya dari tanggal acuan.
3. Simpan invoice, rincian biaya, dan audit log.
4. Ubah trial menjadi active jika diperlukan, lalu majukan tanggal billing.
5. Commit seluruh perubahan, atau rollback jika ada kegagalan.

Unique index `(subscription_id, period_start)` menjadi perlindungan tambahan
terhadap invoice ganda. Invoice yang sudah dibatalkan juga tetap memakai kunci
periode tersebut; pembatalan tidak otomatis memicu invoice pengganti.
Pemrosesan ulang melewati periode yang sudah selesai. Jika data tidak konsisten
hingga melanggar unique index, renewal melaporkan kegagalan dan tidak menebak
perbaikan tanggal.

Backlog diproses per periode sampai batas jumlah percobaan tercapai. Subscription
yang gagal dilewati selama run tersebut agar subscription lain tetap diproses.
Periode yang belum selesai dapat dicoba lagi pada run berikutnya. Baris yang sedang
dikunci proses lain dilewati dan dapat diambil pada run berikutnya.

## Kupon dan pembayaran

- Kupon yang masih terpasang berlaku pada renewal jika sudah diterapkan dan belum
  kedaluwarsa pada awal periode. Kupon yang kedaluwarsa tepat pada awal periode
  tidak digunakan. Invoice pertama dapat memakai kupon yang baru dipasang dalam
  transaksi pembuatan subscription.
- Jika beberapa kupon terpasang, kupon diproses berurutan menurut ID kupon.
  Setiap diskon dihitung dari saldo setelah diskon sebelumnya dan dibatasi agar
  total tidak negatif.
- `used_count` menghitung pemasangan kupon, termasuk pada trial; renewal tidak
  menaikkannya lagi. Kuota pemasangan yang penuh tidak membatalkan kupon yang
  sudah terpasang.
- Invoice positif berstatus `pending`; tanggal jatuh tempo adalah awal periode.
  Invoice nol rupiah langsung `paid` dengan `paid_at`, tanpa membuat transaksi
  pembayaran fiktif.
- Tagihan yang belum dibayar tidak menghentikan renewal pada tahap ini.
  Grace period, suspend otomatis, reminder, dan retry pembayaran belum tersedia.

## Pause, resume, cancel

- Paused dan cancelled tidak diambil oleh renewal. Invoice yang sudah terbit
  tetap disimpan dan tidak otomatis dibatalkan atau direfund.
- Resume sebelum jadwal berikutnya mempertahankan jadwal. Jika tanggal sudah
  lewat, tanggal billing dan acuan digeser ke waktu resume sehingga masa pause
  tidak ditagihkan sebagai backlog. Prorata/refund masa pause belum didukung.
- Perubahan status, add-on, dan kupon memakai kunci baris subscription yang sama
  dengan renewal. Dengan demikian perubahan tanggal hasil renewal tidak tertimpa
  oleh status yang dibaca sebelum renewal selesai.
- Perubahan add-on/kupon memengaruhi invoice berikutnya, tidak mengedit invoice
  yang sudah terbit. Riwayat add-on yang sudah dihapus belum tersedia untuk
  merekonstruksi backlog secara historis.

## Migrasi dan data lama

Migrasi `Migrations/2026-09-09-26_add_invoice_periods.sql` menambah kolom jadwal,
periode, constraint, dan index. Terapkan migrasi sebelum menjalankan backend
versi ini. Migrasi telah diuji di schema sementara dan diterapkan pada database lokal
`pay_langgan`, schema `public`, pada 9 September 2026. Riwayat versi 23-25 yang
sebelumnya tertinggal diselaraskan setelah struktur kolom, constraint, dan index
diverifikasi. Pencatatan tersebut dan migrasi 26 di-commit dalam satu transaksi.
Database lain tetap perlu dimigrasi sebelum memakai backend versi ini.

Data lama sengaja memiliki `billing_anchor` dan `billing_cycle` NULL. Renewal
melewati subscription lama yang jatuh tempo dan melaporkan `legacy_skipped`.
Invoice lama tetap bisa dibaca dengan periode NULL.

Sebelum mengaktifkan renewal untuk data lama, rekonsiliasi invoice yang sudah
terbit (termasuk invoice trial), tanggal periode, dan tanggal berikutnya; kemudian
isi tanggal acuan dan siklus yang sesuai. Jangan mengisi semua tanggal secara
massal berdasarkan asumsi, karena dapat menagih ulang periode lama.

## Menjalankan dan menguji

Dari folder `Backend`, dengan database yang sudah dimigrasi:

```powershell
go run ./cmd/billing -limit 100
```

Untuk demonstrasi pada database uji, waktu dapat ditentukan:

```powershell
go run ./cmd/billing -at "2026-10-31T00:00:00Z" -limit 100
```

`-at` adalah batas pemrosesan dan tetap menulis invoice, bukan dry run. CLI
mengeluarkan JSON berisi `invoices_created`, `legacy_skipped`, dan `failures`.
Kegagalan koneksi atau renewal memberi exit code 1. CLI ini merupakan perintah
operator lintas tenant, bukan endpoint publik.

```powershell
go test ./...
$env:BILLING_INTEGRATION = "1"
go test ./internal/services/billing -count=1 -v
Remove-Item Env:BILLING_INTEGRATION
```

Integration test membaca konfigurasi `Backend/.env`, membuat schema unik
`billing_test_*`, menerapkan seluruh migrasi, lalu menghapus schema uji. Akun
PostgreSQL perlu izin CREATE SCHEMA. Tabel pada schema aplikasi tidak dimodifikasi.

Pengujian mencakup tanggal akhir bulan/tahun kabisat, diskon, batas nominal,
catch-up, replay, dua worker bersamaan, unique constraint, snapshot, trial,
pengecualian paused/cancelled/legacy, batas batch, rollback setelah gagal menulis
rincian, kupon kedaluwarsa, invoice nol rupiah, invoice awal, dan resume.

## Scheduler berkala

Worker terpisah dari HTTP API; jalankan dari folder `Backend`:

```powershell
go run ./cmd/worker
```

Atau gunakan `make worker`. Konfigurasi dibaca dari environment/`.env`:

| Variabel | Default | Arti |
|---|---|---|
| BILLING_INTERVAL | 1m | Waktu tunggu setelah satu batch selesai |
| BILLING_BATCH_TIMEOUT | 30s | Batas waktu setiap batch |
| BILLING_BATCH_LIMIT | 100 | Maksimal percobaan periode dalam satu batch (1?10000) |

Worker langsung menjalankan batch saat startup. Batch berikutnya dimulai setelah
interval tunggu; tidak ada batch tumpang tindih di satu worker. Timeout diteruskan
ke query invoice, rincian, dan pembaruan subscription. Batch yang gagal dicatat
dan dicoba lagi pada putaran berikutnya. Periode yang sudah commit tetap tersimpan;
transaksi yang belum selesai dibatalkan. Proteksi database tetap berlaku jika
beberapa instance worker dijalankan.

Log JSON memuat waktu cutoff, durasi, jumlah invoice, legacy yang dilewati, dan
jumlah kegagalan. Kegagalan per subscription dicatat dengan ID dan pesan error.
`Ctrl+C` atau SIGTERM membatalkan batch aktif dan menghentikan worker, lalu menutup
koneksi database. Konfigurasi tidak valid atau kegagalan koneksi awal menyebabkan
exit code 1. Gangguan database setelah startup dicoba lagi pada interval berikutnya.

Interval adalah polling dengan jeda setelah batch, bukan jadwal cron berdasarkan
jam kalender. Proses harus tetap berjalan; jika server mati, backlog dilanjutkan
ketika worker dimulai lagi. Worker tidak dipasang otomatis sebagai Windows Service
atau layanan deployment. Gunakan process manager pada tahap deployment.

Unit test scheduler menggunakan waktu simulasi Go: jeda setelah batch panjang,
kelanjutan setelah error/partial failure, deadline batch, dan shutdown saat menunggu
atau bekerja. Integration test renewal tetap memverifikasi transaksi PostgreSQL.

Tahap berikutnya: rekonsiliasi data lama yang akan dipakai untuk demo, lalu
integrasi Xendit sandbox dan webhook pembayaran.
