const features = [
  {
    title: 'Manajemen Langganan',
    description:
      'Kelola siklus hidup langganan pelanggan secara penuh — aktif, trial, pause, hingga cancel — dengan riwayat lengkap.',
  },
  {
    title: 'Penagihan Otomatis',
    description:
      'Invoice dibuat otomatis sesuai siklus bulanan, 6 bulanan, atau tahunan lewat penjadwalan cronjob, tanpa intervensi manual.',
  },
  {
    title: 'Payment Gateway Xendit',
    description:
      'Auto-debit kartu kredit dengan tokenisasi, sehingga penagihan berulang berjalan tanpa pelanggan input ulang kartu.',
  },
  {
    title: 'Dashboard Pendapatan',
    description:
      'Pantau MRR, ARR, pelanggan aktif, dan tren pertumbuhan secara real-time dalam satu dashboard analitik.',
  },
  {
    title: 'Multi-Tenant & Multi-Role',
    description:
      'Satu platform untuk banyak tenant, dengan peran Super Admin, Admin, Sales, dan Finance yang terpisah rapi.',
  },
  {
    title: 'Audit Log & Monitoring',
    description:
      'Setiap perubahan langganan, invoice, dan pembayaran tercatat untuk keamanan dan penelusuran riwayat.',
  },
]

export default function Features() {
  return (
    <section id="fitur" className="bg-surface/60 py-20 md:py-28">
      <div className="mx-auto max-w-6xl px-5">
        <div className="max-w-xl">
          <span className="chip">Fitur</span>
          <h2 className="mt-4 font-display text-[32px] font-bold leading-tight tracking-[-0.02em] text-primary md:text-[36px]">
            Semua yang dibutuhkan bisnis SaaS Anda, dalam satu tempat.
          </h2>
          <p className="mt-4 text-[16px] leading-relaxed text-secondary">
            Dari pendaftaran pelanggan hingga analisis pendapatan, PayLanggan
            menghubungkan setiap tahap siklus bisnis langganan.
          </p>
        </div>

        <div className="mt-12 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature) => (
            <div key={feature.title} className="card">
              <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-md bg-primary/[0.06]">
                <span className="h-2.5 w-2.5 rounded-full bg-tertiary" />
              </div>
              <h3 className="font-display text-[18px] font-semibold text-primary">
                {feature.title}
              </h3>
              <p className="mt-2 text-[14px] leading-relaxed text-secondary">
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
