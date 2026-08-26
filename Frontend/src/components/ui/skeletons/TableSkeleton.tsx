import Skeleton from '../Skeleton'

interface TableSkeletonProps {
  columns: string[]
  rows?: number
}

export default function TableSkeleton({ columns, rows = 5 }: TableSkeletonProps) {
  return (
    <div className="glass overflow-hidden rounded-2xl">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[560px] text-left text-[13px]">
          <thead>
            <tr className="border-b border-white/50">
              {columns.map((column) => (
                <th
                  key={column}
                  className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted"
                >
                  {column}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {Array.from({ length: rows }).map((_, rowIndex) => (
              <tr key={rowIndex} className="border-b border-white/30 last:border-0">
                {columns.map((column) => (
                  <td key={column} className="px-5 py-4">
                    <Skeleton className="h-3.5 w-full max-w-[140px]" />
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
