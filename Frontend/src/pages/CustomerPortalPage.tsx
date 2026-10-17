import { useEffect, useRef, useState, type FormEvent } from 'react'
import { useLocation, useParams } from 'react-router-dom'
import { ArrowRight, CreditCard, FileText, Mail, RefreshCw } from 'lucide-react'
import Logo from '../components/ui/Logo'
import ThemeToggle from '../components/ui/ThemeToggle'
import Input from '../components/ui/Input'
import SubmitButton from '../components/ui/SubmitButton'
import StatusBadge from '../components/ui/StatusBadge'
import Modal from '../components/ui/Modal'
import { portalRequest, PortalError, type PortalOverview, type PortalInvoice, type PortalInvoiceDetail } from '../lib/portalApi'
import { formatCurrency, subscriptionStatusLabel } from '../lib/format'

const buttonClass = 'inline-flex items-center justify-center gap-2 rounded-full border border-border px-4 py-2.5 text-sm font-semibold text-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand disabled:opacity-50'
const invoiceLabels: Record<string, string> = { pending: 'Menunggu pembayaran', paid: 'Lunas', failed: 'Gagal', cancelled: 'Dibatalkan' }
function date(value: string | null) { return value ? new Date(value).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) : '—' }
function storedSession(site: string) { try { return sessionStorage.getItem(`paylanggan:portal:${site}`) ?? '' } catch { return '' } }

