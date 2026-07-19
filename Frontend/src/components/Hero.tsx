export default function Hero() {
  return (
    <section className="relative overflow-hidden bg-neutral">
      <div className="mx-auto grid max-w-6xl gap-16 px-5 pb-20 pt-16 md:grid-cols-2 md:items-center md:pb-32 md:pt-24">
        <div>
          <span className="chip bg-success/40 text-primary">Baru · Integrasi Xendit</span>

          <h1 className="mt-5 font-display text-[40px] font-bold leading-[1.08] tracking-[-0.03em] text-primary md:text-[48px]">
            Satu platform untuk langganan, penagihan, dan pendapatan SaaS Anda.
          </h1>

          <p className="mt-6 max-w-md text-[18px] leading-relaxed text-secondary">
            PayLanggan menyatukan pelanggan, paket layanan, penagihan berulang,
            pembayaran digital, dan analitik pendapatan dalam satu dashboard
            multi-tenant — tanpa spreadsheet, tanpa penagihan manual.
          </p>

          <div className="mt-8 flex flex-wrap items-center gap-4">
            <a href="#signup" className="btn-primary">
              Mulai Uji Coba Gratis
            </a>
            <a href="#demo" className="btn-secondary">
              Lihat Demo
            </a>
          </div>

          <p className="mt-5 text-[13px] text-muted">
            Tanpa kartu kredit · Siap dalam 5 menit
          </p>
        </div>

        <div className="relative">
          <div className="absolute -inset-6 -z-10 rounded-xl bg-surface md:-inset-10" />
          <div className="card mx-auto max-w-sm space-y-5">
            <div className="flex items-center justify-between">
              <p className="text-[13px] font-semibold text-muted">Ringkasan Pendapatan</p>
              <span className="chip">Bulan ini</span>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="rounded-md bg-surface p-4">
                <p className="text-[11px] font-medium uppercase tracking-wide text-muted">MRR</p>
                <p className="mt-1 font-display text-[22px] font-bold text-primary">
                  Rp48,2jt
                </p>
                <p className="mt-1 text-[12px] font-semibold text-[#3f7d1c]">
                  +12,4%
                </p>
              </div>
              <div className="rounded-md bg-surface p-4">
                <p className="text-[11px] font-medium uppercase tracking-wide text-muted">ARR</p>
                <p className="mt-1 font-display text-[22px] font-bold text-primary">
                  Rp578jt
                </p>
                <p className="mt-1 text-[12px] font-semibold text-[#3f7d1c]">
                  +8,1%
                </p>
              </div>
            </div>

            <div className="space-y-2 rounded-md border border-border p-4">
              <div className="flex items-center justify-between text-[13px]">
                <span className="text-secondary">Langganan Aktif</span>
                <span className="font-semibold text-primary">1.284</span>
              </div>
              <div className="h-2 w-full rounded-full bg-surface">
                <div className="h-2 w-[78%] rounded-full bg-tertiary" />
              </div>
              <div className="flex items-center justify-between text-[13px]">
                <span className="text-secondary">Invoice Jatuh Tempo</span>
                <span className="font-semibold text-primary">32</span>
              </div>
              <div className="h-2 w-full rounded-full bg-surface">
                <div className="h-2 w-[24%] rounded-full bg-accent" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
