import { useEffect, useRef, useState, type FormEvent } from 'react'
import { flushSync } from 'react-dom'
import { AlertCircle, CheckCircle2, Clock3, CreditCard, ShieldCheck, ArrowLeft, ArrowRight } from 'lucide-react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { XenditComponents, type XenditSubmissionEndEvent } from 'xendit-components-web'
import Input from '../components/ui/Input'
import Logo from '../components/ui/Logo'
import SubmitButton from '../components/ui/SubmitButton'
import CheckoutStepper from '../components/payment/CheckoutStepper'
import CheckoutOptions from '../components/payment/CheckoutOptions'
import CheckoutScrollArea from '../components/payment/CheckoutScrollArea'
import MidtransCardForm from '../components/payment/MidtransCardForm'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import { ApiError, getPublicCheckout, cancelPublicCheckout, verifyMidtransCard, startPublicCheckout, getEmbedPlan, createEmbedCheckout, type PublicCheckout, type StartedCheckout, type EmbedPlanPreview, type EmbedCheckoutOrder, type CheckoutAddOnSelection } from '../lib/api'
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
  const [preview, setPreview] = useState<EmbedPlanPreview | null>(null)
  const [completedOrder, setCompletedOrder] = useState<EmbedCheckoutOrder | null>(null)
  const [addOns, setAddOns] = useState<CheckoutAddOnSelection[]>(() => {
    try {
      const value: unknown = JSON.parse(new URLSearchParams(window.location.search).get('add_ons') ?? '[]')
      if (Array.isArray(value) && value.length <= 20 && value.every((item: unknown) => typeof item === 'object' && item !== null && 'add_on_id' in item && 'quantity' in item && Number.isInteger(item.add_on_id) && Number.isInteger(item.quantity))) return value as CheckoutAddOnSelection[]
    } catch { /* Backend will validate selections again. */ }
    return []
  })
  const [couponCode, setCouponCode] = useState(() => new URLSearchParams(window.location.search).get('coupon_code') ?? '')
  const [quoteBusy, setQuoteBusy] = useState(false)
  const [quoteError, setQuoteError] = useState<string | null>(null)
  const orderRequest = useRef<string>('')
  const [searchParams] = useSearchParams()
  const [portalSite] = useState(() => {
    const requested = searchParams.get('portal') ?? ''
    const valid = (value: string) => /^pl_site_[A-Za-z0-9_-]+$/.test(value)
    try {
      if (token && valid(requested)) sessionStorage.setItem(`paylanggan:checkout-portal:${token}`, requested)
      const saved = token ? sessionStorage.getItem(`paylanggan:checkout-portal:${token}`) ?? '' : ''
      return valid(requested) ? requested : valid(saved) ? saved : ''
    } catch { return valid(requested) ? requested : '' }
  })
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
  const [cardConsent, setCardConsent] = useState(false)
  const [verificationDelayed, setVerificationDelayed] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [isStarting, setIsStarting] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isCardReady, setIsCardReady] = useState(false)
  const [isSnapReady, setIsSnapReady] = useState(false)
  const [isVerifying, setIsVerifying] = useState(returningFromXendit)
  const [error, setError] = useState<string | null>(null)
  const [isAuthenticating, setIsAuthenticating] = useState(false)
  const [cancelDialogOpen, setCancelDialogOpen] = useState(false)
  const [isCancelling, setIsCancelling] = useState(false)
  const [cancelError, setCancelError] = useState<string | null>(null)
  const [cancelled, setCancelled] = useState(false)
  const actionContainer = useRef<HTMLDivElement>(null)
  const cardContainer = useRef<HTMLDivElement>(null)
  const componentsRef = useRef<XenditComponents | null>(null)
  const paidNotified = useRef(false)
  const isTrialRegistration = Boolean(checkout?.card_registration || (!token && preview?.trial_days))
  const trialComplete = Boolean(checkout?.card_registration && checkout.checkout_status === 'completed')
  const trialAutoDebit = checkout?.payment_provider === 'xendit' || checkout?.payment_provider === 'midtrans'
  const trialCycleLabel = checkout?.billing_cycle === 'yearly' ? 'tahun' : 'bulan'
  const trialBillingDateValue = checkout?.trial_ends_at ? new Date(checkout.trial_ends_at) : new Date()
  if (!checkout?.trial_ends_at) trialBillingDateValue.setDate(trialBillingDateValue.getDate() + (checkout?.trial_days ?? preview?.trial_days ?? 0))
  const trialBillingDate = trialBillingDateValue.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
  const trialDebitExpiry = new Date()
  trialDebitExpiry.setFullYear(trialDebitExpiry.getFullYear() + 10)
  const trialDebitExpiryLabel = trialDebitExpiry.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
  const isPaid = checkout?.invoice_status === 'paid'
  const isUnavailable = checkout?.checkout_status === 'expired' || checkout?.checkout_status === 'failed' || checkout?.checkout_status === 'cancelled' || cancelled

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
    if ((!trialComplete && checkout?.invoice_status !== 'paid') || !parentOrigin || paidNotified.current || window.parent === window || !isAllowedParentOrigin(parentOrigin)) return
    try {
      window.parent.postMessage({ type: 'paylanggan:checkout-status', status: trialComplete ? 'trial' : 'paid', trialEndsAt: checkout?.trial_ends_at }, parentOrigin)
      paidNotified.current = true
      try { window.sessionStorage.removeItem(parentOriginKey) } catch { /* Storage may be unavailable. */ }
    } catch { /* Ignore an unavailable parent window. */ }
  }, [checkout?.invoice_status, checkout?.trial_ends_at, trialComplete, parentOrigin, parentOriginKey])

  useEffect(() => {
    let active = true
    if (!token) {
      const controller = new AbortController()
      setQuoteBusy(true); setQuoteError(null)
      const timer = window.setTimeout(() => {
        void getEmbedPlan(searchParams.get('site') ?? '', searchParams.get('plan') ?? '', addOns, couponCode.trim(), controller.signal)
          .then((data) => { if (active) setPreview(data) })
          .catch(async (err: unknown) => {
            if (!active) return
            setQuoteError(err instanceof ApiError ? err.message : 'Paket belum dapat diperiksa. Coba lagi.')
            try { const base = await getEmbedPlan(searchParams.get('site') ?? '', searchParams.get('plan') ?? '', [], '', controller.signal); if (active) setPreview(base) } catch { /* Keep original error. */ }
          })
          .finally(() => { if (active) { setIsLoading(false); setQuoteBusy(false) } })
      }, 250)
      return () => { active = false; controller.abort(); window.clearTimeout(timer) }
    }
    getPublicCheckout(token)
      .then(async (data) => {
        if (!active) return
        setCheckout(data)
        setName(data.customer_name)
        setEmail(data.customer_email ?? '')
        setMobile(data.customer_mobile ?? '')
        if (((returningFromXendit || parentOrigin) && data.invoice_status === 'pending' || data.card_registration) && data.checkout_status === 'active') {
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
  }, [token, returningFromXendit, parentOrigin, searchParams, addOns, couponCode])

  useEffect(() => { orderRequest.current = ''; }, [addOns, couponCode])

  useEffect(() => {
    if (!session || session.provider !== 'xendit' || !session.components_sdk_key || isVerifying || !cardContainer.current || (checkout?.invoice_status === 'paid' || trialComplete)) return
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
  }, [session, token, returningFromXendit, isVerifying, checkout?.invoice_status, trialComplete])

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
    setVerificationDelayed(false)
    if (!isTrialRegistration || !isVerifying || trialComplete) return
    const timer = window.setTimeout(() => setVerificationDelayed(true), 60000)
    return () => window.clearTimeout(timer)
  }, [isTrialRegistration, isVerifying, trialComplete])

  useEffect(() => {
    if (trialComplete || verificationDelayed || (!session && !isVerifying) || checkout?.invoice_status === 'paid' || checkout?.checkout_status === 'expired') return
    let active = true
    const timer = window.setInterval(() => {
      void getPublicCheckout(token).then((data) => { if (active) setCheckout(data) }).catch(() => undefined)
    }, 3000)
    return () => { active = false; window.clearInterval(timer) }
  }, [token, session, isVerifying, checkout?.invoice_status, checkout?.checkout_status, trialComplete, verificationDelayed])

  useEffect(() => {
    if (session?.card_registration || !isVerifying || session?.provider !== 'midtrans' || session.mode !== 'CORE' || checkout?.invoice_status === 'paid' || checkout?.checkout_status === 'expired') return
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
  }, [token, session?.provider, session?.mode, session?.card_registration, isVerifying, checkout?.invoice_status, checkout?.checkout_status])

  async function handleStart(event: FormEvent) {
    event.preventDefault()
    if (isStarting) return
    if (!token && (quoteBusy || quoteError || completedOrder)) return
    if (session) { setStep(3); return }
    setIsStarting(true)
    setError(null)
    try {
      let paymentToken = token
      if (!paymentToken) {
        if (!orderRequest.current) orderRequest.current = window.crypto.randomUUID()
        const order = await createEmbedCheckout(searchParams.get('site') ?? '', searchParams.get('plan') ?? '', { name, email, mobile }, { add_ons: addOns, coupon_code: couponCode.trim(), request_id: orderRequest.current })
        if (order.subscription_status === 'trial' || order.invoice_status === 'paid') {
          setCompletedOrder(order); setStep(4)
          if (parentOrigin && isAllowedParentOrigin(parentOrigin) && window.parent !== window) window.parent.postMessage({ type: 'paylanggan:checkout-status', status: order.subscription_status === 'trial' ? 'trial' : 'paid', externalOrderId: order.external_order_id, trialEndsAt: order.trial_ends_at }, parentOrigin)
          return
        }
        if (!order.checkout_url) throw new Error('Tautan pembayaran belum tersedia. Silakan coba lagi.')
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
      const started = await startPublicCheckout(paymentToken, { name, email, mobile, card_consent: cardConsent, parent_origin: parentOrigin ?? undefined })
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
      setError(err instanceof Error ? err.message : 'Pembayaran belum dapat dimulai. Silakan coba lagi.')
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

  async function handleCancel() {
    if (isCancelling) return
    setIsCancelling(true)
    setCancelError(null)
    try {
      if (token) setCheckout(await cancelPublicCheckout(token))
      orderRequest.current = ''
      setCancelled(true)
      setSession(null)
      setError(null)
      setCancelDialogOpen(false)
      if (parentOrigin && isAllowedParentOrigin(parentOrigin) && window.parent !== window) {
        window.parent.postMessage({ type: 'paylanggan:checkout-cancelled' }, parentOrigin)
      }
    } catch (err) {
      setCancelError(err instanceof ApiError ? err.message : 'Checkout belum dapat dibatalkan. Periksa status pembayaran dan coba lagi.')
    } finally {
      setIsCancelling(false)
    }
  }

  const details = checkout ?? preview
  const currentStep = isPaid || trialComplete || completedOrder ? 4 : step
  const backButton = <button type="button" disabled={isStarting || isSubmitting || isAuthenticating || isVerifying} onClick={() => setStep((value) => Math.max(1, value - 1))} className="inline-flex min-h-11 items-center gap-2 rounded-xl border border-border px-4 text-sm font-semibold disabled:opacity-40"><ArrowLeft size={16} />Kembali</button>
  return (
    <main className="checkout-shell flex h-dvh min-h-0 flex-col overflow-hidden text-ink">
      <header className="checkout-header relative z-20 shrink-0 border-b border-border bg-white/80 px-6 py-4 dark:bg-slate-900">
        <div className="mx-auto max-w-lg"><Logo /></div>
      </header>
      <div className="checkout-stepper-header z-10 shrink-0"><div className="mx-auto max-w-lg px-6"><CheckoutStepper currentStep={currentStep} skipCard={Boolean(completedOrder)} /></div></div>
      <CheckoutScrollArea step={currentStep}>
        {isLoading ? <p role="status">Memuat informasi pembayaran...</p> : (error || quoteError) && !details ? <p role="alert" className="text-red-600">{error || quoteError}</p> : completedOrder ? (
          <section className="space-y-5 py-6 text-center"><CheckCircle2 className="mx-auto text-emerald-600" size={48} /><h1 className="font-display text-2xl font-bold">{completedOrder.subscription_status === 'trial' ? 'Trial berhasil dimulai' : 'Langganan berhasil dibuat'}</h1><p className="font-semibold">{details?.plan_name}</p>{completedOrder.subscription_status === 'trial' && completedOrder.trial_ends_at ? <p className="text-sm text-body">Trial berlaku sampai {new Date(completedOrder.trial_ends_at).toLocaleDateString('id-ID')}. Tidak ada pembayaran hari ini.</p> : <p className="text-sm text-body">Total tagihan Rp 0. Tidak diperlukan pembayaran kartu.</p>}</section>
        ) : trialComplete ? (
          <section className="space-y-5 py-6 text-center"><CheckCircle2 className="mx-auto text-emerald-600" size={48} /><h1 className="font-display text-2xl font-bold">Trial berhasil dimulai</h1><p className="font-semibold">{checkout?.plan_name}</p><p className="text-sm text-body">Kartu berhasil didaftarkan. Tidak ada biaya paket yang ditagihkan hari ini.</p>{checkout?.trial_ends_at && <p className="text-sm text-body">Trial berlaku sampai {new Date(checkout.trial_ends_at).toLocaleDateString('id-ID')}.</p>}</section>
        ) : isPaid ? (
          <section className="space-y-5 py-6 text-center"><CheckCircle2 className="mx-auto text-emerald-600" size={48} /><h1 className="font-display text-2xl font-bold">Pembayaran berhasil</h1><p className="text-sm text-body">Tagihan {checkout?.invoice_number} telah lunas.</p><div className="rounded-2xl border border-border bg-white/70 p-5"><p className="font-semibold">{details?.plan_name}</p><p className="mt-2 text-2xl font-bold">{formatCurrency(details?.amount ?? 0)}</p><p className="mt-2 text-sm text-body">{checkout?.customer_name}</p></div>{portalSite && <Link to={`/portal/${encodeURIComponent(portalSite)}`} className="inline-flex rounded-full bg-brand px-5 py-3 text-sm font-semibold text-white focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand">Kembali ke portal pelanggan</Link>}</section>
        ) : cancelled || checkout?.checkout_status === 'cancelled' ? <div className="space-y-3 py-6"><CheckCircle2 className="text-brand" /><h1 className="text-xl font-bold">Checkout dibatalkan</h1><p className="text-sm text-body">Anda dapat kembali ke website penjual dan memulai checkout baru. Data pelanggan yang sudah ada tetap tersimpan.</p></div> : isUnavailable ? <div className="space-y-3"><AlertCircle className="text-red-600" /><h1 className="text-xl font-bold">Pembayaran tidak tersedia</h1><p className="text-sm text-body">Hubungi penyedia layanan untuk mendapatkan tautan pembayaran baru.</p></div> : details ? (
          <>
            <p className="mb-1 text-xs font-semibold uppercase tracking-wide text-brand">{details.business_name}</p>
            <h1 className="font-display text-xl font-bold leading-7">{step === 1 ? 'Informasi paket' : step === 2 ? 'Informasi pelanggan' : isAuthenticating ? 'Verifikasi kartu' : isTrialRegistration ? 'Pendaftaran kartu' : 'Pembayaran kartu'}</h1>
            {step === 1 && <section className="mt-4 space-y-4">
              <div className="rounded-2xl border border-border bg-white/70 p-5"><h2 className="text-lg font-semibold">{details.plan_name}</h2>{preview && <p className="mt-1 text-sm text-body">{preview.billing_cycle === 'yearly' ? 'Langganan tahunan' : 'Langganan bulanan'}</p>}
                {checkout?.items?.map((item, index) => <div key={index} className="mt-4 flex justify-between gap-4 text-sm"><span className="text-body">{item.type === 'discount' ? 'Kupon / diskon: ' : item.type === 'add_on' ? 'Add-on: ' : ''}{item.description}{item.quantity > 1 ? ` × ${item.quantity}` : ''}</span><span className="shrink-0 font-medium">{formatCurrency(item.subtotal)}</span></div>)}
                {!token && preview?.items?.filter((item) => item.type !== 'plan').map((item, index) => <div key={index} className="mt-3 flex justify-between gap-4 text-sm"><span>{item.name}{item.quantity > 1 ? ` × ${item.quantity}` : ''}</span><span>{formatCurrency(item.subtotal)}</span></div>)}
                <div className="mt-5 flex items-center justify-between gap-3 border-t border-border pt-4"><span className="text-sm text-body">{isTrialRegistration ? 'Harga setelah trial' : 'Total pembayaran'}</span><strong className="text-xl">{formatCurrency(details.amount)}</strong></div>
                {isTrialRegistration && <div className="mt-4 rounded-xl bg-brand/10 p-3 text-sm"><strong>Trial gratis {checkout?.trial_days ?? preview?.trial_days} hari · Bayar hari ini Rp 0</strong><p className="mt-2 text-body">Daftarkan kartu untuk memulai trial. Biaya paket tidak ditagihkan hari ini. Kupon berlaku selama belum kedaluwarsa.</p></div>}
              </div>
              {!token && preview && <CheckoutOptions preview={preview} addOns={addOns} couponCode={couponCode} busy={quoteBusy} error={quoteError} onAddOns={(items) => { setQuoteBusy(true); setAddOns(items) }} onCoupon={(code) => { setQuoteBusy(true); setCouponCode(code) }} />}
              <p className="text-sm text-body">Periksa paket dan rincian tagihan sebelum melanjutkan.</p>
            </section>}
            {step === 2 && <form id="checkout-customer-form" onSubmit={handleStart} className="mt-4 space-y-4">
              <p className="text-sm text-body">Gunakan kontak yang dapat dihubungi untuk langganan Anda.</p>
              <Input id="checkout-name" label="Nama lengkap" value={name} onChange={(event) => setName(event.target.value)} required autoComplete="name" readOnly={Boolean(session)} />
              <Input id="checkout-email" label="Email" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required autoComplete="email" readOnly={Boolean(session)} />
              <Input id="checkout-mobile" label="Nomor ponsel" type="tel" value={mobile} onChange={(event) => setMobile(event.target.value)} placeholder="+628123456789" required autoComplete="tel" readOnly={Boolean(session)} />
              {isTrialRegistration && !session && <label className="flex items-start gap-3 text-sm text-body"><input type="checkbox" required checked={cardConsent} onChange={(event) => setCardConsent(event.target.checked)} className="mt-1 size-4 accent-brand" />{trialAutoDebit
                ? <>Saya menyetujui penyimpanan token kartu di {checkout?.payment_provider === 'midtrans' ? 'Midtrans' : 'Xendit'} dan debit otomatis sebesar <strong>{formatCurrency(details.amount)} / {trialCycleLabel}</strong> mulai {trialBillingDate}, selama langganan aktif. Persetujuan berlaku maksimal sampai {trialDebitExpiryLabel} atau sampai saya membatalkan. Tidak ada biaya paket hari ini.</>
                : <>Saya menyetujui pendaftaran kartu untuk memulai trial. Tidak ada biaya paket yang ditagihkan hari ini.</>}</label>}
              {session && <p className="text-xs text-muted">Kontak telah digunakan untuk menyiapkan sesi pembayaran ini.</p>}
              {error && <p role="alert" className="text-sm text-red-600">{error}</p>}
            </form>}
            {session && <section hidden={step !== 3} className="mt-2 space-y-3">
              <div className="flex justify-between gap-4 rounded-xl border border-border bg-white/60 px-4 py-2 text-sm"><span>{details.plan_name}</span><strong className="shrink-0">{isTrialRegistration ? 'Hari ini Rp 0' : formatCurrency(details.amount)}</strong></div>
              {isVerifying ? <div role="status" className="flex items-start gap-3 py-6 text-brand"><Clock3 size={22} /><p className="text-sm">{isTrialRegistration ? 'Kami sedang memeriksa pendaftaran kartu Anda. Trial dimulai setelah proses ini selesai.' : 'Kami sedang memastikan pembayaran Anda. Status akan diperbarui secara otomatis.'}</p>{verificationDelayed && <button type="button" className="shrink-0 rounded-lg border border-border px-3 py-2 text-sm font-semibold focus-visible:outline-2 focus-visible:outline-brand" onClick={() => { void getPublicCheckout(token).then(setCheckout).catch(() => setError('Status belum dapat diperiksa. Coba lagi.')) }}>Periksa status</button>}{verificationDelayed && error && <p role="alert" className="text-sm text-red-600">{error}</p>}</div> : <>
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
      <footer className="checkout-footer z-10 shrink-0">
        <div className="mx-auto w-full max-w-lg px-6 py-3">
          {!isLoading && details && !isPaid && !trialComplete && !completedOrder && !isUnavailable && <div className="checkout-actions flex items-center gap-3">
            {step > 1 && backButton}
            {step === 1 && <SubmitButton type="button" onClick={() => setStep(2)} disabled={!token && (quoteBusy || Boolean(quoteError))} className="w-full justify-center gap-2">Lanjut<ArrowRight size={16} /></SubmitButton>}
            {step === 2 && <SubmitButton type="submit" form="checkout-customer-form" disabled={isStarting || (!token && (quoteBusy || Boolean(quoteError)))} className="flex-1 justify-center gap-2">{isStarting ? 'Menyiapkan...' : 'Lanjut'}<ArrowRight size={16} /></SubmitButton>}
            {step === 3 && session?.provider === 'xendit' && !isAuthenticating && !isVerifying && <SubmitButton type="button" onClick={handlePay} disabled={!isCardReady || isSubmitting} className="flex-1 justify-center gap-2"><CreditCard size={17} />{isSubmitting ? 'Memproses...' : isTrialRegistration ? 'Verifikasi kartu dan mulai trial' : `Bayar ${formatCurrency(details.amount)}`}</SubmitButton>}
            <div ref={setCardActions} className="min-w-0 flex-1 empty:hidden" />
          </div>}
          {!isLoading && !isPaid && !trialComplete && !completedOrder && !isUnavailable && <button type="button" disabled={isStarting || isSubmitting || isAuthenticating || isVerifying || isCancelling} onClick={() => { setCancelError(null); setCancelDialogOpen(true) }} className="mx-auto mt-2 block rounded-lg px-3 py-1 text-xs font-semibold text-body underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-brand disabled:opacity-40">Batalkan checkout</button>}
          <p className="mt-2 flex items-center justify-center gap-2 text-xs text-body"><ShieldCheck size={14} />{isTrialRegistration ? 'Data kartu diproses oleh penyedia pembayaran' : 'Pembayaran terlindungi'}</p>
        </div>
      </footer>
      <ConfirmDialog open={cancelDialogOpen} onClose={() => setCancelDialogOpen(false)} onConfirm={() => { void handleCancel() }} title="Batalkan checkout?" description="Checkout yang belum selesai akan dibatalkan. Anda dapat memulai lagi dari website penjual. Data pelanggan yang sudah ada tidak dihapus." confirmLabel="Ya, batalkan checkout" busy={isCancelling} error={cancelError} />
    </main>
  )
}
