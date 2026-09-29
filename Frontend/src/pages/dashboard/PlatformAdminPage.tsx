import { useEffect, useState } from 'react'
import { AlertCircle, Building2, CreditCard, Download, Wallet, XCircle } from 'lucide-react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import StatCard from '../../components/dashboard/StatCard'
import GlassPanel from '../../components/ui/GlassPanel'
import { useAuth } from '../../context/AuthContext'
import { ApiError, getPlatformDashboardSummary, type PlatformDashboardSummary } from '../../lib/api'
import { formatCurrency } from '../../lib/format'
import { exportPlatformReportPdf } from '../../lib/reportPdf'

export default function PlatformAdminPage() {
  const { auth } = useAuth()
  const [summary, setSummary] = useState<PlatformDashboardSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [exporting, setExporting] = useState(false)

  useEffect(() => {
    if (!auth?.token) return
    setLoading(true)
    getPlatformDashboardSummary(auth.token)
      .then(setSummary)
      .catch(err => setError(err instanceof ApiError ? err.message : 'Gagal memuat statistik PayLanggan.'))
      .finally(() => setLoading(false))
  }, [auth?.token])

  async function exportPdf() {
    if (!auth?.token) return
    setExporting(true)
    setError(null)
    try {
      const latest = await getPlatformDashboardSummary(auth.token)
      setSummary(latest)
      await exportPlatformReportPdf(latest)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal mengunduh laporan PDF.')
    } finally {
      setExporting(false)
    }
  }

  return <div className="space-y-6">
    <div className="flex flex-wrap items-start justify-between gap-4">
      <DashboardPageHeader title="Ringkasan PayLanggan" description="Pendapatan dan tagihan bisnis kepada PayLanggan." />
      <button type="button" onClick={() => void exportPdf()} disabled={loading || exporting || !summary} className="inline-flex items-center gap-2 rounded-full bg-brand px-4 py-2 text-[13px] font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50">
        <Download size={16} />{exporting ? 'Menyiapkan PDF...' : 'Ekspor PDF'}
      </button>
    </div>
    {error && <p role="alert" className="text-sm text-red-600">{error}</p>}
    <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
      <StatCard icon={Building2} label="Bisnis Aktif" caption="Langganan platform aktif" value={summary?.active_businesses.toLocaleString('id-ID')} isLoading={loading} />
      <StatCard icon={AlertCircle} label="Bisnis Menunggak" caption="Langganan platform past due" value={summary?.past_due_businesses.toLocaleString('id-ID')} isLoading={loading} />
      <StatCard icon={Wallet} label="Pendapatan Bulan Ini" caption="Tagihan platform yang lunas bulan ini" value={summary ? formatCurrency(summary.revenue_this_month) : undefined} isLoading={loading} />
      <StatCard icon={CreditCard} label="Tagihan Belum Dibayar" caption="Total invoice platform pending" value={summary ? formatCurrency(summary.outstanding_amount) : undefined} isLoading={loading} />
      <StatCard icon={XCircle} label="Debit Otomatis Gagal" caption="Kegagalan bulan ini" value={summary?.failed_autopay_this_month.toLocaleString('id-ID')} isLoading={loading} />
    </div>
    <GlassPanel><p className="text-xs text-muted">Statistik ini hanya mencakup langganan bisnis kepada PayLanggan. Pembayaran pelanggan milik tiap bisnis tersedia pada dashboard bisnis masing-masing.</p></GlassPanel>
  </div>
}
