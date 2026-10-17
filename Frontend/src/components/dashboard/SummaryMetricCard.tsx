import type { ReactNode } from 'react'
import GlassPanel from '../ui/GlassPanel'
import Skeleton from '../ui/Skeleton'
import InfoPopover from './InfoPopover'
import { statInfo, type StatInfoKey } from '../../data/statInfo'

export default function SummaryMetricCard({ statistic, timezone, caption, total, loading, error, onRetry, children, className = '' }: { statistic: StatInfoKey; timezone: string; caption: string; total: string; loading: boolean; error: string | null; onRetry: () => void; children: ReactNode; className?: string }) {
  return <GlassPanel className={`flex min-h-[calc(clamp(240px,22vw,340px)+208px)] min-w-0 flex-col !shadow-sm ${className}`}>
    <div className="flex items-center justify-between gap-3"><h2 className="text-sm font-semibold">{statInfo[statistic].title}</h2><InfoPopover statistic={statistic} timezone={timezone} /></div>
    <p className="mt-1 text-xs leading-5 text-body">{caption}</p>
    {loading ? <div role="status" aria-label={`Memuat ${statInfo[statistic].title}`} className="mt-4"><Skeleton className="h-9 w-36" /><Skeleton className="mt-6 h-[clamp(240px,22vw,340px)] w-full" /></div> : error ? <div role="alert" className="flex min-h-[340px] flex-1 flex-col items-center justify-center gap-4 text-center"><p className="text-sm text-red-600">{error}</p><button type="button" onClick={onRetry} className="rounded-full border border-border px-4 py-2 text-sm font-semibold focus-visible:outline-2 focus-visible:outline-brand">Coba lagi</button></div> : <><p className="mt-4 break-words text-[clamp(24px,1.7vw,36px)] font-bold leading-tight">{total}</p><div className="mt-4 flex min-h-0 flex-1 flex-col">{children}</div></>}
  </GlassPanel>
}
