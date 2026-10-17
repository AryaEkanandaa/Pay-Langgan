# Revisi modal checkout PayLanggan

Stepper terhubung; area form saja yang scroll; footer aksi tetap terlihat. Tombol submit tetap terhubung ke form asli melalui atribut form sehingga validasi native dan handler pembayaran dipertahankan. Tombol Midtrans dirender ke footer dengan React portal. Tidak ada dependency atau endpoint baru.

Modal drop-in dibatasi min(90dvh, 880px). Di mobile logo diletakkan di bawah kontrol header supaya tidak tumpang tindih. Pada tinggi rendah isi dapat scroll tanpa scrollbar, dengan fade saat masih ada isi di bawah. Form Midtrans normal muat tanpa scroll pada desktop setinggi 800px; error dan halaman autentikasi bank dapat membutuhkan scroll. Isi iframe milik penyedia/bank tetap mengikuti penyedia.

Pemeriksaan: TypeScript, lint, build, browser dengan API/SDK simulasi; semua langkah, back/next, validasi form, error sesi, footer Xendit/Midtrans, autentikasi inline, Tab/Escape, reduced motion, dark mode, lebar 360/768/1440 dan tinggi 600/800/1080. Tidak menjalankan transaksi nyata.

## Kode lengkap komponen yang berubah

### Frontend/src/components/payment/CheckoutStepper.tsx

```tsx
import { Check } from 'lucide-react'

export default function CheckoutStepper({ currentStep }: { currentStep: number }) {
  return (
    <div className="relative mx-2">
    <div aria-hidden="true" className="checkout-step-connectors">
      {[1, 2, 3].map(number => <span key={number} className="checkout-step-line"><span className={currentStep > number ? 'is-complete' : ''} /></span>)}
    </div>
    <ol aria-label="Langkah checkout" className="checkout-stepper relative grid grid-cols-4 py-3">
      {['Paket', 'Pelanggan', 'Kartu', 'Selesai'].map((label, index) => {
        const number = index + 1
        const done = currentStep > number
        const active = currentStep === number
        return (
          <li key={label} aria-current={active ? 'step' : undefined} className="relative text-center">
            <span className={`checkout-step-circle relative z-10 mx-auto flex size-8 items-center justify-center rounded-full text-xs font-bold ${done || active ? 'bg-brand text-white' : 'border border-slate-400 bg-white text-slate-600 dark:bg-slate-900 dark:text-slate-300'} ${active ? 'checkout-step-active' : ''}`}>
              {done ? <Check size={16} aria-hidden="true" /> : number}
            </span>
            <span className={`mt-2 block text-xs ${active ? 'font-bold text-brand dark:text-blue-300' : done ? 'font-semibold text-ink' : 'font-medium text-slate-600 dark:text-slate-300'}`}>{label}</span>
            <span className="sr-only">{done ? 'Selesai' : active ? 'Sedang berlangsung' : 'Belum selesai'}</span>
          </li>
        )
      })}
    </ol>
    </div>
  )
}

```

### Frontend/src/components/payment/CheckoutScrollArea.tsx

```tsx
import { useEffect, useRef, useState, type ReactNode } from 'react'

export default function CheckoutScrollArea({ children, step }: { children: ReactNode; step: number }) {
  const container = useRef<HTMLDivElement>(null)
  const content = useRef<HTMLDivElement>(null)
  const [hasMore, setHasMore] = useState(false)
  useEffect(() => { container.current?.scrollTo({ top: 0 }) }, [step])
  useEffect(() => {
    const node = container.current
    const inner = content.current
    if (!node || !inner) return
    const update = () => setHasMore(node.scrollHeight - node.clientHeight - node.scrollTop > 2)
    const observer = new ResizeObserver(update)
    observer.observe(node)
    observer.observe(inner)
    node.addEventListener('scroll', update, { passive: true })
    update()
    return () => { observer.disconnect(); node.removeEventListener('scroll', update) }
  }, [])
  return <div className="relative min-h-0 flex-1">
    <div ref={container} tabIndex={0} aria-label="Isi checkout" className="checkout-scroll h-full overflow-y-auto focus-visible:outline focus-visible:outline-2 focus-visible:outline-inset focus-visible:outline-brand">
      <div ref={content} className="mx-auto w-full max-w-lg px-6 pb-1 pt-2">{children}</div>
    </div>
    {hasMore && <div aria-hidden="true" className="checkout-scroll-fade pointer-events-none absolute inset-x-0 bottom-0 h-5" />}
  </div>
}

```

### Frontend/src/pages/CheckoutPage.tsx

