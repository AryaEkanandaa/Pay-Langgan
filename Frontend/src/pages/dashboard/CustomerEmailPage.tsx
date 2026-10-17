import { useEffect, useState } from 'react'
import { Mail, RefreshCw } from 'lucide-react'
import { useAuth } from '../../context/AuthContext'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import Select from '../../components/ui/Select'
import Modal from '../../components/ui/Modal'
import StatusBadge from '../../components/ui/StatusBadge'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import { getCustomerEmailHistory, getCustomerEmailPreview, getEmailReminderPreferences, setEmailReminderPreferences, retryCustomerEmail, type CustomerEmailLog, type CustomerEmailKind, type CustomerEmailStatus, type CustomerEmailPage as EmailPage, type EmailReminderPreferences } from '../../lib/api'

const kinds: Record<CustomerEmailKind, string> = { invoice_issued: 'Invoice', payment_succeeded: 'Pembayaran berhasil', trial_started: 'Trial dimulai', trial_ending: 'Trial segera berakhir', invoice_due: 'Pengingat jatuh tempo' }
const statuses: Record<CustomerEmailStatus, string> = { queued: 'Dalam antrean', sending: 'Mengirim', sent: 'Terkirim', previewed: 'Preview lokal', failed: 'Gagal', cancelled: 'Dilewati' }
const actionClass = 'rounded-full border border-border px-4 py-2 text-sm font-semibold text-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand disabled:opacity-50'
function date(value?: string) { return value ? new Date(value).toLocaleString('id-ID') : '—' }

