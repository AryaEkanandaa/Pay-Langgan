import { useEffect, useRef, useState } from 'react'
import { LoaderCircle, Mail, RotateCcw } from 'lucide-react'
import { getInvoiceEmail, queueInvoiceEmail, type InvoiceEmailState } from '../../lib/api'

interface SendEmailButtonProps {
  token: string
  invoiceId: number
  state: InvoiceEmailState | null
  onStateChange: (state: InvoiceEmailState) => void
  disabled?: boolean
}

const POLL_LIMIT = 90_000

export default function SendEmailButton({ token, invoiceId, state, onStateChange, disabled = false }: SendEmailButtonProps) {
  const status = state?.delivery?.status
  const [requesting, setRequesting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [celebrating, setCelebrating] = useState(false)
  const previousStatus = useRef(status)
  const posting = useRef<AbortController | null>(null)
  const deadline = useRef<number | null>(null)

  useEffect(() => () => { posting.current?.abort(); posting.current = null }, [])
  useEffect(() => {
    if (status === 'sent' && previousStatus.current !== 'sent') {
      setCelebrating(true)
      const timer = window.setTimeout(() => setCelebrating(false), 3000)
      previousStatus.current = status
      return () => window.clearTimeout(timer)
    }
    previousStatus.current = status
  }, [status])

  useEffect(() => {
    if (status !== 'queued' && status !== 'sending') { deadline.current = null; return }
    deadline.current ??= Date.now() + POLL_LIMIT
    const controller = new AbortController()
    let timer: number | undefined
    const timeout = window.setTimeout(() => {
      controller.abort()
      setError('Status pengiriman belum selesai. Tutup dan buka kembali detail untuk memeriksa. Email tidak dikirim ulang.')
    }, Math.max(0, deadline.current - Date.now()))
    async function poll() {
      try {
        const next = await getInvoiceEmail(token, invoiceId, controller.signal)
        if (controller.signal.aborted) return
        setError(null)
        onStateChange(next)
        if (next.delivery?.status !== 'queued' && next.delivery?.status !== 'sending') return
      } catch {
        if (controller.signal.aborted) return
        setError('Koneksi terputus. Sedang memeriksa status pengiriman kembali.')
      }
      timer = window.setTimeout(() => void poll(), 1000)
    }
    timer = window.setTimeout(() => void poll(), 1000)
    return () => { controller.abort(); window.clearTimeout(timer); window.clearTimeout(timeout) }
  }, [token, invoiceId, status, onStateChange])

  async function send() {
    if (posting.current || disabled || !state || state.mode === 'disabled' || status === 'queued' || status === 'sending' || status === 'sent') return
    const controller = new AbortController()
    posting.current = controller
    setRequesting(true); setError(null)
    const timer = window.setTimeout(() => controller.abort(), 20_000)
    try {
      const next = await queueInvoiceEmail(token, invoiceId, controller.signal)
      if (!controller.signal.aborted) onStateChange(next)
    } catch {
      if (posting.current === controller) setError('Permintaan belum dapat dipastikan. Coba lagi untuk memeriksa atau melanjutkan antrean yang sama.')
    } finally {
      window.clearTimeout(timer)
      if (posting.current === controller) { posting.current = null; setRequesting(false) }
    }
  }

  const phase = requesting ? 'requesting' : celebrating ? 'sent' : status === 'queued' ? 'queued' : status === 'sending' ? 'processing' : status === 'failed' || (error && status !== 'sent') ? 'failed' : 'idle'
  const label = { idle: state?.mode === 'preview' ? 'Buat preview lokal' : 'Kirim email', requesting: 'Menyiapkan…', queued: 'Dalam antrean', processing: 'Mengirim…', sent: 'Terkirim', failed: 'Gagal, coba lagi' }[phase]
  const busy = requesting || status === 'queued' || status === 'sending'
  return <div className="min-w-0">
    <div className="send-email-slot">
      <button type="button" className="send-email-button" data-phase={phase} onClick={() => void send()} disabled={disabled || !state || state.mode === 'disabled' || busy || status === 'sent'} aria-label={label} aria-busy={busy}>
        <span key={phase} className="send-email-content" aria-hidden="true">
          {phase === 'queued' ? <span className="send-email-dots"><i /><i /><i /></span> : phase === 'processing' || phase === 'requesting' ? <LoaderCircle size={16} className="send-email-spinner" /> : phase === 'sent' ? <svg viewBox="0 0 24 24" width="18" height="18" fill="none"><path className="send-email-check" d="m5 12 4 4L19 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" /></svg> : phase === 'failed' ? <RotateCcw size={16} /> : <Mail size={16} />}
          {label}
        </span>
      </button>
    </div>
    <span role="status" aria-live="polite" aria-atomic="true" className="sr-only">{status === 'sent' ? 'Email terkirim.' : label}</span>
    {error && <p role="status" aria-live="polite" className="mt-2 max-w-xs text-xs leading-5 text-red-600 dark:text-red-300">{error}</p>}
  </div>
}