```tsx
import { useEffect, useRef, useState, type FormEvent } from 'react'
import { flushSync } from 'react-dom'
import { AlertCircle, CheckCircle2, Clock3, CreditCard, ShieldCheck, ArrowLeft, ArrowRight } from 'lucide-react'
import { useParams, useSearchParams } from 'react-router-dom'
import { XenditComponents, type XenditSubmissionEndEvent } from 'xendit-components-web'
import Input from '../components/ui/Input'
import Logo from '../components/ui/Logo'
import SubmitButton from '../components/ui/SubmitButton'
import CheckoutStepper from '../components/payment/CheckoutStepper'
import CheckoutScrollArea from '../components/payment/CheckoutScrollArea'
import MidtransCardForm from '../components/payment/MidtransCardForm'
import { ApiError, getPublicCheckout, verifyMidtransCard, startPublicCheckout, getEmbedPlan, createEmbedCheckout, type PublicCheckout, type StartedCheckout } from '../lib/api'
import { formatCurrency } from '../lib/format'

declare global {
  interface Window {
    snap?: {
      embed: (token: string, callbacks: {
        embedId: string
        enabledPayments: string[]
        selectedPaymentType: string
        language: 'id'
        onSuccess?: (result: unknown) => void
        onPending?: (result: unknown) => void
        onError?: (result: unknown) => void
        onClose?: () => void
      }) => void
    }
  }
}

function isAllowedParentOrigin(value: string): boolean {
  try {
    const origin = new URL(value)
    return origin.origin === value && (origin.protocol === 'https:' ||
      (origin.protocol === 'http:' && (origin.hostname === 'localhost' || origin.hostname === '127.0.0.1')))
  } catch {
    return false
  }
}

export default function CheckoutPage() {
  const { token: routeToken = '' } = useParams()
  const [token, setToken] = useState(routeToken)
  const [step, setStep] = useState(1)
  const [cardActions, setCardActions] = useState<HTMLDivElement | null>(null)
  const [preview, setPreview] = useState<{ business_name: string; plan_name: string; billing_cycle: string; amount: number } | null>(null)
  const [searchParams] = useSearchParams()
  const returningFromXendit = searchParams.has('token_request_id')
  const parentOriginParam = searchParams.get('parent_origin')
  const parentOriginKey = `paylanggan:parent-origin:${token}`
  let parentOrigin = parentOriginParam
  if (!parentOrigin && window.parent !== window) {
    try { parentOrigin = window.sessionStorage.getItem(parentOriginKey) } catch { /* Storage may be unavailable. */ }
  }
  const [checkout, setCheckout] = useState<PublicCheckout | null>(null)
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [mobile, setMobile] = useState('')
  const [session, setSession] = useState<StartedCheckout | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isStarting, setIsStarting] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isCardReady, setIsCardReady] = useState(false)
  const [isSnapReady, setIsSnapReady] = useState(false)
  const [isVerifying, setIsVerifying] = useState(returningFromXendit)
  const [error, setError] = useState<string | null>(null)
  const [isAuthenticating, setIsAuthenticating] = useState(false)
  const actionContainer = useRef<HTMLDivElement>(null)
  const cardContainer = useRef<HTMLDivElement>(null)
  const componentsRef = useRef<XenditComponents | null>(null)
  const paidNotified = useRef(false)
  const isPaid = checkout?.invoice_status === 'paid'
  const isUnavailable = checkout?.checkout_status === 'expired' || checkout?.checkout_status === 'failed'

  useEffect(() => {
    if (!parentOrigin || window.parent === window || !isAllowedParentOrigin(parentOrigin)) return
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') window.parent.postMessage({ type: 'paylanggan:checkout-close' }, parentOrigin)
    }
    window.addEventListener('keydown', handleEscape)
    return () => window.removeEventListener('keydown', handleEscape)
  }, [parentOrigin])

  useEffect(() => {
    if (!parentOriginParam || window.parent === window || !isAllowedParentOrigin(parentOriginParam)) return
    try {
      window.sessionStorage.setItem(parentOriginKey, parentOriginParam)
    } catch { /* Storage may be unavailable. */ }
  }, [parentOriginParam, parentOriginKey])

  useEffect(() => {
    if (checkout?.invoice_status !== 'paid' || !parentOrigin || paidNotified.current || window.parent === window || !isAllowedParentOrigin(parentOrigin)) return
    try {
      window.parent.postMessage({ type: 'paylanggan:checkout-status', status: 'paid' }, parentOrigin)
      paidNotified.current = true
      try { window.sessionStorage.removeItem(parentOriginKey) } catch { /* Storage may be unavailable. */ }
    } catch { /* Ignore an unavailable parent window. */ }
  }, [checkout?.invoice_status, parentOrigin, parentOriginKey])

  useEffect(() => {
    let active = true
    if (!token) {
      getEmbedPlan(searchParams.get('site') ?? '', searchParams.get('plan') ?? '')
        .then((data) => { if (active) setPreview(data) })
        .catch(() => { if (active) setError('Paket tidak ditemukan atau sudah tidak tersedia.') })
        .finally(() => { if (active) setIsLoading(false) })
      return () => { active = false }
    }
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
              setSession(resumed)
              setStep(3)
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
  }, [token, returningFromXendit, parentOrigin, searchParams])

  useEffect(() => {
    if (!session || session.provider !== 'xendit' || !session.components_sdk_key || isVerifying || !cardContainer.current || checkout?.invoice_status === 'paid') return
    const target = cardContainer.current
    const components = new XenditComponents({
      componentsSdkKey: session.components_sdk_key,
      // Xendit only accepts resume after redirecting back with its query parameters.
      // Reopening an active checkout is a regular SDK initialization.
      resume: returningFromXendit,
    })
    componentsRef.current = components
    const actionTarget = actionContainer.current
    let actionElement: HTMLElement | null = null
    const onActionBegin = () => flushSync(() => setIsAuthenticating(true))
    const onActionEnd = () => setIsAuthenticating(false)
    components.addEventListener('action-begin', onActionBegin)
    components.addEventListener('action-end', onActionEnd)
    let cardElement: HTMLElement | null = null
    setIsCardReady(false)
    const onInit = () => {
      // Action components require initialized session data. Create before submit
      // so 3DS uses this inline container instead of the SDK's default modal.
      actionElement = components.createActionContainerComponent()
      actionElement.style.minHeight = '440px'
      actionElement.style.width = '100%'
      actionTarget?.replaceChildren(actionElement)
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
    const onBegin = () => {
      setError(null)
      setIsSubmitting(true)
    }
    const onEnd = (event: XenditSubmissionEndEvent) => {
      setIsAuthenticating(false)
      setIsSubmitting(false)
      if (event.developerErrorMessage?.type === 'NETWORK_ERROR') {
        setError('Status pembayaran belum dapat dipastikan. Tunggu pembaruan tagihan sebelum mencoba lagi.')
      } else if (event.developerErrorMessage || event.userErrorMessage?.length) {
        setError('Percobaan pembayaran belum berhasil. Periksa pesan dari penyedia pembayaran, lalu coba lagi.')
      } else if (event.reason === 'ACTION_ABORTED' || event.reason === 'REQUEST_ABORTED') {
        setError('Pembayaran dibatalkan. Anda dapat mencoba lagi.')
      }
    }
    const onComplete = () => {
      setIsSubmitting(false)
      setError(null)
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
      components.removeEventListener('action-begin', onActionBegin)
      components.removeEventListener('action-end', onActionEnd)
      if (actionElement) components.destroyComponent(actionElement)
      actionTarget?.replaceChildren()
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
    if (session?.provider !== 'midtrans' || session.mode !== 'SNAP' || !session.snap_script_url || !session.snap_client_key) {
      setIsSnapReady(false)
      return
    }
    const scriptId = 'paylanggan-midtrans-snap'
    const existing = document.getElementById(scriptId) as HTMLScriptElement | null
    if (existing && existing.src === session.snap_script_url && window.snap) {
      setIsSnapReady(true)
      return
    }
    existing?.remove()
    window.snap = undefined
    setIsSnapReady(false)
    const script = document.createElement('script')
    script.id = scriptId
    script.src = session.snap_script_url
    script.async = true
    script.setAttribute('data-client-key', session.snap_client_key)
    script.onload = () => setIsSnapReady(true)
    script.onerror = () => setError('Formulir pembayaran belum dapat dimuat. Periksa koneksi lalu coba lagi.')
    document.head.appendChild(script)
    return () => {
      script.onload = null
      script.onerror = null
    }
  }, [session?.provider, session?.mode, session?.snap_script_url, session?.snap_client_key])

  useEffect(() => {
    if ((!session && !isVerifying) || checkout?.invoice_status === 'paid' || checkout?.checkout_status === 'expired') return
    const timer = window.setInterval(() => {
      void getPublicCheckout(token).then(setCheckout).catch(() => undefined)
    }, 3000)
    return () => window.clearInterval(timer)
  }, [token, session, isVerifying, checkout?.invoice_status, checkout?.checkout_status])

  useEffect(() => {
    if (!isVerifying || session?.provider !== 'midtrans' || session.mode !== 'CORE' || checkout?.invoice_status === 'paid' || checkout?.checkout_status === 'expired') return
    let active = true
    let timer: number | undefined
    let attempts = 0
    const verify = async () => {
      attempts += 1
      try {
        const latest = await verifyMidtransCard(token)
        if (!active) return
        setCheckout(latest)
        if (latest.invoice_status === 'paid' || latest.checkout_status === 'expired' || latest.checkout_status === 'failed') return
      } catch {
        // Keep database polling and the worker as fallback during network errors.
      }
      if (active && attempts < 6) timer = window.setTimeout(() => { void verify() }, 5000)
    }
    void verify()
    return () => { active = false; window.clearTimeout(timer) }
  }, [token, session?.provider, session?.mode, isVerifying, checkout?.invoice_status, checkout?.checkout_status])

  async function handleStart(event: FormEvent) {
    event.preventDefault()
    if (isStarting) return
    if (session) { setStep(3); return }
    setIsStarting(true)
    setError(null)
    try {
      let paymentToken = token
      if (!paymentToken) {
        const order = await createEmbedCheckout(searchParams.get('site') ?? '', searchParams.get('plan') ?? '', { name, email, mobile })
        const url = new URL(order.checkout_url)
        paymentToken = url.pathname.split('/')[2]
        if (!paymentToken) throw new Error('Tautan pembayaran belum tersedia.')
        setToken(paymentToken)
        setCheckout(await getPublicCheckout(paymentToken))
        // Keep the actual checkout link for a reload and the full-page control.
        const nextURL = new URL(window.location.href)
        nextURL.pathname = `/checkout/${paymentToken}`
        nextURL.searchParams.delete('site')
        nextURL.searchParams.delete('plan')
        window.history.replaceState(null, '', nextURL)
        if (parentOrigin && isAllowedParentOrigin(parentOrigin)) window.parent.postMessage({ type: 'paylanggan:checkout-link', checkoutUrl: order.checkout_url, externalOrderId: order.external_order_id }, parentOrigin)
      }
      const started = await startPublicCheckout(paymentToken, { name, email, mobile, parent_origin: parentOrigin ?? undefined })
      if (started.provider === 'midtrans' ?
        (started.mode === 'CORE' ? !started.client_key || !started.environment : started.mode !== 'SNAP' || !started.snap_token || !started.snap_client_key || !started.snap_script_url) :
        started.mode !== 'COMPONENTS' || !started.components_sdk_key) {
        throw new Error('Pembayaran belum dapat disiapkan. Silakan coba lagi.')
      }
      setSession(started)
      setStep(3)
      setIsVerifying(false)
    } catch (err) {
      // A webhook may have completed the invoice while the session was loading.
      if (token) {
        try {
          const latest = await getPublicCheckout(token)
          setCheckout(latest)
          if (latest.invoice_status === 'paid') { setError(null); return }
        } catch { /* Keep the original error if status cannot be retrieved. */ }
      }
      setError(err instanceof ApiError ? 'Sesi pembayaran belum dapat disiapkan. Silakan coba lagi setelah beberapa saat.' : err instanceof Error ? err.message : 'Pembayaran belum dapat dimulai. Silakan coba lagi.')
    } finally {
      setIsStarting(false)
    }
  }

  useEffect(() => {
    if (session?.provider === 'midtrans' && session.mode === 'SNAP' && isSnapReady && session.snap_token && window.snap &&
      !isVerifying && !isPaid && !isUnavailable) {
      try {
        window.snap.embed(session.snap_token, {
          embedId: 'paylanggan-midtrans-card',
          enabledPayments: ['credit_card'],
          selectedPaymentType: 'credit_card',
          language: 'id',
          onSuccess: () => {
            setIsSubmitting(false)
            setIsVerifying(true)
            void getPublicCheckout(token).then(setCheckout).catch(() => undefined)
          },
          onPending: () => {
            setIsSubmitting(false)
            setError('Pembayaran sedang diproses. Status tagihan akan diperbarui setelah konfirmasi diterima.')
          },
          onError: () => {
            setIsSubmitting(false)
            setError('Pembayaran belum berhasil. Periksa metode pembayaran dan coba lagi.')
          },
          onClose: () => setIsSubmitting(false),
        })
      } catch {
        setIsSubmitting(false)
        setError('Formulir kartu belum dapat dibuka. Muat ulang halaman lalu coba lagi.')
      }
    }
  }, [session?.provider, session?.mode, session?.snap_token, isSnapReady, isVerifying, isPaid, isUnavailable, token])

  function handlePay() {
    if (!componentsRef.current || !isCardReady || isSubmitting) return
    setError(null)
    try {
      // The SDK checks the action container's dimensions synchronously.
      // Make it visible before submit so it cannot fall back to a modal.
      flushSync(() => setIsAuthenticating(true))
      componentsRef.current.submit()
    } catch {
      setIsAuthenticating(false)
      setError('Periksa kembali data kartu lalu coba lagi.')
    }
  }

  const details = checkout ?? preview
  const currentStep = isPaid ? 4 : step
  const backButton = <button type="button" disabled={isStarting || isSubmitting || isAuthenticating || isVerifying} onClick={() => setStep((value) => Math.max(1, value - 1))} className="inline-flex min-h-11 items-center gap-2 rounded-xl border border-border px-4 text-sm font-semibold disabled:opacity-40"><ArrowLeft size={16} />Kembali</button>
  return (
    <main className="checkout-shell flex h-dvh min-h-0 flex-col overflow-hidden bg-page text-ink">
      <header className="checkout-header relative z-20 shrink-0 border-b border-border bg-white/80 px-6 py-4 dark:bg-slate-900">
        <div className="mx-auto max-w-lg"><Logo /></div>
      </header>
      <div className="checkout-stepper-header z-10 shrink-0"><div className="mx-auto max-w-lg px-6"><CheckoutStepper currentStep={currentStep} /></div></div>
      <CheckoutScrollArea step={currentStep}>
        {isLoading ? <p role="status">Memuat informasi pembayaran...</p> : error && !details ? <p role="alert" className="text-red-600">{error}</p> : isPaid ? (
          <section className="space-y-5 py-6 text-center"><CheckCircle2 className="mx-auto text-emerald-600" size={48} /><h1 className="font-display text-2xl font-bold">Pembayaran berhasil</h1><p className="text-sm text-body">Tagihan {checkout?.invoice_number} telah lunas.</p><div className="rounded-2xl border border-border bg-white/70 p-5"><p className="font-semibold">{details?.plan_name}</p><p className="mt-2 text-2xl font-bold">{formatCurrency(details?.amount ?? 0)}</p><p className="mt-2 text-sm text-body">{checkout?.customer_name}</p></div></section>
        ) : isUnavailable ? <div className="space-y-3"><AlertCircle className="text-red-600" /><h1 className="text-xl font-bold">Pembayaran tidak tersedia</h1><p className="text-sm text-body">Hubungi penyedia layanan untuk mendapatkan tautan pembayaran baru.</p></div> : details ? (
          <>
            <p className="mb-1 text-xs font-semibold uppercase tracking-wide text-brand">{details.business_name}</p>
            <h1 className="font-display text-xl font-bold leading-7">{step === 1 ? 'Informasi paket' : step === 2 ? 'Informasi pelanggan' : isAuthenticating ? 'Verifikasi kartu' : 'Pembayaran kartu'}</h1>
            {step === 1 && <section className="mt-4 space-y-4">
              <div className="rounded-2xl border border-border bg-white/70 p-5"><h2 className="text-lg font-semibold">{details.plan_name}</h2>{preview && <p className="mt-1 text-sm text-body">{preview.billing_cycle === 'yearly' ? 'Langganan tahunan' : 'Langganan bulanan'}</p>}
                {checkout?.items?.map((item, index) => <div key={index} className="mt-4 flex justify-between gap-4 text-sm"><span className="text-body">{item.type === 'discount' ? 'Kupon / diskon: ' : item.type === 'add_on' ? 'Add-on: ' : ''}{item.description}{item.quantity > 1 ? ` × ${item.quantity}` : ''}</span><span className="shrink-0 font-medium">{formatCurrency(item.subtotal)}</span></div>)}
                <div className="mt-5 flex items-center justify-between gap-3 border-t border-border pt-4"><span className="text-sm text-body">Total pembayaran</span><strong className="text-xl">{formatCurrency(details.amount)}</strong></div>
              </div>
              <p className="text-sm text-body">Periksa paket dan rincian tagihan sebelum melanjutkan.</p>
            </section>}
            {step === 2 && <form id="checkout-customer-form" onSubmit={handleStart} className="mt-4 space-y-4">
              <p className="text-sm text-body">Gunakan kontak yang dapat dihubungi untuk langganan Anda.</p>
              <Input id="checkout-name" label="Nama lengkap" value={name} onChange={(event) => setName(event.target.value)} required autoComplete="name" readOnly={Boolean(session)} />
              <Input id="checkout-email" label="Email" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required autoComplete="email" readOnly={Boolean(session)} />
              <Input id="checkout-mobile" label="Nomor ponsel" type="tel" value={mobile} onChange={(event) => setMobile(event.target.value)} placeholder="+628123456789" required autoComplete="tel" readOnly={Boolean(session)} />
              {session && <p className="text-xs text-muted">Kontak telah digunakan untuk menyiapkan sesi pembayaran ini.</p>}
              {error && <p role="alert" className="text-sm text-red-600">{error}</p>}
            </form>}
            {session && <section hidden={step !== 3} className="mt-2 space-y-3">
              <div className="flex justify-between gap-4 rounded-xl border border-border bg-white/60 px-4 py-2 text-sm"><span>{details.plan_name}</span><strong className="shrink-0">{formatCurrency(details.amount)}</strong></div>
              {isVerifying ? <div role="status" className="flex items-start gap-3 py-6 text-brand"><Clock3 size={22} /><p className="text-sm">Kami sedang memastikan pembayaran Anda. Status akan diperbarui secara otomatis.</p></div> : <>
                {session.provider === 'xendit' ? <>
                  <div ref={cardContainer} hidden={isAuthenticating} className="checkout-card-fields" aria-label="Formulir pembayaran kartu" />
                  <div ref={actionContainer} hidden={!isAuthenticating} className="min-h-[440px] w-full overflow-hidden rounded-xl" aria-label="Verifikasi pembayaran kartu" />
                </> : session.mode === 'CORE' ? <MidtransCardForm token={token} session={session} name={name} email={email} mobile={mobile} amount={details.amount} actionContainer={cardActions} actionsVisible={step === 3} onBusyChange={setIsSubmitting} onVerify={() => { setIsVerifying(true); void getPublicCheckout(token).then(setCheckout).catch(() => undefined) }} /> : <div id="paylanggan-midtrans-card" className="min-h-[600px] w-full overflow-hidden rounded-2xl bg-white" />}
                {error && <p role="alert" className="text-sm text-red-600">{error}</p>}

              </>}
            </section>}
          </>
        ) : null}
      </CheckoutScrollArea>
      <footer className="checkout-footer z-10 shrink-0 border-t border-border">
        <div className="mx-auto w-full max-w-lg px-6 py-3">
          {!isLoading && details && !isPaid && !isUnavailable && <div className="checkout-actions flex items-center gap-3">
            {step > 1 && backButton}
            {step === 1 && <SubmitButton type="button" onClick={() => setStep(2)} className="w-full justify-center gap-2">Lanjut<ArrowRight size={16} /></SubmitButton>}
            {step === 2 && <SubmitButton type="submit" form="checkout-customer-form" disabled={isStarting} className="flex-1 justify-center gap-2">{isStarting ? 'Menyiapkan...' : 'Lanjut'}<ArrowRight size={16} /></SubmitButton>}
            {step === 3 && session?.provider === 'xendit' && !isAuthenticating && !isVerifying && <SubmitButton type="button" onClick={handlePay} disabled={!isCardReady || isSubmitting} className="flex-1 justify-center gap-2"><CreditCard size={17} />{isSubmitting ? 'Memproses pembayaran...' : `Bayar ${formatCurrency(details.amount)}`}</SubmitButton>}
            <div ref={setCardActions} className="min-w-0 flex-1 empty:hidden" />
          </div>}
          <p className="mt-2 flex items-center justify-center gap-2 text-xs text-body"><ShieldCheck size={14} />Pembayaran terlindungi</p>
        </div>
      </footer>
    </main>
  )
}

```

