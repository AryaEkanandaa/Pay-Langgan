// Data contoh (placeholder) untuk chart dashboard.
// Ganti dengan data asli setelah endpoint statistik/revenue tersedia di backend.

export interface MonthlyPerformancePoint {
  month: string
  value: number
}

export const monthlyPerformanceData: MonthlyPerformancePoint[] = [
  { month: 'Jan', value: 8200 },
  { month: 'Feb', value: 10400 },
  { month: 'Mar', value: 9100 },
  { month: 'Apr', value: 12800 },
  { month: 'Mei', value: 15870 },
  { month: 'Jun', value: 13200 },
  { month: 'Jul', value: 14600 },
]

export interface YearlyBreakdownPoint {
  year: string
  income: number
  spending: number
}

export const yearlyBreakdownData: YearlyBreakdownPoint[] = [
  { year: '2022', income: 32, spending: 14 },
  { year: '2023', income: 38, spending: 18 },
  { year: '2024', income: 29, spending: 22 },
  { year: '2025', income: 45, spending: 20 },
  { year: '2026', income: 41, spending: 16 },
]

export const yearlyIncomeSharePct = 67
export const yearlySpendingSharePct = 33
