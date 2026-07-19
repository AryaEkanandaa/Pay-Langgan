export default function CTA() {
  return (
    <section id="tentang" className="px-5 py-4">
      <div className="mx-auto max-w-6xl overflow-hidden rounded-xl bg-primary px-8 py-16 text-center md:px-16">
        <h2 className="font-display text-[28px] font-bold leading-tight tracking-[-0.02em] text-neutral md:text-[34px]">
          Siap merapikan langganan dan pendapatan bisnis Anda?
        </h2>
        <p className="mx-auto mt-4 max-w-lg text-[16px] leading-relaxed text-accent">
          Bergabung dengan penyedia layanan SaaS yang sudah mengotomatisasi
          penagihan mereka bersama PayLanggan.
        </p>
        <div className="mt-8 flex flex-wrap items-center justify-center gap-4">
          <a href="#signup" className="btn-primary">
            Mulai Uji Coba Gratis
          </a>
          <a href="#demo" className="btn-secondary bg-transparent text-neutral ring-neutral/40">
            Jadwalkan Demo
          </a>
        </div>
      </div>
    </section>
  )
}
