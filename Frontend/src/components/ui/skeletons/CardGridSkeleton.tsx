import GlassPanel from '../GlassPanel'
import Skeleton from '../Skeleton'

interface CardGridSkeletonProps {
  count?: number
}

export default function CardGridSkeleton({ count = 6 }: CardGridSkeletonProps) {
  return (
    <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: count }).map((_, index) => (
        <GlassPanel key={index}>
          <Skeleton className="h-10 w-10 rounded-lg" />
          <Skeleton className="mt-4 h-4 w-2/3" />
          <Skeleton className="mt-2 h-3 w-full" />
          <Skeleton className="mt-1 h-3 w-4/5" />
          <div className="mt-4 flex items-center justify-between">
            <Skeleton className="h-5 w-16" />
            <Skeleton className="h-8 w-20 rounded-full" />
          </div>
        </GlassPanel>
      ))}
    </div>
  )
}
