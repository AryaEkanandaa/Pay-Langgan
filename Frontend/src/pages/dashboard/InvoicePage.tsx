import { useCallback, useEffect, useRef, useState } from 'react'
import { Check, Copy, Eye, Link2, Printer } from 'lucide-react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import Modal from '../../components/ui/Modal'
import SendEmailButton from '../../components/billing/SendEmailButton'
import Select from '../../components/ui/Select'
import SubmitButton from '../../components/ui/SubmitButton'
import TableSkeleton from '../../components/ui/skeletons/TableSkeleton'
import { useAuth } from '../../context/AuthContext'
import {
  ApiError,
  getInvoiceEmail,
  previewInvoiceEmail,
  type InvoiceEmailState,
  createCheckoutLink,
  getInvoice,
  getMyBusiness,
  listInvoices,
  markInvoicePaid,
  type Invoice,
  type InvoiceDetail,
  type InvoiceStatus,
  type CheckoutLink,
} from '../../lib/api'
import { formatCurrency } from '../../lib/format'
import { printInvoiceDocument } from '../../lib/printInvoice'

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
  const [checkoutLink, setCheckoutLink] = useState<CheckoutLink | null>(null)
  const [emailState, setEmailState] = useState<InvoiceEmailState | null>(null)
  const [emailPreview, setEmailPreview] = useState<{subject: string; html: string; logo_warning?: string} | null>(null)
  const [emailError, setEmailError] = useState<string | null>(null)
  const [emailBusy, setEmailBusy] = useState(false)
  const [isPrinting, setIsPrinting] = useState(false)
  const [printError, setPrintError] = useState<string | null>(null)
  const [linkCopied, setLinkCopied] = useState(false)
  const detailRequest = useRef<AbortController | null>(null)
  useEffect(() => () => detailRequest.current?.abort(), [])

  function closeDetail() {
    detailRequest.current?.abort()
    setSelectedInvoice(null)
    setDetailLoading(false)
  }

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
    detailRequest.current?.abort()
    const controller = new AbortController()
    detailRequest.current = controller
    setEmailState(null)
    setEmailError(null)
    setEmailPreview(null)
    setPrintError(null)
    setSelectedInvoice(null)
    setCheckoutLink(null)
    setLinkCopied(false)
    setDetailLoading(true)
    try {
      const detail = await getInvoice(token, invoice.id, controller.signal)
      if (controller.signal.aborted) return
      setSelectedInvoice(detail)
      const email = await getInvoiceEmail(token, invoice.id, controller.signal)
      if (!controller.signal.aborted) setEmailState(email)
    } catch (err) {
      if (controller.signal.aborted) return
      setError(err instanceof ApiError ? err.message : 'Gagal memuat detail invoice.')
    } finally {
      if (!controller.signal.aborted) setDetailLoading(false)
    }
  }

  async function handleEmailPreview() {
    if (!token || !selectedInvoice) return
    setEmailBusy(true)
    setEmailError(null)
    try { setEmailPreview(await previewInvoiceEmail(token, selectedInvoice.id)) }
    catch (err) { setEmailError(err instanceof ApiError ? err.message : 'Gagal memproses email invoice.') }
    finally { setEmailBusy(false) }
  }

  async function handlePrint() {
    if (!token || !selectedInvoice || isPrinting) return
    setIsPrinting(true)
    setPrintError(null)
    try {
      const business = await getMyBusiness(token)
      await printInvoiceDocument(selectedInvoice, business)
    } catch (err) {
      setPrintError(err instanceof Error ? err.message : 'Gagal menyiapkan cetak invoice.')
    } finally { setIsPrinting(false) }
  }

  async function handleCreateCheckout(invoice: InvoiceDetail) {
    if (!token) return
    setActionID(invoice.id)
    setError(null)
    try {
      setCheckoutLink(await createCheckoutLink(token, invoice.id))
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal membuat link pembayaran.')
    } finally {
      setActionID(null)
    }
  }

  async function copyCheckoutLink() {
    if (!checkoutLink) return
    try {
      await navigator.clipboard.writeText(checkoutLink.checkout_url)
      setLinkCopied(true)
    } catch {
      setError('Browser tidak mengizinkan penyalinan otomatis.')
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
                        {invoice.status === 'pending' && invoice.checkout_status === 'expired' && <p className="mt-2 text-xs font-medium text-body">Checkout kedaluwarsa</p>}
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
        onClose={closeDetail}
        title={selectedInvoice?.invoice_number ?? 'Detail Invoice'}
      >
        {detailLoading ? (
          <p className="text-[13px] text-body">Memuat detail invoice...</p>
        ) : selectedInvoice ? (
          <div className="space-y-5">
            {selectedInvoice.status === 'pending' && selectedInvoice.checkout_status === 'expired' && <p className="rounded-xl bg-amber-50 p-3 text-sm text-amber-900 dark:bg-amber-950 dark:text-amber-200">Checkout sebelumnya sudah kedaluwarsa. Invoice belum lunas; pelanggan dapat melanjutkan pembayaran melalui portal atau tautan pembayaran baru.</p>}
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

            <section className="rounded-xl border border-border bg-white/30 p-4">
              <h3 className="text-[14px] font-semibold text-ink">Email invoice</h3>
              <p className="mt-1 text-[12px] text-muted">{emailState?.mode === 'smtp' ? 'Email dikirim oleh PayLanggan ke alamat pelanggan.' : emailState?.mode === 'preview' ? 'Mode preview: belum ada email yang dikirim ke pelanggan.' : 'Pengiriman email belum diaktifkan.'}</p>
              {emailState?.delivery && <p role="status" aria-live="polite" aria-atomic="true" className="mt-2 break-words text-[13px] text-body">{{queued: 'Dalam antrean', sending: 'Sedang diproses', sent: 'Terkirim', previewed: 'Preview dibuat (belum dikirim)', failed: 'Pengiriman gagal'}[emailState.delivery.status]} - {emailState.delivery.recipient}</p>}
              {emailError && <p role="alert" className="mt-2 text-[13px] text-red-600">{emailError}</p>}
              <div className="mt-3 flex flex-wrap gap-2">
                <SubmitButton type="button" variant="secondary" disabled={emailBusy} onClick={() => void handleEmailPreview()}>Preview email</SubmitButton>
                {canMarkPaid && token && <SendEmailButton key={selectedInvoice.id} token={token} invoiceId={selectedInvoice.id} state={emailState} onStateChange={setEmailState} disabled={emailBusy} />}
              </div>
              {emailPreview && <div className="mt-4">{emailPreview.logo_warning && <p role="status" className="mb-3 rounded-lg bg-amber-50 p-3 text-xs leading-5 text-amber-900 dark:bg-amber-950 dark:text-amber-200">{emailPreview.logo_warning}</p>}<p className="mb-2 text-[12px] text-body">Subjek: {emailPreview.subject}</p><iframe title="Preview email invoice" sandbox="" srcDoc={emailPreview.html} className="h-[520px] w-full rounded-xl border border-border bg-white" /></div>}
            </section>
            {checkoutLink && (
              <div className="border-t border-white/40 pt-4">
                <label htmlFor="checkout-link" className="text-[12px] font-medium text-body">Link pembayaran</label>
                <div className="mt-2 flex items-center gap-2">
                  <input id="checkout-link" readOnly value={checkoutLink.checkout_url} className="min-w-0 flex-1 rounded-lg border border-border bg-white/50 px-3 py-2 text-[12px] text-ink outline-none" />
                  <button type="button" onClick={copyCheckoutLink} title="Salin link pembayaran" aria-label="Salin link pembayaran" className="glass flex size-9 shrink-0 items-center justify-center rounded-lg text-brand">
                    {linkCopied ? <Check size={16} /> : <Copy size={16} />}
                  </button>
                </div>
              </div>
            )}

            {printError && <p role="alert" className="text-[13px] text-red-600">{printError}</p>}
            <div className="flex flex-wrap justify-end gap-2 border-t border-white/40 pt-4">
              <button type="button" onClick={handlePrint} disabled={isPrinting} className="glass inline-flex items-center gap-2 rounded-full px-4 py-2 text-[13px] font-semibold text-ink disabled:opacity-50">
                <Printer size={15} /> {isPrinting ? 'Menyiapkan...' : 'Cetak'}
              </button>
              {canMarkPaid && selectedInvoice.status === 'pending' && (
                <>
                  <SubmitButton type="button" variant="secondary" onClick={() => handleCreateCheckout(selectedInvoice)} disabled={actionID === selectedInvoice.id}>
                    <Link2 size={15} /> {checkoutLink ? 'Buat Link Baru' : 'Buat Link Bayar'}
                  </SubmitButton>
                  <SubmitButton type="button" onClick={() => handleMarkPaid(selectedInvoice)} disabled={actionID === selectedInvoice.id}>
                    {actionID === selectedInvoice.id ? 'Memproses...' : 'Tandai Sudah Dibayar'}
                  </SubmitButton>
                </>
              )}
            </div>
          </div>
        ) : null}
      </Modal>
    </div>
  )
}
