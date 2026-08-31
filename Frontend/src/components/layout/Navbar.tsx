import { Link } from 'react-router-dom'
import Logo from '../ui/Logo'
import { buttonBaseClass, buttonVariantClass } from '../ui/buttonStyles'
import { navLinks } from '../../data/navigation'
import { useAuth } from '../../context/AuthContext'
import ThemeToggle from '../ui/ThemeToggle'

export default function Navbar() {
  const { auth, logout } = useAuth()

  return (
    <header className="glass sticky top-0 z-50">
      <div className="mx-auto flex h-[72px] max-w-6xl items-center justify-between px-5">
        <Logo />

        <nav className="hidden items-center gap-8 md:flex">
          {navLinks.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="text-[14px] font-medium text-body transition-colors hover:text-ink"
            >
              {link.label}
            </a>
          ))}
        </nav>

        <div className="flex items-center gap-3">
          <ThemeToggle />
          {auth ? (
            <>
              <Link
                to="/dashboard"
                className="hidden text-[14px] font-semibold text-ink sm:inline"
              >
                Halo, {auth.user.name.split(' ')[0]}
              </Link>
              <button
                onClick={logout}
                className={`${buttonBaseClass} ${buttonVariantClass.secondary}`}
              >
                Keluar
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="hidden text-[14px] font-semibold text-ink sm:inline">
                Masuk
              </Link>
              <Link to="/register" className={`${buttonBaseClass} ${buttonVariantClass.primary}`}>
                Daftar
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  )
}
