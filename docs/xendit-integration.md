# Integrasi Xendit PayLanggan

## Keputusan ruang lingkup

Integrasi pertama menggunakan Xendit dan hanya menerima kartu kredit. Midtrans
belum dihubungkan ke alur pembayaran. Ada dua alur bisnis yang nantinya terpisah:

1. Bisnis membayar biaya platform PayLanggan melalui Xendit.
2. Customer bisnis membayar subscription melalui hosted page PayLanggan dan
   Payment Session Xendit.

Alur pelanggan memiliki halaman checkout publik, penyimpanan sesi, Payment Session,
dan webhook. Tagihan bisnis kepada PayLanggan memakai Payment Session terpisah
dan webhook platform; token kartu yang disimpan dipakai untuk perpanjangan otomatis.

## Pembayaran bisnis kepada PayLanggan

Pada **Pengaturan → Tagihan PayLanggan**, admin bisnis menyetujui biaya
Rp300.000/bulan dan debit otomatis. Backend membuat Payment Session Xendit dengan
`mode=COMPONENTS`, kanal `CARDS`, `allow_save_payment_method=FORCED`, dan
`card_on_file_type=RECURRING`. Frontend menampilkan formulir kartu Xendit langsung
di halaman PayLanggan; tidak membuka `payment_link_url` untuk sesi baru.

`POST /api/v1/platform/subscription/checkout` memulai sesi dan mengirim
`components_sdk_key` sekali kepada admin bisnis. `GET` pada alamat yang sama
melanjutkan sesi aktif setelah halaman dimuat ulang, tanpa menyimpan SDK key di
database. Webhook `payment_session.completed` atau rekonsiliasi
`go run ./cmd/platform-worker` mengubah invoice menjadi `paid` dan menyimpan token
untuk debit berikutnya. Tampilan berhasil saja tidak mengubah status invoice.

Untuk uji Xendit Test Mode, `APP_PUBLIC_URL` dan `XENDIT_RETURN_BASE_URL` harus
menunjuk origin HTTPS frontend yang sama, misalnya tunnel frontend yang sedang
aktif. Webhook platform harus menuju
`https://<tunnel-backend>/api/v1/webhooks/xendit/platform-payment-sessions`.
Sesi lama yang dibuat dengan `PAYMENT_LINK` tetap tersimpan sampai selesai atau
kedaluwarsa; setelah itu admin dapat memulai sesi Components baru.

## Payment Session kartu

PayLanggan membuat sesi dari backend melalui `POST /sessions`. Konfigurasi yang
dipakai:

- `session_type`: `PAY`
- `mode`: `COMPONENTS`
- `country`: `ID`
- `currency`: `IDR`
- `capture_method`: `AUTOMATIC`
- `allowed_payment_channels`: hanya `CARDS`
- `allow_save_payment_method`: `FORCED`

- `components_configuration.origins`: origin dari `APP_PUBLIC_URL` (scheme, host, port)
- `components_configuration.return_url`: `XENDIT_RETURN_BASE_URL/checkout/:token`

Backend mengembalikan `components_sdk_key` untuk inisialisasi Xendit Components
di halaman PayLanggan. Xendit menangani nomor kartu, CVV, dan autentikasi kartu. PayLanggan tidak menerima atau
menyimpan data tersebut. Token pembayaran baru boleh disimpan setelah webhook
Xendit yang sah atau rekonsiliasi API mengonfirmasi sesi selesai.

Client memvalidasi batas kartu Indonesia sebesar Rp5.000 sampai Rp200.000.000,
email, nomor telepon E.164, dan return URL HTTPS sebelum mengirim permintaan.
Request HTTP memiliki timeout 15 detik dan response dibatasi 1 MiB. POST tidak
dicoba ulang otomatis karena hasil timeout dapat ambigu. Checkout tetap `creating`
dan memblokir link baru sampai identitas/hasil sesi ditemukan. Setiap checkout memakai
`reference_id` unik dan webhook session dideduplikasi berdasarkan event serta ID
session Xendit.

## Hosted checkout dan webhook

Admin atau Finance membuat link melalui
`POST /api/v1/invoices/:id/checkout-links`. Token mentah hanya dikembalikan saat
link dibuat; database menyimpan hash SHA-256. Customer dapat membuka
`/checkout/:token` tanpa login dan memulai sesi melalui
`POST /api/v1/checkout/:token/session`.

Body tetap berupa `name`, `email`, dan `mobile`. Data respons sukses sekarang:

