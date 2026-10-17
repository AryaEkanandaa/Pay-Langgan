import { useEffect, useRef, useState, type FormEvent } from 'react'
import { Link, useLocation } from 'react-router-dom'
import AuthSplitLayout from '../components/layout/AuthSplitLayout'
import Input from '../components/ui/Input'
import SubmitButton from '../components/ui/SubmitButton'
import { ApiError, resetPassword } from '../lib/api'
import { useAuth } from '../context/AuthContext'

export default function ResetPassword() {
  // Fragment tokens never reach the frontend HTTP server or its request logs.
  const { hash } = useLocation()
  const token = new URLSearchParams(hash.slice(1)).get('token') ?? ''
  const [password, setPassword] = useState('')
  const [confirmation, setConfirmation] = useState('')
  const [busy, setBusy] = useState(false)
  const [done, setDone] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const pending = useRef<AbortController | null>(null)
  const { logout } = useAuth()
  const validToken = /^[A-Za-z0-9_-]{43}$/.test(token)

  useEffect(() => {
    const meta = document.createElement('meta')
    meta.name = 'referrer'; meta.content = 'no-referrer'; document.head.append(meta)
    function stop() { pending.current?.abort(); meta.remove() }
    return stop
  }, [])

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (busy || !validToken) return
    if (password !== confirmation) { setError('Konfirmasi password belum sama.'); return }
    const bytes = new TextEncoder().encode(password).length
    if (Array.from(password).length < 8 || bytes > 72) { setError('Gunakan password minimal 8 karakter, maksimal 72 byte.'); return }
    const controller = new AbortController(); pending.current = controller
    setBusy(true); setError(null)
    try {
      await resetPassword(token, password, confirmation, controller.signal)
      if (!controller.signal.aborted) {
        setDone(true); setPassword(''); setConfirmation(''); logout()
        window.history.replaceState(window.history.state, '', window.location.pathname)
      }
    } catch (err) {
      if (!controller.signal.aborted) setError(err instanceof ApiError ? err.message : 'Password belum dapat diperbarui. Periksa koneksi Anda lalu coba lagi.')
    } finally { if (!controller.signal.aborted) setBusy(false) }
  }

  return <AuthSplitLayout title={done ? 'Password diperbarui' : 'Buat password baru'} subtitle={done ? 'Gunakan password baru untuk masuk kembali ke akun Anda.' : 'Pilih password baru untuk akun PayLanggan Anda.'} panelTitle="Akun Anda, terlindungi." panelDescription="Tautan reset hanya berlaku selama 15 menit dan tidak dapat digunakan kembali setelah password diperbarui." footer={<Link to="/login" className="font-semibold text-neutral hover:underline">Kembali ke halaman masuk</Link>}>
    {done ? <div className="space-y-5"><p role="status" aria-live="polite" className="rounded-xl border border-emerald-400/30 bg-emerald-950/40 p-4 text-[13px] leading-6 text-emerald-100">Password berhasil diperbarui. Sesi login lama telah dinonaktifkan.</p><Link to="/login" className="inline-flex w-full justify-center rounded-full bg-brand px-5 py-3 text-sm font-semibold text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand">Masuk dengan password baru</Link></div> : !validToken ? <div className="space-y-4"><p role="alert" className="text-sm leading-6 text-red-300">Tautan reset tidak lengkap atau tidak valid. Silakan minta tautan baru.</p><Link to="/forgot-password" className="font-semibold text-neutral underline">Minta tautan baru</Link></div> : <form onSubmit={submit} className="space-y-4">
      <Input id="new-password" label="Password baru" type="password" variant="dark" autoComplete="new-password" required minLength={8} maxLength={72} value={password} onChange={event => setPassword(event.target.value)} disabled={busy} />
      <Input id="confirm-password" label="Konfirmasi password baru" type="password" variant="dark" autoComplete="new-password" required minLength={8} maxLength={72} value={confirmation} onChange={event => setConfirmation(event.target.value)} disabled={busy} />
      <p className="text-[12px] leading-5 text-neutral/70">Minimal 8 karakter. Gunakan password yang unik untuk akun ini.</p>
      {error && <div role="alert" className="space-y-2 text-[13px] text-red-300"><p>{error}</p><Link to="/forgot-password" className="font-semibold text-neutral underline">Minta tautan baru</Link></div>}
      <SubmitButton type="submit" className="w-full" disabled={busy}>{busy ? 'Menyimpan…' : 'Simpan password baru'}</SubmitButton>
    </form>}
  </AuthSplitLayout>
}