### Frontend/src/components/payment/MidtransCardForm.tsx

```tsx
import { useEffect, useId, useRef, useState, type FormEvent } from 'react'
import { createPortal } from 'react-dom'
import { CreditCard, ShieldCheck } from 'lucide-react'
import SubmitButton from '../ui/SubmitButton'
import { chargeMidtransCard, type StartedCheckout } from '../../lib/api'
import { formatCurrency } from '../../lib/format'

type SDKResult = { token_id?: string; status_code?: string }
type AuthOptions = {
  performAuthentication: (url: string) => void
  onSuccess: (result: SDKResult) => void
  onFailure: (result: SDKResult) => void
  onPending: (result: SDKResult) => void
}
declare global {
  interface Window {
    MidtransNew3ds?: {
      getCardToken: (card: { card_number: string; card_exp_month: string; card_exp_year: string; card_cvv: string }, callbacks: {
        onSuccess: (result: SDKResult) => void; onFailure: (result: SDKResult) => void
      }) => void
      authenticate: (url: string, options: AuthOptions) => void
    }
  }
}

function validAuthenticationURL(value: string) {
  try {
    const url = new URL(value)
    return url.protocol === 'https:' && !url.username && !url.password && !url.port &&
      ['card.midtrans.com', 'api.midtrans.com', 'api.sandbox.midtrans.com', 'api.veritrans.co.id', 'api.sandbox.veritrans.co.id'].includes(url.hostname)
  } catch { return false }
}

export default function MidtransCardForm({ token, session, name, email, mobile, amount, onVerify, onBusyChange, actionContainer, actionsVisible = true }: {
  token: string; session: StartedCheckout; name: string; email: string; mobile: string; amount: number; onVerify: () => void; onBusyChange: (busy: boolean) => void
  actionContainer?: HTMLElement | null
  actionsVisible?: boolean
}) {
  const formId = useId()
  const [number, setNumber] = useState('')
  const [expiry, setExpiry] = useState('')
  const [cvv, setCvv] = useState('')
  const [firstName, setFirstName] = useState(name.split(' ')[0] ?? '')
  const [lastName, setLastName] = useState(name.split(' ').slice(1).join(' '))
  const [contactEmail, setContactEmail] = useState(email)
  const [contactMobile, setContactMobile] = useState(mobile)
  const [ready, setReady] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [authenticationURL, setAuthenticationURL] = useState<string | null>(null)
  const resumed = useRef(false)
  const submitting = useRef(false)
  const mounted = useRef(true)

  useEffect(() => { onBusyChange(busy) }, [busy, onBusyChange])

  useEffect(() => {
    mounted.current = true
    return () => { mounted.current = false }
  }, [])

  useEffect(() => {
    let active = true
    const id = 'midtrans-script'
    let script = document.getElementById(id) as HTMLScriptElement | null
    const loaded = () => { if (active) setReady(Boolean(window.MidtransNew3ds)) }
    const failed = () => { if (active) setError('Formulir kartu belum dapat dimuat. Muat ulang halaman lalu coba lagi.') }
    if (script?.getAttribute('data-client-key') !== session.client_key || script?.getAttribute('data-environment') !== session.environment) {
      script?.remove()
      window.MidtransNew3ds = undefined
      script = null
    }
    if (!script) {
      script = document.createElement('script')
      script.id = id
      script.src = 'https://api.midtrans.com/v2/assets/js/midtrans-new-3ds.min.js'
      script.setAttribute('data-client-key', session.client_key ?? '')
      script.setAttribute('data-environment', session.environment ?? 'sandbox')
      script.async = true
      script.addEventListener('load', loaded)
      script.addEventListener('error', failed)
      document.head.appendChild(script)
    } else if (window.MidtransNew3ds) loaded()
    else {
      script.addEventListener('load', loaded)
      script.addEventListener('error', failed)
    }
    return () => {
      active = false
      script?.removeEventListener('load', loaded)
      script?.removeEventListener('error', failed)
    }
  }, [session.client_key, session.environment])

  function authenticate(url: string) {
    if (!window.MidtransNew3ds || !validAuthenticationURL(url)) {
      setError('Verifikasi kartu belum dapat dibuka. Tunggu konfirmasi pembayaran.')
      onVerify()
      return
    }
    setBusy(true)
    const finished = () => {
      if (!mounted.current) return
      setAuthenticationURL(null)
      setBusy(false)
      onVerify()
    }
    window.MidtransNew3ds.authenticate(url, {
      performAuthentication: (challengeURL) => {
        if (!mounted.current) return
        if (!validAuthenticationURL(challengeURL)) { finished(); return }
        setAuthenticationURL(challengeURL)
      },
      onSuccess: finished,
      onPending: finished,
      onFailure: finished,
    })
  }

  useEffect(() => {
    if (!ready || resumed.current || !session.payment_session_id) return
    resumed.current = true
    submitting.current = true
    if (session.authentication_url) {
      try { authenticate(session.authentication_url) } catch { onVerify() }
    } else onVerify()
    // Resume only once for this checkout; callbacks never mark an invoice paid.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [ready, session.payment_session_id, session.authentication_url])

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (!ready || !window.MidtransNew3ds || submitting.current) return
    const digits = expiry.replace(/\D/g, '')
    const month = Number(digits.slice(0, 2))
    const year = 2000 + Number(digits.slice(2))
    if (digits.length !== 4 || month < 1 || month > 12 || new Date(year, month, 1) <= new Date()) {
      setError('Periksa kembali masa berlaku kartu.')
      return
    }
    submitting.current = true
    setBusy(true)
    setError(null)
    let cardToken: string
    try {
      cardToken = await new Promise<string>((resolve, reject) => {
        window.MidtransNew3ds!.getCardToken({
          card_number: number.replace(/\D/g, ''), card_exp_month: String(month), card_exp_year: String(year), card_cvv: cvv,
        }, {
          onSuccess: (result) => result.token_id ? resolve(result.token_id) : reject(new Error('invalid token')),
          onFailure: () => reject(new Error('tokenization failed')),
        })
      })
    } catch {
      if (mounted.current) {
        submitting.current = false
        setBusy(false)
        setError('Detail kartu belum dapat diverifikasi. Periksa nomor kartu, masa berlaku, dan CVV.')
      }
      return
    }
    // Keep raw card details only in this form, and clear them after tokenization.
    setNumber('')
    setExpiry('')
    setCvv('')
    try {
      const result = await chargeMidtransCard(token, {
        token_id: cardToken, name: `${firstName} ${lastName}`.trim(), email: contactEmail, mobile: contactMobile,
      })
      if (!mounted.current) return
      if (result.authentication_url) authenticate(result.authentication_url)
      else onVerify()
    } catch {
      if (mounted.current) {
        setError('Status pembayaran sedang diperiksa. Tunggu konfirmasi sebelum mencoba pembayaran lain.')
        setBusy(false)
        onVerify()
      }
    }
  }

  const input = 'min-h-11 min-w-0 w-full bg-white px-3 py-2.5 text-[14px] leading-5 text-slate-900 outline-none placeholder:text-slate-400 focus:ring-2 focus:ring-inset focus:ring-brand disabled:opacity-60'
  const group = 'overflow-hidden rounded-[6px] border border-slate-200 bg-white'
  const label = 'mb-1 block text-[13px] leading-5 text-ink'
  if (authenticationURL) return (
    <section aria-label="Verifikasi kartu" className="space-y-4">
      <div className="flex items-center gap-2 text-brand"><ShieldCheck size={20} /><h2 className="font-semibold">Verifikasi kartu</h2></div>
      <p className="text-sm text-body">Selesaikan verifikasi dari bank penerbit kartu Anda.</p>
      <iframe title="Verifikasi pembayaran kartu" src={authenticationURL} className="h-[440px] w-full rounded-xl border border-border bg-white" allow="payment" />
      <button type="button" className="text-sm font-semibold text-brand" onClick={() => { setAuthenticationURL(null); onVerify() }}>Periksa status pembayaran</button>
    </section>
  )
  const payButton = <SubmitButton type="submit" form={formId} disabled={!ready || busy || submitting.current} className="w-full justify-center gap-2">
    <CreditCard size={17} aria-hidden="true" />{busy ? 'Memproses pembayaran...' : `Bayar ${formatCurrency(amount)}`}
  </SubmitButton>
  return (
    <form id={formId} onSubmit={submit} className="midtrans-card-form space-y-3" aria-label="Formulir pembayaran kartu kredit">
      <div>
        <label className={label} htmlFor="midtrans-card-number">Detail Kartu</label>
        <div className={group}>
          <div className="relative">
            <input id="midtrans-card-number" className={`${input} pr-12`} value={number} onChange={(event) => setNumber(event.target.value.replace(/\D/g, '').slice(0, 19).replace(/(.{4})/g, '$1 ').trim())} placeholder="1234 1234 1234 1234" inputMode="numeric" autoComplete="cc-number" required pattern="[0-9 ]{15,23}" disabled={busy} />
            <CreditCard className="pointer-events-none absolute right-3 top-3 text-slate-400" size={20} aria-hidden="true" />
          </div>
          <div className="grid grid-cols-2 border-t border-slate-200">
            <input aria-label="Masa berlaku kartu" className={`${input} border-r border-slate-200`} value={expiry} onChange={(event) => { const value = event.target.value.replace(/\D/g, '').slice(0, 4); setExpiry(value.length > 2 ? `${value.slice(0, 2)}/${value.slice(2)}` : value) }} placeholder="MM/YY" inputMode="numeric" autoComplete="cc-exp" required pattern="[0-9]{2}/[0-9]{2}" disabled={busy} />
            <input aria-label="CVV" className={input} value={cvv} onChange={(event) => setCvv(event.target.value.replace(/\D/g, '').slice(0, 4))} placeholder="CVV" type="password" inputMode="numeric" autoComplete="cc-csc" required pattern="[0-9]{3,4}" disabled={busy} />
          </div>
        </div>
      </div>
      <div><label className={label} htmlFor="midtrans-card-first-name">Nama Pemegang Kartu</label>
        <div className={`${group} grid grid-cols-2`}>
          <input id="midtrans-card-first-name" className={`${input} border-r border-slate-200`} value={firstName} onChange={(event) => setFirstName(event.target.value)} placeholder="Nama Depan" autoComplete="cc-given-name" required maxLength={50} disabled={busy} />
          <input aria-label="Nama belakang pemegang kartu" className={input} value={lastName} onChange={(event) => setLastName(event.target.value)} placeholder="Nama Belakang" autoComplete="cc-family-name" maxLength={49} disabled={busy} />
        </div>
      </div>
      <div><label className={label} htmlFor="midtrans-card-email">Alamat Email</label><div className={group}>
        <input id="midtrans-card-email" className={input} type="email" value={contactEmail} onChange={(event) => setContactEmail(event.target.value)} autoComplete="email" required maxLength={254} disabled={busy} />
      </div></div>
      <div><label className={label} htmlFor="midtrans-card-mobile">Nomor Handphone</label><div className={`${group} flex`}>
        <span className="flex shrink-0 items-center gap-2 border-r border-slate-200 px-3 text-[14px] text-slate-900"><span aria-label="Indonesia" className="h-3 w-4 overflow-hidden rounded-sm border border-slate-200 bg-white"><span className="block h-1/2 bg-red-600" /></span> +62</span>
        <input id="midtrans-card-mobile" className={input} type="tel" value={contactMobile.replace(/^\+62/, '')} onChange={(event) => setContactMobile(`+62${event.target.value.replace(/\D/g, '').replace(/^0/, '')}`)} placeholder="812 345 678" autoComplete="tel-national" required pattern="[0-9]{8,13}" disabled={busy} />
      </div></div>
      {error && <p role="alert" className="text-[13px] text-red-600">{error}</p>}
      {actionsVisible && (actionContainer ? createPortal(payButton, actionContainer) : payButton)}
    </form>
  )
}

```

