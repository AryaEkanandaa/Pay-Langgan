import { useEffect, useState } from 'react'
import { AlertCircle, ArrowUpRight, CalendarDays, CreditCard, Download, Repeat, Ticket, UserMinus, UserPlus, Users, Wallet } from 'lucide-react'
import { Link } from 'react-router-dom'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import GlassPanel from '../../components/ui/GlassPanel'
import StatCard from '../../components/dashboard/StatCard'
import PerformanceChart from '../../components/dashboard/PerformanceChart'
import YearlyBreakdownChart from '../../components/dashboard/YearlyBreakdownChart'
import { useAuth } from '../../context/AuthContext'
import { ApiError, getDashboardSummary, type DashboardSummary } from '../../lib/api'
import { formatCurrency } from '../../lib/format'
import { exportBusinessReportPdf } from '../../lib/reportPdf'

export default function DashboardHome() {
  const { auth } = useAuth()
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [exporting, setExporting] = useState(false)

  useEffect(() => {
    if (!auth?.token) return
    setIsLoading(true)
    getDashboardSummary(auth.token)
      .then(setSummary)
      .catch(err => setError(err instanceof ApiError ? err.message : 'Gagal memuat statistik dashboard.'))
      .finally(() => setIsLoading(false))
  }, [auth?.token])

  const isSales = auth?.user.role === 'sales'
  const isFinance = auth?.user.role === 'finance'
  const today = new Date().toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })

  async function exportPdf() {
    if (!auth?.token) return
    setExporting(true)
    setError(null)
    try {
      const latest = await getDashboardSummary(auth.token)
      setSummary(latest)
      await exportBusinessReportPdf(latest, auth.user.business_id, isSales)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal mengunduh laporan PDF.')
    } finally {
      setExporting(false)
    }
  }

  return <div>
    <div className="mb-8 flex flex-wrap items-start justify-between gap-4">
      <DashboardPageHeader title={`Halo, ${auth?.user.name.split(' ')[0]}`} description={isSales
        ? 'Pantau pelanggan dan langganan yang perlu ditindaklanjuti.'
        : 'Pantau langganan, tagihan, dan pembayaran bisnis Anda.'} />
      <div className="flex flex-wrap items-center gap-3">
        <span className="glass flex items-center gap-2 rounded-full px-4 py-2 text-[13px] font-medium text-ink"><CalendarDays size={16} />{today}</span>
        <button type="button" onClick={() => void exportPdf()} disabled={isLoading || exporting || !summary} className="inline-flex items-center gap-2 rounded-full bg-brand px-4 py-2 text-[13px] font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50">
          <Download size={16} />{exporting ? 'Menyiapkan PDF...' : 'Ekspor PDF'}
        </button>
      </div>
    </div>
    {error && <p role="alert" className="mb-4 text-sm text-red-600">{error}</p>}

    <p className="mb-4 text-[13px] font-semibold uppercase tracking-wide text-muted">Ringkasan</p>
    <div className={`grid gap-5 sm:grid-cols-2 ${isSales ? 'lg:grid-cols-3' : 'lg:grid-cols-5'}`}>
      <StatCard icon={Users} label="Pelanggan Aktif" caption="Memiliki langganan trial atau aktif" value={summary?.active_customers.toLocaleString('id-ID')} isLoading={isLoading} />
      <StatCard icon={Repeat} label="Langganan Berjalan" caption="Trial dan langganan aktif" value={summary?.active_subscriptions.toLocaleString('id-ID')} isLoading={isLoading} />
      {!isSales && <>
        <StatCard icon={CreditCard} label="MRR" caption="Estimasi bulanan dari langganan aktif" value={summary ? formatCurrency(summary.mrr) : undefined} isLoading={isLoading} />
        <StatCard icon={Ticket} label="Invoice Jatuh Tempo" caption="Invoice pending yang lewat jatuh tempo" value={summary?.due_invoices.toLocaleString('id-ID')} isLoading={isLoading} />
      </>}
      <Link to={isSales ? '/dashboard/pelanggan' : isFinance ? '/dashboard/invoices' : '/dashboard/katalog'} className="flex flex-col justify-between rounded-2xl bg-gradient-to-br from-brand to-brand-dark p-6 text-white shadow-lg">
        <p className="font-display text-[17px] font-bold">{isSales ? 'Kelola Pelanggan' : isFinance ? 'Tinjau Invoice' : 'Kelola Katalog'}</p>
        <span className="mt-4 flex h-9 w-9 items-center justify-center rounded-full bg-white/15"><ArrowUpRight size={18} /></span>
      </Link>
    </div>

    <div className={`mt-5 grid gap-5 sm:grid-cols-2 ${isSales ? 'lg:grid-cols-2' : 'lg:grid-cols-4'}`}>
      <StatCard icon={UserPlus} label="Langganan Baru" caption="Dibuat bulan ini" value={summary?.new_subscriptions.toLocaleString('id-ID')} isLoading={isLoading} />
      <StatCard icon={UserMinus} label="Langganan Berhenti" caption="Dibatalkan bulan ini" value={summary?.cancelled_subscriptions.toLocaleString('id-ID')} isLoading={isLoading} />
      {!isSales && <>
        <StatCard icon={Wallet} label="Pendapatan Diterima" caption="Pembayaran sukses bulan ini" value={summary ? formatCurrency(summary.revenue_this_month) : undefined} isLoading={isLoading} />
        <StatCard icon={AlertCircle} label="Tagihan Belum Dibayar" caption={summary ? `Lewat jatuh tempo: ${formatCurrency(summary.overdue_amount)}` : 'Invoice pending'} value={summary ? formatCurrency(summary.outstanding_amount) : undefined} isLoading={isLoading} />
      </>}
    </div>

    {isSales ? <GlassPanel className="mt-6">
      <p className="text-sm font-semibold text-ink">Akses Cepat Sales</p>
      <div className="mt-4 flex gap-4 text-sm"><Link className="text-brand hover:underline" to="/dashboard/pelanggan">Pelanggan</Link><Link className="text-brand hover:underline" to="/dashboard/langganan">Langganan</Link></div>
    </GlassPanel> : <>
      <div className="mt-6 grid gap-5 lg:grid-cols-3">
        <GlassPanel className="lg:col-span-2">
          <p className="text-sm font-semibold text-ink">Pendapatan Diterima per Bulan</p>
          <p className="mt-1 text-xs text-muted">Pembayaran sukses, 7 bulan terakhir</p>
          <div className="mt-4"><PerformanceChart data={summary?.monthly_performance ?? []} /></div>
        </GlassPanel>
        <GlassPanel>
          <p className="text-sm font-semibold text-ink">Pembayaran Bulan Ini</p>
          <div className="mt-5 grid grid-cols-2 gap-4">
            <div><p className="text-xs text-muted">Berhasil</p><p className="text-2xl font-bold text-ink">{summary?.payments_succeeded ?? '—'}</p></div>
            <div><p className="text-xs text-muted">Gagal</p><p className="text-2xl font-bold text-ink">{summary?.payments_failed ?? '—'}</p></div>
            <div><p className="text-xs text-muted">Debit otomatis berhasil</p><p className="text-2xl font-bold text-ink">{summary?.autopay_succeeded ?? '—'}</p></div>
            <div><p className="text-xs text-muted">Debit otomatis gagal</p><p className="text-2xl font-bold text-ink">{summary?.autopay_failed ?? '—'}</p></div>
          </div>
        </GlassPanel>
      </div>
      <div className="mt-6 grid gap-5 lg:grid-cols-2">
        <GlassPanel><p className="text-sm font-semibold text-ink">Pendapatan Diterima per Tahun</p><div className="mt-4"><YearlyBreakdownChart data={summary?.yearly_breakdown ?? []} /></div></GlassPanel>
        <GlassPanel>
          <p className="text-sm font-semibold text-ink">Pembayaran Menurut Paket Bulan Ini</p>
          <p className="mt-1 text-xs text-muted">Termasuk add-on dan diskon pada invoice paket</p>
          {!summary?.revenue_by_plan.length ? <p className="mt-5 text-sm text-muted">Belum ada pembayaran paket pada bulan ini.</p> :
            <div className="mt-4 space-y-3">{summary.revenue_by_plan.map(plan => <div key={plan.plan_id} className="flex justify-between gap-3 border-b border-border pb-3 text-sm last:border-0"><span className="font-medium text-ink">{plan.name}</span><span className="font-semibold text-ink">{formatCurrency(plan.amount)}</span></div>)}</div>}
        </GlassPanel>
      </div>
    </>}
    <GlassPanel className="mt-6"><p className="text-xs text-muted">Pendapatan diterima dihitung dari pembayaran sukses. MRR adalah estimasi bulanan dari langganan aktif, bukan uang yang sudah diterima.</p></GlassPanel>
  </div>
}