```json
{
  "payment_session_id": "ps-example",
  "mode": "COMPONENTS",
  "components_sdk_key": "session-specific-client-key",
  "status": "active",
  "expires_at": "2026-09-23T12:30:00Z"
}
```

Respons memakai `Cache-Control: no-store`. SDK key tidak disimpan di database
atau log. Pemanggilan ulang pada checkout aktif mengambil sesi yang sama melalui
`GET /sessions/:id`, tanpa membuat payment kedua. Masa berlaku lokal dibatasi
oleh masa berlaku sesi Xendit. Sesi tidak aktif, invoice yang sudah dibayar, dan
sesi hosted lama menghasilkan conflict, bukan SDK key. Sesi hosted lama perlu
diselesaikan atau menerima webhook expired sebelum link baru dibuat.

Frontend checkout kini menginisialisasi Xendit Components memakai
`components_sdk_key`, menampilkan komponen kartu, lalu memeriksa status invoice
melalui endpoint checkout publik. Setelah redirect autentikasi kartu, frontend
mengambil ulang sesi aktif dan mencoba melanjutkan SDK. Return URL dan event SDK
bukan bukti pembayaran; invoice baru ditampilkan lunas setelah backend menerima
webhook terautentikasi atau menyelesaikan rekonsiliasi dengan API Xendit.

Jika checkout aktif pertama kali dibuka di halaman penuh lalu dibuka lewat
drop-in, URL kembali pada sesi Xendit tetap memakai nilai saat sesi dibuat.
Halaman checkout menyimpan `parent_origin` untuk tab iframe tersebut di
`sessionStorage` berdasarkan token checkout. Setelah kembali dari redirect,
halaman dapat mengirim status lunas ke website induk meskipun URL kembali tidak
memuat `parent_origin`. Nilai ini hanya dipakai untuk notifikasi UI; status
pembayaran tetap berasal dari backend.

Percobaan kartu yang ditolak dapat menghasilkan event SDK `submission-end`
dengan detail kesalahan tanpa mengakhiri Payment Session. Frontend menampilkan
pesan agar pelanggan dapat mencoba lagi; invoice tetap pending. Endpoint webhook
di bawah ini menerima event siklus hidup Payment Session saja. Jika diperlukan
riwayat setiap percobaan gagal di backend, webhook Payments API seperti
`payment.failure` perlu dikonfigurasi dan ditangani pada endpoint terpisah,
bukan dikirim ke endpoint Payment Session.

