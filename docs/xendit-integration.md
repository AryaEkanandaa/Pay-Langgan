# Integrasi Xendit PayLanggan

## Keputusan ruang lingkup

Integrasi pertama menggunakan Xendit dan hanya menerima kartu kredit. Midtrans
belum dihubungkan ke alur pembayaran. Ada dua alur bisnis yang nantinya terpisah:

1. Bisnis membayar biaya platform PayLanggan melalui Xendit.
2. Customer bisnis membayar subscription melalui hosted page PayLanggan dan
   Payment Session Xendit.

Tahap yang sudah diterapkan adalah client Payment Session Xendit sebagai fondasi
alur kedua. Hosted page, penyimpanan sesi ke database, webhook, dan platform
billing belum dihubungkan pada tahap ini.

## Payment Session kartu

PayLanggan membuat sesi dari backend melalui `POST /sessions`. Konfigurasi yang
dipakai:

- `session_type`: `PAY`
- `mode`: `PAYMENT_LINK`
- `country`: `ID`
- `currency`: `IDR`
- `capture_method`: `AUTOMATIC`
- `allowed_payment_channels`: hanya `CARDS`
- `allow_save_payment_method`: `FORCED`

Customer akan diarahkan ke URL checkout yang dikembalikan Xendit. Xendit
menangani nomor kartu, CVV, dan autentikasi kartu. PayLanggan tidak menerima atau
menyimpan data tersebut. Token pembayaran baru boleh disimpan setelah webhook
Xendit yang sah mengonfirmasi sesi selesai.

Client memvalidasi batas kartu Indonesia sebesar Rp5.000 sampai Rp200.000.000,
email, nomor telepon E.164, dan return URL HTTPS sebelum mengirim permintaan.
Request HTTP memiliki timeout 15 detik dan response dibatasi 1 MiB. POST tidak
dicoba ulang otomatis karena hasil timeout dapat ambigu; rekonsiliasi sesi dan
idempotensi database ditambahkan saat hosted checkout dihubungkan.

## Konfigurasi

```env
XENDIT_SECRET_KEY=
XENDIT_CALLBACK_TOKEN=
XENDIT_BASE_URL=https://api.xendit.co
XENDIT_MOCK=true
```

`XENDIT_MOCK=true` menghasilkan sesi lokal semu tanpa menghubungi Xendit. Untuk
sandbox Xendit yang sesungguhnya, isi development secret key dan callback token,
siapkan return URL HTTPS, kemudian ubah mock menjadi false. Secret key dan
callback token tidak boleh dikirim ke frontend atau disimpan di Git.

Mode non-mock menolak pembuatan client jika secret key/callback token kosong atau
base URL tidak memakai HTTPS. Callback diverifikasi secara constant-time melalui
header `x-callback-token`; verifikasi payload dan idempotensi event akan dilakukan
oleh webhook service pada tahap berikutnya.

## Tahap berikutnya

1. Membuat model dan CRUD hosted page untuk Admin/Sales.
2. Membuat halaman publik berdasarkan token acak, tanpa login customer.
3. Menyimpan checkout dan `reference_id` unik sebelum meminta Payment Session.
4. Mengarahkan customer ke `payment_link_url` Xendit.
5. Memproses webhook secara idempoten untuk menyimpan `payment_token_id`, payment,
   dan status invoice.
6. Menggunakan token tersimpan untuk tagihan kartu periode selanjutnya.
7. Menambahkan platform billing bisnis ke PayLanggan melalui konfigurasi Xendit
   yang terpisah dari pembayaran customer bisnis.
