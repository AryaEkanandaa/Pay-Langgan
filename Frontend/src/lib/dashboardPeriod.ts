import type { DashboardRangePoint, DashboardRangeStats } from './api'

export const dashboardRanges = [{ id: 'harian', label: 'Harian' }, { id: '3b', label: '3 bulan' }, { id: '6b', label: '6 bulan' }, { id: '12b', label: '12 bulan' }] as const
export type DashboardRange = typeof dashboardRanges[number]['id']
export interface DatePeriod { from: string; to: string }
export const rangeGranularity = { harian: 'day', '3b': 'week', '6b': 'month', '12b': 'month' } as const
export function businessToday(timezone: string, date = new Date()) {
  const parts = new Intl.DateTimeFormat('en-CA', { timeZone: timezone, year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(date)
  return ['year', 'month', 'day'].map(type => parts.find(part => part.type === type)?.value).join('-')
}
export function addDays(day: string, count: number) {
  const date = new Date(`${day}T12:00:00Z`)
  date.setUTCDate(date.getUTCDate() + count)
  return date.toISOString().slice(0, 10)
}
export function addMonths(day: string, count: number) {
  const date = new Date(`${day}T12:00:00Z`), original = date.getUTCDate()
  date.setUTCDate(1); date.setUTCMonth(date.getUTCMonth() + count)
  const last = new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth() + 1, 0)).getUTCDate()
  date.setUTCDate(Math.min(original, last))
  return date.toISOString().slice(0, 10)
}
export function periodFor(range: DashboardRange, today: string): DatePeriod {
  return { from: range === 'harian' ? addDays(today, -29) : addDays(addMonths(today, -Number.parseInt(range)), 1), to: today }
}
export function periodLabel(period: DatePeriod) {
  const formatter = new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short', year: 'numeric', timeZone: 'UTC' })
  return `${formatter.format(new Date(`${period.from}T12:00:00Z`))} – ${formatter.format(new Date(`${period.to}T12:00:00Z`))}`
}
export function periodDays(period: DatePeriod) { return Math.round((Date.parse(period.to) - Date.parse(period.from)) / 86400000) + 1 }
export function periodError(period: DatePeriod, today: string) {
  if (!period.from || !period.to || !/^\d{4}-\d{2}-\d{2}$/.test(period.from) || !/^\d{4}-\d{2}-\d{2}$/.test(period.to) || !Number.isFinite(Date.parse(period.from)) || !Number.isFinite(Date.parse(period.to))) return 'Isi tanggal Dari dan Sampai yang valid.'
  if ([period.from, period.to].some(date => new Date(date).toISOString().slice(0, 10) !== date)) return 'Isi tanggal Dari dan Sampai yang valid.'
  if (period.from > period.to) return 'Tanggal Dari tidak boleh setelah Sampai.'
  if (period.to > today) return 'Tanggal tidak boleh di masa depan.'
  if (period.from < addMonths(period.to, -24)) return 'Rentang maksimum 24 bulan.'
  return null
}
export function completeSeries(stats: DashboardRangeStats): DashboardRangePoint[] {
  let date = stats.granularity === 'month' ? `${stats.from.slice(0, 7)}-01` : stats.from
  if (stats.granularity === 'week') date = addDays(date, -((new Date(`${date}T12:00:00Z`).getUTCDay() + 6) % 7))
  const values = new Map(stats.series.map(point => [point.date, point]))
  const result: DashboardRangePoint[] = []
  while (date <= stats.to) {
    result.push(values.get(date) ?? { date, revenue: 0, new_subscriptions: 0, cancelled_subscriptions: 0, payments_succeeded: 0, payments_failed: 0, autopay_succeeded: 0, autopay_failed: 0 })
    date = stats.granularity === 'month' ? addMonths(date, 1) : addDays(date, stats.granularity === 'week' ? 7 : 1)
  }
  return result
}
