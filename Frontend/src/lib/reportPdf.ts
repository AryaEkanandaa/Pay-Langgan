import type { DashboardSummary, PlatformDashboardSummary } from './api'
import { formatCurrency } from './format'
import { periodLabel, type DatePeriod } from './dashboardPeriod'

type PdfLib = typeof import('pdf-lib')
type PdfDocument = import('pdf-lib').PDFDocument
type PdfPage = import('pdf-lib').PDFPage
type PdfFont = import('pdf-lib').PDFFont

const PAGE_WIDTH = 595.28
const PAGE_HEIGHT = 841.89
const LEFT = 48
const RIGHT = PAGE_WIDTH - LEFT

function safeText(value: string) {
  // Standard PDF fonts use WinAnsi; normalize names before drawing them.
  return value.normalize('NFKD').replace(/[\u0300-\u036f]/g, '').replace(/[^\x20-\x7E]/g, '?')
}

class ReportWriter {
  private page: PdfPage
  private readonly pdf: PdfDocument
  private readonly regular: PdfFont
  private readonly bold: PdfFont
  private y = PAGE_HEIGHT - 48
  private pageNumber = 1
  private readonly ink
  private readonly muted
  private readonly brand
  private readonly light

  constructor(pdf: PdfDocument, regular: PdfFont, bold: PdfFont, lib: PdfLib) {
    this.pdf = pdf
    this.regular = regular
    this.bold = bold
    this.ink = lib.rgb(0.12, 0.16, 0.25)
    this.muted = lib.rgb(0.43, 0.49, 0.59)
    this.brand = lib.rgb(0.21, 0.34, 0.96)
    this.light = lib.rgb(0.91, 0.93, 0.97)
    this.page = pdf.addPage([PAGE_WIDTH, PAGE_HEIGHT])
    this.footer()
  }

  private footer() {
    this.page.drawLine({ start: { x: LEFT, y: 43 }, end: { x: RIGHT, y: 43 }, thickness: 0.7, color: this.light })
    this.page.drawText('PayLanggan - laporan statistik', { x: LEFT, y: 28, size: 8, font: this.regular, color: this.muted })
    this.page.drawText(String(this.pageNumber), { x: RIGHT - 8, y: 28, size: 8, font: this.regular, color: this.muted })
  }

  private space(height: number) {
    if (this.y - height >= 65) return
    this.page = this.pdf.addPage([PAGE_WIDTH, PAGE_HEIGHT])
    this.pageNumber += 1
    this.y = PAGE_HEIGHT - 48
    this.footer()
  }

  title(title: string, subtitle: string) {
    this.page.drawRectangle({ x: LEFT, y: this.y + 7, width: 6, height: 27, color: this.brand })
    this.page.drawText(safeText(title), { x: LEFT + 18, y: this.y + 15, size: 20, font: this.bold, color: this.ink })
    this.y -= 13
    this.text(subtitle, 9, true)
    this.y -= 12
  }

  section(title: string) {
    this.space(38)
    this.y -= 13
    this.page.drawText(safeText(title), { x: LEFT, y: this.y, size: 12, font: this.bold, color: this.brand })
    this.y -= 10
    this.page.drawLine({ start: { x: LEFT, y: this.y }, end: { x: RIGHT, y: this.y }, thickness: 0.7, color: this.light })
    this.y -= 15
  }

  row(label: string, value: string) {
    this.space(20)
    let displayLabel = safeText(label)
    while (displayLabel.length > 3 && this.regular.widthOfTextAtSize(displayLabel, 10) > 240) {
      displayLabel = displayLabel.slice(0, -1)
    }
    if (displayLabel !== safeText(label)) displayLabel = displayLabel.slice(0, -3) + '...'
    this.page.drawText(displayLabel, { x: LEFT, y: this.y, size: 10, font: this.regular, color: this.ink })
    const display = safeText(value)
    const width = this.bold.widthOfTextAtSize(display, 10)
    this.page.drawText(display, { x: Math.max(LEFT + 220, RIGHT - width), y: this.y, size: 10, font: this.bold, color: this.ink })
    this.y -= 20
  }

  text(value: string, size = 9, muted = false) {
    const maxWidth = RIGHT - LEFT
    const words = safeText(value).split(/\s+/)
    let line = ''
    for (const word of words) {
      const candidate = line ? `${line} ${word}` : word
      if (line && this.regular.widthOfTextAtSize(candidate, size) > maxWidth) {
        this.space(size + 6)
        this.page.drawText(line, { x: LEFT, y: this.y, size, font: this.regular, color: muted ? this.muted : this.ink })
        this.y -= size + 6
        line = word
      } else line = candidate
    }
    if (line) {
      this.space(size + 6)
      this.page.drawText(line, { x: LEFT, y: this.y, size, font: this.regular, color: muted ? this.muted : this.ink })
      this.y -= size + 6
    }
  }
}

