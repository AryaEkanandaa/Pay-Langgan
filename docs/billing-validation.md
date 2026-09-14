# Validasi billing lokal ? 9 September 2026

## Migrasi

Target: PostgreSQL lokal, database `pay_langgan`, schema `public`.
Riwayat Goose awal berhenti pada versi 22. Struktur versi 23?25 sudah ada:
`subscriptions.created_at`, scope business pada kupon beserta index/FK,
dan constraint role/business pada users. Struktur tersebut diperiksa sebelum
versi 23?25 dicatat sebagai applied.

Pencatatan 23?25 dan penerapan SQL migrasi 26 dilakukan dalam satu transaksi,
dengan lock pada tabel riwayat, lock timeout 5 detik, dan statement timeout
30 detik. Transaksi berhasil di-commit. Tidak ada perubahan tagihan historis.
Jumlah data sebelum/sesudah: 10 subscription dan 10 invoice.
Sepuluh subscription lama tetap memiliki billing_anchor NULL.

## Hasil pengujian

Perintah dari folder Backend:

```powershell
$env:BILLING_INTEGRATION = "1"
go test ./internal/services/billing -run TestRenewalPostgres -count=1 -v
```

Hasil: PASS. Enam skenario PostgreSQL lulus:

- Catch-up, rincian biaya, kupon, dan pemrosesan ulang tanpa invoice tambahan.
- Dua worker bersamaan.
- Trial, paused/cancelled, legacy, dan batas batch.
- Rollback ketika penulisan rincian gagal, sementara subscription lain diproses.
- Kupon kedaluwarsa dan invoice nol rupiah.
- Pembuatan subscription/invoice awal serta resume.

Data pengujian dibuat pada schema sementara yang dihapus setelah tes selesai.

## Renewal pada schema aplikasi

```powershell
go run ./cmd/billing -limit 100
```

Hasil, exit code 0:

```json
{"invoices_created":0,"legacy_skipped":2,"failures":[]}
```

Dua dari sepuluh subscription lama memenuhi filter status dan jatuh tempo.
Keduanya dilewati karena belum direkonsiliasi. Hasil ini membuktikan CLI dapat
berjalan pada schema aplikasi yang telah dimigrasi; pembentukan invoice baru
sudah diverifikasi melalui schema uji, bukan dengan menagih data lama.

## Kelanjutan

Fondasi siap untuk implementasi scheduler berkala. Untuk demo pada schema
aplikasi, gunakan subscription baru atau rekonsiliasi periode subscription lama
sebelum mengaktifkan renewal untuk data tersebut. Scheduler dan Xendit belum
ditambahkan pada tahap validasi ini.
