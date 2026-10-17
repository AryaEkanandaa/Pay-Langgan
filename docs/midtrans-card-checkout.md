# Checkout kartu Midtrans

Checkout baru menggunakan Midtrans Core API setelah migrasi 36. Formulir dibuat
di PayLanggan agar susunan detail kartu, nama pemegang kartu, email, telepon,
dan tombol bayar mengikuti checkout Xendit. Checkout Snap yang sudah dibuat
sebelumnya tetap dapat diselesaikan melalui alur lamanya.

## Alur

1. Bisnis memilih Midtrans dan menyimpan Merchant ID, Client Key, serta Server Key
   melalui Pengaturan → Payment Gateway. Gunakan seluruh kredensial dari mode yang sama.
2. Website bisnis memakai drop-in script dan tombol plan yang sama seperti sebelumnya.
3. `POST /api/v1/checkout/:token/session` mengirim mode `CORE`, Client Key publik,
   dan environment. Persiapan formulir belum membuat charge.
4. Browser mengirim detail kartu langsung ke SDK resmi Midtrans
   `MidtransNew3ds.getCardToken`. Nomor kartu dan CVV tidak dikirim ke backend PayLanggan.
5. `POST /api/v1/checkout/:token/card-payment` menerima `token_id`, nama, email, dan
   telepon. Nominal dan order ID dibaca dari invoice; pengguna tidak dapat menetapkannya.
6. Backend membuat transaksi `credit_card` dengan `authentication=true` melalui
   `/v2/charge`, menggunakan Server Key milik bisnis tersebut.
7. Bila diperlukan, SDK membuka OTP/3DS dalam modal. Callback browser hanya
   memulai pengecekan, dan tidak melunasi invoice.
8. Invoice dilunasi setelah webhook bertanda tangan valid atau respons API Midtrans
   yang terautentikasi cocok dengan order, nominal, akun, dan metode pembayaran.

Nomor kartu dan CVV tidak disimpan di database atau browser storage.
Endpoint token menolak field kartu mentah. Order ID disimpan sebelum charge;
permintaan berulang diblokir. Worker memeriksa `/v2/:order_id/status` saat respons
charge terputus, sehingga tidak membuat charge baru untuk pesanan yang sama.

## Uji Sandbox

Setelah gateway bisnis disimpan sebagai Midtrans Sandbox, buat **pesanan baru**
di demo. Gunakan kartu dan OTP sesuai dokumentasi Midtrans. Periksa pembayaran
di dashboard Sandbox dan invoice di PayLanggan. URL webhook bisnis tetap sama.

Tes browser otomatis memakai SDK dan respons simulasi; pembayaran Sandbox nyata
tetap perlu diuji dengan akun bisnis yang terhubung.

```powershell
cd Backend
$env:PAYMENT_INTEGRATION='1'
go test ./internal/services/payment -run 'TestMidtrans|TestCheckoutPostgres' -count=1
go test ./pkg/payment ./internal/handlers/payment
Remove-Item Env:PAYMENT_INTEGRATION
```

Tes database memakai schema sementara yang terpisah dari data bisnis.

## Produksi

Konfirmasi aktivasi Core API/kartu dan persyaratan keamanan integrasi kartu dengan
Midtrans sebelum memakai akun produksi. Formulir OTP/3DS tetap mengikuti bank
atau Midtrans, walaupun formulir utama memakai desain PayLanggan.

Referensi resmi:

- [Integrasi kartu](https://docs.midtrans.com/docs/coreapi-card-payment-integration)
- [Core API dan aktivasi produksi](https://docs.midtrans.com/docs/custom-interface-core-api)
- [Pengujian Sandbox](https://docs.midtrans.com/docs/testing-payment-on-sandbox)
