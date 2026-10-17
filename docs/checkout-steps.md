# Alur checkout drop-in

Modal pembayaran menggunakan satu kolom, maksimal 560 px.

1. **Paket**: menampilkan nama paket, periode untuk pembelian baru, dan total dari backend. Untuk invoice yang sudah dibuat, rincian add-on dan diskon/kupon ditampilkan dari `invoice_items`.
2. **Pelanggan**: nama, email, dan nomor ponsel. Pembelian baru baru membuat order setelah tahap ini dikirim.
3. **Kartu**: formulir gateway dan verifikasi 3DS ditampilkan di area checkout yang sama.
4. **Selesai**: ditampilkan setelah backend memastikan invoice berstatus `paid`.

Tombol Kembali dan Lanjut mempertahankan order dan sesi yang sama. Kontak yang sudah digunakan untuk sesi gateway ditampilkan sebagai baca saja saat kembali. Navigasi kembali dinonaktifkan selama pembayaran/verifikasi diproses.

`paylanggan.js` membuka `/checkout/embed?site=...&plan=...` untuk pembelian baru. Endpoint publik `GET /api/v1/embed/sites/:site/plans/:plan` hanya membaca paket milik bisnis terkait dan tidak membuat pelanggan atau order. Sesudah order dibuat, tautan halaman penuh dan cache order widget diperbarui ke `/checkout/:token`.

## 3DS

- Xendit: `createActionContainerComponent()` didaftarkan sebelum submit. Area dibuat terlihat dan diberi tinggi minimum sebelum SDK dipanggil, agar SDK tidak menggunakan modal bawaan.
- Midtrans Core: callback `performAuthentication` menampilkan iframe bank di dalam tahap Kartu. Domain resmi `card.midtrans.com` serta domain API yang masih digunakan diterima oleh validasi URL.
- Tampilan isi verifikasi tetap berasal dari gateway/bank. Callback browser memicu pengecekan; callback tersebut tidak membuat invoice lunas.
- Sesi Midtrans Snap lama tetap menggunakan tampilan Snap untuk kompatibilitas.

Tahap paket menampilkan add-on dan kupon yang **sudah masuk tagihan**. Perubahan ini tidak menambahkan pemilihan add-on atau penerapan kode kupon baru pada checkout publik.

Referensi: [Xendit Components](https://github.com/xendit/xendit-components-web), [Midtrans 3DS](https://docs.midtrans.com/reference/card-feature-3d-secure-3ds), [domain 3DS Midtrans](https://docs.midtrans.com/reference/deployment).
