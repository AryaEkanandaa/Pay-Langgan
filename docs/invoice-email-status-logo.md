# Status pengiriman dan penyimpanan logo email

## Hasil pemeriksaan

Pekerjaan email Kelana Workspace untuk invoice ID 22 menyimpan `logo_url` dengan awalan `data:image/jpeg;base64,`. Gambar JPEG valid, berukuran 20.124 byte dan 440 × 454 piksel. Ini bukan URL HTTPS: tidak ada alamat gambar yang bisa diuji dengan HTTP 200. Gambar rusak berasal dari penyisipan data URI ke HTML email, bukan dari logo kosong atau endpoint yang memerlukan login.

Email baru tidak menyisipkan data URI, SVG, localhost, atau domain trycloudflare sebagai logo. Logo yang belum memiliki alamat permanen dihilangkan dari email; nama bisnis tetap ditampilkan. Preview menampilkan penjelasan jika logo belum siap. Email lama yang sudah diterima tidak dapat diubah.

## Yang sudah disiapkan

- Upload PNG/JPG dari Tampilan Invoice diekstrak ke penyimpanan file backend, maksimum 1 MB dan dimensi maksimum 4096 × 4096. Nama file memakai hash isi gambar.
- Endpoint publik `GET /public-assets/invoice-logos/:file` hanya melayani file PNG/JPG tersebut, tanpa login, dengan Content-Type gambar dan cache immutable.
- API dan worker memakai persiapan logo yang sama. Worker menyimpan URL dan dimensi yang benar-benar dipakai ke snapshot pekerjaan. Preview pekerjaan tersebut memakai snapshot, sehingga perubahan pengaturan bisnis tidak mengganti logo email yang sudah terkirim.
- URL PNG/JPG publik yang dimasukkan langsung di Tampilan Invoice juga didukung. URL harus permanen, HTTPS, tanpa login, dan tidak kedaluwarsa. Validasi kode tidak membuktikan kepemilikan domain atau ketersediaan gambar; akses anonim tetap perlu diperiksa.
- Tidak ada akun storage/CDN, bucket, atau domain permanen yang otomatis dibuat. Penyimpanan lokal sudah siap; publikasi permanen masih memerlukan hosting/domain Anda.

## Cara mengaktifkan URL permanen

Pilihan sederhana adalah backend dengan domain tetap dan disk persisten:

1. Siapkan hosting backend dan domain/subdomain tetap, misalnya `api.domain-anda.com`, dengan sertifikat HTTPS.
2. Mount direktori penyimpanan persisten ke backend. API dan worker email harus memakai direktori yang sama. Simpan cadangan direktori ini dan jangan hapus file yang sudah dirujuk email lama.
3. Isi `Backend/.env` pada server:

   ```dotenv
   INVOICE_ASSET_DIR=/data/paylanggan/invoice-logos
   PUBLIC_ASSET_URL=https://api.domain-anda.com/public-assets/invoice-logos
   ```

   Pada Windows development, direktori yang disiapkan adalah `./storage/invoice-logos`, relatif terhadap folder Backend. `PUBLIC_ASSET_URL` saat ini sengaja kosong karena Anda belum memiliki alamat permanen.

4. Arahkan domain melalui reverse proxy ke API dan izinkan GET publik untuk `/public-assets/invoice-logos/`. Jangan pasang halaman login atau proteksi hotlink pada jalur tersebut.
5. Restart API dan worker email setelah mengubah env. Frontend tidak memerlukan secret storage.
6. Buka Preview email untuk invoice baru/belum dikirim. Salin URL gambar dari HTML dan buka di jendela privat. Respons harus HTTP 200 dengan `Content-Type: image/png` atau `image/jpeg`.

Jika memakai CDN, sinkronkan isi `INVOICE_ASSET_DIR` ke storage publik dengan nama file yang sama dan isi `PUBLIC_ASSET_URL` dengan prefix CDN. Upload otomatis ke S3/R2 belum disambungkan; jangan mengarahkan env ke bucket kosong. Alternatif tanpa sinkronisasi: unggah PNG/JPG secara manual ke storage publik permanen, lalu masukkan URL lengkapnya ke field URL logo pada Tampilan Invoice.

Jangan mengisi base URL dengan trycloudflare. Tunnel tetap dapat dipakai untuk menguji endpoint publik, tetapi URL tersebut tidak ditanamkan sebagai logo email.

## Status tombol kirim email