### Frontend/public/paylanggan.js

```js
(function () {
  'use strict'

  var script = document.currentScript
  if (!script || !script.src) return
  var checkoutOrigin = new URL(script.src).origin
  var active = null

  function close() {
    if (!active) return
    var state = active
    active = null
    window.removeEventListener('message', state.onMessage)
    document.removeEventListener('keydown', state.onKeydown)
    state.overlay.remove()
    document.body.style.overflow = state.previousOverflow
    if (state.previousFocus && state.previousFocus.focus) state.previousFocus.focus()
    if (typeof state.onClose === 'function') state.onClose()
  }

  function open(options) {
    if (!options || typeof options.checkoutUrl !== 'string') throw new Error('checkoutUrl is required')
    var url = new URL(options.checkoutUrl)
    if (url.origin !== checkoutOrigin || !/^\/checkout\/[A-Za-z0-9_-]+$/.test(url.pathname) && url.pathname !== '/checkout/embed') {
      throw new Error('Invalid PayLanggan checkout URL')
    }
    close()
    url.searchParams.set('parent_origin', window.location.origin)

    var overlay = document.createElement('div')
    overlay.className = 'pl-dropin-overlay'
    overlay.setAttribute('role', 'dialog')
    overlay.setAttribute('aria-modal', 'true')
    overlay.setAttribute('aria-label', 'Pembayaran PayLanggan')
    var panel = document.createElement('div')
    panel.className = 'pl-dropin-panel'
    var button = document.createElement('button')
    button.type = 'button'
    button.className = 'pl-dropin-close'
    button.setAttribute('aria-label', 'Tutup pembayaran')
    button.innerHTML = '<svg viewBox="0 0 20 20" width="18" height="18" aria-hidden="true"><path d="m5 5 10 10M15 5 5 15" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>'
    var fullPage = document.createElement('a')
    fullPage.className = 'pl-dropin-full-page'
    fullPage.href = options.checkoutUrl
    fullPage.target = '_blank'
    fullPage.rel = 'noopener noreferrer'
    fullPage.textContent = 'Buka halaman penuh'
    var frame = document.createElement('iframe')
    frame.title = 'Pembayaran PayLanggan'
    frame.src = url.toString()
    frame.allow = 'payment *'
    var controls = document.createElement('div')
    controls.className = 'pl-dropin-controls'
    controls.appendChild(fullPage)
    controls.appendChild(button)
    panel.appendChild(controls)
    panel.appendChild(frame)
    overlay.appendChild(panel)

    function onMessage(event) {
      if (event.origin !== checkoutOrigin || event.source !== frame.contentWindow) return
      if (event.data && event.data.type === 'paylanggan:checkout-link') {
        try {
          var next = new URL(event.data.checkoutUrl)
          if (next.origin === checkoutOrigin && /^\/checkout\/[A-Za-z0-9_-]+$/.test(next.pathname)) {
            fullPage.href = next.toString()
            if (typeof options.onCheckoutLink === 'function') options.onCheckoutLink(next.toString(), event.data.externalOrderId)
          }
        } catch { /* Ignore invalid links. */ }
        return
      }
      if (event.data && event.data.type === 'paylanggan:checkout-close') { close(); return }
      if (!event.data || event.data.type !== 'paylanggan:checkout-status' || event.data.status !== 'paid') return
      if (typeof options.onPaid === 'function') options.onPaid()
    }
    function onKeydown(event) {
      if (event.key === 'Escape') close()
      if (event.key === 'Tab') {
        // The iframe contains its own focusable checkout controls.
        if (document.activeElement === button && event.shiftKey) { event.preventDefault(); frame.focus() }
        else if (document.activeElement === frame && !event.shiftKey) { event.preventDefault(); button.focus() }
      }
    }
    button.addEventListener('click', close)
    overlay.addEventListener('click', function (event) { if (event.target === overlay) close() })
    window.addEventListener('message', onMessage)
    document.addEventListener('keydown', onKeydown)
    var previousFocus = document.activeElement
    var previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    document.body.appendChild(overlay)
    active = { overlay: overlay, onMessage: onMessage, onKeydown: onKeydown,
      previousFocus: previousFocus, previousOverflow: previousOverflow, onClose: options.onClose }
    button.focus()
    return { close: close }
  }

  function openEmbed(options) {
    var planId = options && String(options.planId || '')
    var publicPlan = /^pl_plan_[0-9a-f]{24}$/.test(planId)
    var legacyPlan = /^[1-9][0-9]*$/.test(planId) && Number.isSafeInteger(Number(planId))
    if (!options || !/^pl_site_[A-Za-z0-9_-]+$/.test(options.siteId || '') || (!publicPlan && !legacyPlan)) {
      throw new Error('Valid siteId and planId are required')
    }
    var api = new URL(options.apiUrl || checkoutOrigin)
    if (api.protocol !== 'https:' && !(api.protocol === 'http:' && (api.hostname === 'localhost' || api.hostname === '127.0.0.1'))) {
      throw new Error('PayLanggan API must use HTTPS')
    }
    var storageKey = 'paylanggan:embed:' + options.siteId + ':' + planId
    function checkoutToken(value) {
      var url = new URL(value)
      if (url.origin !== checkoutOrigin || !/^\/checkout\/[A-Za-z0-9_-]+$/.test(url.pathname) && url.pathname !== '/checkout/embed') throw new Error('Invalid PayLanggan checkout URL')
      return url.pathname.split('/')[2]
    }
    function displayOrder(order) {
      checkoutToken(order.checkout_url)
      try { sessionStorage.setItem(storageKey, JSON.stringify(order)) } catch { /* Storage may be unavailable. */ }
      return open({
        checkoutUrl: order.checkout_url,
        onPaid: function () {
          try { sessionStorage.removeItem(storageKey) } catch {}
          document.dispatchEvent(new CustomEvent('paylanggan:paid', { detail: { externalOrderId: order.external_order_id } }))
          if (typeof options.onPaid === 'function') options.onPaid(order.external_order_id)
        },
        onClose: options.onClose,
      })
    }
    function showForm() {
      var url = new URL('/checkout/embed', checkoutOrigin)
      url.searchParams.set('site', options.siteId)
      url.searchParams.set('plan', planId)
      var orderId
      return open({ checkoutUrl: url.toString(), onClose: options.onClose,
        onCheckoutLink: function (link, externalOrderId) {
          orderId = externalOrderId
          try { sessionStorage.setItem(storageKey, JSON.stringify({ checkout_url: link, external_order_id: orderId })) } catch { /* Storage may be unavailable. */ }
        },
        onPaid: function () {
          try { sessionStorage.removeItem(storageKey) } catch { /* Storage may be unavailable. */ }
          document.dispatchEvent(new CustomEvent('paylanggan:paid', { detail: { externalOrderId: orderId } }))
          if (typeof options.onPaid === 'function') options.onPaid(orderId)
        },
      })
    }
    var saved = null
    try { saved = JSON.parse(sessionStorage.getItem(storageKey) || 'null') } catch {}
    if (!saved || !saved.checkout_url) return Promise.resolve(showForm())
    var token
    try { token = checkoutToken(saved.checkout_url) }
    catch { try { sessionStorage.removeItem(storageKey) } catch {} return Promise.resolve(showForm()) }
    return fetch(new URL('/api/v1/checkout/' + encodeURIComponent(token), api), { credentials: 'omit', cache: 'no-store' })
      .then(function (response) { if (!response.ok) throw new Error('Status pesanan belum dapat diperiksa.'); return response.json() })
      .then(function (body) {
        if (body.data && body.data.invoice_status === 'pending' && ['created', 'active'].indexOf(body.data.checkout_status) !== -1) return displayOrder(saved)
        try { sessionStorage.removeItem(storageKey) } catch {}
        return showForm()
      })
  }

  var style = document.createElement('style')
  style.textContent = '.pl-dropin-overlay{position:fixed;inset:0;z-index:2147483647;display:flex;align-items:center;justify-content:center;padding:16px;background:rgba(15,23,42,.7)}' +
    '.pl-dropin-panel{position:relative;width:min(100%,560px);height:min(90dvh,880px);max-height:min(90dvh,880px);overflow:hidden;border-radius:16px;background:#fff;box-shadow:0 24px 80px rgba(0,0,0,.28)}' +
    '.pl-dropin-panel iframe{display:block;width:100%;height:100%;border:0}' +
    '.pl-dropin-controls{position:absolute;top:10px;right:10px;z-index:2;display:flex;align-items:center;gap:8px;padding:4px;border-radius:24px;background:rgba(255,255,255,.97);box-shadow:0 2px 12px rgba(15,23,42,.16)}' +
    '.pl-dropin-close{position:static;display:grid;place-items:center;flex:0 0 36px;width:36px;height:36px;padding:0;border:0;border-radius:50%;background:#f1f5f9;color:#111827;font:26px/1 system-ui,sans-serif;cursor:pointer}' +
    '.pl-dropin-close:hover{background:#e2e8f0}.pl-dropin-close:focus-visible,.pl-dropin-full-page:focus-visible{outline:2px solid #3155f5;outline-offset:2px}' +
    '.pl-dropin-full-page{display:inline-flex;align-items:center;justify-content:center;min-height:36px;padding:0 12px;border-radius:18px;background:#fff;color:#2455eb;font:600 12px system-ui,sans-serif;white-space:nowrap;text-decoration:none}' +
    '@media(max-width:640px){.pl-dropin-overlay{padding:0}.pl-dropin-panel{height:min(90dvh,880px);width:100%;border-radius:16px}.pl-dropin-overlay{padding:8px}}' +
    '.pl-embed-form-panel{box-sizing:border-box;width:min(100%,460px);max-height:90vh;overflow:auto;border:1px solid rgba(148,163,184,.2);border-radius:22px;background:#fff;padding:24px 28px 26px;color:#172033;font:14px/1.5 system-ui,-apple-system,"Segoe UI",sans-serif;box-shadow:0 28px 90px rgba(15,23,42,.3)}' +
    '.pl-embed-form-panel *{box-sizing:border-box}.pl-embed-header{display:flex;align-items:center;justify-content:space-between;padding-bottom:20px;border-bottom:1px solid #eef1f6}' +
    '.pl-embed-brand{display:inline-flex;align-items:center;gap:9px;font-size:15px;font-weight:750;letter-spacing:-.03em;color:#111827}.pl-embed-logo{display:block;width:30px;height:30px;object-fit:contain}' +
    '.pl-embed-form-panel button.pl-embed-close{appearance:none;box-sizing:border-box;position:relative;display:flex;flex:0 0 36px;align-items:center;justify-content:center;width:36px!important;min-width:36px;height:36px;min-height:36px;margin:0!important;padding:0!important;overflow:hidden;border:0;border-radius:50%;background:#f1f5f9;color:#475569;line-height:1;cursor:pointer}' +
    '.pl-embed-form-panel button.pl-embed-close:hover{background:#e2e8f0;color:#111827}.pl-embed-close svg{display:block;flex:none;width:16px;height:16px;overflow:visible}.pl-embed-close path{fill:none;stroke:currentColor;stroke-width:1.8;stroke-linecap:round}' +
    '.pl-embed-intro{padding:23px 0 4px}.pl-embed-eyebrow{color:#3155f5;font-size:10px;font-weight:750;letter-spacing:.12em}.pl-embed-form-panel h2{margin:7px 0 4px;font-size:25px;line-height:1.2;letter-spacing:-.04em}.pl-embed-form-panel .pl-embed-intro p{margin:0;color:#64748b;font-size:13px}' +
    '.pl-embed-form-panel label{display:block;margin:16px 0 0;color:#263247;font-size:12px;font-weight:650}.pl-embed-form-panel input{display:block;width:100%;height:46px;margin-top:7px;border:1px solid #d7deea;border-radius:11px;background:#fff;padding:0 13px;color:#111827;font:14px system-ui,sans-serif;outline:none;transition:border-color .15s,box-shadow .15s}.pl-embed-form-panel input::placeholder{color:#9aa5b5}.pl-embed-form-panel input:focus{border-color:#4564ff;box-shadow:0 0 0 3px rgba(49,85,245,.13)}' +
    '.pl-embed-form-panel button:focus-visible{outline:3px solid rgba(49,85,245,.28);outline-offset:2px}.pl-embed-form-panel .pl-embed-submit,.pl-embed-form-panel .pl-embed-cancel{display:block;width:100%;min-height:45px;margin-top:18px;padding:11px 14px;border:0;border-radius:11px;font:650 13px system-ui,sans-serif;cursor:pointer;transition:background .15s,transform .15s}.pl-embed-form-panel .pl-embed-submit{background:#3155f5;color:#fff;box-shadow:0 8px 18px rgba(49,85,245,.2)}.pl-embed-form-panel .pl-embed-submit:hover{background:#2447e8}.pl-embed-form-panel .pl-embed-cancel{min-height:40px;margin-top:8px;background:transparent;color:#64748b}.pl-embed-form-panel .pl-embed-cancel:hover{background:#f4f6fa;color:#334155}.pl-embed-form-panel button:disabled{opacity:.6;cursor:wait}.pl-embed-form-panel .pl-embed-error{margin:10px 0 0;color:#b42318;font-size:12px;min-height:0}.pl-embed-form-panel .pl-embed-error:empty{display:none}'
  document.head.appendChild(style)
  window.PayLanggan = Object.freeze({ open: open, openEmbed: openEmbed, close: close })
  var siteId = script.getAttribute('data-pl-site')
  var apiUrl = script.getAttribute('data-pl-api-url')
  if (siteId && apiUrl) {
    document.addEventListener('click', function (event) {
      var button = event.target && event.target.closest && event.target.closest('[data-pl-checkout]')
      if (!button) return
      event.preventDefault()
      openEmbed({ siteId: siteId, planId: button.getAttribute('data-pl-plan'), apiUrl: apiUrl }).catch(function (reason) {
        document.dispatchEvent(new CustomEvent('paylanggan:error', { detail: { message: reason.message || 'Pembayaran belum tersedia.' } }))
      })
    })
  }
})()

```