async function createWriter() {
  const lib = await import('pdf-lib')
  const pdf = await lib.PDFDocument.create()
  pdf.setAuthor('PayLanggan')
  pdf.setCreator('PayLanggan Dashboard')
  const regular = await pdf.embedFont(lib.StandardFonts.Helvetica)
  const bold = await pdf.embedFont(lib.StandardFonts.HelveticaBold)
  return { pdf, writer: new ReportWriter(pdf, regular, bold, lib) }
}

async function download(pdf: PdfDocument, filename: string) {
  const bytes = await pdf.save()
  const blob = new Blob([new Uint8Array(bytes).buffer], { type: 'application/pdf' })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 60_000)
}

function reportDate() {
  return new Date().toLocaleString('id-ID', { dateStyle: 'long', timeStyle: 'short', timeZone: 'Asia/Makassar' }) + ' WITA'
}

function fileDate() {
  return new Date().toISOString().slice(0, 10)
}

function monthLabel(index: number, total: number) {
  const now = new Date()
  const month = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() - (total - 1 - index), 1))
  return month.toLocaleDateString('id-ID', { month: 'long', year: 'numeric', timeZone: 'UTC' })
}

export async function exportBusinessReportPdf(summary: DashboardSummary, businessID: string, salesOnly: boolean, period?: DatePeriod & { timezone: string }) {
  if (period) {
    const range = summary.range_stats
    if (!range) throw new Error('Statistik rentang belum tersedia.')
    const { pdf, writer } = await createWriter()
    const total = (key: 'revenue' | 'new_subscriptions' | 'cancelled_subscriptions' | 'payments_succeeded' | 'payments_failed' | 'autopay_succeeded' | 'autopay_failed') => range.series.reduce((sum, point) => sum + point[key], 0)
    pdf.setTitle('Ringkasan PayLanggan')
    const createdAt = new Date().toLocaleString('id-ID', { dateStyle: 'long', timeStyle: 'short', timeZone: range.timezone })
    writer.title('Ringkasan PayLanggan', `ID bisnis: ${businessID} | Dibuat: ${createdAt} (${range.timezone})`)
    writer.text(`Periode: ${periodLabel(period).replace(' – ', ' - ')}. Zona waktu: ${range.timezone}. Tanggal akhir termasuk dalam rentang.`, 9, true)
    writer.section('Kondisi saat ini - per tanggal ekspor')
    writer.row('Pelanggan aktif', String(summary.active_customers))
    writer.row('Langganan berjalan', String(summary.active_subscriptions))
    if (!salesOnly) {
      writer.row('MRR', formatCurrency(summary.mrr))
      writer.row('Invoice jatuh tempo', String(summary.due_invoices))
      writer.row('Tagihan belum dibayar', formatCurrency(summary.outstanding_amount))
      writer.row('Di antaranya lewat jatuh tempo', formatCurrency(summary.overdue_amount))
    }
    writer.section('Statistik pada periode terpilih')
    writer.row('Langganan baru', String(total('new_subscriptions')))
    writer.row('Langganan berhenti', String(total('cancelled_subscriptions')))
    if (!salesOnly) {
      writer.row('Pendapatan diterima', formatCurrency(total('revenue')))
      writer.row('Pembayaran berhasil', String(total('payments_succeeded')))
      writer.row('Pembayaran gagal', String(total('payments_failed')))
      writer.row('Debit otomatis berhasil', String(total('autopay_succeeded')))
      writer.row('Debit otomatis gagal', String(total('autopay_failed')))
      writer.section('Pembayaran menurut paket')
      if (!range.revenue_by_plan.length) writer.text('Belum ada data pada rentang ini.')
      range.revenue_by_plan.forEach(point => writer.row(point.name, formatCurrency(point.amount)))
      writer.section('Pendapatan per tahun - tidak terpengaruh filter')
      summary.yearly_breakdown.forEach(point => writer.row(point.year, formatCurrency(point.income)))
      writer.text('Tahun kalender UTC. Pendapatan berasal dari pembayaran sukses, belum dikurangi refund atau biaya gateway.', 8, true)
    }
    await download(pdf, `ringkasan-paylanggan_${period.from}_${period.to}.pdf`)
    return
  }
  const { pdf, writer } = await createWriter()
  pdf.setTitle('Laporan statistik bisnis PayLanggan')
  writer.title('Laporan Statistik Bisnis', `ID bisnis: ${businessID} | Dibuat: ${reportDate()}`)
  writer.text('Periode metrik bulan ini mengikuti kalender UTC. Angka yang ditampilkan adalah data saat laporan dibuat.', 8, true)

  writer.section('Langganan dan pelanggan')
  writer.row('Pelanggan aktif', summary.active_customers.toLocaleString('id-ID'))
  writer.row('Langganan berjalan (termasuk trial)', summary.active_subscriptions.toLocaleString('id-ID'))
  writer.row('Langganan baru bulan ini', summary.new_subscriptions.toLocaleString('id-ID'))
  writer.row('Langganan berhenti bulan ini', summary.cancelled_subscriptions.toLocaleString('id-ID'))

  if (!salesOnly) {
    writer.section('Pendapatan dan tagihan')
    writer.row('Pendapatan diterima bulan ini', formatCurrency(summary.revenue_this_month))
    writer.row('MRR dari langganan aktif', formatCurrency(summary.mrr))
    writer.row('Tagihan belum dibayar', formatCurrency(summary.outstanding_amount))
    writer.row('Di antaranya lewat jatuh tempo', formatCurrency(summary.overdue_amount))
    writer.row('Jumlah invoice jatuh tempo', summary.due_invoices.toLocaleString('id-ID'))

    writer.section('Pembayaran bulan ini')
    writer.row('Pembayaran berhasil', summary.payments_succeeded.toLocaleString('id-ID'))
    writer.row('Pembayaran gagal', summary.payments_failed.toLocaleString('id-ID'))
    writer.row('Debit otomatis berhasil', summary.autopay_succeeded.toLocaleString('id-ID'))
    writer.row('Debit otomatis gagal', summary.autopay_failed.toLocaleString('id-ID'))

    writer.section('Pendapatan diterima per bulan')
    summary.monthly_performance.forEach((point, index) => writer.row(monthLabel(index, summary.monthly_performance.length), formatCurrency(point.value)))
    writer.section('Pendapatan diterima per tahun')
    summary.yearly_breakdown.forEach(point => writer.row(point.year, formatCurrency(point.income)))
    writer.section('Pembayaran menurut paket bulan ini')
    if (summary.revenue_by_plan.length === 0) writer.text('Belum ada pembayaran paket bulan ini.', 9, true)
    summary.revenue_by_plan.forEach(point => writer.row(point.name, formatCurrency(point.amount)))
    writer.section('Catatan perhitungan')
    writer.text('Pendapatan diterima berasal dari pembayaran berstatus sukses. MRR adalah estimasi dari langganan aktif dan bukan uang yang sudah diterima. Angka ini belum memperhitungkan refund atau biaya operasional.', 8, true)
  }
  await download(pdf, `paylanggan-laporan-bisnis-${fileDate()}.pdf`)
}

