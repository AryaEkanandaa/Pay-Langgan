const links = [
  { label: 'Fitur', href: '#fitur' },
  { label: 'Harga', href: '#harga' },
  { label: 'Tentang', href: '#tentang' },
]

export default function Navbar() {
  return (
    <header className="sticky top-0 z-50 border-b border-border/70 bg-neutral/80 backdrop-blur">
      <div className="mx-auto flex h-[72px] max-w-6xl items-center justify-between px-5">
        <a href="#" className="flex items-center gap-2">
          <span className="flex h-8 w-8 items-center justify-center rounded-md bg-primary">
            <span className="h-2 w-2 rounded-full bg-tertiary" />
          </span>
          <span className="font-display text-[18px] font-bold tracking-tight text-primary">
            PayLanggan
          </span>
        </a>

        <nav className="hidden items-center gap-8 md:flex">
          {links.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="text-[14px] font-medium text-secondary transition-colors hover:text-primary"
            >
              {link.label}
            </a>
          ))}
        </nav>

        <div className="flex items-center gap-3">
          <a href="#login" className="hidden text-[14px] font-semibold text-primary sm:inline">
            Masuk
          </a>
          <a href="#signup" className="btn-primary">
            Coba Gratis
          </a>
        </div>
      </div>
    </header>
  )
}
