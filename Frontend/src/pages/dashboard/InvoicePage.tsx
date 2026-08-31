import { useCallback, useEffect, useState } from 'react'
import { Eye, Printer } from 'lucide-react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import Modal from '../../components/ui/Modal'
import Select from '../../components/ui/Select'
import SubmitButton from '../../components/ui/SubmitButton'
import TableSkeleton from '../../components/ui/skeletons/TableSkeleton'
import { useAuth } from '../../context/AuthContext'
import {
  ApiError,
  getInvoice,
  listInvoices,
  markInvoicePaid,
  type Invoice,
  type InvoiceDetail,
  type InvoiceStatus,
} from '../../lib/api'
import { formatCurrency } from '../../lib/format'

const PAGE_SIZE = 10

const statusConfig: Record<InvoiceStatus, { label: string; className: string }> = {
  pending: { label: 'Menunggu Pembayaran', className: 'bg-amber-100 text-amber-700' },
  paid: { label: 'Lunas', className: 'bg-emerald-100 text-emerald-700' },
  failed: { label: 'Gagal', className: 'bg-red-100 text-red-700' },
  cancelled: { label: 'Dibatalkan', className: 'bg-slate-100 text-slate-700' },
}

function formatDate(value: string | null) {
  return value ? new Date(value).toLocaleDateString('id-ID') : '—'
}