### Frontend/src/index.css (CSS checkout yang ditambahkan)

```css
/* Only checkout content scrolls; progress and actions stay visible. */
.checkout-scroll { scrollbar-width: none; -ms-overflow-style: none; overscroll-behavior-y: contain; }
.checkout-scroll::-webkit-scrollbar { display: none; }
.checkout-step-connectors { position: absolute; z-index: 1; top: 27px; left: calc(12.5% + 24px); right: calc(12.5% + 24px); display: flex; gap: 48px; pointer-events: none; }
.checkout-step-line { flex: 1; height: 2px; overflow: hidden; background: #94a3b8; }
.checkout-step-line > span { display: block; height: 100%; background: #3654ff; transform: scaleX(0); transform-origin: left; transition: transform 250ms ease; }
.checkout-step-line > span.is-complete { transform: scaleX(1); }
.checkout-step-active { box-shadow: 0 0 0 4px rgba(54,84,255,0.18); }
.checkout-stepper-header, .checkout-footer { background: rgba(244,246,255,0.98); }
.checkout-footer { padding-bottom: env(safe-area-inset-bottom, 0px); }
.checkout-scroll-fade { background: linear-gradient(transparent, #f0f3ff); }
html.dark .checkout-stepper-header, html.dark .checkout-footer { background: #13192a; }
html.dark .checkout-scroll-fade { background: linear-gradient(transparent, #13192a); }
@media (max-width: 479px) {
  .checkout-header { min-height: 108px; padding-top: 60px; }
  .checkout-actions { flex-direction: column-reverse; align-items: stretch; }
  .checkout-actions > button { justify-content: center; }
  .checkout-actions > div { width: 100%; }
}
@media (prefers-reduced-motion: reduce) { .checkout-step-line > span { transition: none; } }
.checkout-card-fields .xendit-channel-form { gap: 12px; }
.checkout-card-fields .xendit-channel-form-field-group,
.checkout-card-fields .xendit-channel-form-field { gap: 4px; }
.checkout-card-fields .xendit-channel-form-field-group label { font-size: 13px; line-height: 20px; }
html.dark .checkout-card-fields { --xendit-color-background: #1e293b; --xendit-color-text: #eef2ff; --xendit-color-text-secondary: #c4ccde; --xendit-color-border: #64748b; }
html.dark .checkout-shell { background-color: #13192a; background-image: radial-gradient(at 0% 0%, rgba(54,84,255,0.16), transparent 65%); }
html.dark .checkout-shell input { color: #eef2ff !important; background-color: #1e293b !important; }
html.dark .midtrans-card-form .bg-white { background-color: #1e293b; }
html.dark .midtrans-card-form .border-slate-200 { border-color: #64748b; }
html.dark .midtrans-card-form .text-slate-900 { color: #eef2ff; }

```
