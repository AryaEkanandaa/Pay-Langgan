import type { PlatformInvoice } from '../../lib/api'
import { formatCurrency } from '../../lib/format'
import GlassPanel from '../ui/GlassPanel'
import { billingDate } from './format'

interface InvoiceHistoryProps {
  invoices: PlatformInvoice[]
  resumableInvoiceId?: number
  busy: boolean
  onResume: () => void
}

function Status({ status }: { status: PlatformInvoice['status'] }) {
  return <span className={`inline-block rounded-full px-3 py-1 text-xs font-semibold ${status === 'paid' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200' : 'bg-amber-100 text-amber-900 dark:bg-amber-950 dark:text-amber-200'}`}>{status === 'paid' ? 'Lunas' : 'Menunggu pembayaran'}</span>
}

export default function InvoiceHistory({ invoices, resumableInvoiceId, busy, onResume }: InvoiceHistoryProps) {
  function action(invoice: PlatformInvoice) {
    return invoice.status === 'pending' && invoice.id === resumableInvoiceId
      ? <button type="button" disabled={busy} onClick={onResume} aria-label={`Lanjutkan pembayaran invoice ${invoice.id}`} className="rounded-lg px-2 py-2 font-semibold text-brand focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand disabled:opacity-50">Lanjutkan pembayaran</button>
      : <span className="text-body">—</span>
  }
  return (
    <GlassPanel className="flex min-h-0 min-w-0 flex-col !shadow-sm" style={{ maxHeight: 480 }}>
      <h3 className="shrink-0 text-base font-semibold text-ink">Riwayat tagihan</h3>
      {invoices.length === 0 ? <div className="py-8 text-center"><p className="font-semibold text-ink">Belum ada tagihan</p><p className="mt-2 text-sm text-body">Tagihan langganan PayLanggan akan ditampilkan di sini setelah dibuat.</p></div> : <div tabIndex={0} role="region" aria-label="Riwayat tagihan langganan" className="mt-4 min-h-0 flex-1 overflow-y-auto focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand">
        <div className="hidden overflow-x-auto md:block">
          <table className="w-full text-left text-sm">
            <caption className="sr-only">Riwayat tagihan langganan PayLanggan</caption>
            <thead className="text-body"><tr>{['Tanggal', 'Nomor invoice', 'Jumlah', 'Status', 'Aksi'].map(label => <th key={label} scope="col" className="px-3 py-3 font-semibold">{label}</th>)}</tr></thead>
            <tbody>{invoices.map(invoice => <tr key={invoice.id} className="border-t border-slate-200 align-top dark:border-slate-700">
              <td className="px-3 py-4 text-body">{billingDate(invoice.created_at)}<p className="mt-2 text-xs">Periode: {billingDate(invoice.period_start)} – {billingDate(invoice.period_end)}</p><p className="mt-2 text-xs">Dibayar: {billingDate(invoice.paid_at)}</p></td>
              <td className="px-3 py-4 font-mono text-ink">#{invoice.id}</td>
              <td className="whitespace-nowrap px-3 py-4 font-semibold text-ink">{formatCurrency(invoice.amount)}</td>
              <td className="px-3 py-4"><Status status={invoice.status} /></td><td className="px-3 py-4">{action(invoice)}</td>
            </tr>)}</tbody>
          </table>
        </div>
        <ul className="space-y-4 md:hidden">{invoices.map(invoice => <li key={invoice.id} className="rounded-xl border border-slate-200 p-4 dark:border-slate-700">
          <div className="flex flex-wrap items-center justify-between gap-2"><span className="font-mono font-semibold text-ink">#{invoice.id}</span><Status status={invoice.status} /></div>
          <p className="mt-4 font-semibold text-ink">{formatCurrency(invoice.amount)}</p>
          <dl className="mt-3 space-y-2 text-sm text-body"><div><dt className="inline">Tanggal: </dt><dd className="inline">{billingDate(invoice.created_at)}</dd></div><div><dt className="inline">Periode: </dt><dd className="inline">{billingDate(invoice.period_start)} – {billingDate(invoice.period_end)}</dd></div><div><dt className="inline">Dibayar: </dt><dd className="inline">{billingDate(invoice.paid_at)}</dd></div></dl>
          {invoice.id === resumableInvoiceId && <div className="mt-3">{action(invoice)}</div>}
        </li>)}</ul>
      </div>}
    </GlassPanel>
  )
}
