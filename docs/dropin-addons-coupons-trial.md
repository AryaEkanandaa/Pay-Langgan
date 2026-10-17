# Add-on, kupon, dan trial pada drop-in

## Menggunakan dashboard

1. Buat plan di **Katalog → Plan**. Isi **Masa Trial (hari)** jika ingin trial gratis; `0` berarti langsung berbayar.
2. Buat add-on di **Katalog → Add-on**. Produk dan periode add-on harus sama dengan plan yang dipilih.
3. Buat kode kupon di **Kupon**. Diskon nominal atau persentase, kedaluwarsa dan batas penggunaan tetap mengikuti konfigurasi yang ada.
4. Buka **Pengaturan → Integrasi Website**, pilih produk dan plan. Plan trial kini ikut tersedia.
5. Opsional: pilih add-on awal, jumlah dan kode kupon awal. Salin ulang kode tombol jika ingin mengganti pilihan awal; skrip header tetap sama selama URL PayLanggan dan ID integrasi tidak berubah.

Pelanggan dapat menyesuaikan add-on dan kode kupon pada langkah **Paket**. Total diperiksa backend sebelum langkah **Pelanggan** dapat dilanjutkan. Pembuatan order memeriksa ulang harga dan kuota kupon dalam transaksi database.

Contoh kode tombol:

```html
<button
  data-pl-checkout
  data-pl-plan="ID_PUBLIK_PLAN"
  data-pl-add-ons='[{"add_on_id":12,"quantity":2}]'
  data-pl-coupon="HEMAT10">
  Langganan
</button>
```

ID add-on pada contoh harus diganti dengan ID katalog bisnis sendiri. Generator dashboard mengisinya otomatis. Atribut add-on dan kupon boleh dihilangkan; pelanggan tetap dapat memilihnya di checkout. Harga tidak dipasang di tombol.

Skrip header masih memakai `src` ke `/paylanggan.js`, `data-pl-site` untuk ID integrasi publik dan `data-pl-api-url` untuk backend. Secret key tidak boleh dipasang di browser. Demo `examples/dropin-simple/index.html` tetap berupa HTML sederhana.

## Alur berbayar dan diskon

- Total invoice adalah harga plan + harga add-on × jumlah − diskon.
- Add-on yang ditawarkan terbatas pada produk dan periode plan yang sama, milik bisnis yang sama.
- Kupon yang tidak ada, kedaluwarsa, atau habis kuota ditolak. Menghapus/mengganti kode memeriksa ulang total.
- Batas total pembayaran kartu tetap Rp 5.000–Rp 200.000.000 dalam rupiah bulat. Total nol didukung: invoice langsung lunas, tanpa meminta kartu atau membuat transaksi gateway.
- Add-on disimpan pada langganan dan ikut invoice periode berikutnya. Kupon juga mengikuti perilaku langganan yang sudah ada: berlaku pada invoice berikutnya selama masih valid. Ini bukan kupon khusus satu invoice.
- Kuota kupon dihitung ketika kupon dipasang pada langganan, termasuk pendaftaran trial; retry order yang sama tidak menambah penggunaan lagi.
- Invoice yang sudah terbit tidak dihitung ulang dan tidak ada prorata baru pada perubahan ini.
- Permintaan API dengan `external_order_id` sama dan isi pembelian berbeda ditolak. Drop-in memakai `request_id` untuk retry selama formulir yang sama terbuka.

## Alur trial

1. Langkah **Paket** menampilkan durasi trial, Rp 0 hari ini, dan estimasi harga setelah trial bersama add-on dan kupon.
2. Pelanggan mengisi identitas dan menyetujui pendaftaran kartu. Backend membuat order dan langganan `pending`, tanpa invoice atau email konfirmasi trial.
3. Langkah **Kartu** menggunakan gateway aktif bisnis: Xendit membuat sesi `SAVE` bernilai Rp 0; Midtrans memakai `registerCard` di browser. Data kartu langsung dikirim ke gateway.
4. Xendit: backend membaca hasil sesi dari Xendit dan mengaktifkan trial hanya setelah sesi `SAVE` selesai serta token tersimpan. Midtrans: hasil `registerCard` dicatat sebagai **terdaftar**, karena pendaftaran ini tidak membuktikan OTP/3DS bank.
5. Setelah pendaftaran kartu selesai, trial dimulai dan tanggal akhirnya dihitung dari waktu tersebut. Email konfirmasi trial baru masuk antrean. Tidak ada invoice atau pemotongan biaya paket pada pendaftaran.
6. Jika sesi ditinggalkan hingga kedaluwarsa, langganan `pending` dibatalkan dan kuota kupon yang dipesan dilepas oleh worker. Checkout baru dapat dimulai lagi.
7. Worker penagihan membuat invoice pertama setelah trial berakhir. Penagihan otomatis dari token trial belum diaktifkan; pelanggan membayar invoice melalui alur pembayaran kartu yang tersedia.

Midtrans membutuhkan izin fitur Register Card pada akun bisnis. Pendebitan berkala Midtrans juga memerlukan dukungan recurring MID dari bank dan konfigurasi terpisah.

Satu pelanggan berdasarkan email dalam bisnis yang sama hanya mendapat satu trial untuk plan yang sama. Mengganti email dapat melewati batas ini; pencegahan penyalahgunaan identitas memerlukan verifikasi email/akun dan belum ditambahkan.