export default function InvoicePage() {
  const { auth } = useAuth()
  const token = auth?.token
  const [invoices, setInvoices] = useState<Invoice[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<InvoiceStatus | undefined>()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedInvoice, setSelectedInvoice] = useState<InvoiceDetail | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [actionID, setActionID] = useState<number | null>(null)

  const loadInvoices = useCallback(async () => {
    if (!token) return
    setIsLoading(true)
    setError(null)
    try {
      const result = await listInvoices(token, {
        page,
        limit: PAGE_SIZE,
        search: search || undefined,
        status,
      })
      setInvoices(result.items)
      setTotal(result.total)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memuat data invoice.')
    } finally {
      setIsLoading(false)
    }
  }, [token, page, search, status])

  useEffect(() => {
    loadInvoices()
  }, [loadInvoices])

  async function openDetail(invoice: Invoice) {
    if (!token) return
    setSelectedInvoice(null)
    setDetailLoading(true)
    try {
      setSelectedInvoice(await getInvoice(token, invoice.id))
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memuat detail invoice.')
    } finally {
      setDetailLoading(false)
    }
  }

  async function handleMarkPaid(invoice: Invoice) {
    if (!token || !window.confirm(`Tandai ${invoice.invoice_number} sebagai lunas?`)) return
    setActionID(invoice.id)
    setError(null)
    try {
      const updated = await markInvoicePaid(token, invoice.id)
      setSelectedInvoice(updated)
      await loadInvoices()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memperbarui pembayaran invoice.')
    } finally {
      setActionID(null)
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))
  const canMarkPaid = auth?.user.role === 'admin' || auth?.user.role === 'finance'

  if (!auth) return null

  return (
    <div>
      <div className="mb-8">
        <DashboardPageHeader
          title="Invoice"
          description="Pantau tagihan, jatuh tempo, dan status pembayaran pelanggan."
        />
      </div>

      <div className="mb-4 flex flex-wrap items-center gap-3">
        <input
          type="search"
          placeholder="Cari nomor invoice atau pelanggan..."
          value={search}
          onChange={(event) => {
            setPage(1)
            setSearch(event.target.value)
          }}
          className="w-full max-w-sm rounded-xl border border-white/60 bg-white/50 px-4 py-2.5 text-[14px] text-ink outline-none backdrop-blur-xl transition-colors placeholder:text-muted focus:border-brand/50 focus:bg-white/70 focus:ring-2 focus:ring-brand/30"
        />
        <Select
          label="Filter status invoice"
          hideLabel
          value={status ?? ''}
          onChange={(event) => {
            setPage(1)
            setStatus((event.target.value || undefined) as InvoiceStatus | undefined)
          }}
          className="w-full sm:w-[240px]"
        >
          <option value="">Semua status</option>
          <option value="pending">Menunggu Pembayaran</option>
          <option value="paid">Lunas</option>
          <option value="failed">Gagal</option>
          <option value="cancelled">Dibatalkan</option>
        </Select>
      </div>

      {error && <p className="mb-4 text-[13px] text-red-600">{error}</p>}

      {isLoading ? (
        <TableSkeleton columns={['Invoice', 'Pelanggan', 'Plan', 'Jatuh Tempo', 'Total', 'Status', '']} />
      ) : invoices.length === 0 ? (
        <div className="glass flex min-h-[240px] flex-col items-center justify-center rounded-2xl px-6 text-center">
          <p className="text-[15px] font-semibold text-ink">Belum ada invoice</p>
        </div>
      ) : (
        <div className="glass overflow-hidden rounded-2xl">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[980px] text-left text-[13px]">
              <thead>
                <tr className="border-b border-white/50">
                  {['Invoice', 'Pelanggan', 'Plan', 'Jatuh Tempo', 'Total', 'Status', 'Aksi'].map((heading) => (
                    <th key={heading} className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">
                      {heading}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {invoices.map((invoice) => {
                  const statusInfo = statusConfig[invoice.status]
                  return (
                    <tr key={invoice.id} className="border-b border-white/30 last:border-0">
                      <td className="px-5 py-4 font-semibold tracking-wide text-ink">{invoice.invoice_number}</td>
                      <td className="px-5 py-4">
                        <p className="font-medium text-ink">{invoice.customer_name}</p>
                        <p className="mt-1 text-[12px] text-muted">{invoice.customer_email || '—'}</p>
                      </td>
                      <td className="px-5 py-4 text-body">{invoice.plan_name}</td>
                      <td className="px-5 py-4 text-body">{formatDate(invoice.due_date)}</td>
                      <td className="px-5 py-4 font-semibold text-ink">{formatCurrency(invoice.amount)}</td>
                      <td className="px-5 py-4">
                        <span className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${statusInfo.className}`}>
                          {statusInfo.label}
                        </span>
                      </td>
                      <td className="px-5 py-4 text-right">
                        <button type="button" onClick={() => openDetail(invoice)} className="inline-flex items-center gap-1.5 text-[13px] font-semibold text-brand hover:underline">
                          <Eye size={15} /> Detail
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {!isLoading && invoices.length > 0 && totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between text-[13px] text-body">
          <span>Halaman {page} dari {totalPages} ({total} invoice)</span>
          <div className="flex gap-2">
            <button type="button" onClick={() => setPage((current) => Math.max(1, current - 1))} disabled={page <= 1} className="glass rounded-full px-4 py-1.5 font-medium text-ink disabled:cursor-not-allowed disabled:opacity-50">Sebelumnya</button>
            <button type="button" onClick={() => setPage((current) => Math.min(totalPages, current + 1))} disabled={page >= totalPages} className="glass rounded-full px-4 py-1.5 font-medium text-ink disabled:cursor-not-allowed disabled:opacity-50">Berikutnya</button>
          </div>
        </div>
      )}

      <Modal
        open={detailLoading || selectedInvoice !== null}
        onClose={() => setSelectedInvoice(null)}
        title={selectedInvoice?.invoice_number ?? 'Detail Invoice'}
      >
        {detailLoading ? (
          <p className="text-[13px] text-body">Memuat detail invoice...</p>
        ) : selectedInvoice ? (
          <div className="space-y-5">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-[13px] text-body">Ditagihkan kepada</p>
                <p className="mt-1 font-semibold text-ink">{selectedInvoice.customer_name}</p>
                <p className="text-[13px] text-muted">{selectedInvoice.customer_email || '—'}</p>
              </div>
              <span className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${statusConfig[selectedInvoice.status].className}`}>
                {statusConfig[selectedInvoice.status].label}
              </span>
            </div>

            <div className="rounded-xl bg-white/40 p-4">
              <div className="flex items-center justify-between text-[13px]">
                <span className="text-body">{selectedInvoice.plan_name}</span>
                <span className="font-semibold text-ink">{formatCurrency(selectedInvoice.amount)}</span>
              </div>
              {selectedInvoice.items.length > 0 && (
                <div className="mt-3 space-y-2 border-t border-white/40 pt-3">
                  {selectedInvoice.items.map((item) => (
                    <div key={item.id} className="flex items-center justify-between text-[12px]">
                      <span className="text-body">{item.description} × {item.quantity}</span>
                      <span className="text-ink">{formatCurrency(item.subtotal)}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>

            <div className="grid grid-cols-2 gap-3 text-[12px]">
              <div><p className="text-muted">Tanggal terbit</p><p className="mt-1 font-medium text-ink">{formatDate(selectedInvoice.created_at)}</p></div>
              <div><p className="text-muted">Jatuh tempo</p><p className="mt-1 font-medium text-ink">{formatDate(selectedInvoice.due_date)}</p></div>
              <div><p className="text-muted">Dibayar pada</p><p className="mt-1 font-medium text-ink">{formatDate(selectedInvoice.paid_at)}</p></div>
              <div><p className="text-muted">Metode pembayaran</p><p className="mt-1 font-medium capitalize text-ink">{selectedInvoice.payment?.provider || '—'}</p></div>
            </div>

            <div className="flex flex-wrap justify-end gap-2 border-t border-white/40 pt-4">
              <button type="button" onClick={() => window.print()} className="glass inline-flex items-center gap-2 rounded-full px-4 py-2 text-[13px] font-semibold text-ink">
                <Printer size={15} /> Cetak
              </button>
              {canMarkPaid && selectedInvoice.status === 'pending' && (
                <SubmitButton type="button" onClick={() => handleMarkPaid(selectedInvoice)} disabled={actionID === selectedInvoice.id}>
                  {actionID === selectedInvoice.id ? 'Memproses...' : 'Tandai Sudah Dibayar'}
                </SubmitButton>
              )}
            </div>
          </div>
        ) : null}
      </Modal>
    </div>
  )
}
