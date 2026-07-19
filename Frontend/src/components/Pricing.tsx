const tiers = [
  {
    name: 'Starter',
    price: 'Rp299rb',
    period: '/bulan',
    description: 'Untuk tenant tunggal yang baru mulai mengelola langganan.',
    features: [
      'Hingga 100 pelanggan aktif',
      'Manajemen paket & langganan',
      '1 pengguna Admin',
      'Invoice otomatis bulanan',
    ],
    highlighted: false,
  },
  {
    name: 'Growth',
    price: 'Rp799rb',
    period: '/bulan',
    description: 'Untuk tim yang butuh penagihan otomatis penuh dan analitik.',
    features: [
      'Hingga 2.500 pelanggan aktif',
      'Penagihan berulang via cronjob',
      'Payment gateway Xendit',
      'Dashboard MRR & ARR',
      'Role Sales & Finance',
    ],
    highlighted: true,
  },
  {
    name: 'Enterprise',
    price: 'Kustom',
    period: '',
    description: 'Untuk banyak tenant dengan kebutuhan audit dan skala besar.',
    features: [
      'Pelanggan aktif tanpa batas',
      'Multi-tenant tanpa batas',
      'Audit log & monitoring penuh',
      'Dukungan prioritas',
    ],
    highlighted: false,
  },
]

export default function Pricing() {
  return (
    <section id="harga" className="py-20 md:py-28">
      <div className="mx-auto max-w-6xl px-5">
        <div className="mx-auto max-w-xl text-center">
          <span className="chip">Harga</span>
          <h2 className="mt-4 font-display text-[32px] font-bold leading-tight tracking-[-0.02em] text-primary md:text-[36px]">
            Paket yang tumbuh bersama bisnis Anda.
          </h2>
          <p className="mt-4 text-[16px] leading-relaxed text-secondary">
            Mulai gratis 14 hari, upgrade kapan saja tanpa kehilangan data langganan.
          </p>
        </div>

        <div className="mt-12 grid gap-6 md:grid-cols-3">
          {tiers.map((tier) => (
            <div
              key={tier.name}
              className={
                tier.highlighted
                  ? 'card relative border-2 border-tertiary md:-translate-y-3'
                  : 'card border border-border'
              }
            >
              {tier.highlighted && (
                <span className="chip absolute -top-3 left-1/2 -translate-x-1/2 bg-tertiary text-neutral">
                  Paling Populer
                </span>
              )}

              <h3 className="font-display text-[18px] font-semibold text-primary">
                {tier.name}
              </h3>
              <p className="mt-2 text-[13px] text-secondary">{tier.description}</p>

              <p className="mt-6 flex items-baseline gap-1">
                <span className="font-display text-[32px] font-bold text-primary">
                  {tier.price}
                </span>
                {tier.period && (
                  <span className="text-[13px] text-muted">{tier.period}</span>
                )}
              </p>

              <a
                href="#signup"
                className={tier.highlighted ? 'btn-primary mt-6 w-full' : 'btn-secondary mt-6 w-full'}
              >
                {tier.name === 'Enterprise' ? 'Hubungi Kami' : 'Mulai Uji Coba'}
              </a>

              <ul className="mt-6 space-y-3">
                {tier.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-2 text-[13px] text-secondary">
                    <span className="mt-[3px] h-[6px] w-[6px] flex-shrink-0 rounded-full bg-tertiary" />
                    {feature}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
