import GlassPanel from '../../components/ui/GlassPanel'
import { useAuth } from '../../context/AuthContext'

export default function PlatformAdminPage() {
  const { auth } = useAuth()

  return (
    <GlassPanel className="max-w-2xl">
      <p className="text-[11px] font-semibold uppercase tracking-wide text-muted">Platform</p>
      <h1 className="mt-2 font-display text-[26px] font-bold text-ink">Selamat datang, Super Admin</h1>
      <p className="mt-3 text-[14px] leading-6 text-body">
        Akun {auth?.user.email} berada di level platform. Modul pengelolaan seluruh bisnis dapat ditambahkan
        tanpa mencampur data tenant bisnis.
      </p>
    </GlassPanel>
  )
}
