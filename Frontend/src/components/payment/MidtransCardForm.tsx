import { useEffect, useId, useRef, useState, type FormEvent } from 'react'
import { createPortal } from 'react-dom'
import { CreditCard, ShieldCheck } from 'lucide-react'
import SubmitButton from '../ui/SubmitButton'
import { chargeMidtransCard, registerMidtransTrialCard, type StartedCheckout } from '../../lib/api'
import { formatCurrency } from '../../lib/format'

type SDKResult = { token_id?: string; status_code?: string; saved_token_id?: string; transaction_id?: string }
type AuthOptions = {
  performAuthentication: (url: string) => void
  onSuccess: (result: SDKResult) => void
  onFailure: (result: SDKResult) => void
  onPending: (result: SDKResult) => void
}
declare global {
  interface Window {
    MidtransNew3ds?: {
      registerCard: (card: { card_number: string; card_exp_month: string; card_exp_year: string; card_cvv: string }, callbacks: {
        onSuccess: (result: SDKResult) => void; onFailure: (result: SDKResult) => void
      }) => void
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
  const registrationReceipt = useRef<{ saved_token_id: string; transaction_id: string } | null>(null)
  const isRegistration = Boolean(session.card_registration)

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
    if (isRegistration || !ready || resumed.current || !session.payment_session_id) return
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
    if (isRegistration) {
      try {
        const receipt = registrationReceipt.current ?? await new Promise<{ saved_token_id: string; transaction_id: string }>((resolve, reject) => {
          const timeout = window.setTimeout(() => reject(new Error('registration timeout')), 45000)
          try {
            window.MidtransNew3ds!.registerCard({ card_number: number.replace(/\D/g, ''), card_exp_month: String(month), card_exp_year: String(year), card_cvv: cvv }, {
              onSuccess: (result) => {
                window.clearTimeout(timeout)
                if (result.status_code === '200' && result.saved_token_id && result.transaction_id) resolve({ saved_token_id: result.saved_token_id, transaction_id: result.transaction_id })
                else reject(new Error('invalid registration receipt'))
              },
              onFailure: () => { window.clearTimeout(timeout); reject(new Error('registration failed')) },
            })
          } catch (err) { window.clearTimeout(timeout); reject(err) }
        })
        if (!mounted.current) return
        registrationReceipt.current = receipt
        await registerMidtransTrialCard(token, receipt)
        if (!mounted.current) return
        setNumber(''); setExpiry(''); setCvv('')
        registrationReceipt.current = null
        setBusy(false)
        onVerify()
      } catch {
        if (mounted.current) {
          submitting.current = false; setBusy(false)
          setError('Pendaftaran kartu belum selesai. Periksa data kartu dan coba lagi. Tidak ada biaya paket yang ditagihkan.')
        }
      }
      return
    }
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
    <CreditCard size={17} aria-hidden="true" />{busy ? (isRegistration ? 'Mendaftarkan kartu...' : 'Memproses pembayaran...') : isRegistration ? 'Daftarkan kartu dan mulai trial' : `Bayar ${formatCurrency(amount)}`}
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
