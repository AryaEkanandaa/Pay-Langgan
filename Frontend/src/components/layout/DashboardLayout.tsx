import { Link, Outlet, useLocation } from 'react-router-dom'
import { LogOut } from 'lucide-react'
import Logo from '../ui/Logo'
import { dashboardNavItems } from '../../data/dashboardNav'
import { useAuth } from '../../context/AuthContext'

export default function DashboardLayout() {
  const { auth, logout } = useAuth()
  const location = useLocation()

  return (
    <div className="flex min-h-screen">
      <aside className="hidden w-64 flex-col bg-gradient-to-b from-ink to-deep px-5 py-6 md:flex">
        <Logo className="mb-8" variant="light" />

        <nav className="flex flex-1 flex-col gap-1">
          {dashboardNavItems.map((item) => {
            const isActive = location.pathname === item.to
            const Icon = item.icon
            return (
              <Link
                key={item.to}
                to={item.to}
                className={`flex items-center gap-3 rounded-xl px-3 py-2.5 text-[14px] font-medium transition-colors ${
                  isActive ? 'bg-brand text-white' : 'text-white/60 hover:bg-white/10 hover:text-white'
                }`}
              >
                <Icon size={18} strokeWidth={2} />
                {item.label}
              </Link>
            )
          })}
        </nav>

        <div className="mt-6 border-t border-white/10 pt-4">
          <p className="truncate text-[13px] font-semibold text-white">{auth?.user.name}</p>
          <p className="truncate text-[12px] text-white/50">{auth?.user.email}</p>
          <button
            onClick={logout}
            className="mt-3 flex w-full items-center gap-2 rounded-xl px-3 py-2 text-left text-[13px] font-medium text-white/60 transition-colors hover:bg-white/10 hover:text-white"
          >
            <LogOut size={16} strokeWidth={2} />
            Keluar
          </button>
        </div>
      </aside>

      <div className="flex flex-1 flex-col">
        <header className="flex items-center justify-between bg-gradient-to-b from-ink to-deep px-5 py-4 md:hidden">
          <Logo variant="light" />
          <button onClick={logout} className="text-[13px] font-semibold text-white">
            Keluar
          </button>
        </header>

        <nav className="flex gap-1 overflow-x-auto bg-gradient-to-b from-ink to-deep px-4 py-2 md:hidden">
          {dashboardNavItems.map((item) => {
            const isActive = location.pathname === item.to
            return (
              <Link
                key={item.to}
                to={item.to}
                className={`whitespace-nowrap rounded-full px-3 py-1.5 text-[13px] font-medium transition-colors ${
                  isActive ? 'bg-brand text-white' : 'text-white/60 hover:bg-white/10 hover:text-white'
                }`}
              >
                {item.label}
              </Link>
            )
          })}
        </nav>

        <main className="flex-1 px-5 py-8 md:px-10 md:py-10">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
