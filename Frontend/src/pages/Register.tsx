import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import AuthSplitLayout from '../components/layout/AuthSplitLayout'
import Input from '../components/ui/Input'
import SubmitButton from '../components/ui/SubmitButton'
import GoogleButton from '../components/ui/GoogleButton'
import { signup, ApiError } from '../lib/api'
import { useAuth } from '../context/AuthContext'

export default function Register() {
  const navigate = useNavigate()
  const { login: setAuth } = useAuth()

  const [businessName, setBusinessName] = useState('')
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setError(null)
    setIsSubmitting(true)

    try {
      const result = await signup({
        business_name: businessName,
        name,
        email,
        password,
      })
      setAuth(result)
      navigate('/dashboard')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Terjadi kesalahan, coba lagi.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <AuthSplitLayout
      title="Daftarkan Bisnis Anda"
      subtitle="Buat akun untuk mulai mengelola katalog, pelanggan, dan langganan."
      panelTitle="Satu platform untuk seluruh siklus langganan."
      panelDescription="Dari katalog produk sampai pelanggan aktif, semuanya tercatat rapi dalam satu tempat."
      footer={
        <>
          Sudah punya akun?{' '}
          <Link to="/login" className="font-semibold text-neutral hover:underline">
            Masuk
          </Link>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          id="business_name"
          label="Nama Bisnis"
          type="text"
          variant="dark"
          placeholder="PT Bisnis Anda"
          autoComplete="organization"
          required
          value={businessName}
          onChange={(event) => setBusinessName(event.target.value)}
        />
        <Input
          id="name"
          label="Nama Anda"
          type="text"
          variant="dark"
          placeholder="Nama lengkap"
          autoComplete="name"
          required
          value={name}
          onChange={(event) => setName(event.target.value)}
        />
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
          placeholder="Minimal 6 karakter"
          autoComplete="new-password"
          minLength={6}
          required
          value={password}
          onChange={(event) => setPassword(event.target.value)}
        />

        {error && <p className="text-[13px] text-red-400">{error}</p>}

        <SubmitButton type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? 'Memproses...' : 'Daftar'}
        </SubmitButton>
      </form>

      <div className="my-6 flex items-center gap-3">
        <span className="h-px flex-1 bg-white/10" />
        <span className="text-[12px] text-neutral/40">atau lanjutkan dengan</span>
        <span className="h-px flex-1 bg-white/10" />
      </div>

      <GoogleButton label="Daftar dengan Google" />
    </AuthSplitLayout>
  )
}
