const assert = require('node:assert/strict')
const fs = require('node:fs')
const ts = require('typescript')
const pdfLib = require('pdf-lib')

const source = fs.readFileSync('src/lib/reportPdf.ts', 'utf8')
const compiled = ts.transpileModule(source, {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020 },
}).outputText

let downloadedBlob
let downloadedName
const exportsObject = {}
const mockRequire = id => {
    if (id === 'pdf-lib') return pdfLib
    if (id === './format') return { formatCurrency: value => `Rp${Number(value).toLocaleString('id-ID')}` }
    throw new Error(`Unexpected import: ${id}`)
  }
const mockURL = { createObjectURL: blob => { downloadedBlob = blob; return 'blob:test' }, revokeObjectURL: () => {} }
const mockDocument = {
    body: { appendChild: () => {} },
    createElement: () => ({ click: () => {}, remove: () => {}, set download(value) { downloadedName = value } }),
}
const mockWindow = { setTimeout: () => {} }
new Function('exports', 'require', 'Blob', 'URL', 'document', 'window', compiled)(
  exportsObject, mockRequire, Blob, mockURL, mockDocument, mockWindow,
)

async function verifyPdf() {
  assert(downloadedBlob, 'PDF was not downloaded')
  assert(downloadedName.endsWith('.pdf'), 'Downloaded file has no .pdf extension')
  const bytes = new Uint8Array(await downloadedBlob.arrayBuffer())
  assert.equal(Buffer.from(bytes.subarray(0, 5)).toString(), '%PDF-')
  const pdf = await pdfLib.PDFDocument.load(bytes)
  assert(pdf.getPageCount() >= 1)
  downloadedBlob = undefined
}

async function main() {
  await exportsObject.exportBusinessReportPdf({
    active_customers: 2, active_subscriptions: 2, new_subscriptions: 1, cancelled_subscriptions: 0,
    revenue_this_month: 100000, mrr: 100000, outstanding_amount: 200000,
    overdue_amount: 100000, due_invoices: 1, payments_succeeded: 1, payments_failed: 0,
    autopay_succeeded: 0, autopay_failed: 0,
    monthly_performance: Array.from({ length: 7 }, (_, i) => ({ month: `Bulan ${i}`, value: i * 10000 })),
    yearly_breakdown: Array.from({ length: 5 }, (_, i) => ({ year: String(2022 + i), income: i * 100000 })),
    revenue_by_plan: [{ plan_id: 1, name: 'Paket Langganan Élite dengan Nama yang Sangat Panjang 🛒', amount: 100000 }],
  }, 'biz-demo', false)
  await verifyPdf()

  await exportsObject.exportPlatformReportPdf({
    active_businesses: 1, past_due_businesses: 1, revenue_this_month: 300000,
    outstanding_amount: 300000, failed_autopay_this_month: 1,
  })
  await verifyPdf()
  process.stdout.write('Business and platform PDF reports are valid.\n')
}

main().catch(error => { console.error(error); process.exitCode = 1 })
