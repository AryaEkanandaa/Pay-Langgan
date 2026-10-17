import { useContext, useEffect, useState } from 'react'
import { Copy, Eye, RefreshCw, Webhook } from 'lucide-react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import Modal from '../../components/ui/Modal'
import Select from '../../components/ui/Select'
import SubmitButton from '../../components/ui/SubmitButton'
import TableSkeleton from '../../components/ui/skeletons/TableSkeleton'
import { useAuth } from '../../context/AuthContext'
import { getNewWebhookCount, getWebhookLogs, type WebhookLogDTO, type WebhookLogPageDTO } from '../../lib/api'
import { WebhookNotificationsContext } from '../../hooks/useNewWebhookCount'

const columns = ['Waktu diterima', 'Gateway / event', 'Referensi', 'Status', 'HTTP', 'Aksi']
const statusConfig = {
  accepted: { label: 'Diterima', className: 'bg-emerald-100 text-emerald-700' },
  duplicate: { label: 'Duplikat', className: 'bg-amber-100 text-amber-700' },
  failed: { label: 'Gagal', className: 'bg-red-100 text-red-700' },
}
const errors: Record<string, string> = {
  authentication_failed: 'Verifikasi webhook gagal. Periksa token Xendit atau konfigurasi akun Midtrans.',
  invalid_event: 'Event tidak valid atau tidak cocok dengan transaksi bisnis ini.',
  state_conflict: 'Status transaksi tidak mengizinkan event ini diproses.',
  not_found: 'Konfigurasi atau transaksi tidak ditemukan.',
  processing_failed: 'Terjadi kesalahan saat memproses event. Periksa log backend.',
}
const date = (value: string) => new Date(value).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'medium' })

function StatusBadge({ status }: { status: WebhookLogDTO['status'] }) {
  const config = statusConfig[status]
  return <span className={`inline-flex rounded-full px-2.5 py-1 text-[11px] font-semibold ${config.className}`}>{config.label}</span>
}