export default function CustomerPortalPage() {
  const { site = '' } = useParams()
  const location = useLocation()
  const [name, setName] = useState('')
  const [token, setToken] = useState(() => storedSession(site))
  const [magic, setMagic] = useState(() => new URLSearchParams(window.location.hash.slice(1)).get('token') ?? '')
  const [email, setEmail] = useState('')
  const [data, setData] = useState<PortalOverview | null>(null)
  const [subPage, setSubPage] = useState(1)
  const [invPage, setInvPage] = useState(1)
  const [revision, setRevision] = useState(0)
  const [busy, setBusy] = useState(false)
  const [loading, setLoading] = useState(true)
  const [unavailable, setUnavailable] = useState(false)
  const [error, setError] = useState('')
  const [notice, setNotice] = useState('')
  const [detail, setDetail] = useState<PortalInvoice | null>(null)
  const [detailData, setDetailData] = useState<PortalInvoiceDetail | null>(null)
  const [detailError, setDetailError] = useState('')
  const mutation = useRef<AbortController | null>(null)
  function saveSession(value: string) {
    try { if (value) sessionStorage.setItem(`paylanggan:portal:${site}`, value); else sessionStorage.removeItem(`paylanggan:portal:${site}`) } catch { /* Memory-only session when storage is unavailable. */ }
    setToken(value)
    if (!value) { setData(null); setDetail(null) }
  }
  useEffect(() => {
    const meta = document.createElement('meta'); meta.name = 'referrer'; meta.content = 'no-referrer'; document.head.append(meta)
    return () => { meta.remove(); mutation.current?.abort() }
  }, [])
  useEffect(() => {
    // Also handle email links opened into an already mounted portal tab.
    const value = new URLSearchParams(location.hash.slice(1)).get('token') ?? ''
    if (value) setMagic(value)
    if (window.location.hash) window.history.replaceState(window.history.state, '', window.location.pathname)
  }, [location.hash])
  useEffect(() => {
    const controller = new AbortController()
    setLoading(true); setUnavailable(false)
    void portalRequest<{ name: string }>(site, '', { signal: controller.signal }).then((result) => { if (!controller.signal.aborted) setName(result.name) }).catch((err: unknown) => { if (!controller.signal.aborted) { setError(err instanceof Error ? err.message : 'Portal tidak tersedia.'); setUnavailable(true) } }).finally(() => { if (!controller.signal.aborted) setLoading(false) })
    return () => controller.abort()
  }, [site])
  useEffect(() => {
    if (!token) return
    const controller = new AbortController(); let inFlight = false
    const load = async () => {
      if (inFlight || controller.signal.aborted) return
      inFlight = true
      try { const result = await portalRequest<PortalOverview>(site, `/overview?subscription_page=${subPage}&invoice_page=${invPage}`, { token, signal: controller.signal }); if (!controller.signal.aborted) { setData(result); setError('') } }
      catch (err) { if (!controller.signal.aborted) { setError(err instanceof Error ? err.message : 'Gagal memuat portal.'); if (err instanceof PortalError && err.status === 401) { try { sessionStorage.removeItem(`paylanggan:portal:${site}`) } catch { /* Storage unavailable. */ } setToken(''); setData(null); setDetail(null) } } }
      finally { inFlight = false }
    }
    void load()
    const timer = window.setInterval(() => { if (!document.hidden) void load() }, 30000)
    const focus = () => { void load() }; window.addEventListener('focus', focus)
    return () => { controller.abort(); window.clearInterval(timer); window.removeEventListener('focus', focus) }
  }, [site, token, subPage, invPage, revision])
  useEffect(() => {
    if (!detail || !token) return
    const controller = new AbortController(); setDetailData(null); setDetailError('')
    void portalRequest<PortalInvoiceDetail>(site, `/invoices/${detail.id}`, { token, signal: controller.signal }).then((result) => { if (!controller.signal.aborted) setDetailData(result) }).catch((err: unknown) => { if (!controller.signal.aborted) setDetailError(err instanceof Error ? err.message : 'Gagal memuat invoice.') })
    return () => controller.abort()
  }, [detail, site, token])
  async function requestLink(event: FormEvent) {
    event.preventDefault(); if (busy) return
    const controller = new AbortController(); mutation.current = controller; setBusy(true); setError(''); setNotice('')
    try { await portalRequest<null>(site, '/request-link', { body: { email }, signal: controller.signal }); if (!controller.signal.aborted) setNotice('Jika email terdaftar pada bisnis ini, tautan masuk akan dikirim. Periksa inbox dan folder spam.') }
    catch (err) { if (!controller.signal.aborted) setError(err instanceof Error ? err.message : 'Gagal meminta tautan.') }
    finally { if (!controller.signal.aborted) setBusy(false) }
  }
  async function login() {
    if (busy) return
    const controller = new AbortController(); mutation.current = controller; setBusy(true); setError('')
    try { const result = await portalRequest<{ token: string }>(site, '/exchange', { body: { token: magic }, signal: controller.signal }); if (!controller.signal.aborted) { saveSession(result.token); setMagic(''); setNotice('Berhasil masuk ke portal pelanggan.') } }
    catch (err) { if (!controller.signal.aborted) { setError(err instanceof Error ? err.message : 'Tautan tidak valid.'); setMagic('') } }
    finally { if (!controller.signal.aborted) setBusy(false) }
  }
  async function logout() {
    if (busy) return
    const controller = new AbortController(); mutation.current = controller; setBusy(true); setError('')
    try { await portalRequest<null>(site, '/logout', { token, body: {}, signal: controller.signal }); if (!controller.signal.aborted) { saveSession(''); setNotice('Anda sudah keluar.') } }
    catch (err) { if (!controller.signal.aborted) { if (err instanceof PortalError && err.status === 401) saveSession(''); else setError('Belum dapat keluar dari server. Periksa koneksi dan coba lagi.') } }
    finally { if (!controller.signal.aborted) setBusy(false) }
  }
  async function pay(invoice: PortalInvoice) {
    if (busy) return
    const controller = new AbortController(); mutation.current = controller; setBusy(true); setError(''); setDetailError('')
    try {
      const link = await portalRequest<{ checkout_url: string }>(site, `/invoices/${invoice.id}/checkout`, { token, body: {}, signal: controller.signal })
      const target = new URL(link.checkout_url)
      if (!/^\/checkout\/[A-Za-z0-9_-]+$/.test(target.pathname) || (target.protocol !== 'https:' && !(target.protocol === 'http:' && ['localhost', '127.0.0.1'].includes(target.hostname)))) throw new Error('Alamat checkout tidak valid.')
      target.searchParams.set('portal', site)
      if (!controller.signal.aborted) window.location.assign(target.toString())
    } catch (err) { if (!controller.signal.aborted) { const message = err instanceof Error ? err.message : 'Pembayaran belum dapat dilanjutkan.'; setError(message); setDetailError(message) } }
    finally { if (!controller.signal.aborted) setBusy(false) }
  }
  return <div className="min-h-screen px-6 py-6 sm:px-10">
    <header className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-4"><Logo /><div className="flex items-center gap-3"><ThemeToggle />{token && <button type="button" className={buttonClass} disabled={busy} onClick={() => { void logout() }}>Keluar</button>}</div></header>
    <main className={`mx-auto mt-10 ${token && !magic ? 'max-w-6xl' : 'max-w-lg'} space-y-6`}>
      <div><p className="text-sm font-semibold text-brand dark:text-blue-300">{name || 'PayLanggan'}</p><h1 className="mt-2 font-display text-3xl font-bold text-ink">Portal pelanggan</h1><p className="mt-2 text-sm text-body">Lihat langganan, invoice, dan lanjutkan pembayaran Anda.</p></div>
      {error && <p role="alert" className="rounded-xl border border-red-300 bg-red-50 p-4 text-sm text-red-800 dark:bg-red-950/40 dark:text-red-200">{error}</p>}
      {notice && <p role="status" aria-live="polite" className="rounded-xl border border-border p-4 text-sm text-body">{notice}</p>}
      {loading ? <p className="text-body">Memuat portal…</p> : unavailable ? <p className="text-body">Hubungi bisnis untuk mendapatkan tautan portal yang aktif.</p> : magic ? <section className="glass space-y-5 rounded-2xl p-6"><h2 className="text-xl font-semibold text-ink">Masuk dengan tautan email</h2><p className="text-sm text-body">Lanjutkan untuk membuka langganan dan invoice Anda. Tautan hanya dapat digunakan satu kali.</p><SubmitButton type="button" disabled={busy} className="w-full" onClick={() => { void login() }}>{busy ? 'Memverifikasi…' : 'Masuk ke portal'}</SubmitButton></section> : !token ? <form onSubmit={requestLink} className="glass space-y-5 rounded-2xl p-6"><Mail className="text-brand" size={28} /><h2 className="text-xl font-semibold text-ink">Masuk tanpa password</h2><p className="text-sm text-body">Gunakan email yang Anda isi saat checkout. Kami akan mengirim tautan masuk yang berlaku selama 15 menit.</p><Input id="portal-email" type="email" label="Email pelanggan" value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="email" maxLength={100} required disabled={busy} /><SubmitButton type="submit" className="w-full" disabled={busy}>{busy ? 'Meminta tautan…' : 'Kirim tautan masuk'}</SubmitButton></form> : !data ? <p className="text-body">Memuat langganan dan invoice…</p> : <>
        <div className="flex flex-wrap items-center justify-between gap-4"><p className="break-all text-sm text-body">Masuk sebagai <strong className="text-ink">{data.email}</strong></p><button type="button" className={buttonClass} onClick={() => setRevision((v) => v + 1)}><RefreshCw size={16} />Perbarui status</button></div>
        <section className="space-y-4" aria-labelledby="portal-subscriptions"><h2 id="portal-subscriptions" className="flex items-center gap-2 text-xl font-semibold text-ink"><CreditCard size={20} />Langganan Anda</h2>{data.subscriptions.length ? <div className="grid gap-4 md:grid-cols-2">{data.subscriptions.map((sub) => <article key={sub.id} className="glass rounded-2xl p-6"><div className="flex flex-wrap justify-between gap-3"><h3 className="font-semibold text-ink">{sub.plan}</h3><StatusBadge tone={sub.status === 'active' ? 'success' : 'neutral'}>{subscriptionStatusLabel[sub.status] ?? sub.status}</StatusBadge></div><p className="mt-4 text-lg font-bold text-ink">{formatCurrency(sub.price)} <span className="text-sm font-normal text-body">/ {sub.cycle === 'yearly' ? 'tahun' : 'bulan'}</span></p><p className="mt-1 text-xs text-body">Harga dasar plan. Rincian add-on dan diskon mengikuti invoice.</p><dl className="mt-5 grid grid-cols-2 gap-4 text-sm"><div><dt className="text-body">Mulai</dt><dd className="mt-1 text-ink">{date(sub.start_date)}</dd></div><div><dt className="text-body">{sub.status === 'trial' ? 'Trial berakhir' : 'Tagihan berikutnya'}</dt><dd className="mt-1 text-ink">{date(sub.status === 'trial' ? sub.trial_ends_at : sub.status === 'active' ? sub.next_billing_date : null)}</dd></div></dl></article>)}</div> : <p className="glass rounded-2xl p-6 text-sm text-body">Belum ada langganan untuk email ini.</p>}<Pagination page={subPage} total={data.subscription_total} setPage={setSubPage} label="langganan" /></section>
        <section className="space-y-4" aria-labelledby="portal-invoices"><h2 id="portal-invoices" className="flex items-center gap-2 text-xl font-semibold text-ink"><FileText size={20} />Invoice Anda</h2>{data.invoices.length ? <div className="space-y-3">{data.invoices.map((inv) => <article key={inv.id} className="glass flex flex-wrap items-center justify-between gap-5 rounded-2xl p-6"><div className="min-w-0"><p className="break-all font-semibold text-ink">{inv.invoice_number}</p><p className="mt-1 text-sm text-body">{inv.plan_name} · Jatuh tempo {date(inv.due_date)}</p><div className="mt-3"><StatusBadge tone={inv.status === 'paid' ? 'success' : 'neutral'}>{invoiceLabels[inv.status] ?? inv.status}</StatusBadge></div></div><div className="flex flex-wrap items-center gap-3"><p className="mr-2 font-bold text-ink">{formatCurrency(inv.amount)}</p><button type="button" className={buttonClass} onClick={() => setDetail(inv)}>Detail invoice</button>{inv.status === 'pending' && <SubmitButton type="button" disabled={busy} onClick={() => { void pay(inv) }}>Lanjutkan pembayaran<ArrowRight size={16} className="ml-2 inline" /></SubmitButton>}</div></article>)}</div> : <p className="glass rounded-2xl p-6 text-sm text-body">Belum ada invoice untuk email ini.</p>}<Pagination page={invPage} total={data.invoice_total} setPage={setInvPage} label="invoice" /></section>
      </>}
      <p className="pb-6 text-center text-xs text-body">Dikelola melalui PayLanggan. Hubungi {name || 'bisnis'} untuk bantuan langganan.</p>
    </main>
    <Modal open={Boolean(detail)} onClose={() => setDetail(null)} title="Detail invoice" trapFocus>{detailError && <p role="alert" className="my-4 text-sm text-red-700 dark:text-red-300">{detailError}</p>}{detailData ? <div className="mt-5 space-y-5"><div className="flex flex-wrap justify-between gap-3"><p className="break-all font-semibold text-ink">{detailData.invoice.invoice_number}</p><StatusBadge tone={detailData.invoice.status === 'paid' ? 'success' : 'neutral'}>{invoiceLabels[detailData.invoice.status] ?? detailData.invoice.status}</StatusBadge></div><p className="text-sm text-body">Jatuh tempo {date(detailData.invoice.due_date)} · Dibayar {date(detailData.invoice.paid_at)}</p><ul className="divide-y divide-border">{detailData.items.map((item) => <li key={item.id} className="flex justify-between gap-4 py-3 text-sm"><span className="text-body">{item.description} × {item.quantity}</span><span className="shrink-0 text-ink">{formatCurrency(item.subtotal)}</span></li>)}</ul><div className="flex justify-between border-t border-border pt-4 font-bold text-ink"><span>Total</span><span>{formatCurrency(detailData.invoice.amount)}</span></div>{detailData.invoice.status === 'pending' && <SubmitButton type="button" disabled={busy} className="w-full" onClick={() => { void pay(detailData.invoice) }}>Lanjutkan pembayaran</SubmitButton>}</div> : !detailError && <p className="mt-4 text-body">Memuat invoice…</p>}</Modal>
  </div>
}
function Pagination({ page, total, setPage, label }: { page: number; total: number; setPage: (value: number) => void; label: string }) {
  if (total <= 20) return null
  return <nav aria-label={`Halaman ${label}`} className="flex flex-wrap items-center justify-between gap-3 text-sm text-body"><span>Halaman {page} dari {Math.ceil(total / 20)}</span><div className="flex gap-2"><button type="button" className={buttonClass} disabled={page === 1} onClick={() => setPage(page - 1)}>Sebelumnya</button><button type="button" className={buttonClass} disabled={page * 20 >= total} onClick={() => setPage(page + 1)}>Berikutnya</button></div></nav>
}
