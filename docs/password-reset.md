# Lupa password PayLanggan

## Cara menggunakan

1. Buka halaman masuk, lalu pilih **Lupa Password?**.
2. Masukkan email yang terdaftar pada akun.
3. Periksa inbox dan spam untuk email **Atur ulang password PayLanggan**.
4. Buka tautan, isi password baru dan konfirmasinya, lalu masuk kembali.

Password minimal 8 karakter dan maksimal 72 byte. Tautan berlaku 15 menit sejak permintaan dibuat dan hanya dapat dipakai sekali. Reset yang berhasil membatalkan tautan reset lain dan sesi login lama akun tersebut. Password pengguna lain tidak berubah.

Alamat `admin@kelanaworkspace.local` tidak dapat menerima email internet. Untuk mencoba akun Kelana, alamat akun harus berupa email nyata yang Anda kuasai; implementasi ini tidak mengganti alamat akun secara otomatis.

## Konfigurasi dan proses

- Migrasi `2026-10-08-40_password_reset.sql` menambahkan antrean reset dan versi sesi pengguna. Sudah diterapkan pada database lokal tanpa mengubah password yang ada.
- Gunakan konfigurasi SMTP Gmail yang sudah ada: `EMAIL_DELIVERY_MODE=smtp`, host, port, pengirim, username, dan app password. Jangan memasukkan app password ke Git.
- `APP_PUBLIC_URL` adalah URL HTTPS frontend yang aktif. Ketika tunnel frontend berubah, perbarui nilai ini dan jalankan ulang API serta worker. Tautan email yang sudah dikirim tetap menunjuk domain lama.
- Pertahankan `JWT_SECRET` secara aman. Rahasia ini digunakan untuk JWT dan rekonstruksi token antrean; menggantinya akan memengaruhi sesi serta permintaan reset yang belum dikirim.
- Dari direktori `Backend`, jalankan API dengan `go run ./cmd/api` dan worker pada terminal lain dengan `go run ./cmd/email-worker`. Worker yang sama menangani invoice dan reset password.
- Mode `preview` menulis HTML ke `EMAIL_PREVIEW_DIR` tanpa pengiriman SMTP. File preview mengandung tautan rahasia; jangan dibagikan atau dimasukkan ke Git.

## Alur implementasi

`POST /api/v1/forgot-password` menerima email dan memberikan pesan generik yang sama untuk email terdaftar maupun tidak terdaftar. Permintaan yang valid masuk antrean database; worker mengirim email dan mencoba ulang kegagalan hingga tiga percobaan.

`POST /api/v1/reset-password` menerima token, password dan `password_confirmation`. Token dikirim melalui fragmen URL (`#token=...`) sehingga tidak masuk ke request HTTP halaman frontend. Backend menyimpan hash token, bukan token teks biasa; antrean menyimpan nonce acak untuk rekonstruksi menggunakan HMAC. Nonce dibersihkan setelah pengiriman selesai atau dibatalkan.

Ada pembatasan satu permintaan per menit dan lima per jam untuk akun yang sama, serta pembatasan berdasarkan IP koneksi langsung. Di belakang tunnel, beberapa pengunjung dapat berbagi IP proxy dan batas yang sama. Deployment produksi perlu menyesuaikan konfigurasi proxy tepercaya sebelum menggunakan IP pengunjung asli; jangan langsung mempercayai header IP dari internet.

## File utama

- `Frontend/src/pages/ForgotPassword.tsx`, `ResetPassword.tsx`: formulir permintaan dan reset, validasi, pembatalan request, serta hasil sukses/gagal.
- `Frontend/src/pages/Login.tsx`, `App.tsx`, `lib/api.ts`: tautan, route publik, dan pemanggilan API.
- `Backend/internal/handlers/identity/password_reset.go`: endpoint, batas request, dan pemeriksaan versi sesi.
- `Backend/internal/services/identity/password_reset.go`: token, validasi password, template email dan worker reset.
- `Backend/internal/repositories/identity/password_reset.go`: antrean persisten, konsumsi token atomik dan pembatalan sesi.
- `Backend/internal/services/mailer/smtp.go`: pengiriman SMTP bersama; pengiriman invoice memakai fungsi yang sama.
- `Backend/cmd/email-worker/main.go`: memproses kedua antrean.
- Model pengguna, JWT, middleware, auth repository/service dan routes: membawa serta memeriksa versi sesi setelah reset.

## Pengujian

Backend: `AUTH_INTEGRATION=1` mengaktifkan pengujian reset pada schema database terisolasi. Pengujian mencakup token kedaluwarsa, penggunaan ulang, dua reset bersamaan, retry worker, akun lain, serta penolakan JWT lama. Pengujian menggunakan preview, bukan email nyata.

Frontend diuji dengan API tiruan: tautan login, antrean request, token kosong, konfirmasi berbeda, token kedaluwarsa, sukses, penghapusan fragmen, lebar 360/768/1440px dan mode gelap. TypeScript dan build produksi juga diperiksa.

Untuk pengujian pengiriman nyata, gunakan akun terdaftar dengan email milik Anda, jalankan alur di atas, lalu pastikan password lama gagal dan password baru berhasil. Tidak ada email tambahan yang dikirim otomatis selama pengujian implementasi.