**Aktivasi setelah pembayaran (migrasi 44):** checkout berbayar membuat langganan `pending` (Menunggu pembayaran). Trial tetap `trial` sampai berakhir; invoice pertama setelah trial membuat status `pending` jika belum lunas. Pembayaran terverifikasi mengaktifkan langganan. Invoice Rp0 langsung lunas dan aktif. Langganan yang dibatalkan/dijeda tidak diaktifkan oleh pembayaran terlambat. Event browser bukan bukti pembayaran untuk otorisasi server.

## API dan event

Endpoint preview yang sudah ada menerima query tambahan:

```text
GET /api/v1/embed/sites/:site/plans/:plan?add_ons=JSON&coupon_code=KODE
```

Response mencakup `trial_days`, `due_now`, `amount` (harga setelah trial jika trial aktif), `items`, dan daftar `add_ons` yang kompatibel. Preview tidak membuat pelanggan, order, invoice atau mengurangi kuota kupon.

Endpoint order embed menerima `add_ons`, `coupon_code`, dan `request_id` opsional. API merchant melalui server menerima `add_ons` dan `coupon_code` bersama `external_order_id`. Plan publik tetap didukung oleh embed; tombol lama dengan ID plan angka tetap berfungsi.

Saat menunggu pendaftaran kartu, order trial mempunyai `invoice_id: 0`, `invoice_status: "not_issued"`, `subscription_status: "pending"`, `trial_ends_at: null`, dan `checkout_url` menuju langkah kartu. Setelah kartu berhasil didaftarkan, status berubah menjadi `trial` dan `trial_ends_at` terisi. Setelah invoice pertama diterbitkan, status order terhubung ke invoice tersebut.

SDK juga menerima:

```js
PayLanggan.openEmbed({
  siteId: 'ID_INTEGRASI',
  planId: 'ID_PUBLIK_PLAN',
  apiUrl: 'https://BACKEND',
  addOns: [{ add_on_id: 12, quantity: 2 }],
  couponCode: 'HEMAT10',
  onTrial: (result) => console.log(result.trialEndsAt),
})
```

Pendaftaran trial memancarkan event DOM `paylanggan:trial-started`; tidak memancarkan `paylanggan:paid`. Pembayaran atau invoice nol memakai event `paylanggan:paid`. Event browser bukan bukti pembayaran untuk otorisasi server.

## Proses yang diperlukan

Dari direktori `Backend`, terminal terpisah:

```powershell
go run ./cmd/api
go run ./cmd/worker
go run ./cmd/email-worker
```

Worker penagihan memeriksa jadwal setiap menit secara default (`BILLING_INTERVAL=1m`). Langganan lama tanpa `billing_anchor` tetap dilewati; jangan menganggap semuanya otomatis diperbaiki oleh migrasi ini.

Migrasi `2026-10-08-41_merchant_checkout_options.sql` sudah diterapkan pada database lokal. Migrasi hanya menambah hubungan subscription dan fingerprint request pada merchant order; tidak menghapus invoice/pembayaran.

Ketika tunnel berubah, perbarui `APP_PUBLIC_URL`, `API_PUBLIC_URL`, `XENDIT_RETURN_BASE_URL`, `VITE_API_URL`, `VITE_ALLOWED_HOST`, serta header skrip website bisnis. Muat ulang API/worker/frontend. Salin URL webhook terbaru dari halaman Payment Gateway dan daftarkan pada dashboard gateway di mode akun yang sesuai.

## Verifikasi

Pengujian integrasi merchant dan billing lolos: harga invoice sesuai preview, isolasi bisnis, periode add-on, kuota kupon serentak, idempotensi, total nol, trial tanpa invoice awal, penolakan trial berulang dan invoice pertama setelah trial.

Pengujian browser memakai API tiruan pada 360/768/1440px, terang/gelap: pilihan add-on, jumlah, kupon salah lalu benar, total, pendaftaran kartu trial, total nol dan order berbayar masuk endpoint sesi yang sudah ada. Tidak membuat transaksi gateway atau mengirim email uji nyata.

Suite integrasi payment yang lebih luas masih gagal pada pengujian autopay (`TestAutopay...`); fitur autopay tidak diubah dalam pekerjaan ini. Hasil pengujian merchant/billing tidak menyatakan seluruh suite autopay lolos.
# Visibilitas add-on di checkout

Pada **Katalog → Add-on**, pilih produk, nama, harga, dan siklus tagihan, lalu atur **Tampilkan di checkout** pada bagian **Tampilan pelanggan**.

- Aktif: pelanggan dapat memilih add-on pada langkah Paket, jika produk dan siklus tagihan cocok dengan plan.
- Nonaktif: add-on tidak ditampilkan dalam checkout maupun generator tombol integrasi. Backend menolak ID add-on tersebut untuk quote dan order baru, termasuk jika dikirim langsung melalui API integrasi.
- Add-on tetap tersedia untuk pengelolaan admin. Langganan, invoice, dan order yang sudah dibuat tidak diubah; retry order yang sudah ada tetap mengikuti idempotensi sebelumnya.
- Add-on lama tetap terlihat setelah migrasi 42. Field API katalog `visible_in_checkout` opsional saat create (default true); jika tidak dikirim saat update, nilai sebelumnya dipertahankan.
- Ini pengaturan checkout, bukan portal pelanggan. Nama produk ditentukan ketika membuat add-on dan tidak dapat dipindah melalui formulir edit.
