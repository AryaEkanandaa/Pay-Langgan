import type { ReactNode } from 'react'

interface SettingsRowProps {
  title: ReactNode
  description?: ReactNode
  children: ReactNode
  stacked?: boolean
}

export default function SettingsRow({ title, description, children, stacked = false }: SettingsRowProps) {
  return (
    <div className={`grid min-w-0 items-center gap-4 border-b border-slate-200 py-4 first:pt-0 last:border-b-0 last:pb-0 ${stacked ? '' : 'sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]'} dark:border-slate-700`}>
      <div className="min-w-0"><div className="text-sm font-semibold text-ink">{title}</div>{description && <div className="mt-1 break-words text-sm leading-6 text-body">{description}</div>}</div>
      <div className={`min-w-0 break-words text-sm text-ink ${stacked ? '[&>span]:text-center' : 'sm:text-right'}`}>{children}</div>
    </div>
  )
}
