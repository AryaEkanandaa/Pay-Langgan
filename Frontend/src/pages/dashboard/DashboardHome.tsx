import { useEffect, useState } from 'react'
import { ArrowUpRight, CalendarDays, CreditCard, Repeat, Ticket, Users } from 'lucide-react'
import { Link } from 'react-router-dom'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import GlassPanel from '../../components/ui/GlassPanel'
import StatCard from '../../components/dashboard/StatCard'
import PerformanceChart from '../../components/dashboard/PerformanceChart'
import YearlyBreakdownChart from '../../components/dashboard/YearlyBreakdownChart'
import { useAuth } from '../../context/AuthContext'
import { ApiError, getDashboardSummary, type DashboardSummary } from '../../lib/api'
import { formatCurrency } from '../../lib/format'

export default function DashboardHome() {
  const { auth } = useAuth()
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!auth?.token) return
    setIsLoading(true)
    setError(null)
    getDashboardSummary(auth.token)
      .then(setSummary)
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : 'Gagal memuat statistik dashboard.')
      })
      .finally(() => setIsLoading(false))
  }, [auth?.token])

  const today = new Date().toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
  const isSales = auth?.user.role === 'sales'
  const isFinance = auth?.user.role === 'finance'
  const pageDescription = isSales
    ? 'Pantau pelanggan dan langganan yang perlu ditindaklanjuti oleh tim Sales.'
    : isFinance
      ? 'Pantau pendapatan, invoice, dan kondisi keuangan bisnis Anda.'
      : 'Pantau aktivitas langganan, pelanggan, dan keuangan bisnis Anda di sini.'

  return (
    <div>
      <div className="mb-8 flex flex-wrap items-start justify-between gap-4">
        <DashboardPageHeader
          title={`Halo, ${auth?.user.name.split(' ')[0]}`}
          description={pageDescription}
        />
        <span className="glass flex items-center gap-2 rounded-full px-4 py-2 text-[13px] font-medium text-ink">
          <CalendarDays size={16} strokeWidth={2} />
          {today}
        </span>
      </div>

      <p className="mb-4 text-[13px] font-semibold uppercase tracking-wide text-muted">Ringkasan</p>

      {error && <p className="mb-4 text-[13px] text-red-600">{error}</p>}

      <div className={`grid gap-5 sm:grid-cols-2 ${isSales ? 'lg:grid-cols-3' : 'lg:grid-cols-5'}`}>
        <StatCard
          icon={Users}
          label="Pelanggan Aktif"
          caption="Pelanggan dengan langganan berjalan"
          value={summary?.active_customers.toLocaleString('id-ID')}
          isLoading={isLoading}
        />
        <StatCard
          icon={Repeat}
          label="Langganan Aktif"
          caption="Trial dan langganan berjalan"
          value={summary?.active_subscriptions.toLocaleString('id-ID')}
          isLoading={isLoading}
        />
        {!isSales && (
          <>
            <StatCard
              icon={CreditCard}
              label="MRR Bulan Ini"
              caption="Estimasi pendapatan berulang"
              value={summary ? formatCurrency(summary.mrr) : undefined}
              isLoading={isLoading}
            />
            <StatCard
              icon={Ticket}
              label="Invoice Jatuh Tempo"
              caption="Invoice pending yang sudah jatuh tempo"
              value={summary?.due_invoices.toLocaleString('id-ID')}
              isLoading={isLoading}
            />
          </>
        )}

        <Link
          to={isSales ? '/dashboard/pelanggan' : isFinance ? '/dashboard/invoices' : '/dashboard/katalog'}
          className="flex flex-col justify-between rounded-2xl bg-gradient-to-br from-brand to-brand-dark p-6 text-white shadow-[0_8px_32px_-8px_rgba(54,84,255,0.5)] transition-transform hover:scale-[1.02]"
        >
          <p className="font-display text-[17px] font-bold leading-tight">
            {isSales ? 'Kelola Pelanggan Anda' : isFinance ? 'Tinjau Invoice Anda' : 'Kelola Katalog Anda'}
          </p>
          <span className="mt-4 flex h-9 w-9 items-center justify-center rounded-full bg-white/15">
            <ArrowUpRight size={18} strokeWidth={2} />
          </span>
        </Link>
      </div>

      {isSales ? (
        <GlassPanel className="mt-6">
          <p className="text-[14px] font-semibold text-ink">Akses Cepat Sales</p>
          <p className="mt-1 text-[13px] text-body">Lanjutkan pekerjaan dari area yang paling sering digunakan.</p>
          <div className="mt-5 grid gap-4 md:grid-cols-2">
            <Link to="/dashboard/pelanggan" className="rounded-xl border border-white/50 bg-white/35 p-4 transition-colors hover:bg-white/60 dark:hover:bg-white/10">
              <p className="text-[13px] font-semibold text-ink">Pelanggan</p>
              <p className="mt-1 text-[12px] text-muted">Kelola data dan hubungan dengan pelanggan.</p>
            </Link>
            <Link to="/dashboard/langganan" className="rounded-xl border border-white/50 bg-white/35 p-4 transition-colors hover:bg-white/60 dark:hover:bg-white/10">
              <p className="text-[13px] font-semibold text-ink">Langganan</p>
              <p className="mt-1 text-[12px] text-muted">Pantau status dan tindak lanjuti langganan.</p>
            </Link>
          </div>
        </GlassPanel>
      ) : (
      <div className="mt-6 grid gap-5 lg:grid-cols-3">
        <GlassPanel className="lg:col-span-2">
          <div className="flex items-center justify-between">
            <p className="text-[14px] font-semibold text-ink">Performa Bulanan</p>
            <span className="glass rounded-full px-3 py-1 text-[12px] font-medium text-body">Bulan Ini</span>
          </div>
          <div className="mt-4">
            <PerformanceChart data={summary?.monthly_performance ?? []} />
          </div>
        </GlassPanel>

        <GlassPanel>
          <p className="text-[14px] font-semibold text-ink">Pendapatan vs Pengeluaran</p>
          <div className="mt-4 flex items-center gap-5">
            <div>
              <p className="font-display text-[24px] font-bold text-ink">
                {summary ? `${summary.income_share_pct}%` : '—'}
              </p>
              <p className="flex items-center gap-1.5 text-[12px] text-muted">
                <span className="h-2 w-2 rounded-full bg-brand" /> Pendapatan
              </p>
            </div>
            <div>
              <p className="font-display text-[24px] font-bold text-ink">
                {summary ? `${summary.spending_share_pct}%` : '—'}
              </p>
              <p className="flex items-center gap-1.5 text-[12px] text-muted">
                <span className="h-2 w-2 rounded-full bg-amber" /> Pengeluaran
              </p>
            </div>
          </div>
          <div className="mt-2">
            <YearlyBreakdownChart data={summary?.yearly_breakdown ?? []} />
          </div>
        </GlassPanel>
      </div>
      )}

      <GlassPanel className="mt-6">
        <p className="text-[13px] font-semibold text-ink">Aktivitas Terbaru</p>

          <div className="mt-4 rounded-xl bg-white/35 p-4 text-[13px] text-muted">
            {isSales
              ? 'Ringkasan ini menampilkan metrik operasional untuk membantu tim Sales memprioritaskan pelanggan dan langganan.'
              : 'Statistik pendapatan dihitung dari payment sukses. Riwayat aktivitas detail akan tersedia setelah modul audit dan billing dikembangkan.'}
        </div>
      </GlassPanel>
    </div>
  )
}
