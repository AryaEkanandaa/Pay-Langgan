import { useId } from 'react'
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { formatCurrency } from '../../lib/format'

export interface SummaryChartPoint { date: string; value: number; label: string }
const compact = new Intl.NumberFormat('id-ID', { notation: 'compact', maximumFractionDigits: 1 })
export default function SummaryChart({ points, money = false, legend }: { points: SummaryChartPoint[]; money?: boolean; legend: string }) {
  const id = useId().replace(/:/g, ''), empty = points.every(point => point.value === 0)
  return <div className="relative flex min-h-0 flex-1 flex-col">
    <div className="summary-chart h-[clamp(240px,22vw,340px)] w-full" aria-label={`${legend}, ${money ? 'nominal rupiah' : 'jumlah'}`}>
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={points} accessibilityLayer margin={{ top: 24, right: 12, bottom: 0, left: 0 }}>
          <defs><linearGradient id={id} x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stopColor="#3654ff" stopOpacity={0.2} /><stop offset="100%" stopColor="#3654ff" stopOpacity={0.02} /></linearGradient></defs>
          <CartesianGrid vertical={false} stroke="var(--color-border)" strokeOpacity={0.6} />
          <XAxis dataKey="label" axisLine={false} tickLine={false} minTickGap={24} tick={{ fill: 'var(--color-body)', fontSize: 12 }} />
          <YAxis width={64} axisLine={false} tickLine={false} allowDecimals={false} domain={[0, 'auto']} tickFormatter={(value: number) => compact.format(value)} tick={{ fill: 'var(--color-body)', fontSize: 12 }} />
          <Tooltip contentStyle={{ backgroundColor: 'var(--summary-tooltip-bg)', borderColor: 'var(--summary-tooltip-border)', borderRadius: 12, color: 'var(--summary-tooltip-text)' }} labelStyle={{ color: 'var(--summary-tooltip-text)' }} formatter={value => [money ? formatCurrency(Number(value)) : Number(value).toLocaleString('id-ID'), legend]} />
          <Area type="monotone" dataKey="value" stroke="#3654ff" strokeWidth={2} fill={`url(#${id})`} dot={points.length === 1} isAnimationActive={false} />
        </AreaChart>
      </ResponsiveContainer>
    </div>
    {empty && <p className="pointer-events-none absolute inset-x-0 top-1/2 text-center text-sm text-body">Belum ada data pada rentang ini</p>}
    <p className="mt-3 flex items-center justify-center gap-2 text-xs text-body"><span className="h-0.5 w-4 bg-brand" aria-hidden="true" />{legend}{money ? ' (Rupiah)' : ''}</p>
  </div>
}
