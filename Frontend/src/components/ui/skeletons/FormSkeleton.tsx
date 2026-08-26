import GlassPanel from '../GlassPanel'
import Skeleton from '../Skeleton'

interface FormSkeletonProps {
  fields: string[]
}

export default function FormSkeleton({ fields }: FormSkeletonProps) {
  return (
    <GlassPanel className="max-w-xl">
      <div className="space-y-5">
        {fields.map((field) => (
          <div key={field}>
            <p className="text-[13px] font-medium text-body">{field}</p>
            <Skeleton className="mt-1.5 h-10 w-full rounded-xl" />
          </div>
        ))}
      </div>
      <Skeleton className="mt-8 h-11 w-32 rounded-full" />
    </GlassPanel>
  )
}
