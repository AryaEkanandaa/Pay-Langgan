import type { ReactNode } from 'react'

export default function StatusBadge({ children, tone = 'neutral' }: { children: ReactNode; tone?: 'success' | 'warning' | 'neutral' }) {
  const colors = { success: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200', warning: 'bg-amber-100 text-amber-900 dark:bg-amber-950 dark:text-amber-200', neutral: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-200' }
  return <span className={`inline-block rounded-full px-3 py-1 text-xs font-semibold ${colors[tone]}`}>{children}</span>
}
