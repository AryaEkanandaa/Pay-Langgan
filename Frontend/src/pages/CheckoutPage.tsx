import { useEffect, useRef, useState, type FormEvent } from 'react'
import { AlertCircle, CheckCircle2, Clock3, CreditCard, ShieldCheck } from 'lucide-react'
import { useParams, useSearchParams } from 'react-router-dom'
import { XenditComponents } from 'xendit-components-web'
import Input from '../components/ui/Input'
import Logo from '../components/ui/Logo'
import SubmitButton from '../components/ui/SubmitButton'
import { ApiError, getPublicCheckout, startPublicCheckout, type PublicCheckout, type StartedCheckout } from '../lib/api'
import { formatCurrency } from '../lib/format'

export default function CheckoutPage() {
  const { token = '' } = useParams()
  const [searchParams] = useSearchParams()
  const returningFromXendit = [...searchParams.keys()].some((key) => key !== 'parent_origin')
  const parentOrigin = searchParams.get('parent_origin')
  const [checkout, setCheckout] = useState<PublicCheckout | null>(null)
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [mobile, setMobile] = useState('')
  const [session, setSession] = useState<StartedCheckout | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isStarting, setIsStarting] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isCardReady, setIsCardReady] = useState(false)
  const [isVerifying, setIsVerifying] = useState(returningFromXendit)
  const [error, setError] = useState<string | null>(null)
  const cardContainer = useRef<HTMLDivElement>(null)
  const componentsRef = useRef<XenditComponents | null>(null)
  const paidNotified = useRef(false)
  const resumedSession = useRef(false)

  useEffect(() => {
    if (checkout?.invoice_status !== 'paid' || !parentOrigin || paidNotified.current || window.parent === window) return
    try {
      const origin = new URL(parentOrigin)
      if (origin.origin !== parentOrigin || (origin.protocol !== 'https:' && origin.hostname !== 'localhost' && origin.hostname !== '127.0.0.1')) return
      window.parent.postMessage({ type: 'paylanggan:checkout-status', status: 'paid' }, parentOrigin)
      paidNotified.current = true
    } catch { /* Ignore invalid parent origin. */ }
  }, [checkout?.invoice_status, parentOrigin])

  useEffect(() => {
    let active = true
    getPublicCheckout(token)
      .then(async (data) => {
        if (!active) return
        setCheckout(data)
        setName(data.customer_name)
        setEmail(data.customer_email ?? '')
        setMobile(data.customer_mobile ?? '')
        if ((returningFromXendit || parentOrigin) && data.invoice_status === 'pending' && data.checkout_status === 'active') {
          try {
            const resumed = await startPublicCheckout(token, {
              name: data.customer_name,
              email: data.customer_email ?? '',
              mobile: data.customer_mobile ?? '',
              parent_origin: parentOrigin ?? undefined,
            })
            if (active) {
              resumedSession.current = true
              setSession(resumed)
              setIsVerifying(false)
            }
          } catch {
            // Payment may have completed before the webhook reaches this backend.
          }
        }
      })
      .catch((err) => {
        if (active) setError(err instanceof ApiError ? 'Tautan pembayaran tidak ditemukan atau sudah tidak berlaku.' : 'Tagihan belum dapat dimuat. Silakan coba lagi.')
      })
      .finally(() => {
        if (active) setIsLoading(false)
      })
    return () => { active = false }
  }, [token, returningFromXendit, parentOrigin])

  useEffect(() => {
    if (!session || isVerifying || !cardContainer.current || checkout?.invoice_status === 'paid') return
    const target = cardContainer.current
    const components = new XenditComponents({
      componentsSdkKey: session.components_sdk_key,
      resume: returningFromXendit || resumedSession.current,
    })
    componentsRef.current = components
    let cardElement: HTMLElement | null = null
    setIsCardReady(false)
    const onInit = () => {
      const card = components.getActiveChannels({ filter: 'CARDS' })[0]
      if (!card) {
        setError('Pembayaran dengan kartu sedang tidak tersedia. Silakan coba lagi nanti.')
        return
      }
      cardElement = components.createChannelComponent(card)
      target.replaceChildren(cardElement)
    }
    const onReady = () => setIsCardReady(true)
    const onNotReady = () => setIsCardReady(false)
    const onBegin = () => setIsSubmitting(true)
    const onEnd = () => setIsSubmitting(false)
    const onComplete = () => {
      setIsSubmitting(false)
      setIsVerifying(true)
      void getPublicCheckout(token).then(setCheckout).catch(() => undefined)
    }
    const onExpired = () => {
      setIsSubmitting(false)
      setError('Waktu pembayaran telah berakhir. Muat ulang halaman untuk memeriksa status tagihan.')
      void getPublicCheckout(token).then(setCheckout).catch(() => undefined)
    }
    const onFatal = () => {
      setIsSubmitting(false)
      setError('Formulir pembayaran belum dapat ditampilkan. Muat ulang halaman dan coba lagi.')
    }
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
  }, [session, token, returningFromXendit, isVerifying, checkout?.invoice_status])

  useEffect(() => {
    if ((!session && !isVerifying) || checkout?.invoice_status === 'paid' || checkout?.checkout_status === 'expired') return
    const timer = window.setInterval(() => {
      void getPublicCheckout(token).then(setCheckout).catch(() => undefined)
    }, 3000)
    return () => window.clearInterval(timer)
  }, [token, session, isVerifying, checkout?.invoice_status, checkout?.checkout_status])

  async function handleStart(event: FormEvent) {
    event.preventDefault()
    if (isStarting) return
    setIsStarting(true)
    setError(null)
    try {
      const started = await startPublicCheckout(token, { name, email, mobile, parent_origin: parentOrigin ?? undefined })
      if (started.mode !== 'COMPONENTS' || !started.components_sdk_key) {
        throw new Error('Pembayaran belum dapat disiapkan. Silakan coba lagi.')
      }
      setSession(started)
      resumedSession.current = false
      setIsVerifying(false)
    } catch (err) {
      setError(err instanceof ApiError ? 'Pembayaran belum dapat dimulai. Periksa kembali data kontak Anda atau coba lagi nanti.' : err instanceof Error ? err.message : 'Pembayaran belum dapat dimulai. Silakan coba lagi.')
    } finally {
      setIsStarting(false)
    }
  }

  function handlePay() {
    if (!componentsRef.current || !isCardReady || isSubmitting) return
    setError(null)
    try {
      componentsRef.current.submit()
    } catch {
      setError('Periksa kembali data kartu lalu coba lagi.')
    }
  }

  const isPaid = checkout?.invoice_status === 'paid'
  const isUnavailable = checkout?.checkout_status === 'expired' || checkout?.checkout_status === 'failed'

  return (
    <main className="min-h-screen bg-page text-ink">
      <header className="border-b border-border bg-white/70 px-5 py-4 backdrop-blur-xl dark:bg-slate-950/70">
        <div className="mx-auto flex max-w-5xl items-center justify-between gap-4">
          <Logo />
          <span className="inline-flex items-center gap-1.5 text-[12px] text-muted">
            <ShieldCheck size={15} aria-hidden="true" /> Pembayaran terlindungi
          </span>
        </div>
      </header>

      <div className="mx-auto grid max-w-5xl gap-10 px-5 py-10 md:grid-cols-[minmax(0,1fr)_360px] md:py-16">
        <section>
          {isLoading ? (
            <p className="text-[14px] text-body">Memuat tagihan...</p>
          ) : error && !checkout ? (
            <div className="flex items-start gap-3 text-red-600"><AlertCircle size={20} /><div><h1 className="font-display text-2xl font-bold">Halaman pembayaran tidak tersedia</h1><p className="mt-2 text-[14px]">{error}</p></div></div>
          ) : isPaid ? (
            <div className="flex items-start gap-4"><CheckCircle2 className="text-emerald-700" size={28} /><div><h1 className="font-display text-3xl font-bold">Pembayaran berhasil</h1><p className="mt-3 text-[15px] text-body">Tagihan {checkout?.invoice_number} telah lunas. Terima kasih atas pembayaran Anda.</p></div></div>
          ) : isUnavailable ? (
            <div className="flex items-start gap-4"><AlertCircle className="text-red-600" size={24} /><div><h1 className="font-display text-2xl font-bold">Waktu pembayaran telah berakhir</h1><p className="mt-3 text-[15px] text-body">Hubungi penyedia layanan untuk mendapatkan tautan pembayaran baru.</p></div></div>
          ) : checkout ? (
            <>
              <p className="text-[12px] font-semibold uppercase text-brand">{checkout.business_name}</p>
              <h1 className="mt-3 font-display text-3xl font-bold">Selesaikan pembayaran</h1>
              {isVerifying ? (
                <div className="mt-6 flex items-start gap-3 text-brand"><Clock3 size={22} /><p>Kami sedang memastikan pembayaran Anda. Status tagihan akan diperbarui secara otomatis.</p></div>
              ) : session ? (
                <div className="mt-8 max-w-xl space-y-5">
                  <div ref={cardContainer} className="checkout-card-fields" aria-label="Formulir pembayaran kartu" />
                  {error && <p role="alert" className="text-[13px] text-red-600">{error}</p>}
                  <SubmitButton type="button" onClick={handlePay} disabled={!isCardReady || isSubmitting} className="gap-2">
                    <CreditCard size={17} aria-hidden="true" />
                    {isSubmitting ? 'Memproses pembayaran...' : `Bayar ${formatCurrency(checkout.amount)}`}
                  </SubmitButton>
                </div>
              ) : (
                <form onSubmit={handleStart} className="mt-8 max-w-xl space-y-5">
                  <p className="text-[15px] text-body">Periksa data kontak Anda sebelum melanjutkan pembayaran.</p>
                  <Input id="checkout-name" label="Nama" value={name} onChange={(event) => setName(event.target.value)} required autoComplete="name" />
                  <Input id="checkout-email" label="Email" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required autoComplete="email" />
                  <Input id="checkout-mobile" label="Nomor ponsel" type="tel" value={mobile} onChange={(event) => setMobile(event.target.value)} placeholder="+628123456789" required autoComplete="tel" />
                  {error && <p role="alert" className="text-[13px] text-red-600">{error}</p>}
                  <SubmitButton type="submit" disabled={isStarting} className="gap-2">
                    <CreditCard size={17} aria-hidden="true" />
                    {isStarting ? 'Menyiapkan pembayaran...' : 'Lanjut ke pembayaran kartu'}
                  </SubmitButton>
                </form>
              )}
            </>
          ) : null}
        </section>

        {checkout && (
          <aside className="border-t border-border pt-7 md:border-l md:border-t-0 md:pl-8 md:pt-0">
            <p className="text-[12px] font-semibold uppercase text-muted">Ringkasan tagihan</p>
            <dl className="mt-5 space-y-4 text-[14px]">
              <div><dt className="text-muted">Nomor tagihan</dt><dd className="mt-1 font-semibold">{checkout.invoice_number}</dd></div>
              <div><dt className="text-muted">Paket</dt><dd className="mt-1 font-semibold">{checkout.plan_name}</dd></div>
              <div><dt className="text-muted">Nama pelanggan</dt><dd className="mt-1 font-semibold">{checkout.customer_name}</dd></div>
              <div className="border-t border-border pt-4"><dt className="text-muted">Total pembayaran</dt><dd className="mt-1 font-display text-2xl font-bold">{formatCurrency(checkout.amount)}</dd></div>
            </dl>
          </aside>
        )}
      </div>
    </main>
  )
}
