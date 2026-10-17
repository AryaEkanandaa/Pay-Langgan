import type { BusinessDTO, InvoiceDetail } from './api'
import { formatCurrency } from './format'

function escapeHTML(value: string | number | null | undefined) {
  return String(value ?? '').replace(/[&<>"']/g, (character) => ({
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;',
  })[character]!)
}

function date(value: string | null) {
  return value ? new Date(value).toLocaleDateString('id-ID') : '—'
}

export function invoicePrintHTML(invoice: InvoiceDetail, business: BusinessDTO): string {
  const raw = business.meta?.invoice
  const settings = raw && typeof raw === 'object' && !Array.isArray(raw) ? raw as Record<string, unknown> : {}
  const text = (key: string) => typeof settings[key] === 'string' ? settings[key] as string : ''
  const color = /^#[\da-f]{6}$/i.test(text('accent_color')) ? text('accent_color') : '#3654ff'
  const logo = text('logo_url')
  const safeLogo = /^https:\/\/[^\s]+$/i.test(logo) || /^data:image\/(png|jpeg|webp);base64,[A-Za-z0-9+/=]+$/.test(logo)
  const money = (amount: number) => escapeHTML(formatCurrency(amount))
  const status = { pending: 'Menunggu Pembayaran', paid: 'Lunas', failed: 'Gagal', cancelled: 'Dibatalkan' }[invoice.status]
  const rows = invoice.items.length ? invoice.items.map((item) => `<tr><td>${escapeHTML(item.description)}</td><td class="quantity">${item.quantity}</td><td class="amount">${money(item.unit_price)}</td><td class="amount">${money(item.subtotal)}</td></tr>`).join('') : `<tr><td>${escapeHTML(invoice.plan_name)}</td><td class="quantity">1</td><td class="amount">${money(invoice.amount)}</td><td class="amount">${money(invoice.amount)}</td></tr>`

  return `<!doctype html><html lang="id"><head><meta charset="utf-8"><title>Invoice ${escapeHTML(invoice.invoice_number)}</title><style>
    @page { size: A4; margin: 16mm; }
    * { box-sizing: border-box; }
    body { margin: 0; color: #172033; background: white; font: 13px/1.5 Arial, sans-serif; }
    .header { display: flex; justify-content: space-between; gap: 24px; padding-bottom: 22px; border-bottom: 3px solid ${color}; }
    .logo { max-width: 150px; max-height: 65px; object-fit: contain; margin-bottom: 10px; }
    h1 { margin: 0; font-size: 23px; color: ${color}; }
    h2 { margin: 0; font-size: 18px; }
    p { margin: 4px 0; }
    .muted { color: #64748b; }
    .address, .note { white-space: pre-line; overflow-wrap: anywhere; }
    .identity { max-width: 55%; }
    .number { text-align: right; overflow-wrap: anywhere; }
    .status { display: inline-block; margin-top: 8px; font-weight: bold; }
    .details { display: flex; justify-content: space-between; gap: 24px; margin: 24px 0; }
    .dates { text-align: right; }
    table { width: 100%; border-collapse: collapse; margin-top: 24px; }
    th { background: #f1f5f9; text-align: left; font-size: 11px; text-transform: uppercase; }
    th, td { padding: 12px 8px; border-bottom: 1px solid #e2e8f0; }
    td:first-child { overflow-wrap: anywhere; }
    .quantity { text-align: center; }
    .amount { text-align: right; white-space: nowrap; }
    .total { text-align: right; margin: 20px 0 28px; font-size: 20px; font-weight: bold; color: ${color}; }
    .footer { margin-top: 30px; padding-top: 16px; border-top: 1px solid #e2e8f0; }
    thead { display: table-header-group; }
    tr, .header, .details, .total { break-inside: avoid; }
    @media print { body { print-color-adjust: exact; -webkit-print-color-adjust: exact; } }
  </style></head><body>
    <header class="header"><div class="identity">${safeLogo ? `<img class="logo" src="${escapeHTML(logo)}" alt="Logo bisnis">` : ''}<h2>${escapeHTML(business.name)}</h2><p class="address muted">${escapeHTML(text('address'))}</p></div><div class="number"><h1>INVOICE</h1><p><strong>${escapeHTML(invoice.invoice_number)}</strong></p><span class="status">${escapeHTML(status)}</span></div></header>
    <section class="details"><div><p class="muted">Ditagihkan kepada</p><h2>${escapeHTML(invoice.customer_name)}</h2><p>${escapeHTML(invoice.customer_email)}</p></div><div class="dates"><p>Terbit: ${escapeHTML(date(invoice.created_at))}</p><p>Jatuh tempo: ${escapeHTML(date(invoice.due_date))}</p>${invoice.paid_at ? `<p>Dibayar: ${escapeHTML(date(invoice.paid_at))}</p>` : ''}${invoice.payment ? `<p>Metode: ${escapeHTML(invoice.payment.provider)}</p>` : ''}</div></section>
    <table><thead><tr><th>Rincian</th><th class="quantity">Jumlah</th><th class="amount">Harga satuan</th><th class="amount">Subtotal</th></tr></thead><tbody>${rows}</tbody></table>
    <p class="total">Total ${money(invoice.amount)}</p><footer class="footer"><p class="note">${escapeHTML(text('footer_note'))}</p><p class="muted">Invoice ${escapeHTML(business.name)} melalui PayLanggan.</p></footer>
  </body></html>`
}

// Print a standalone document so the dashboard and modal never enter the print tree.
export async function printInvoiceDocument(invoice: InvoiceDetail, business: BusinessDTO): Promise<void> {
  const frame = document.createElement('iframe')
  frame.title = 'Dokumen cetak invoice'
  frame.setAttribute('aria-hidden', 'true')
  frame.style.cssText = 'position:fixed;left:-10000px;top:0;width:210mm;height:297mm;border:0;'
  const loaded = new Promise<void>((resolve, reject) => {
    const timer = window.setTimeout(() => reject(new Error('Dokumen cetak belum dapat dimuat.')), 15000)
    frame.onload = () => { window.clearTimeout(timer); resolve() }
  })
  frame.srcdoc = invoicePrintHTML(invoice, business)
  document.body.appendChild(frame)
  try {
    await loaded
    const printWindow = frame.contentWindow
    if (!printWindow) throw new Error('Browser belum dapat membuka cetak invoice.')
    // Images must finish loading before Chrome captures the print document.
    await Promise.all(Array.from(printWindow.document.images).map((image) => image.decode().catch(() => undefined)))
    await printWindow.document.fonts.ready
    printWindow.addEventListener('afterprint', () => frame.remove(), { once: true })
    printWindow.focus()
    printWindow.print()
  } catch (error) {
    frame.remove()
    throw error
  }
}
