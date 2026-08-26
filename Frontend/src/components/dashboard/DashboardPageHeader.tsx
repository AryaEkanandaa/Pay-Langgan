interface DashboardPageHeaderProps {
  title: string
  description: string
}

export default function DashboardPageHeader({ title, description }: DashboardPageHeaderProps) {
  return (
    <div className="mb-8">
      <h1 className="font-display text-[24px] font-bold text-ink">{title}</h1>
      <p className="mt-1 text-[14px] text-body">{description}</p>
    </div>
  )
}
