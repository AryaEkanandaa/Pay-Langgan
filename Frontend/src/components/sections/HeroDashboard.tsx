import { ArrowUpRight, Check, ChevronDown, FileText, Layers, LockKeyhole, Users } from 'lucide-react'

const customers = [
  { initials: 'AS', name: 'Alya Studio', plan: 'Pro', date: '12 Okt', status: 'Aktif', color: 'bg-blue-50 text-blue-600 dark:bg-blue-500/15 dark:text-blue-300' },
  { initials: 'KT', name: 'Karya Tim', plan: 'Business', date: '15 Okt', status: 'Aktif', color: 'bg-violet-50 text-violet-600 dark:bg-violet-500/15 dark:text-violet-300' },
  { initials: 'NR', name: 'Nara Digital', plan: 'Basic', date: '18 Okt', status: 'Trial', color: 'bg-orange-50 text-orange-600 dark:bg-orange-500/15 dark:text-orange-300' },
]

export default function HeroDashboard() {
  return (
    <figure className="relative mx-auto min-w-0 w-full max-w-[510px] pb-4">
      <div className="overflow-hidden rounded-[20px] border border-slate-200/80 bg-white shadow-[0_24px_70px_-24px_rgba(39,54,118,0.3)] dark:border-slate-700 dark:bg-[#171e30]">
        <div className="flex items-center gap-4 border-b border-border bg-slate-50/80 px-4 py-3 dark:bg-slate-800/50">
          <div className="flex gap-1.5" aria-hidden="true">
            <span className="size-2 rounded-full bg-rose-300" />
            <span className="size-2 rounded-full bg-amber/60" />
            <span className="size-2 rounded-full bg-emerald-300" />
          </div>
          <div className="flex flex-1 items-center justify-center gap-1.5 text-[10px] text-muted">
            <LockKeyhole size={10} aria-hidden="true" />
            <span>paylanggan / dashboard</span>
          </div>
          <Layers size={13} className="text-muted" aria-hidden="true" />
        </div>

        <div className="p-4 sm:p-6">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-[10px] font-medium uppercase tracking-[0.14em] text-muted">Workspace Anda</p>
              <p className="mt-1 font-display text-lg font-bold tracking-tight text-ink sm:text-xl">Ringkasan langganan</p>
            </div>
            <span className="flex shrink-0 items-center gap-1 rounded-lg border border-border px-2 py-1.5 text-[10px] text-body">
              Oktober <ChevronDown size={11} aria-hidden="true" />
            </span>
          </div>

          <div className="mt-5 grid grid-cols-2 gap-3">
            <div className="rounded-xl border border-brand/10 bg-brand/[0.045] p-3 sm:p-4 dark:bg-brand/10">
              <div className="flex items-center justify-between gap-1 text-[11px] text-body">
                <span>Langganan aktif</span><Users size={14} className="text-brand dark:text-indigo-300" aria-hidden="true" />
              </div>
              <p className="mt-2 font-display text-[30px] font-bold leading-none tracking-tight text-ink">128</p>
              <p className="mt-2 flex items-center gap-1 text-[10px] text-emerald-700 dark:text-emerald-300"><ArrowUpRight size={12} aria-hidden="true" />12 baru bulan ini</p>
            </div>
            <div className="rounded-xl border border-border p-3 sm:p-4">
              <div className="flex items-center justify-between gap-1 text-[11px] text-body">
                <span>Invoice bulan ini</span><FileText size={14} className="text-muted" aria-hidden="true" />
              </div>
              <p className="mt-2 font-display text-[30px] font-bold leading-none tracking-tight text-ink">36</p>
              <p className="mt-2 text-[10px] text-muted">Tercatat dalam satu tempat</p>
            </div>
          </div>

          <div className="mt-6 flex items-center justify-between">
            <p className="text-xs font-semibold text-ink">Langganan pelanggan</p>
            <span className="rounded-md bg-slate-100 px-2 py-1 text-[9px] text-body dark:bg-slate-800">3 ditampilkan</span>
          </div>
          <table className="mt-3 w-full table-fixed text-left text-[11px]">
            <caption className="sr-only">Contoh pelanggan dan tanggal tagihan berikutnya</caption>
            <thead className="border-b border-border text-[9px] font-medium uppercase tracking-wider text-muted">
              <tr>
                <th scope="col" className="w-[49%] pb-2 font-medium">Pelanggan</th>
                <th scope="col" className="w-[27%] pb-2 font-medium">Tagihan</th>
                <th scope="col" className="pb-2 text-right font-medium">Status</th>
              </tr>
            </thead>
            <tbody>
              {customers.map((customer) => (
                <tr key={customer.name} className="border-b border-border last:border-0">
                  <td className="py-3">
                    <div className="flex items-center gap-2">
                      <span className={`flex size-7 shrink-0 items-center justify-center rounded-lg text-[9px] font-semibold ${customer.color}`} aria-hidden="true">{customer.initials}</span>
                      <div className="min-w-0">
                        <p className="truncate font-semibold text-ink">{customer.name}</p>
                        <p className="mt-0.5 text-[9px] text-muted">{customer.plan} · Bulanan</p>
                      </div>
                    </div>
                  </td>
                  <td className="text-body">{customer.date}</td>
                  <td className="text-right">
                    <span className={`inline-flex items-center gap-1 rounded-full px-2 py-1 text-[9px] font-medium ${customer.status === 'Aktif' ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300' : 'bg-amber/10 text-yellow-800 dark:text-amber'}`}>
                      <span className="size-1 rounded-full bg-current" aria-hidden="true" />{customer.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="relative ml-5 mr-3 -mt-2 flex items-center gap-3 rounded-xl border border-slate-200 bg-white p-3 shadow-[0_12px_30px_-12px_rgba(39,54,118,0.25)] sm:ml-12 sm:mr-[-12px] sm:p-4 dark:border-slate-700 dark:bg-[#202a40]">
        <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-emerald-50 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-300"><Check size={18} aria-hidden="true" /></span>
        <div>
          <p className="text-[11px] font-semibold text-ink sm:text-xs">Invoice periode berikutnya berhasil dibuat</p>
          <p className="mt-1 text-[10px] text-muted">Alya Studio · Paket Pro · Bulanan</p>
        </div>
      </div>
      <figcaption className="mt-5 text-center text-[10px] tracking-wide text-muted">Ilustrasi dashboard · Menggunakan data contoh</figcaption>
    </figure>
  )
}
