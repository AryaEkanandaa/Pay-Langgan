import { useEffect, useMemo, useRef, useState } from 'react'
import { CalendarDays, Download } from 'lucide-react'
import { Link, useSearchParams } from 'react-router-dom'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import InfoPopover from '../../components/dashboard/InfoPopover'
import SummaryChart from '../../components/dashboard/SummaryChart'
import SummaryMetricCard from '../../components/dashboard/SummaryMetricCard'
import SummaryExportDialog from '../../components/dashboard/SummaryExportDialog'
import Skeleton from '../../components/ui/Skeleton'
import { useAuth } from '../../context/AuthContext'
import { getDashboardSummary, getMyBusiness, type DashboardSummary, type DashboardRangePoint } from '../../lib/api'
import { formatCurrency } from '../../lib/format'
import { businessToday, completeSeries, dashboardRanges, periodFor, periodLabel, rangeGranularity, type DashboardRange } from '../../lib/dashboardPeriod'
import type { StatInfoKey } from '../../data/statInfo'

export default function DashboardHome() {
  const { auth } = useAuth(), [params, setParams] = useSearchParams()
  const rawRange = params.get('rentang'), range: DashboardRange = dashboardRanges.some(item => item.id === rawRange) ? rawRange as DashboardRange : 'harian'
  const [snapshot, setSnapshot] = useState<DashboardSummary | null>(null), [filtered, setFiltered] = useState<DashboardSummary | null>(null)
  const [timezone, setTimezone] = useState<string | null>(null), [baseLoading, setBaseLoading] = useState(true), [rangeLoading, setRangeLoading] = useState(true)
  const [baseError, setBaseError] = useState<string | null>(null), [rangeError, setRangeError] = useState<string | null>(null)
  const [retryBase, setRetryBase] = useState(0), [retryRange, setRetryRange] = useState(0), [exportOpen, setExportOpen] = useState(false)
  const tabs = useRef<Array<HTMLButtonElement | null>>([])
  const [clock, setClock] = useState(() => new Date())
  const zone = timezone ?? 'Asia/Makassar', today = businessToday(zone, clock), period = useMemo(() => periodFor(range, today), [range, today])
  const todayLabel = new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short', year: 'numeric', timeZone: zone }).format(clock)
  useEffect(() => {
    const refreshDate = () => setClock(new Date())
    const onVisible = () => { if (document.visibilityState === 'visible') refreshDate() }
    const interval = window.setInterval(refreshDate, 30_000)
    window.addEventListener('focus', refreshDate)
    document.addEventListener('visibilitychange', onVisible)
    return () => {
      window.clearInterval(interval)
      window.removeEventListener('focus', refreshDate)
      document.removeEventListener('visibilitychange', onVisible)
    }
  }, [])
  const isSales = auth?.user.role === 'sales', isFinance = auth?.user.role === 'finance'
  useEffect(() => {
    if (rawRange && !dashboardRanges.some(item => item.id === rawRange)) {
      const next = new URLSearchParams(params); next.set('rentang', 'harian'); setParams(next, { replace: true })
    }
  }, [rawRange, params, setParams])
  useEffect(() => {
    if (!auth?.token) return
    let active = true
    setBaseLoading(true); setBaseError(null)
    Promise.all([getDashboardSummary(auth.token), getMyBusiness(auth.token)]).then(([data, business]) => {
      if (!active) return
      const stored = business.meta?.timezone
      let nextZone = typeof stored === 'string' && stored ? stored : 'Asia/Makassar'
      try { new Intl.DateTimeFormat('id-ID', { timeZone: nextZone }).format(new Date()) }
      catch { nextZone = 'Asia/Makassar' }
      setSnapshot(data); setTimezone(nextZone)
    }).catch(() => { if (active) setBaseError('Gagal memuat statistik dashboard. Coba lagi.') }).finally(() => { if (active) setBaseLoading(false) })
    return () => { active = false }
  }, [auth?.token, retryBase])
  useEffect(() => {
    if (!auth?.token || !timezone) return
    let active = true
    setRangeLoading(true); setRangeError(null); setFiltered(null)
    getDashboardSummary(auth.token, { ...period, granularity: rangeGranularity[range] }).then(data => {
      if (!active) return
      if (!data.range_stats || data.range_stats.from !== period.from || data.range_stats.to !== period.to) throw new Error('Statistik rentang belum tersedia.')
      setFiltered(data)
    }).catch(() => { if (active) setRangeError('Gagal memuat statistik pada rentang ini. Coba lagi.') }).finally(() => { if (active) setRangeLoading(false) })
    // Ignore late responses when changing tabs quickly; snapshot remains independent.
    return () => { active = false }
  }, [auth?.token, timezone, period, range, retryRange])
  const series = useMemo(() => filtered?.range_stats ? completeSeries(filtered.range_stats) : [], [filtered])
  const total = (key: keyof Omit<DashboardRangePoint, 'date'>) => series.reduce((sum, point) => sum + point[key], 0)
  const points = (key: keyof Omit<DashboardRangePoint, 'date'>) => series.map(point => ({ date: point.date, value: point[key], label: new Intl.DateTimeFormat('id-ID', { day: range === '6b' || range === '12b' ? undefined : 'numeric', month: 'short', timeZone: 'UTC' }).format(new Date(`${point.date}T12:00:00Z`)) }))
  const label = periodLabel(period), error = baseError ?? rangeError
  const rangeReady = filtered?.range_stats?.from === period.from && filtered.range_stats.to === period.to && filtered.range_stats.granularity === rangeGranularity[range]
  const busy = !error && (baseLoading || rangeLoading || !rangeReady)
  const retry = () => baseError ? setRetryBase(value => value + 1) : setRetryRange(value => value + 1)
  const cardProps = { timezone: zone, caption: label, loading: busy, error, onRetry: retry }
  const kpis: Array<{ statistic: StatInfoKey; label: string; value: string; caption: string }> = [
    { statistic: 'customers', label: 'Pelanggan Aktif', value: snapshot?.active_customers.toLocaleString('id-ID') ?? '—', caption: 'Memiliki langganan trial atau aktif' },
    { statistic: 'subscriptions', label: 'Langganan Berjalan', value: snapshot?.active_subscriptions.toLocaleString('id-ID') ?? '—', caption: 'Trial dan langganan aktif' },
    ...(!isSales ? [
      { statistic: 'mrr' as const, label: 'MRR', value: snapshot ? formatCurrency(snapshot.mrr) : '—', caption: 'Estimasi bulanan dari langganan aktif' },
      { statistic: 'due' as const, label: 'Invoice Jatuh Tempo', value: snapshot?.due_invoices.toLocaleString('id-ID') ?? '—', caption: 'Invoice pending yang lewat jatuh tempo' },
      { statistic: 'outstanding' as const, label: 'Tagihan Belum Dibayar', value: snapshot ? formatCurrency(snapshot.outstanding_amount) : '—', caption: snapshot ? `Lewat jatuh tempo: ${formatCurrency(snapshot.overdue_amount)}` : 'Semua invoice pending' },
    ] : []),
  ]
  function selectRange(nextRange: DashboardRange) { const next = new URLSearchParams(params); next.set('rentang', nextRange); setParams(next) }
  // TODO: API range_stats must include previous-period aggregates before showing comparisons.
  return <div className="summary-page w-full min-w-0">
    <div className="mb-8 flex flex-wrap items-start justify-between gap-4">
      <DashboardPageHeader title={`Halo, ${auth?.user.name.split(' ')[0] ?? 'Admin'}`} description={isSales ? 'Pantau pelanggan dan langganan yang perlu ditindaklanjuti.' : 'Pantau langganan, tagihan, dan pembayaran bisnis Anda.'} />
      <div className="summary-header-actions flex flex-wrap items-center gap-3">
        <span className="summary-date-chip inline-flex min-h-11 items-center gap-2 rounded-full px-5 text-sm text-body" aria-label={`Tanggal hari ini: ${todayLabel}`}><CalendarDays size={16} aria-hidden="true" /><time dateTime={today}>{todayLabel}</time></span>
        <Link to={isSales ? '/dashboard/pelanggan' : isFinance ? '/dashboard/invoices' : '/dashboard/katalog'} className="glass inline-flex min-h-11 items-center rounded-full px-5 text-sm font-semibold focus-visible:outline-2 focus-visible:outline-brand">{isSales ? 'Kelola Pelanggan' : isFinance ? 'Tinjau Invoice' : 'Kelola Katalog'}</Link>
        <button type="button" onClick={() => setExportOpen(true)} disabled={!snapshot || !timezone} className="inline-flex min-h-11 items-center gap-2 rounded-full bg-brand px-5 text-sm font-semibold text-white disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"><Download size={16} aria-hidden="true" />Ekspor</button>
      </div>
    </div>
    {baseError && <div role="alert" className="mb-4 flex flex-wrap items-center gap-3 text-sm text-red-600">{baseError}<button type="button" onClick={retry} className="rounded-lg border px-3 py-2 font-semibold">Coba lagi</button></div>}
    <p className="mb-3 text-xs font-semibold uppercase tracking-wide text-body">Kondisi saat ini</p>
    <section aria-label="Kondisi saat ini" className={`summary-kpis grid gap-3 ${isSales ? 'summary-sales-kpis' : ''}`}>
      {kpis.map(item => <div key={item.statistic} className="summary-kpi-card min-w-0 rounded-xl p-4"><div className="flex items-center justify-between gap-2"><h2 className="text-xs font-semibold">{item.label}</h2><InfoPopover statistic={item.statistic} timezone={zone} /></div>{baseLoading ? <Skeleton className="mt-3 h-10 w-28" /> : <p className="summary-kpi-number mt-3 break-words text-[clamp(24px,1.7vw,40px)] font-bold leading-tight">{item.value}</p>}<p className="mt-3 text-xs leading-5 text-body">{item.caption}</p></div>)}
    </section>
    <div className="my-6 flex flex-wrap items-center justify-between gap-4">
      <div role="tablist" aria-label="Rentang waktu ringkasan" className="inline-flex max-w-full gap-1 rounded-full border border-border bg-white/50 p-1 dark:bg-slate-900/50">{dashboardRanges.map((item, index) => <button key={item.id} ref={node => { tabs.current[index] = node }} role="tab" type="button" id={`range-${item.id}`} aria-controls="summary-charts" aria-selected={range === item.id} tabIndex={range === item.id ? 0 : -1} onClick={() => selectRange(item.id)} onKeyDown={event => { const direction = event.key === 'ArrowRight' ? 1 : event.key === 'ArrowLeft' ? -1 : 0; if (!direction && event.key !== 'Home' && event.key !== 'End') return; event.preventDefault(); const next = event.key === 'Home' ? 0 : event.key === 'End' ? 3 : (index + direction + 4) % 4; selectRange(dashboardRanges[next].id); tabs.current[next]?.focus() }} className={`min-h-10 whitespace-nowrap rounded-full px-3 text-xs font-semibold focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand sm:px-5 ${range === item.id ? 'bg-brand text-white' : 'text-body hover:bg-brand/10'}`}>{item.label}</button>)}</div>
      <p className="text-sm text-body" aria-live="polite">{label}<span className="mt-1 block text-xs">{zone} · {range === 'harian' ? 'Per hari' : range === '3b' ? 'Per minggu' : 'Per bulan'}</span></p>
    </div>
    <div id="summary-charts" role="tabpanel" aria-labelledby={`range-${range}`} aria-busy={busy} className={`summary-charts grid items-stretch gap-6 ${isSales ? 'summary-sales-charts' : ''}`}>
      {!isSales && <SummaryMetricCard {...cardProps} statistic="revenue" total={formatCurrency(total('revenue'))} className="summary-main-chart"><SummaryChart points={points('revenue')} money legend="Pembayaran sukses" /></SummaryMetricCard>}
      <SummaryMetricCard {...cardProps} statistic="new" total={total('new_subscriptions').toLocaleString('id-ID')}><SummaryChart points={points('new_subscriptions')} legend="Langganan baru" /></SummaryMetricCard>
      <SummaryMetricCard {...cardProps} statistic="cancelled" total={total('cancelled_subscriptions').toLocaleString('id-ID')}><SummaryChart points={points('cancelled_subscriptions')} legend="Langganan dibatalkan" /></SummaryMetricCard>
      {!isSales && <>
        <SummaryMetricCard {...cardProps} statistic="payments" total={`${(total('payments_succeeded') + total('payments_failed')).toLocaleString('id-ID')} pembayaran`}><div className="grid flex-1 grid-cols-2 content-center gap-4">{([{ key: 'payments_succeeded', label: 'Berhasil' }, { key: 'payments_failed', label: 'Gagal' }, { key: 'autopay_succeeded', label: 'Debit otomatis berhasil' }, { key: 'autopay_failed', label: 'Debit otomatis gagal' }] as const).map(item => <div key={item.key} className="rounded-xl border border-border bg-white/30 p-4 dark:bg-slate-800/40"><p className="text-xs leading-5 text-body">{item.label}</p><p className="mt-3 break-words text-3xl font-bold">{total(item.key).toLocaleString('id-ID')}</p></div>)}</div>{!total('payments_succeeded') && !total('payments_failed') && <p className="mt-4 text-center text-sm text-body">Belum ada data pada rentang ini</p>}</SummaryMetricCard>
        <SummaryMetricCard {...cardProps} statistic="plans" total={formatCurrency(filtered?.range_stats?.revenue_by_plan.reduce((sum, item) => sum + item.amount, 0) ?? 0)}>{!filtered?.range_stats?.revenue_by_plan.length ? <div className="flex min-h-[240px] items-center justify-center border-b border-border text-sm text-body">Belum ada data pada rentang ini</div> : <div className="max-h-[340px] space-y-5 overflow-y-auto pr-1">{filtered.range_stats.revenue_by_plan.map(plan => <div key={plan.plan_id}><div className="mb-2 flex flex-wrap justify-between gap-2 text-sm"><span className="min-w-0 break-words font-medium">{plan.name}</span><span className="font-semibold">{formatCurrency(plan.amount)}</span></div><div className="h-2 rounded-full bg-brand/10"><div className="h-2 rounded-full bg-brand" style={{ width: `${Math.max(0, Math.min(100, plan.amount / Math.max(1, ...filtered.range_stats!.revenue_by_plan.map(item => item.amount)) * 100))}%` }} /></div></div>)}</div>}</SummaryMetricCard>
        <SummaryMetricCard timezone={zone} statistic="yearly" caption="Tidak terpengaruh filter · Tahun kalender UTC" total={snapshot ? formatCurrency(snapshot.yearly_breakdown.reduce((sum, point) => sum + point.income, 0)) : '—'} loading={baseLoading} error={baseError} onRetry={() => setRetryBase(value => value + 1)}><SummaryChart points={(snapshot?.yearly_breakdown ?? []).map(point => ({ date: point.year, label: point.year, value: point.income }))} money legend="Pembayaran sukses per tahun" /></SummaryMetricCard>
      </>}
    </div>
    {exportOpen && auth && timezone && <SummaryExportDialog onClose={() => setExportOpen(false)} initialPeriod={period} initialPreset={range === 'harian' ? '1b' : range} timezone={timezone} token={auth.token} businessID={auth.user.business_id} salesOnly={Boolean(isSales)} />}
  </div>
}
