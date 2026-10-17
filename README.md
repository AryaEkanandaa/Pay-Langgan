# PayLanggan

**PayLanggan** adalah platform manajemen langganan dan penagihan untuk bisnis berbasis SaaS. Platform ini membantu bisnis mengelola katalog, pelanggan, siklus langganan, invoice, serta pembayaran dalam satu dashboard.

## Fitur utama

- **Manajemen katalog** — kelola produk, plan, add-on, dan kupon untuk penawaran berlangganan.
- **Siklus langganan** — catat status langganan, masa trial, periode tagihan, dan perubahan status pembayaran.
- **Pelanggan dan invoice** — kelola data pelanggan, terbitkan invoice, serta tampilkan rincian tagihan.
- **Pembayaran** — integrasikan gateway pembayaran Xendit atau Midtrans untuk memproses pembayaran pelanggan.
- **Notifikasi webhook** — terima dan catat pembaruan status pembayaran dari gateway.
- **Email transaksi** — siapkan komunikasi untuk invoice dan aktivitas langganan kepada pelanggan.
- **Dashboard dan laporan** — sajikan ringkasan bisnis dan informasi transaksi untuk membantu pemantauan.
- **Akses berbasis peran** — pisahkan akses Admin, Finance, Sales, dan Super Admin.

## Teknologi

| Bagian | Teknologi |
| --- | --- |
| Frontend | React, TypeScript, Vite |
| Backend | Go, Echo |
| Database | PostgreSQL |
| Autentikasi | JWT |
| Integrasi pembayaran | Xendit, Midtrans |
| Email transaksional | React Email dan SMTP |

## Struktur repositori

```text
PayLanggan/
├── Backend/       # API, layanan domain, worker, dan integrasi
├── Frontend/      # Dashboard bisnis dan alur checkout pelanggan
├── Migrations/    # Skema dan perubahan database
├── docs/          # Catatan desain dan dokumentasi fitur
└── examples/      # Contoh integrasi dan penggunaan
```

## Area produk

PayLanggan mencakup dua pengalaman utama:

1. **Dashboard bisnis** untuk mengelola pelanggan, katalog, kupon, langganan, invoice, gateway pembayaran, dan pengaturan bisnis.
2. **Checkout pelanggan** untuk memilih plan, memasukkan informasi pelanggan, dan menyelesaikan pembayaran atau memulai trial sesuai konfigurasi bisnis.

Backend menyediakan API untuk dashboard dan checkout, sementara worker menangani pekerjaan penagihan berkala dan pemrosesan terkait langganan.

## Keamanan dan isolasi data

Autentikasi dan otorisasi diterapkan berdasarkan peran. Data operasional bisnis dikaitkan dengan identitas bisnis agar akses setiap bisnis tetap terpisah. Kredensial gateway perlu ditangani sebagai data sensitif dan tidak boleh dimasukkan ke dalam repositori.

## Dokumentasi

Dokumentasi fitur dan catatan teknis tersedia di folder [`docs/`](docs/). Contoh integrasi tersedia di folder [`examples/`](examples/).
