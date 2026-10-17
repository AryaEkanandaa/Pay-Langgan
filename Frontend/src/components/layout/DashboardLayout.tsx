import { Fragment, useEffect, useState } from 'react'
import { Link, Outlet, useLocation } from 'react-router-dom'
import { ChevronDown, LogOut } from 'lucide-react'
import Logo from '../ui/Logo'
import ThemeToggle from '../ui/ThemeToggle'
import { dashboardNavItems } from '../../data/dashboardNav'
import { useAuth } from '../../context/AuthContext'
import NavBadge from '../ui/NavBadge'
import { useNewWebhookCount, WebhookNotificationsContext } from '../../hooks/useNewWebhookCount'

export default function DashboardLayout() {
  const { auth, logout } = useAuth()
  const location = useLocation()
  const webhooks = useNewWebhookCount(auth?.token, auth?.user.role === 'admin')
  const [openMenu, setOpenMenu] = useState<string | null>(null)
  const visibleNavItems = dashboardNavItems.filter(
    (item) => !item.allowedRoles || item.allowedRoles.includes(auth?.user.role ?? 'admin'),
  )
  const isPathActive = (to: string) =>
    location.pathname === to || (to !== '/dashboard' && location.pathname.startsWith(`${to}/`))
  const isNavItemActive = (item: (typeof dashboardNavItems)[number]) =>
    isPathActive(item.to) || Boolean(item.children?.some((child) => isPathActive(child.to.split('#')[0])))

  useEffect(() => {
    const matches = (to: string) => location.pathname === to || (to !== '/dashboard' && location.pathname.startsWith(`${to}/`))
    const activeItem = dashboardNavItems.find(item => item.children && (!item.allowedRoles || item.allowedRoles.includes(auth?.user.role ?? 'admin')) && (matches(item.to) || item.children.some(child => matches(child.to.split('#')[0]))))
    setOpenMenu(activeItem?.to ?? null)
  }, [location.pathname, auth?.user.role])

  return (
    <WebhookNotificationsContext.Provider value={webhooks}>
    <div className="flex min-h-screen">
      <aside className="sticky top-0 hidden h-screen max-h-screen w-64 shrink-0 flex-col overflow-y-auto bg-gradient-to-b from-ink to-deep px-5 py-6 md:flex">
        <div className="mb-8 flex items-center justify-between gap-2">
          <Logo variant="light" />
          <ThemeToggle compact />
        </div>

        <nav className="flex flex-1 flex-col gap-1">
          {visibleNavItems.map((item) => {
            const isActive = isNavItemActive(item)
            const Icon = item.icon
            return (
              <div key={item.to}>
                <div className="flex items-center">
                  <Link
                    to={item.to}
                    aria-current={isActive ? 'page' : undefined}
                    aria-label={item.label === 'Webhook Logs' ? `Webhook Logs, ${webhooks.count} baru` : undefined}
                    className={`flex min-w-0 flex-1 items-center gap-3 rounded-xl px-3 py-2.5 text-[14px] font-medium transition-colors ${
                      isActive ? 'bg-brand text-white' : 'text-white/60 hover:bg-white/10 hover:text-white'
                    }`}
                  >
                    <Icon size={18} strokeWidth={2} />
                    <span className="truncate">{item.label}</span>
                    {item.label === 'Webhook Logs' && <span className="ml-auto"><NavBadge count={webhooks.count} /></span>}
                  </Link>
                  {item.children && (
                    <button
                      type="button"
                      aria-label={`${openMenu === item.to ? 'Tutup' : 'Buka'} submenu ${item.label}`}
                      aria-expanded={openMenu === item.to}
                      onClick={() => setOpenMenu((current) => (current === item.to ? null : item.to))}
                      className={`ml-1 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-white/50 transition-colors hover:bg-white/10 hover:text-white ${
                        isActive ? 'text-white/80' : ''
                      }`}
                    >
                      <ChevronDown size={16} className={openMenu === item.to ? 'rotate-180' : ''} />
                    </button>
                  )}
                </div>
                {item.children && openMenu === item.to && (
                  <div className="ml-9 mt-1 space-y-0.5 border-l border-white/10 pl-3">
                    {item.children.map((child) => {
                      const [childPath, childHash] = child.to.split('#')
                      const childActive = isPathActive(childPath) && (!childHash || location.hash === `#${childHash}`)
                      return (
                        <Link
                          key={child.to}
                          to={child.to}
                          className={`block rounded-lg px-3 py-2 text-[13px] transition-colors ${
                            childActive ? 'bg-white/15 text-white' : 'text-white/50 hover:bg-white/10 hover:text-white'
                          }`}
                        >
                          {child.label}
                        </Link>
                      )
                    })}
                  </div>
                )}
              </div>
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

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex items-center justify-between bg-gradient-to-b from-ink to-deep px-5 py-4 md:hidden">
          <Logo variant="light" />
          <div className="flex items-center gap-2">
            <ThemeToggle compact />
            <button onClick={logout} className="text-[13px] font-semibold text-white">
              Keluar
            </button>
          </div>
        </header>

        <nav className="flex gap-1 overflow-x-auto bg-gradient-to-b from-ink to-deep px-4 py-2 md:hidden">
          {visibleNavItems.map((item) => {
            const isActive = isNavItemActive(item)
            return (
              <Fragment key={item.to}>
                <Link
                  to={item.to}
                  aria-current={isActive ? 'page' : undefined}
                  aria-label={item.label === 'Webhook Logs' ? `Webhook Logs, ${webhooks.count} baru` : undefined}
                  className={`whitespace-nowrap rounded-full px-3 py-1.5 text-[13px] font-medium transition-colors ${
                    isActive ? 'bg-brand text-white' : 'text-white/60 hover:bg-white/10 hover:text-white'
                  }`}
                >
                  {item.label}
                  {item.label === 'Webhook Logs' && <span className="ml-2 inline-flex align-middle"><NavBadge count={webhooks.count} /></span>}
                </Link>
                {isActive && item.children?.map((child) => {
                  const [childPath, childHash] = child.to.split('#')
                  const childActive = isPathActive(childPath) && (!childHash || location.hash === `#${childHash}`)
                  return (
                    <Link
                      key={child.to}
                      to={child.to}
                      className={`whitespace-nowrap rounded-full px-3 py-1.5 text-[13px] font-medium ${
                        childActive ? 'bg-white/20 text-white' : 'bg-white/10 text-white/70'
                      }`}
                    >
                      {child.label}
                    </Link>
                  )
                })}
              </Fragment>
            )
          })}
        </nav>

        <main className={`flex-1 py-8 md:py-10 ${['/dashboard', '/dashboard/integrasi', '/dashboard/pengaturan', '/dashboard/pengaturan/payment-gateway'].includes(location.pathname) ? 'px-6 min-[900px]:px-10' : 'px-5 md:px-10'}`}>
          <Outlet />
        </main>
      </div>
    </div>
    </WebhookNotificationsContext.Provider>
  )
}
