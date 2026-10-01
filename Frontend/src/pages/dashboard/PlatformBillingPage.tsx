import { useCallback, useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { XenditComponents, type XenditSubmissionEndEvent } from 'xendit-components-web'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import { useAuth } from '../../context/AuthContext'
import {
  ApiError, cancelPlatformSubscription, getPlatformInvoices, getPlatformPlan,
  getPlatformSubscription, resumePlatformCheckout, startPlatformCheckout,
  type PlatformCheckout, type PlatformInvoice, type PlatformPlan, type PlatformSubscription,
} from '../../lib/api'
import { formatCurrency } from '../../lib/format'

function date(value: string | null | undefined) {
  return value ? new Date(value).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) : '—'
}

const subscriptionStatus: Record<PlatformSubscription['status'], string> = {
  not_subscribed: 'Belum berlangganan',
  pending: 'Menunggu pembayaran',
  active: 'Aktif',
  past_due: 'Pembayaran belum selesai',
  cancelled: 'Perpanjangan dihentikan',
}

export default function PlatformBillingPage() {
  const [searchParams] = useSearchParams()
  const returningFromXendit = searchParams.has('token_request_id')
  const { auth } = useAuth()
  const token = auth?.token
  const [plan, setPlan] = useState<PlatformPlan | null>(null)
  const [subscription, setSubscription] = useState<PlatformSubscription | null>(null)
  const [invoices, setInvoices] = useState<PlatformInvoice[]>([])
  const [consent, setConsent] = useState(false)
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const [session, setSession] = useState<PlatformCheckout | null>(null)
  const [cardReady, setCardReady] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [verifying, setVerifying] = useState(returningFromXendit)
  const cardContainer = useRef<HTMLDivElement>(null)
  const componentsRef = useRef<XenditComponents | null>(null)
  const resumed = useRef(false)

  const refresh = useCallback(async () => {
    if (!token) return
    try {
      const [p, s, i] = await Promise.all([
        getPlatformPlan(token), getPlatformSubscription(token), getPlatformInvoices(token),
      ])
      setPlan(p); setSubscription(s); setInvoices(i)
      setError(null)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memuat tagihan PayLanggan.')
    } finally { setLoading(false) }
  }, [token])

  useEffect(() => { void refresh() }, [refresh])

  useEffect(() => {
    if (!token || !subscription?.checkout_available || resumed.current) return
    resumed.current = true
    void resumePlatformCheckout(token).then((value) => {
      if (value.mode === 'COMPONENTS' && value.components_sdk_key) {
        setSession(value)
        setVerifying(false)
      }
    }).catch(() => { setVerifying(true) })
  }, [token, subscription?.checkout_available])

  useEffect(() => {
    if (!token || (!session && !verifying)) return
    const timer = window.setInterval(() => {
      void Promise.all([getPlatformSubscription(token), getPlatformInvoices(token)]).then(([next, items]) => {
        setSubscription(next)
        setInvoices(items)
        if (next.status === 'active' && !next.pending_invoice) {
          setSession(null)
          setVerifying(false)
          setNotice('Pembayaran berhasil. Langganan PayLanggan aktif.')
        } else if (verifying && !next.checkout_available) {
          setVerifying(false)
          setSession(null)
        }
      }).catch(() => undefined)
    }, 3000)
    return () => window.clearInterval(timer)
  }, [token, session, verifying])

  useEffect(() => {
    if (!session || verifying || !cardContainer.current) return
    const target = cardContainer.current
    const components = new XenditComponents({ componentsSdkKey: session.components_sdk_key, resume: returningFromXendit })
    componentsRef.current = components
    let cardElement: HTMLElement | null = null
    const onInit = () => {
      const card = components.getActiveChannels({ filter: 'CARDS' })[0]
      if (!card) { setError('Pembayaran kartu sedang tidak tersedia.'); return }
      cardElement = components.createChannelComponent(card)
      target.replaceChildren(cardElement)
    }
    const onReady = () => setCardReady(true)
    const onNotReady = () => setCardReady(false)
    const onBegin = () => { setSubmitting(true); setError(null) }
    const onEnd = (event: XenditSubmissionEndEvent) => {
      setSubmitting(false)
      if (event.developerErrorMessage?.type === 'NETWORK_ERROR') {
        setError('Status pembayaran belum dapat dipastikan. Periksa kembali status tagihan sebelum mencoba lagi.')
      } else if (event.developerErrorMessage || event.userErrorMessage?.length) {
        setError('Pembayaran belum berhasil. Periksa data kartu dan coba lagi.')
      }
    }
    const onComplete = () => { setSubmitting(false); setVerifying(true); void refresh() }
    const onExpired = () => { setSubmitting(false); setSession(null); setVerifying(true); void refresh() }
    const onFatal = () => { setSubmitting(false); setError('Formulir kartu belum dapat ditampilkan. Muat ulang halaman dan coba lagi.') }
    components.addEventListener('init', onInit)
    components.addEventListener('submission-ready', onReady)
    components.addEventListener('submission-not-ready', onNotReady)
    components.addEventListener('submission-begin', onBegin)
    components.addEventListener('submission-end', onEnd)
    components.addEventListener('session-complete', onComplete)
    components.addEventListener('session-expired-or-canceled', onExpired)
    components.addEventListener('fatal-error', onFatal)
    return () => {
      components.removeEventListener('init', onInit)
      components.removeEventListener('submission-ready', onReady)
      components.removeEventListener('submission-not-ready', onNotReady)
      components.removeEventListener('submission-begin', onBegin)
      components.removeEventListener('submission-end', onEnd)
      components.removeEventListener('session-complete', onComplete)
      components.removeEventListener('session-expired-or-canceled', onExpired)
      components.removeEventListener('fatal-error', onFatal)
      if (componentsRef.current === components) componentsRef.current = null
      if (cardElement) components.destroyComponent(cardElement)
      target.replaceChildren()
    }
  }, [session, returningFromXendit, verifying, refresh])

  async function checkout() {
    if (!token || !consent) return
    setBusy(true); setError(null)
    try {
      const result = await startPlatformCheckout(token)
      if (result.mode !== 'COMPONENTS' || !result.components_sdk_key) throw new Error('Formulir pembayaran belum tersedia.')
      resumed.current = true
      setSession(result)
      setVerifying(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal membuka checkout.')
    } finally { setBusy(false) }
  }

  async function resumeCheckout() {
    if (!token) return
    setBusy(true); setError(null)
    try {
      const result = await resumePlatformCheckout(token)
      resumed.current = true
      setSession(result)
      setVerifying(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Sesi pembayaran belum dapat dilanjutkan.')
      await refresh()
    } finally { setBusy(false) }
  }

  function pay() {
    if (!componentsRef.current || !cardReady || submitting) return
    try { componentsRef.current.submit() }
    catch { setError('Periksa kembali data kartu Anda.') }
  }

  async function cancel() {
    if (!token || !window.confirm('Hentikan perpanjangan otomatis PayLanggan? Periode yang sudah dibayar tetap tercatat hingga tanggal berakhir.')) return
    setBusy(true); setError(null)
    try {
      await cancelPlatformSubscription(token)
      setNotice('Perpanjangan otomatis telah dihentikan.')
      await refresh()
    } catch (err) { setError(err instanceof ApiError ? err.message : 'Gagal membatalkan langganan.') }
    finally { setBusy(false) }
  }

  if (loading) return <p className="p-6">Memuat tagihan PayLanggan...</p>
  return <div className="space-y-6">
    <DashboardPageHeader title="Tagihan PayLanggan" description="Langganan bisnis Anda untuk menggunakan PayLanggan." />
    {error && <p role="alert" className="rounded-lg bg-red-50 p-4 text-red-700">{error}</p>}
    {notice && <p role="status" className="rounded-lg bg-green-50 p-4 text-green-700">{notice}</p>}
    <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-xl font-semibold">{plan?.name ?? 'PayLanggan Bulanan'}</h2>
      <p className="mt-2 text-2xl font-bold">{formatCurrency(plan?.amount ?? 300000)} <span className="text-base font-normal">/ bulan</span></p>
      <p className="mt-3">Status langganan: <strong>{subscription ? subscriptionStatus[subscription.status] : '—'}</strong></p>
      {subscription?.paid_through && <p>Berlaku sampai: {date(subscription.paid_through)}</p>}
      {subscription?.next_billing_date && <p>Periode tagihan berikutnya: {date(subscription.next_billing_date)}</p>}
      <p>Pembayaran otomatis: {subscription?.auto_renew ? 'Aktif' : 'Belum aktif'}</p>
      {subscription?.legacy_checkout_active && <p className="mt-4 text-sm text-amber-800">Sesi pembayaran lama masih aktif. Tunggu sesi berakhir sebelum membuat pembayaran baru di halaman ini.</p>}
      {subscription?.checkout_available && !session && !verifying && <button type="button" disabled={busy} onClick={resumeCheckout} className="mt-4 rounded-lg bg-blue-700 px-5 py-2 text-white disabled:opacity-50">Lanjutkan pembayaran</button>}
      {verifying && <p role="status" className="mt-4 text-sm text-blue-700">Kami sedang memastikan pembayaran Anda. Status tagihan akan diperbarui otomatis.</p>}
      {session && !verifying && <div className="mt-5 max-w-xl space-y-4">
        <p className="text-sm text-slate-600">Masukkan detail kartu pada formulir pembayaran aman berikut.</p>
        <div ref={cardContainer} className="checkout-card-fields" aria-label="Formulir pembayaran kartu" />
        <button type="button" disabled={!cardReady || submitting} onClick={pay} className="rounded-lg bg-blue-700 px-5 py-2 text-white disabled:opacity-50">{submitting ? 'Memproses pembayaran...' : `Bayar ${formatCurrency(session.amount)}`}</button>
      </div>}
      {(subscription?.status === 'not_subscribed' || subscription?.status === 'pending' || subscription?.status === 'past_due') && !subscription.checkout_available && !subscription.legacy_checkout_active && !session && !verifying && <div className="mt-5 space-y-3">
        <label className="flex items-start gap-3"><input type="checkbox" checked={consent} onChange={e => setConsent(e.target.checked)} className="mt-1" />
          <span>Saya menyetujui pembayaran {formatCurrency(plan?.amount ?? 300000)} sekarang dan pendebitan kartu otomatis sebesar {formatCurrency(plan?.amount ?? 300000)} setiap bulan selama langganan aktif. Saya dapat membatalkan perpanjangan otomatis melalui halaman ini.</span>
        </label>
        <button type="button" disabled={!consent || busy} onClick={checkout} className="rounded-lg bg-blue-700 px-5 py-2 text-white disabled:opacity-50">{busy ? 'Memproses...' : 'Lanjut ke pembayaran'}</button>
      </div>}
      {(subscription?.status === 'active' || subscription?.status === 'past_due') && <button type="button" disabled={busy} onClick={cancel} className="mt-5 rounded-lg border border-red-300 px-4 py-2 text-red-700 disabled:opacity-50">Hentikan perpanjangan otomatis</button>}
    </section>
    <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="mb-4 text-xl font-semibold">Riwayat tagihan</h2>
      {invoices.length === 0 ? <p>Belum ada tagihan.</p> : <div className="overflow-x-auto"><table className="w-full text-left text-sm"><thead><tr><th className="p-2">Periode</th><th className="p-2">Jumlah</th><th className="p-2">Status</th><th className="p-2">Dibayar</th></tr></thead><tbody>{invoices.map(item => <tr key={item.id} className="border-t"><td className="p-2">{date(item.period_start)} – {date(item.period_end)}</td><td className="p-2">{formatCurrency(item.amount)}</td><td className="p-2">{item.status === 'paid' ? 'Lunas' : 'Menunggu pembayaran'}</td><td className="p-2">{date(item.paid_at)}</td></tr>)}</tbody></table></div>}
    </section>
  </div>
}
