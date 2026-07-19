import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import AuthSplitLayout from '../components/layout/AuthSplitLayout'
import Input from '../components/ui/Input'
import SubmitButton from '../components/ui/SubmitButton'
import GoogleButton from '../components/ui/GoogleButton'
import { login, ApiError } from '../lib/api'
import { useAuth } from '../context/AuthContext'

export default function Login() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuth()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [rememberMe, setRememberMe] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setError(null)
    setIsSubmitting(true)

    try {
      const result = await login({ email, password })
      setAuth(result, rememberMe)
      navigate('/')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Terjadi kesalahan, coba lagi.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <AuthSplitLayout
      title="Selamat Datang Kembali"
      subtitle="Masukkan email dan password untuk melanjutkan."
      panelTitle="Kelola langganan dari satu dashboard."
      panelDescription="Katalog produk, pelanggan, dan siklus langganan — semuanya terhubung dalam satu platform multi-tenant."
      footer={
        <>
          Belum punya akun?{' '}
          <Link to="/register" className="font-semibold text-neutral hover:underline">
            Daftar
          </Link>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          id="email"
          label="Email"
          type="email"
          variant="dark"
          placeholder="nama@perusahaan.com"
          autoComplete="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
        />
        <Input
          id="password"
          label="Password"
          type="password"
          variant="dark"
          placeholder="••••••••"
          autoComplete="current-password"
          required
          value={password}
          onChange={(event) => setPassword(event.target.value)}
        />

        <div className="flex items-center justify-between text-[13px]">
          <label className="flex items-center gap-2 text-neutral/70">
            <input
              type="checkbox"
              checked={rememberMe}
              onChange={(event) => setRememberMe(event.target.checked)}
              className="h-4 w-4 rounded border-white/30 bg-white/10 accent-brand"
            />
            Ingat saya
          </label>
          <a href="#" className="font-semibold text-neutral/70 hover:text-neutral">
            Lupa Password?
          </a>
        </div>

        {error && <p className="text-[13px] text-red-400">{error}</p>}

        <SubmitButton type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? 'Memproses...' : 'Masuk'}
        </SubmitButton>
      </form>

      <div className="my-6 flex items-center gap-3">
        <span className="h-px flex-1 bg-white/10" />
        <span className="text-[12px] text-neutral/40">atau lanjutkan dengan</span>
        <span className="h-px flex-1 bg-white/10" />
      </div>

      <GoogleButton label="Masuk dengan Google" />
    </AuthSplitLayout>
  )
}