- POST email menyiapkan satu pekerjaan per invoice; status backend `queued` menjadi “Dalam antrean”, `sending` menjadi “Mengirim…”, `sent` menjadi “Terkirim”, dan `failed` menjadi “Gagal, coba lagi”.
- Frontend memeriksa endpoint status yang sudah ada setiap sekitar satu detik tanpa permintaan yang tumpang tindih. Tidak ada timer yang memalsukan keberhasilan.
- Polling berhenti saat modal ditutup, komponen dilepas, atau setelah 90 detik. Timeout hanya menghentikan pemantauan UI, bukan membatalkan pekerjaan backend. Buka kembali detail untuk memeriksa status; jangan membuat invoice ganda.
- Tampilan sukses kembali ke tampilan awal setelah tiga detik, tetapi tombol tetap nonaktif untuk invoice yang sudah terkirim. Teks penerima tetap menampilkan hasil terakhir.
- Worker menggunakan retry yang sudah ada. Status dapat kembali ke antrean selama menunggu percobaan otomatis; gagal final menyediakan coba lagi. SMTP menerima pesan tidak menjamin pesan masuk inbox.
- Animasi mendukung reduced motion, label status untuk pembaca layar, dan ruang tombol tetap agar elemen sekitarnya tidak bergeser.

## Cara menguji di Gmail

1. Hubungkan URL gambar permanen terlebih dahulu dan pastikan pemeriksaan HTTP anonim di atas berhasil.
2. Gunakan invoice baru/belum dikirim dengan alamat email milik Anda. Preview email harus menampilkan logo dan URL yang sama dengan yang akan dikirim.
3. Jalankan API dan `go run ./cmd/email-worker` dari folder Backend, lalu klik Kirim email sekali. Amati antrean, pengiriman, dan sukses sesuai status backend. Tahap yang sangat cepat mungkin hanya terlihat sebentar.
4. Pastikan baris status berubah menjadi “Terkirim - alamat penerima”, lalu periksa inbox/spam Gmail dan logo.
5. Tutup modal ketika masih antrean: polling harus berhenti. Buka lagi untuk melanjutkan pemeriksaan. Gunakan kegagalan terkontrol pada lingkungan uji untuk memeriksa retry; jangan merusak kredensial SMTP produksi.

Invoice yang sudah sent tidak bisa dikirim ulang melalui tombol ini. Tidak ada email pelanggan tambahan yang dikirim otomatis selama pekerjaan perbaikan ini.

## File utama

| File | Perubahan |
| --- | --- |
| `Frontend/src/components/billing/SendEmailButton.tsx` | Tombol reusable, animasi, polling nyata, timeout dan pembatalan. |
| `Frontend/src/pages/dashboard/InvoicePage.tsx` | Integrasi tombol, status penerima, warning logo dan pembatalan fetch detail. |
| `Frontend/src/lib/api.ts` | AbortSignal dan tipe warning preview. |
| `Frontend/src/index.css` | Morph, dots, spinner, centang, dark mode dan reduced motion. |
| `Frontend/src/pages/dashboard/InvoiceCustomizationPage.tsx` | Upload PNG/JPG dan arahan URL permanen. |
| `Backend/internal/services/billing/email_assets.go` | Validasi gambar, penyimpanan atomik, URL permanen dan fallback. |
| `Backend/internal/services/billing/email_template.go` | Hanya logo HTTPS PNG/JPG, dimensi eksplisit dan display:block. |
| `Backend/internal/services/billing/email_queue.go` | Pembacaan snapshot untuk preview. |
| `Backend/internal/services/billing/email_worker.go` | Persiapan logo dan pembekuan snapshot sebelum pengiriman. |
| `Backend/internal/handlers/billing/email_handler.go` | Preview dan halaman invoice memakai resolver logo yang sama. |
| `Backend/internal/handlers/billing/invoice_asset_handler.go` | Akses publik terbatas untuk file gambar. |
| `Backend/internal/config/config.go`, `Backend/cmd/api/main.go` | Konfigurasi storage dan registrasi route gambar. |
| `Backend/.env.example`, `.gitignore` | Contoh env dan pengecualian file logo dari Git. |
| `email_assets_test.go`, `email_template_test.go`, `invoice_asset_handler_test.go` | Uji raster, fallback, URL, snapshot dan HTTP anonim. |

Tidak ada dependency atau migrasi tambahan untuk perubahan status tombol dan logo ini.
