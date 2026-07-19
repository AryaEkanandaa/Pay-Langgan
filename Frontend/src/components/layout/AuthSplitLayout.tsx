import type { ReactNode } from 'react'
import Logo from '../ui/Logo'
import GlassPanel from '../ui/GlassPanel'
import Starburst from '../ui/Starburst'
import { statusRows } from '../../data/subscriptionPreview'

interface AuthSplitLayoutProps {
  title: string
  subtitle: string
  children: ReactNode
  footer: ReactNode
  panelTitle: string
  panelDescription: string
}

export default function AuthSplitLayout({
  title,
  subtitle,
  children,
  footer,
  panelTitle,
  panelDescription,
}: AuthSplitLayoutProps) {
  return (
    <div className="flex min-h-screen items-center justify-center px-4 py-10">
      <div className="relative w-full max-w-5xl">
        <div className="grid overflow-hidden rounded-[2.5rem] shadow-[0_40px_80px_-30px_rgba(11,13,23,0.55)] md:grid-cols-2">
          <div className="flex flex-col justify-center border border-white/10 bg-ink px-8 py-12 sm:px-12">
            <Logo variant="light" className="mb-10" />

            <h1 className="font-display text-[26px] font-bold text-neutral">{title}</h1>
            <p className="mt-2 text-[14px] text-neutral/60">{subtitle}</p>

            <div className="mt-8">{children}</div>

            <p className="mt-6 text-center text-[13px] text-neutral/60">{footer}</p>
          </div>

          <div className="relative hidden overflow-hidden border-y border-r border-white/5 bg-ink px-10 py-14 md:flex md:flex-col md:justify-center">
            <div className="pointer-events-none absolute -right-12 -top-12 h-72 w-72 rounded-full bg-violet/40 blur-3xl" />
            <Starburst className="pointer-events-none absolute -right-8 top-8 h-52 w-52 text-violet/50" />

            <div className="relative z-10 max-w-[260px]">
              <span className="font-display text-[48px] leading-none text-brand/50">&ldquo;</span>
              <h2 className="-mt-3 font-display text-[26px] font-bold leading-tight text-neutral">
                {panelTitle}
              </h2>
              <p className="mt-4 text-[14px] leading-relaxed text-neutral/60">
                {panelDescription}
              </p>
            </div>
          </div>
        </div>

        {/* <GlassPanel className="absolute bottom-6 right-8 hidden w-56 border-white/60 bg-white/95 shadow-xl md:block">
          <p className="text-[11px] font-semibold uppercase tracking-wide text-muted">
            Status Langganan
          </p>
          <div className="mt-3 space-y-2.5">
            {statusRows.slice(0, 2).map((row) => (
              <div key={row.label} className="space-y-1">
                <div className="flex items-center justify-between text-[12px]">
                  <span className="text-body">{row.label}</span>
                  <span className="font-semibold text-ink">{row.value}%</span>
                </div>
                <div className="h-1.5 w-full rounded-full bg-surface">
                  <div
                    className={`h-1.5 rounded-full ${row.className}`}
                    style={{ width: `${row.value}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </GlassPanel> */}
      </div>
    </div>
  )
}
