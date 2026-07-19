const columns = [
  {
    title: 'Produk',
    links: ['Fitur', 'Harga', 'Integrasi Xendit', 'Dashboard Pendapatan'],
  },
  {
    title: 'Perusahaan',
    links: ['Tentang Kami', 'Karier', 'Kontak'],
  },
  {
    title: 'Sumber Daya',
    links: ['Dokumentasi', 'Status Sistem', 'Dukungan'],
  },
]

export default function Footer() {
  return (
    <footer className="border-t border-border bg-neutral">
      <div className="mx-auto max-w-6xl px-5 py-16">
        <div className="grid gap-10 md:grid-cols-[1.3fr_1fr_1fr_1fr]">
          <div>
            <div className="flex items-center gap-2">
              <span className="flex h-8 w-8 items-center justify-center rounded-md bg-primary">
                <span className="h-2 w-2 rounded-full bg-tertiary" />
              </span>
              <span className="font-display text-[18px] font-bold text-primary">
                PayLanggan
              </span>
            </div>
            <p className="mt-4 max-w-xs text-[13px] leading-relaxed text-secondary">
              Sistem manajemen langganan dan pendapatan layanan SaaS berbasis
              web, dirancang untuk penyedia layanan multi-tenant.
            </p>
          </div>

          {columns.map((column) => (
            <div key={column.title}>
              <p className="text-[12px] font-semibold uppercase tracking-wide text-muted">
                {column.title}
              </p>
              <ul className="mt-4 space-y-3">
                {column.links.map((link) => (
                  <li key={link}>
                    <a
                      href="#"
                      className="text-[14px] text-secondary transition-colors hover:text-primary"
                    >
                      {link}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="mt-14 flex flex-col items-center justify-between gap-4 border-t border-border pt-8 text-[13px] text-muted md:flex-row">
          <p>© {new Date().getFullYear()} PayLanggan. Seluruh hak cipta dilindungi.</p>
          <div className="flex items-center gap-6">
            <a href="#" className="hover:text-primary">
              Kebijakan Privasi
            </a>
            <a href="#" className="hover:text-primary">
              Syarat Layanan
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}
