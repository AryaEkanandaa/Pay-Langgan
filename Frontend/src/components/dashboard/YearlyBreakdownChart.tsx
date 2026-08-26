import { Bar, BarChart, ResponsiveContainer, XAxis } from 'recharts'
import type { YearlyBreakdownPoint } from '../../data/dashboardMock'

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
        <Bar dataKey="spending" fill="var(--color-amber)" radius={[6, 6, 0, 0]} maxBarSize={16} />
      </BarChart>
    </ResponsiveContainer>
  )
}
