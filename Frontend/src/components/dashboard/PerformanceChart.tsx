import {
  Bar,
  BarChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  type TooltipContentProps,
} from 'recharts'
import type { NameType, ValueType } from 'recharts/types/component/DefaultTooltipContent'
import type { DashboardSummary } from '../../lib/api'

type MonthlyPerformancePoint = DashboardSummary['monthly_performance'][number]

interface PerformanceChartProps {
  data: MonthlyPerformancePoint[]
}

function ChartTooltip({ active, payload }: TooltipContentProps<ValueType, NameType>) {
  if (!active || !payload?.length) return null
  const value = Number(payload[0].value)

  return (
    <div className="glass flex items-center gap-2 rounded-full bg-ink/90 px-4 py-2 text-white shadow-lg">
      <span className="flex h-6 w-6 items-center justify-center rounded-full bg-brand text-[11px]">●</span>
      <div className="text-left">
        <p className="text-[13px] font-bold leading-tight">{value.toLocaleString('id-ID')}</p>
        <p className="text-[10px] text-white/60">Views / hr</p>
      </div>
    </div>
  )
}

export default function PerformanceChart({ data }: PerformanceChartProps) {
  return (
    <ResponsiveContainer width="100%" height={260}>
      <BarChart data={data} barCategoryGap="30%">
        <XAxis
          dataKey="month"
          axisLine={false}
          tickLine={false}
          tick={{ fill: 'var(--color-muted)', fontSize: 12 }}
        />
        <YAxis
          axisLine={false}
          tickLine={false}
          tick={{ fill: 'var(--color-muted)', fontSize: 12 }}
          tickFormatter={(value: number) => (value >= 1000 ? `${value / 1000}K` : String(value))}
        />
        <Tooltip content={ChartTooltip} cursor={{ fill: 'rgba(54, 84, 255, 0.08)' }} />
        <Bar dataKey="value" fill="var(--color-brand)" radius={[6, 6, 0, 0]} maxBarSize={22} />
      </BarChart>
    </ResponsiveContainer>
  )
}
