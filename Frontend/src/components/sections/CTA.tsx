import { Link } from 'react-router-dom'

export default function CTA() {
  return (
    <div className="landing-reveal w-full px-5 py-4">
      <div className="glass-dark mx-auto max-w-6xl rounded-3xl px-8 py-16 text-center md:px-16">
        <h2 className="font-display text-[28px] font-bold leading-tight tracking-[-0.02em] text-neutral md:text-[34px]">
          Siap merapikan langganan dan penagihan bisnis Anda?
        </h2>
        <p className="mx-auto mt-4 max-w-lg text-[16px] leading-relaxed text-neutral/70">
          Daftarkan bisnis Anda dan lihat bagaimana PayLanggan membantu tim
          Anda mengelola pelanggan, paket, dan penagihan dalam satu tempat.
        </p>
        <div className="mt-8 flex flex-wrap items-center justify-center gap-4">
          <Link
            to="/register"
            className="inline-flex h-11 items-center justify-center rounded-full bg-brand px-6 text-[14px] font-semibold text-neutral shadow-[0_10px_30px_-10px_rgba(54,84,255,0.55)] transition-all hover:bg-brand-dark active:scale-[0.98]"
          >
            Daftar Sekarang
          </Link>
          <a
            href="#fitur"
            className="inline-flex h-11 items-center justify-center rounded-full border border-white/20 bg-white/5 px-6 text-[14px] font-semibold text-neutral backdrop-blur-xl transition-colors hover:bg-white/10"
          >
            Lihat Fitur
          </a>
        </div>
      </div>
    </div>
  )
}
