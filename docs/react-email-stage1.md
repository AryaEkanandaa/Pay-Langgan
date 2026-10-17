# Template email transaksional PayLanggan — tahap 1

Tahap ini berisi komponen bersama dan tiga template: `payment-success`, `invoice-issued`, dan `payment-failed`. Semua sumber lengkap ada di `Frontend/src/emails/`. Template lain dalam daftar kebutuhan belum dibuat.

## Menjalankan

Jalankan dari folder `Frontend` dengan Node.js 20 atau lebih baru:

```bash
npm install
npm run email:dev
npm run email:test
npm run email:serve
```

Preview React Email tersedia di `http://localhost:3001`. Masing-masing template memiliki `PreviewProps` dengan data contoh. Data ini hanya untuk preview dan tidak mengirim email.

`email:serve` menjalankan renderer privat di `http://127.0.0.1:3101`. Worker Go mengirim HTML hasil renderer tersebut melalui SMTP yang sudah ada. Di `Backend/.env`, isi:

```dotenv
EMAIL_RENDERER_URL=http://127.0.0.1:3101
EMAIL_DELIVERY_MODE=smtp
```

Jalankan `go run ./cmd/email-worker` dari folder `Backend` setelah renderer siap. Worker memeriksa antrean setiap 5 detik. Jika renderer berhenti, pengiriman diantre ulang dan tidak diam-diam memakai template lama. Mode `preview` menyimpan HTML di `EMAIL_PREVIEW_DIR` tanpa mengirim SMTP. Hanya alamat loopback yang diterima untuk `EMAIL_RENDERER_URL`.

Untuk menguji SMTP tanpa membuat pelanggan atau invoice, jalankan satu kali dari folder `Backend` setelah renderer aktif:

```powershell
$env:EMAIL_TEST_RECIPIENT='alamat-anda@example.com'
go run ./cmd/email-smoke
Remove-Item Env:EMAIL_TEST_RECIPIENT
```

Email tersebut diberi subjek `[Uji]`, berisi data contoh dan total Rp 0. Perintah melaporkan penerimaan oleh server SMTP; periksa kotak masuk penerima untuk memastikan pengantaran akhir. Setelah komputer atau layanan dimulai ulang, jalankan lagi renderer dan worker sebelum menguji pengiriman dari dashboard atau checkout.

## Data dan render

`renderEmail(type, props)` menerima props sesuai jenis template dan menghasilkan `{ subject, html, text }`. Contoh:

```ts
import { renderEmail } from './src/emails/render'
import type { InvoiceIssuedProps } from './src/emails/types'

const props: InvoiceIssuedProps = {
  brand: { name: 'Kelana Workspace', brandColor: '#3B4EFF' },
  customer: { name: 'Arya Ekananda', email: 'arya@example.com' },
  invoice: {
    number: 'INV-20261015-001', status: 'pending',
    issuedAt: '2026-10-15T00:00:00+08:00', dueDate: '2026-10-22T00:00:00+08:00',
    items: [{ name: 'Profesional Bulanan', quantity: 1, unitPrice: 100000, subtotal: 100000 }],
    subtotal: 100000, total: 100000, currency: 'IDR',
    url: 'https://contoh-domain-publik.example/invoices/INV-20261015-001',
  },
}

const email = await renderEmail('invoice-issued', props)
```

`invoice.url`, `receiptUrl`, `actionUrl`, `supportUrl`, dan `logoUrl` harus berupa HTTPS publik. Logo email hanya menerima PNG/JPG/JPEG. Jika URL logo tidak ada atau tidak valid, nama bisnis tetap tampil tanpa gambar. URL contoh `.example` tidak boleh dipakai dalam email produksi. Untuk logo, unggah PNG/JPG ke penyimpanan publik dengan domain stabil dan pastikan respons anonim HTTP 200 dengan `Content-Type: image/png` atau `image/jpeg`.

## Contoh pengiriman dan lampiran PDF

Contoh berikut adalah pola integrasi, belum dipasang ke backend Go PayLanggan. Pilih satu penyedia email sesuai arsitektur pengiriman yang akan dipakai. PDF harus dibuat dari invoice yang sama dan nama file aman.

### Nodemailer

```ts
const email = await renderEmail('payment-success', props)
await transporter.sendMail({
  from: 'PayLanggan <paylanggan@example.com>',
  to: props.customer.email,
  subject: email.subject,
  html: email.html,
  text: email.text,
  attachments: [{ filename: 'invoice.pdf', content: pdfBytes, contentType: 'application/pdf' }],
})
```

### Resend

```ts
const email = await renderEmail('payment-success', props)
await resend.emails.send({
  from: 'PayLanggan <paylanggan@example.com>',
  to: [props.customer.email],
  subject: email.subject,
  html: email.html,
  text: email.text,
  attachments: [{ filename: 'invoice.pdf', content: pdfBytes }],
})
```

### Amazon SES v2

```ts
const email = await renderEmail('payment-success', props)
await ses.send(new SendEmailCommand({
  FromEmailAddress: 'PayLanggan <paylanggan@example.com>',
  Destination: { ToAddresses: [props.customer.email] },
  Content: {
    Simple: {
      Subject: { Data: email.subject, Charset: 'UTF-8' },
      Body: {
        Html: { Data: email.html, Charset: 'UTF-8' },
        Text: { Data: email.text, Charset: 'UTF-8' },
      },
      Attachments: [{ FileName: 'invoice.pdf', RawContent: pdfBytes, ContentType: 'application/pdf' }],
    },
  },
}))
```

## Template tahap 1

| Jenis | Pemicu | Pola subject | CTA |
| --- | --- | --- | --- |
| `invoice-issued` | Invoice berbayar diterbitkan | Invoice Rp {total} - {nomor} - {bisnis} | Bayar sekarang |
| `payment-success` | Pembayaran terverifikasi sukses | Pembayaran berhasil Rp {total} - {nomor} - {bisnis} | Lihat invoice dan pembayaran |
| `payment-failed` | Pembayaran gagal | Pembayaran gagal Rp {total} - {nomor} - {bisnis} | Perbarui metode pembayaran |

Sebanyak 12 jenis lain masih menunggu tahap berikutnya. Invoice terbit dan pembayaran berhasil memakai antrean yang sudah ada. Email pembayaran gagal masuk antrean setelah kegagalan debit otomatis yang definitif; invoice tetap menunggu pembayaran. Kegagalan percobaan kartu checkout yang dapat langsung dicoba ulang belum memicu email gagal. Renderer Go tetap digunakan jika `EMAIL_RENDERER_URL` dikosongkan, sehingga instalasi yang belum menjalankan Node masih bisa mengirim template lama.

Logo email membutuhkan `PUBLIC_ASSET_URL` dari penyimpanan gambar PNG/JPG publik yang stabil. Bila belum dikonfigurasi, email tetap bisa dikirim dengan nama bisnis tanpa logo. URL tunnel `trycloudflare.com` tidak dipakai untuk logo karena sifatnya sementara.
