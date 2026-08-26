import { ArrowUpRight, CalendarDays, CreditCard, Repeat, Ticket, Users } from 'lucide-react'
import { Link } from 'react-router-dom'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import GlassPanel from '../../components/ui/GlassPanel'
import Skeleton from '../../components/ui/Skeleton'
import StatCard from '../../components/dashboard/StatCard'
import PerformanceChart from '../../components/dashboard/PerformanceChart'
import YearlyBreakdownChart from '../../components/dashboard/YearlyBreakdownChart'
import { useAuth } from '../../context/AuthContext'
import {
  monthlyPerformanceData,
  yearlyBreakdownData,
  yearlyIncomeSharePct,
  yearlySpendingSharePct,
} from '../../data/dashboardMock'

const stats = [
  { icon: Users, label: 'Pelanggan Aktif', caption: 'Total pelanggan aktif' },
  { icon: Repeat, label: 'Langganan Aktif', caption: 'Langganan berjalan' },
  { icon: CreditCard, label: 'MRR Bulan Ini', caption: 'Pendapatan berulang' },
  { icon: Ticket, label: 'Invoice Jatuh Tempo', caption: 'Perlu ditagih' },
]

export default function DashboardHome() {
  const { auth } = useAuth()

  const today = new Date().toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })

  return (
    <div>
      <div className="mb-8 flex flex-wrap items-start justify-between gap-4">
        <DashboardPageHeader
          title={`Halo, ${auth?.user.name.split(' ')[0]}`}
          description="Pantau aktivitas langganan dan pelanggan bisnis Anda di sini."
        />
        <span className="glass flex items-center gap-2 rounded-full px-4 py-2 text-[13px] font-medium text-ink">
          <CalendarDays size={16} strokeWidth={2} />
          {today}
        </span>
      </div>

      <p className="mb-4 text-[13px] font-semibold uppercase tracking-wide text-muted">Ringkasan</p>

      <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-5">
        {stats.map((stat) => (
          <StatCard key={stat.label} icon={stat.icon} label={stat.label} caption={stat.caption} isLoading />
        ))}

        <Link
          to="/dashboard/katalog"
          className="flex flex-col justify-between rounded-2xl bg-gradient-to-br from-brand to-brand-dark p-6 text-white shadow-[0_8px_32px_-8px_rgba(54,84,255,0.5)] transition-transform hover:scale-[1.02]"
        >
          <p className="font-display text-[17px] font-bold leading-tight">Kelola Katalog Anda</p>
          <span className="mt-4 flex h-9 w-9 items-center justify-center rounded-full bg-white/15">
            <ArrowUpRight size={18} strokeWidth={2} />
          </span>
        </Link>
      </div>

      <div className="mt-6 grid gap-5 lg:grid-cols-3">
        <GlassPanel className="lg:col-span-2">
          <div className="flex items-center justify-between">
            <p className="text-[14px] font-semibold text-ink">Performa Bulanan</p>
            <span className="glass rounded-full px-3 py-1 text-[12px] font-medium text-body">Bulan Ini</span>
          </div>
          <div className="mt-4">
            <PerformanceChart data={monthlyPerformanceData} />
          </div>
        </GlassPanel>

        <GlassPanel>
          <p className="text-[14px] font-semibold text-ink">Pendapatan vs Pengeluaran</p>
          <div className="mt-4 flex items-center gap-5">
            <div>
              <p className="font-display text-[24px] font-bold text-ink">{yearlyIncomeSharePct}%</p>
              <p className="flex items-center gap-1.5 text-[12px] text-muted">
                <span className="h-2 w-2 rounded-full bg-brand" /> Pendapatan
              </p>
            </div>
            <div>
              <p className="font-display text-[24px] font-bold text-ink">{yearlySpendingSharePct}%</p>
              <p className="flex items-center gap-1.5 text-[12px] text-muted">
                <span className="h-2 w-2 rounded-full bg-amber" /> Pengeluaran
              </p>
            </div>
          </div>
          <div className="mt-2">
            <YearlyBreakdownChart data={yearlyBreakdownData} />
          </div>
        </GlassPanel>
      </div>

      <GlassPanel className="mt-6">
        <p className="text-[13px] font-semibold text-ink">Aktivitas Terbaru</p>

        <div className="mt-4 space-y-4">
          {[1, 2, 3].map((row) => (
            <div key={row} className="flex items-center gap-3">
              <Skeleton className="h-9 w-9 rounded-full" />
              <div className="flex-1 space-y-2">
                <Skeleton className="h-3 w-1/3" />
                <Skeleton className="h-3 w-1/5" />
              </div>
            </div>
          ))}
        </div>

        <p className="mt-4 text-[12px] text-muted">
          Data akan tampil di sini setelah dashboard terhubung ke API pelanggan &amp;
          langganan.
        </p>
      </GlassPanel>
    </div>
  )
}
