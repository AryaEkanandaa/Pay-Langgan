import { useEffect, useRef, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import AuthSplitLayout from '../components/layout/AuthSplitLayout'
import Input from '../components/ui/Input'
import SubmitButton from '../components/ui/SubmitButton'
import { ApiError, requestPasswordReset } from '../lib/api'

export default function ForgotPassword() {
  const [email, setEmail] = useState('')
  const [busy, setBusy] = useState(false)
  const [sent, setSent] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const pending = useRef<AbortController | null>(null)
  useEffect(() => { function stop() { pending.current?.abort() }; return stop }, [])

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (busy) return
    const controller = new AbortController()
    pending.current = controller
    setBusy(true); setError(null); setSent(false)
    try {
      await requestPasswordReset(email.trim(), controller.signal)
      if (!controller.signal.aborted) setSent(true)
    } catch (err) {
      if (!controller.signal.aborted) setError(err instanceof ApiError ? err.message : 'Permintaan belum dapat dikirim. Periksa koneksi Anda lalu coba lagi.')
    } finally { if (!controller.signal.aborted) setBusy(false) }
  }

  return <AuthSplitLayout title="Lupa password?" subtitle="Masukkan email akun Anda untuk menerima tautan reset password." panelTitle="Kembali kelola bisnis Anda." panelDescription="Atur ulang password melalui email yang terdaftar pada akun PayLanggan Anda." footer={<Link to="/login" className="font-semibold text-neutral hover:underline">Kembali ke halaman masuk</Link>}>
    <form onSubmit={submit} className="space-y-4">
      <Input id="reset-email" label="Email akun" type="email" variant="dark" placeholder="nama@perusahaan.com" autoComplete="email" required maxLength={100} value={email} onChange={event => setEmail(event.target.value)} disabled={busy} />
      {sent && <p role="status" aria-live="polite" className="rounded-xl border border-emerald-400/30 bg-emerald-950/40 p-4 text-[13px] leading-6 text-emerald-100">Jika email terdaftar, tautan untuk mengatur ulang password akan dikirim. Periksa inbox dan folder spam. Tautan berlaku 15 menit.</p>}
      {error && <p role="alert" className="text-[13px] text-red-300">{error}</p>}
      <SubmitButton type="submit" className="w-full" disabled={busy}>{busy ? 'Memproses…' : 'Kirim tautan reset'}</SubmitButton>
      <p className="text-[12px] leading-5 text-neutral/70">Gunakan email yang sama dengan saat mendaftar. Jika belum menerima email, tunggu sebentar sebelum meminta lagi.</p>
    </form>
  </AuthSplitLayout>
}
