import Logo from '../ui/Logo'
import { footerColumns } from '../../data/navigation'

export default function Footer() {
  return (
    <footer className="landing-reveal glass border-x-0 border-b-0" style={{ transitionDelay: '160ms' }}>
      <div className="mx-auto max-w-6xl px-5 py-16">
        <div className="grid gap-10 md:grid-cols-[1.3fr_1fr_1fr_1fr]">
          <div>
            <Logo />
            <p className="mt-4 max-w-xs text-[13px] leading-relaxed text-body">
              Platform manajemen langganan dan penagihan untuk bisnis SaaS
              multi-tenant.
            </p>
          </div>

          {footerColumns.map((column) => (
            <div key={column.title}>
              <p className="text-[12px] font-semibold uppercase tracking-wide text-muted">
                {column.title}
              </p>
              <ul className="mt-4 space-y-3">
                {column.links.map((link) => (
                  <li key={link}>
                    <a
                      href="#"
                      className="text-[14px] text-body transition-colors hover:text-ink"
                    >
                      {link}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="mt-14 flex flex-col items-center justify-between gap-4 border-t border-white/40 pt-8 text-[13px] text-muted md:flex-row">
          <p>© {new Date().getFullYear()} PayLanggan. Seluruh hak cipta dilindungi.</p>
          <div className="flex items-center gap-6">
            <a href="#" className="hover:text-ink">
              Kebijakan Privasi
            </a>
            <a href="#" className="hover:text-ink">
              Syarat Layanan
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}