export default function CustomerEmailPage() {
  const { auth } = useAuth()
  const token = auth?.token
  const canWrite = auth?.user.role === 'admin' || auth?.user.role === 'finance'
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')
  const [kind, setKind] = useState('')
  const [search, setSearch] = useState('')
  const [data, setData] = useState<EmailPage | null>(null)
  const [settings, setSettings] = useState<EmailReminderPreferences | null>(null)
  const [revision, setRevision] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [saving, setSaving] = useState(false)
  const [retry, setRetry] = useState<CustomerEmailLog | null>(null)
  const [preview, setPreview] = useState<CustomerEmailLog | null>(null)
  const [previewData, setPreviewData] = useState<{ subject: string; html: string } | null>(null)
  const [previewError, setPreviewError] = useState('')

  useEffect(() => {
    if (!token) return
    const controller = new AbortController()
    void getEmailReminderPreferences(token, controller.signal).then(setSettings).catch((err: unknown) => { if (!controller.signal.aborted) setError(err instanceof Error ? err.message : 'Gagal memuat pengaturan.') })
    return () => controller.abort()
  }, [token])

  useEffect(() => {
    if (!token) return
    const controller = new AbortController()
    setLoading(true)
    let inFlight = false
    const load = async () => {
      if (inFlight || controller.signal.aborted) return
      inFlight = true
      try { const result = await getCustomerEmailHistory(token, page, status, kind, search, controller.signal); if (!controller.signal.aborted) { setData(result); setError('') } }
      catch (err) { if (!controller.signal.aborted) setError(err instanceof Error ? err.message : 'Gagal memuat riwayat email.') }
      finally { inFlight = false; if (!controller.signal.aborted) setLoading(false) }
    }
    const delay = window.setTimeout(() => { void load() }, 250)
    const timer = window.setInterval(() => { if (!document.hidden) void load() }, 5000)
    return () => { controller.abort(); window.clearTimeout(delay); window.clearInterval(timer) }
  }, [token, page, status, kind, search, revision])

  useEffect(() => {
    if (!token || !preview) return
    const controller = new AbortController()
    setPreviewData(null); setPreviewError('')
    void getCustomerEmailPreview(token, preview, controller.signal).then((result) => { if (!controller.signal.aborted) setPreviewData(result) }).catch((err: unknown) => { if (!controller.signal.aborted) setPreviewError(err instanceof Error ? err.message : 'Preview belum tersedia.') })
    return () => controller.abort()
  }, [token, preview])

  async function changeReminders() {
    if (!token || !settings || saving) return
    setSaving(true); setError('')
    try { const result = await setEmailReminderPreferences(token, !settings.enabled); setSettings(result); setNotice(result.enabled ? 'Pengingat otomatis diaktifkan untuk data baru.' : 'Pengingat otomatis dinonaktifkan.') }
    catch (err) { setError(err instanceof Error ? err.message : 'Gagal menyimpan pengaturan.') }
    finally { setSaving(false) }
  }
  async function confirmRetry() {
    if (!token || !retry || saving) return
    setSaving(true); setError('')
    try { await retryCustomerEmail(token, retry); setRetry(null); setNotice('Email masuk antrean kembali.'); setRevision((v) => v + 1) }
    catch (err) { setError(err instanceof Error ? err.message : 'Gagal mencoba ulang email.') }
    finally { setSaving(false) }
  }
  return <div className="space-y-6">
    <DashboardPageHeader title="Email Pelanggan" description="Pantau pengiriman invoice, konfirmasi pembayaran, trial, dan pengingat pelanggan." />
    <section className="glass rounded-2xl p-6" aria-labelledby="reminder-heading">
      <div className="flex flex-wrap items-start justify-between gap-4"><div><h2 id="reminder-heading" className="font-semibold text-ink">Pengingat otomatis</h2><p className="mt-2 text-sm text-body">Trial: 3 hari sebelum berakhir. Invoice: 1 hari sebelum jatuh tempo.</p></div>
        {canWrite ? <button type="button" role="switch" aria-checked={settings?.enabled ?? false} aria-label="Pengingat otomatis" disabled={!settings || saving} onClick={() => { void changeReminders() }} className={`${actionClass} inline-flex items-center gap-3`}><span className={`relative h-6 w-11 rounded-full ${settings?.enabled ? 'bg-brand' : 'bg-slate-400'}`}><span className={`absolute top-1 size-4 rounded-full bg-white transition-transform motion-reduce:transition-none ${settings?.enabled ? 'left-1 translate-x-5' : 'left-1'}`} /></span>{settings?.enabled ? 'Aktif' : 'Nonaktif'}</button> : <StatusBadge>{settings?.enabled ? 'Aktif' : 'Nonaktif'}</StatusBadge>}
      </div><p className="mt-4 text-sm text-body">Berlaku untuk trial dan invoice yang dibuat sejak pengingat diaktifkan. Pengingat dilewati jika trial berakhir/dibatalkan atau invoice sudah lunas. Invoice yang langsung jatuh tempo saat checkout tidak mendapat pengingat sebelumnya.</p>
      {data && data.mode !== 'smtp' && <p className="mt-3 text-sm text-body">{data.mode === 'preview' ? 'Mode preview: email dibuat sebagai berkas lokal dan belum dikirim ke pelanggan.' : 'Pengiriman email belum aktif pada server.'}</p>}
    </section>
    <div aria-live="polite" className="text-sm text-body">{notice}</div>
    {error && <p role="alert" className="text-sm text-red-700 dark:text-red-300">{error}</p>}
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-[minmax(220px,1fr)_220px_250px_auto]">
      <label className="text-sm text-body">Cari pelanggan atau email<input type="search" value={search} onChange={(e) => { setSearch(e.target.value); setPage(1) }} className="mt-1.5 block w-full rounded-lg border border-border bg-transparent px-4 py-2.5 text-ink focus-visible:outline-brand" /></label>
      <Select label="Status email" value={status} onChange={(e) => { setStatus(e.target.value); setPage(1) }}><option value="">Semua status</option>{Object.entries(statuses).map(([key, value]) => <option key={key} value={key}>{value}</option>)}</Select>
      <Select label="Jenis email" value={kind} onChange={(e) => { setKind(e.target.value); setPage(1) }}><option value="">Semua jenis</option>{Object.entries(kinds).map(([key, value]) => <option key={key} value={key}>{value}</option>)}</Select>
      <button type="button" disabled={loading} onClick={() => setRevision((v) => v + 1)} className={`${actionClass} flex items-center justify-center gap-2 self-end`}><RefreshCw size={16} />Perbarui</button>
    </div>
    <section className="glass overflow-hidden rounded-2xl" aria-label="Riwayat email" aria-busy={loading}>
      {!data && loading ? <p className="p-8 text-body">Memuat riwayat email…</p> : data?.items.length ? <div className="overflow-x-auto"><table className="w-full min-w-[860px] text-left text-sm"><thead className="border-b border-border text-body"><tr>{['Pelanggan', 'Jenis email', 'Status', 'Percobaan', 'Dibuat / Terkirim', 'Aksi'].map((label) => <th key={label} className="px-5 py-4 font-semibold">{label}</th>)}</tr></thead><tbody>{data.items.map((item) => <tr key={`${item.source}-${item.id}`} className="border-b border-border last:border-0"><td className="px-5 py-4"><p className="font-semibold text-ink">{item.customer_name}</p><p className="mt-1 break-all text-body">{item.recipient}</p></td><td className="px-5 py-4 text-ink">{kinds[item.kind]}</td><td className="px-5 py-4"><StatusBadge tone={item.status === 'sent' ? 'success' : item.status === 'failed' ? 'warning' : 'neutral'}>{statuses[item.status]}</StatusBadge>{item.last_error && <p className="mt-2 max-w-48 text-xs text-body">{item.status === 'cancelled' ? 'Pengingat tidak diperlukan lagi.' : 'Pengiriman belum berhasil. Periksa alamat email dan konfigurasi pengirim.'}</p>}</td><td className="px-5 py-4 text-body">{item.attempts}</td><td className="px-5 py-4 text-body"><p>{date(item.created_at)}</p><p className="mt-1">Terkirim: {date(item.sent_at)}</p></td><td className="px-5 py-4"><div className="flex gap-2"><button type="button" className={actionClass} onClick={() => setPreview(item)}>Detail</button>{item.status === 'failed' && canWrite && <button type="button" disabled={data.mode === 'disabled'} className={actionClass} onClick={() => { setError(''); setRetry(item) }}>Coba lagi</button>}</div></td></tr>)}</tbody></table></div> : <div className="p-12 text-center"><Mail className="mx-auto mb-4 text-brand" size={32} /><h2 className="font-semibold text-ink">Belum ada email</h2><p className="mt-2 text-sm text-body">Email yang masuk antrean akan muncul di sini. Sesuaikan filter jika diperlukan.</p></div>}
    </section>
    {data && <div className="flex flex-wrap items-center justify-between gap-4 text-sm text-body"><span>{data.total} email · Halaman {page} dari {Math.max(1, Math.ceil(data.total / 20))}</span><div className="flex gap-2"><button type="button" className={actionClass} disabled={page === 1} onClick={() => setPage((v) => v - 1)}>Sebelumnya</button><button type="button" className={actionClass} disabled={page * 20 >= data.total} onClick={() => setPage((v) => v + 1)}>Berikutnya</button></div></div>}
    <Modal open={Boolean(preview)} onClose={() => setPreview(null)} title="Detail email" size="wide" trapFocus>{preview && <><p className="mb-4 text-sm text-body">{kinds[preview.kind]} · {preview.recipient} · {statuses[preview.status]}</p>{previewError ? <p role="alert" className="text-sm text-body">{previewError}</p> : previewData ? <><h3 className="mb-4 font-semibold text-ink">{previewData.subject}</h3><iframe title="Preview email pelanggan" sandbox="" srcDoc={previewData.html} className="h-[60dvh] w-full rounded-xl border border-border bg-white" /></> : <p className="text-body">Memuat preview…</p>}</>}</Modal>
    <ConfirmDialog open={Boolean(retry)} onClose={() => setRetry(null)} onConfirm={() => { void confirmRetry() }} title="Coba kirim ulang email?" description={`Email akan dimasukkan kembali ke antrean untuk ${retry?.recipient ?? ''}. Pengingat akan diperiksa kembali sebelum dikirim.`} confirmLabel="Masukkan ke antrean" busy={saving} error={error} />
  </div>
}