Untuk memeriksa alur gagal di Xendit Test Mode, buat invoice khusus sandbox
dengan nominal Rp10.051 dan gunakan kartu uji yang tidak memerlukan challenge.
Menurut [daftar simulasi kartu Xendit](https://docs.xendit.co/docs/cards-simulate-card-scenarios),
nominal tersebut mensimulasikan `EXPIRED_CARD` di luar Vietnam. Pastikan halaman
menampilkan percobaan belum berhasil, invoice tetap pending, dan checkout masih
dapat digunakan untuk percobaan berikutnya selama sesi Xendit aktif. Jangan
gunakan invoice operasional untuk pengujian ini.

Xendit mengirim event `payment_session.completed` atau
`payment_session.expired` ke
`POST /api/v1/webhooks/xendit/payment-sessions`. Backend memverifikasi header
`x-callback-token`, session ID, reference ID, mata uang, nominal, dan status.
Webhook yang lolos autentikasi dan pemeriksaan struktur disimpan dahulu di
`webhook_events`, terpisah dari transaksi pembaruan pembayaran. Payload dinormalisasi
ke field yang dibutuhkan; SDK key dan field tambahan tidak ikut disimpan.
Setelah tersimpan, endpoint menjawab sukses meskipun pemrosesan bisnis perlu retry.
Kegagalan penyimpanan inbox tetap menghasilkan respons error agar Xendit mengirim ulang.

Event sukses memperbarui payment, invoice, token, dan `processed=true` dalam satu
transaksi. Jika gagal, efek bisnis di-rollback tetapi inbox beserta jumlah percobaan
dan kode kesalahan tetap tersimpan. Event duplikat tidak membuat payment kedua.
Event expired yang terlambat tidak menurunkan status pembayaran sukses.

## Konfigurasi

```env
XENDIT_SECRET_KEY=
XENDIT_CALLBACK_TOKEN=
XENDIT_BASE_URL=https://api.xendit.co
XENDIT_MOCK=true
XENDIT_RETURN_BASE_URL=https://paylanggan.example
APP_PUBLIC_URL=http://localhost:5173
CHECKOUT_TTL=24h
```

`XENDIT_MOCK=true` menghasilkan sesi dan SDK key lokal semu tanpa menghubungi
Xendit; key ini tidak dapat dipakai di SDK Xendit sesungguhnya. Untuk
sandbox Xendit yang sesungguhnya, isi development secret key dan callback token,
siapkan return URL HTTPS, kemudian ubah mock menjadi false. Secret key dan
callback token tidak boleh dikirim ke frontend atau disimpan di Git.

Mode non-mock menolak pembuatan client jika secret key/callback token kosong atau
base URL tidak memakai HTTPS. Callback diverifikasi secara constant-time melalui
header `x-callback-token`. Return URL Xendit wajib HTTPS, termasuk ketika link
aplikasi lokal masih menggunakan HTTP.

Set `APP_PUBLIC_URL` ke alamat frontend yang benar (termasuk port). Origin
mengizinkan HTTPS dan HTTP localhost/loopback untuk pengembangan, tanpa wildcard.
Kolom legacy `payment_link_url` dibiarkan untuk sesi lama dan bernilai NULL untuk
sesi Components. Fitur recovery membutuhkan migrasi
`2026-09-23-28_payment_recovery.sql` (setelah migrasi checkout nomor 27).

Referensi: [Xendit Components](https://docs.xendit.co/docs/components-overview),
[Create session](https://docs.xendit.co/apidocs/create-session),
[Get session](https://docs.xendit.co/apidocs/get-session).

## Tahap berikutnya

Integrasikan SDK Xendit Components pada frontend sebelum menguji pembayaran end-to-end.

1. Menguji token kartu dan recurring charge dengan akun sandbox Xendit.
2. Menambahkan status pembayaran langsung pada UI setelah webhook diterima.
3. Menambahkan platform billing bisnis ke PayLanggan melalui konfigurasi Xendit
   yang terpisah dari pembayaran customer bisnis.

## Worker rekonsiliasi dan retry webhook

Jalankan migrasi terlebih dahulu (`make migrate-up` dari direktori `Backend`),
kemudian jalankan worker sebagai proses terpisah dari API dan worker billing:

```sh
go run ./cmd/payments
# Satu batch, cocok untuk cron atau pemeriksaan manual:
go run ./cmd/payments -once -limit 50 -timeout 2m
```

Default: jeda 1 menit setelah batch selesai, timeout batch 2 menit, maksimal
50 webhook dan 50 sesi per batch. Flag `-interval`, `-timeout`, dan `-limit`
dapat disesuaikan. Worker berhenti melalui SIGINT/SIGTERM. Log hanya berisi
jumlah hasil, bukan payload pembayaran atau SDK key.

- Sesi `active` dan `creating` yang sudah mempunyai provider ID diperiksa lewat
  `GET /sessions/:id`. Reference, ID, mata uang, dan nominal harus cocok.
- Hasil `COMPLETED` memperbarui invoice/payment/token menggunakan aturan transaksi
  yang sama dengan webhook. Hasil `EXPIRED` melepaskan blokir untuk checkout baru.
- Kedaluwarsa lokal saja, GET 404, dan kegagalan jaringan **bukan** bukti bahwa
  pembayaran gagal. Sesi tetap diblokir sambil menunggu hasil provider.
- Link `created`/`failed` yang belum punya sesi aktif boleh dikedaluwarsakan lokal.
- Worker menggunakan reservasi dua menit untuk rekonsiliasi dan row lock untuk
  inbox. Proses yang berhenti di tengah jalan dapat dilanjutkan worker berikutnya.
- Retry webhook menggunakan backoff 30 detik, 1 menit, 2 menit, dan seterusnya,
  maksimum 1 jam. Setelah 8 kegagalan, event masuk dead letter. Payload yang tidak
  cocok dengan checkout langsung memerlukan pemeriksaan.

`checkout_sessions.last_error`, `review_required`, `reconcile_attempts`, dan
`next_reconcile_at` menyimpan status recovery. Pada inbox, gunakan `attempts`,
`last_error`, `dead_letter_at`, dan `next_attempt_at`. Kode kesalahan sengaja
ringkas agar respons provider tidak membocorkan data sensitif ke log/database.

## Pemulihan timeout dan operasi manual

Kesalahan validasi lokal atau penolakan HTTP yang definitif membuat checkout
`failed`; admin dapat membuat link baru. Timeout, koneksi terputus, respons rusak,
HTTP 408/409/429/5xx ditandai tidak pasti dan tetap `creating`.

Jika ID sudah diterima, backend menyimpannya sebelum aktivasi payment sehingga
worker dapat mengambil alih ketika transaksi aktivasi gagal. Jika respons hilang
seluruhnya, webhook terminal dapat mencocokkan `reference_id` dan menyambungkan
sesi secara otomatis, termasuk ketika webhook datang sebelum respons POST.

Dokumentasi API yang diperiksa menyediakan GET berdasarkan ID sesi, tetapi tidak
mendokumentasikan lookup sesi berdasarkan reference ataupun jaminan idempotency
untuk retry POST. Karena itu worker tidak mengirim ulang POST atau mengarang ID.
Jika tidak ada webhook, cari ID sesi di Xendit lalu jalankan:

```sh
go run ./cmd/payments -recover-checkout 123 -session-id ps-ID_DARI_XENDIT
# Rekonsiliasi manual jika ID sudah tersimpan:
go run ./cmd/payments -recover-checkout 123
# Setelah akar masalah webhook diperbaiki:
go run ./cmd/payments -retry-event 456
```

Perintah ini untuk operator dengan akses backend/database, bukan endpoint publik.
ID dari operator diverifikasi melalui API Xendit sebelum dipasang ke checkout.
Event yang sudah diproses tidak dapat diulang; retry manual tidak melewati validasi.
Sesi dengan hasil yang belum didukung (misalnya `CANCELED`) ditandai untuk diperiksa.

Mode mock menyimpan metadata sesi hanya dalam proses pembuatnya. Worker terpisah
tidak dapat mengambil sesi mock dari proses API; gunakan development credentials
dan `XENDIT_MOCK=false` untuk pengujian worker dengan sandbox Xendit.

Tes integrasi terisolasi (PowerShell, dari `Backend`):

```powershell
$env:PAYMENT_INTEGRATION='1'
go test ./internal/services/payment -v
```

Tes membuat schema PostgreSQL sementara dan membersihkannya setelah selesai;
tidak menjalankan migrasi pada schema aplikasi.

## Rate limit checkout

Migrasi `2026-09-23-29_payment_controls.sql` menambahkan penyimpanan rate limit
bersama di PostgreSQL, sehingga batas tidak direset dengan berpindah instance API.

| Endpoint | Batas per IP per menit | Batas per token checkout per menit |
| --- | ---: | ---: |
| `GET /api/v1/checkout/:token` | 60 | 30 |
| `POST /api/v1/checkout/:token/session` | 10 | 5 |

Kelebihan batas menghasilkan HTTP 429 dan `Retry-After: 60`. Kegagalan penyimpanan
limiter menghasilkan HTTP 503. Bucket menyimpan hash identifier, tidak menyimpan
IP atau token checkout mentah. Worker membersihkan bucket yang berusia >1 jam.
Webhook tidak menggunakan limiter checkout agar pengiriman ulang provider tidak
terhambat. Batas ini memakai jendela satu menit sejak bucket dimulai.

IP diambil dari koneksi langsung (`RemoteAddr`), bukan header `X-Forwarded-For`
yang dapat dipalsukan. Jika API berada di belakang reverse proxy, limit backend
akan terbagi untuk alamat proxy; terapkan limit per pengunjung pada proxy juga.

## Riwayat pembayaran Admin/Finance

Semua endpoint berikut membutuhkan Bearer JWT dengan role `admin` atau `finance`.
Data dibatasi oleh business ID dari JWT, bukan business ID yang dikirim pengguna.

- `GET /api/v1/payments?page=1&limit=20`
- Filter opsional: `invoice_id`, `subscription_id`, `status`.
- `GET /api/v1/customers/:id/payment-methods`
- `GET /api/v1/subscriptions/:id/autopay`
- `PUT /api/v1/subscriptions/:id/autopay`

Riwayat menampilkan checkout (termasuk pembuatan sesi yang gagal), pembayaran
manual, dan percobaan auto-debit tanpa menggandakan payment yang terhubung ke
checkout. Field mencakup invoice, subscription, nominal, `status`, `session_status`,
`failure_code`, `review_required`, `attempted_at`, dan `paid_at`.
`autopay_charge_id` dapat dipakai operator untuk pemulihan charge. Pagination
maksimal 100 baris, urutan terbaru lebih dulu. Status filter yang didukung:
`pending`, `success`, `failed`, `expired`, `created`, `creating`, `active`, `completed`.

Daftar metode pembayaran hanya berisi ID lokal, provider, is_default, dan waktu
penyimpanan. Token kartu, SDK key, callback token, dan payload provider tidak
dikirim melalui endpoint manajemen ini.

## Pembayaran subscription otomatis

Auto-debit menggunakan token tersimpan melalui
[`POST /v3/payment_requests`](https://docs.xendit.co/apidocs/create-payment-request),
dengan `type=PAY`, `card_on_file_type=RECURRING`, `transaction_sequence=SUBSEQUENT`,
dan konfigurasi frekuensi/expiry. Lihat
[panduan recurring Xendit](https://docs.xendit.co/docs/cards-recurring-and-merchant-initiated-transactions)
dan [subsequent transactions](https://docs.xendit.co/docs/subsequent-merchant-initiated-transaction).
Token tersimpan tidak otomatis berarti pelanggan telah menyetujui auto-debit.
Kelayakan kartu/token dan izin recurring merchant perlu diverifikasi di sandbox.

Ada dua syarat aktivasi yang terpisah:

1. `AUTOPAY_ENABLED=true` pada proses worker pembayaran. Default `false`.
2. Persetujuan pelanggan dicatat pada subscription oleh Admin/Finance.

Contoh body `PUT /api/v1/subscriptions/123/autopay`:

```json
{
  "enabled": true,
  "payment_method_id": 456,
  "consent_reference": "signed-consent-document-789",
  "consented_at": "2026-09-23T02:00:00Z",
  "consent_expires_at": "2027-09-23T02:00:00Z",
  "max_amount": 200000
}
```

`consent_reference` harus merujuk bukti persetujuan pelanggan yang benar-benar
dikumpulkan; endpoint ini mencatat bukti, bukan menggantikan proses memperoleh
persetujuan. `payment_method_id` harus milik pelanggan subscription dan provider
Xendit. Subscription harus aktif dan mempunyai jadwal monthly/yearly. Waktu
persetujuan tidak boleh di masa depan, expiry harus di masa depan, dan batas
nominal harus dalam rentang IDR 5.000–200.000.000. Setiap perubahan dicatat di
audit log beserta pengguna yang melakukan perubahan.

Untuk menghentikan debit baru, kirim `{"enabled": false}` ke endpoint yang sama.
Pembayaran yang sudah mulai dikirim tetap direkonsiliasi sampai diketahui hasilnya.
Pause/cancel subscription juga menghentikan pemilihan tagihan baru oleh worker.

Alur operasional:

1. Worker billing (`go run ./cmd/worker`) membuat invoice sesuai jadwal yang ada.
2. Worker pembayaran (`go run ./cmd/payments`) memilih invoice pending yang jatuh
   tempo, periodenya dimulai setelah pengaktifan/pembaruan consent, dan tidak
   memiliki checkout aktif atau charge auto-debit sebelumnya.
3. Worker mencatat payment/attempt dan mengunci klaim sebelum menghubungi Xendit.
   Nominal di luar batas persetujuan dicatat gagal tanpa mengirim debit.
4. Hasil `SUCCEEDED` menandai invoice paid dan mencatat payment ID. Hasil belum
   final diperiksa ulang melalui `GET /v3/payment_requests/:id`.

Kebijakan awal adalah **satu percobaan debit per invoice**. Penolakan final
(`FAILED`, `EXPIRED`, `CANCELED`) menyimpan kode kegagalan dan membiarkan invoice
pending agar dapat dibayar manual. Tidak ada retry debit otomatis setelah penolakan.
Timeout atau respons tidak pasti menjadi `review_required` dan memblokir checkout
serta mark-paid manual sampai hasilnya dipastikan. `REQUIRES_ACTION` memerlukan
tindak lanjut pelanggan; backend menyimpan status ini dan terus memeriksa request,
tetapi belum menyediakan UI untuk menyelesaikan tindakan tersebut.

Sesi yang belum pasti tidak dikirim ulang. Jika respons hilang sebelum request ID
tersimpan, operator dapat mencarinya di Xendit dan menjalankan:

```sh
go run ./cmd/payments -recover-autopay 123 -request-id pr-ID_DARI_XENDIT
```

ID, reference, nominal, dan mata uang diverifikasi sebelum status lokal diperbarui.
Status charge ada di `autopay_charges`, percobaan finansial di `payment_attempts`,
dan konfigurasi/persetujuan aktif di `subscription_autopay`.

Auto-debit saat ini direkonsiliasi melalui polling worker. Endpoint webhook
`/webhooks/xendit/payment-sessions` tetap khusus Payment Sessions, bukan webhook
Payments API v3. Jangan mengarahkan event v3 ke endpoint tersebut.

Pada `XENDIT_MOCK=true`, charge baru langsung disimulasikan sukses dan dapat
mengubah invoice lokal menjadi paid bila auto-debit diaktifkan. Gunakan database
uji terpisah; jangan mengaktifkan auto-debit mock pada data operasional.
