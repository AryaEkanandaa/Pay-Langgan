import { Bar, BarChart, ResponsiveContainer, Tooltip, XAxis } from 'recharts'
import type { DashboardSummary } from '../../lib/api'
import { formatCurrency } from '../../lib/format'

type YearlyBreakdownPoint = DashboardSummary['yearly_breakdown'][number]

interface YearlyBreakdownChartProps {
  data: YearlyBreakdownPoint[]
}

export default function YearlyBreakdownChart({ data }: YearlyBreakdownChartProps) {
  return (
    <ResponsiveContainer width="100%" height={220}>
      <BarChart data={data} barCategoryGap="24%" barGap={4}>
        <XAxis
          dataKey="year"
          axisLine={false}
          tickLine={false}
          tick={{ fill: 'var(--color-ink)', fontSize: 12, fontWeight: 600 }}
        />
        <Bar dataKey="income" fill="var(--color-brand)" radius={[6, 6, 0, 0]} maxBarSize={16} />
        <Tooltip formatter={(value) => formatCurrency(Number(value))} />
      </BarChart>
    </ResponsiveContainer>
  )
}
