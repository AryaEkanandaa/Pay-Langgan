import type { LucideIcon } from 'lucide-react'
import GlassPanel from '../ui/GlassPanel'
import Skeleton from '../ui/Skeleton'

interface StatCardProps {
  icon: LucideIcon
  label: string
  value?: string
  delta?: string
  caption: string
  isLoading?: boolean
}

export default function StatCard({ icon: Icon, label, value, delta, caption, isLoading }: StatCardProps) {
  return (
    <GlassPanel className="flex flex-col gap-4">
      <span className="flex h-11 w-11 items-center justify-center rounded-2xl bg-brand text-white">
        <Icon size={20} strokeWidth={2} />
      </span>

      <p className="text-[13px] font-semibold text-ink">{label}</p>

      {isLoading || value === undefined ? (
        <Skeleton className="h-7 w-20" />
      ) : (
        <div className="flex items-baseline gap-2">
          <span className="font-display text-[22px] font-bold text-ink">{value}</span>
          {delta && (
            <span className="rounded-full bg-success/10 px-2 py-0.5 text-[11px] font-semibold text-success">
              {delta}
            </span>
          )}
        </div>
      )}

      <p className="text-[12px] text-muted">{caption}</p>
    </GlassPanel>
  )
}