export async function exportPlatformReportPdf(summary: PlatformDashboardSummary) {
  const { pdf, writer } = await createWriter()
  pdf.setTitle('Laporan statistik platform PayLanggan')
  writer.title('Laporan Statistik PayLanggan', `Dibuat: ${reportDate()}`)
  writer.text('Periode bulan ini mengikuti kalender UTC. Laporan ini hanya mencakup tagihan bisnis kepada PayLanggan.', 8, true)
  writer.section('Langganan bisnis')
  writer.row('Bisnis berlangganan aktif', summary.active_businesses.toLocaleString('id-ID'))
  writer.row('Bisnis menunggak', summary.past_due_businesses.toLocaleString('id-ID'))
  writer.section('Pendapatan dan penagihan')
  writer.row('Pendapatan diterima bulan ini', formatCurrency(summary.revenue_this_month))
  writer.row('Tagihan bisnis belum dibayar', formatCurrency(summary.outstanding_amount))
  writer.row('Debit otomatis gagal bulan ini', summary.failed_autopay_this_month.toLocaleString('id-ID'))
  writer.section('Catatan perhitungan')
  writer.text('Pendapatan diterima berasal dari invoice platform yang lunas. Tagihan belum dibayar berasal dari invoice platform pending. Angka ini bukan laporan laba bersih.', 8, true)
  await download(pdf, `paylanggan-laporan-platform-${fileDate()}.pdf`)
}
