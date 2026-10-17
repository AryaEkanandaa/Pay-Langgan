# Ringkasan dashboard

## Tampilan

- Strip KPI menampilkan kondisi saat ini dan tidak mengikuti filter tanggal.
- Invoice Jatuh Tempo menghitung invoice pending yang telah jatuh tempo. Tagihan Belum Dibayar menjumlahkan seluruh invoice pending, sehingga kedua statistik tetap terpisah.
- Filter Harian mencakup 30 hari terakhir; filter 3, 6, dan 12 bulan mencakup bulan kalender berjalan mundur. Pilihan disimpan dalam query URL `rentang`.
- Grafik menggunakan granularitas harian, mingguan, atau bulanan. Pendapatan per Tahun tetap memakai lima tahun kalender UTC sesuai perhitungan sebelumnya.
- InfoPopover menjelaskan definisi dan sumber perhitungan setiap statistik.
- Ekspor memakai periode pilihannya sendiri, termasuk rentang satu hari. Snapshot pada PDF ditandai per tanggal ekspor.

## API yang dipakai

Endpoint yang sudah ada tetap digunakan:

```text
GET /api/v1/dashboard/summary
GET /api/v1/dashboard/summary?from=2026-10-06&to=2026-10-06&granularity=day
```

Parameter baru bersifat opsional: `from`, `to` (tanggal ISO), dan `granularity` (`day`, `week`, `month`). Tanpa parameter, response lama tetap berlaku. Dengan parameter, response menyertakan `range_stats` berisi periode, zona waktu, deret statistik, dan pembayaran menurut paket. Tidak ada endpoint baru.

Tanggal akhir termasuk dalam rentang. Backend menerjemahkan batas hari menggunakan `business.meta.timezone`, dengan fallback Asia/Makassar, kemudian membandingkannya dengan timestamp UTC yang tersimpan. Rentang maksimum 24 bulan, tidak boleh terbalik atau berada di masa depan. Pembatasan bisnis dan akses Sales tetap diterapkan.

## Komponen

- `DashboardHome.tsx`: snapshot, filter URL, pemuatan data, grid statistik, serta pengabaian response usang.
- `SummaryMetricCard.tsx`: kartu, skeleton, empty/error state.
- `SummaryChart.tsx`: grafik Recharts dengan format angka Indonesia.
- `InfoPopover.tsx` dan `statInfo.ts`: popup informasi dan definisi statistik terpusat.
- `SummaryExportDialog.tsx` dan `dashboardPeriod.ts`: pilihan periode, validasi tanggal, dan unduh PDF memakai mekanisme ekspor yang sudah ada.

## Verifikasi

- TypeScript, lint pada komponen baru, dan build frontend berhasil.
- Uji Go memeriksa validasi tanggal, batas hari Asia/Makassar, batas 24 bulan, dan tahun kabisat.
- Uji PostgreSQL memakai schema sementara untuk memeriksa agregasi harian/mingguan/bulanan, batas inklusif/eksklusif, pemisahan data bisnis, serta akses Sales.
- Uji browser dengan API mock mencakup lebar 360, 768, 1366, 1440, 1920, 2560, dan 3440px dalam mode terang/gelap; error/retry, response usang, data kosong/besar/satu titik, keyboard, popup informasi, dan PDF satu hari.

## Yang belum tersedia

Perbandingan terhadap periode sebelumnya belum ditampilkan. Komentar TODO pada halaman mencatat kebutuhan agregat periode pembanding dari API. Tidak ada persentase atau tren buatan.

Tidak ada dependency atau migrasi database baru untuk perubahan Ringkasan ini.