export default function WebhookLogsPage() {
  const { auth } = useAuth()
  const token = auth?.token
  const notifications = useContext(WebhookNotificationsContext)
  const markSeen = notifications?.markSeen
  const [visitBaseline, setVisitBaseline] = useState<string | null>(null)
  const [loadedThrough, setLoadedThrough] = useState<string | null>(null)
  const [data, setData] = useState<WebhookLogPageDTO | null>(null)
  const [page, setPage] = useState(1)
  const [provider, setProvider] = useState('')
  const [status, setStatus] = useState('')
  const [reload, setReload] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selected, setSelected] = useState<WebhookLogDTO | null>(null)
  const [copied, setCopied] = useState(false)
  const [copyError, setCopyError] = useState(false)

  useEffect(() => {
    if (!token) return
    const controller = new AbortController()
    setLoading(true)
    setError(null)
    async function load() {
      // Capture server time before the list request. Later arrivals remain unread.
      const unread = await getNewWebhookCount(token!, controller.signal).catch(() => null)
      const result = await getWebhookLogs(token!, page, provider, status, controller.signal)
      if (controller.signal.aborted) return
      if (unread) {
        setVisitBaseline(previous => previous ?? unread.last_seen_at)
        setLoadedThrough(unread.checked_at)
      }
      setData(result)
    }
    void load()
      .catch(() => { if (!controller.signal.aborted) setError('Log webhook belum dapat dimuat. Silakan coba lagi.') })
      .finally(() => { if (!controller.signal.aborted) setLoading(false) })
    return () => controller.abort()
  }, [token, page, provider, status, reload])

  // Run after the list is rendered, preserving the old visit baseline in memory.
  useEffect(() => {
    if (loading || error || !data || !loadedThrough || !markSeen) return
    const controller = new AbortController()
    void markSeen(loadedThrough, controller.signal).catch(() => { /* Retry on refresh; keep the last known badge. */ })
    return () => controller.abort()
  }, [loading, error, data, loadedThrough, markSeen])

  async function copyDetail() {
    try {
      await navigator.clipboard.writeText(JSON.stringify(selected, null, 2))
      setCopied(true)
      setCopyError(false)
    } catch {
      setCopyError(true)
    }
  }

  const totalPages = Math.max(1, Math.ceil((data?.total ?? 0) / (data?.limit ?? 20)))
  const paginationButton = 'glass rounded-full px-4 py-1.5 font-medium text-ink disabled:cursor-not-allowed disabled:opacity-50'

  return (
    <div>
      <div className="mb-8">
        <DashboardPageHeader title="Webhook Logs" description="Pantau notifikasi pembayaran dari Xendit dan Midtrans ke PayLanggan." />
      </div>
      <div className="mb-4 flex flex-wrap items-center gap-3">
        <div className="w-full sm:w-[220px]">
        <Select label="Gateway" hideLabel value={provider} onChange={event => { setProvider(event.target.value); setPage(1) }}>
          <option value="">Semua gateway</option>
          <option value="xendit">Xendit</option>
          <option value="midtrans">Midtrans</option>
        </Select>
        </div>
        <div className="w-full sm:w-[220px]">
        <Select label="Status webhook" hideLabel value={status} onChange={event => { setStatus(event.target.value); setPage(1) }}>
          <option value="">Semua status</option>
          <option value="accepted">Diterima</option>
          <option value="duplicate">Duplikat</option>
          <option value="failed">Gagal</option>
        </Select>
        </div>
        <SubmitButton type="button" variant="secondary" disabled={loading} onClick={() => setReload(value => value + 1)} className="gap-2 sm:ml-auto">
          <RefreshCw size={16} className={loading ? 'animate-spin' : ''} />Perbarui
        </SubmitButton>
      </div>
      <p className="mb-4 text-[12px] text-muted">Status di sini menunjukkan penerimaan webhook, bukan status pembayaran.</p>
      {(notifications?.count ?? 0) > 0 && !loading && <button type="button" className="webhook-new-notice mb-4 rounded-full px-4 py-2 text-sm font-semibold focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand" onClick={() => { setPage(1); setReload(value => value + 1) }} aria-live="polite">{notifications?.count} log baru, muat ulang</button>}
      {error && <p role="alert" className="mb-4 text-[13px] text-red-600">{error}</p>}
      {loading ? (
        <TableSkeleton columns={columns} />
      ) : error ? null : !data?.items.length ? (
        <div className="glass flex min-h-[240px] flex-col items-center justify-center rounded-2xl px-6 text-center">
          <Webhook className="mb-3 text-muted" size={28} />
          <p className="text-[15px] font-semibold text-ink">Belum ada webhook{provider || status ? ' yang sesuai filter' : ''}</p>
          <p className="mt-2 text-[13px] text-muted">Notifikasi yang mencapai endpoint PayLanggan akan muncul di sini.</p>
        </div>
      ) : (
        <div className="glass overflow-hidden rounded-2xl">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[900px] text-left text-[13px]">
              <thead>
                <tr className="border-b border-white/50">
                  {columns.map(label => <th key={label} className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">{label}</th>)}
                </tr>
              </thead>
              <tbody>
                {data.items.map(log => (
                  <tr key={log.id} className={`border-b border-white/30 last:border-0 ${visitBaseline && new Date(log.received_at).getTime() > new Date(visitBaseline).getTime() ? 'webhook-row-new' : ''}`}>
                    <td className="whitespace-nowrap px-5 py-4 text-body">{date(log.received_at)}{visitBaseline && new Date(log.received_at).getTime() > new Date(visitBaseline).getTime() && <span className="webhook-new-label ml-2 inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-[11px] font-semibold"><span aria-hidden="true" className="size-1.5 rounded-full bg-current" />Baru</span>}</td>
                    <td className="px-5 py-4">
                      <p className="font-medium capitalize text-ink">{log.provider}</p>
                      <p className="mt-1 text-[12px] text-muted">{log.event_type || 'Event tidak terbaca'}</p>
                    </td>
                    <td className="max-w-56 px-5 py-4"><p className="truncate font-mono text-[12px] text-body" title={log.reference_id}>{log.reference_id || '—'}</p></td>
                    <td className="px-5 py-4">
                      <StatusBadge status={log.status} />
                      {log.status !== 'failed' && <p className="mt-1 text-[12px] text-muted">{log.processed ? 'Sudah diproses' : 'Belum tercatat diproses'}</p>}
                    </td>
                    <td className="px-5 py-4 font-medium text-ink">{log.http_status}</td>
                    <td className="px-5 py-4 text-right">
                      <button type="button" className="inline-flex items-center gap-1.5 font-semibold text-brand hover:text-brand-dark" onClick={() => { setSelected(log); setCopied(false); setCopyError(false) }}>
                        <Eye size={15} />Detail
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
      {data && !error && (
        <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-[13px]">
          <span className="text-body">Total {data.total} log</span>
          <div className="flex items-center gap-2">
            <button type="button" disabled={loading || page === 1} className={paginationButton} onClick={() => setPage(value => value - 1)}>Sebelumnya</button>
            <span className="px-2 text-muted">{page} / {totalPages}</span>
            <button type="button" disabled={loading || page >= totalPages} className={paginationButton} onClick={() => setPage(value => value + 1)}>Berikutnya</button>
          </div>
        </div>
      )}
      <p className="mt-4 text-[12px] text-muted">Webhook yang gagal mencapai server hanya dapat diperiksa melalui riwayat pengiriman pada dashboard gateway.</p>
      <Modal open={selected !== null} onClose={() => setSelected(null)} title="Detail webhook">
        {selected && (
          <div className="space-y-5">
            <div className="flex items-center justify-between gap-3">
              <p className="text-[13px] font-semibold capitalize text-ink">{selected.provider}</p>
              <StatusBadge status={selected.status} />
            </div>
            <dl className="space-y-4 rounded-xl bg-white/40 p-4 text-[13px]">
              {[
                ['Waktu diterima', date(selected.received_at)],
                ['Event', selected.event_type || '—'],
                ['Referensi transaksi', selected.reference_id || '—'],
                ['Respons HTTP', String(selected.http_status)],
                ['Waktu pemrosesan', `${selected.duration_ms} ms`],
              ].map(([label, value]) => (
                <div key={label}><dt className="text-muted">{label}</dt><dd className="mt-1 break-all font-medium text-ink">{value}</dd></div>
              ))}
            </dl>
            {selected.error_code && <p className="rounded-xl bg-red-100 p-4 text-[13px] text-red-700">{errors[selected.error_code] ?? 'Pemrosesan webhook gagal.'}</p>}
            {copyError && <p role="alert" className="text-[13px] text-red-600">Browser tidak mengizinkan penyalinan otomatis.</p>}
            <SubmitButton type="button" variant="secondary" onClick={copyDetail} className="w-full gap-2"><Copy size={16} />{copied ? 'Tersalin' : 'Salin detail'}</SubmitButton>
          </div>
        )}
      </Modal>
    </div>
  )
}
