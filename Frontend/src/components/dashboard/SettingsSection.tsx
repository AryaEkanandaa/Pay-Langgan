import { useId, type ReactNode } from 'react'

interface SettingsSectionProps {
  title: string
  description: string
  children: ReactNode
  aside?: ReactNode
  stretch?: boolean
}

export default function SettingsSection({ title, description, children, aside, stretch = false }: SettingsSectionProps) {
  const titleId = useId()
  return (
    <section aria-labelledby={titleId} className={`settings-section grid ${stretch ? 'items-stretch [&>div>.glass]:h-full [&>aside>.glass]:h-full [&>aside>.glass]:flex [&>aside>.glass]:flex-col' : 'items-start'} gap-6 min-[900px]:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] min-[1100px]:grid-cols-[280px_minmax(0,1fr)] min-[1100px]:gap-8 min-[1600px]:gap-12 ${aside ? stretch ? 'min-[1600px]:grid-cols-[280px_minmax(0,720px)_minmax(320px,360px)]' : 'min-[1600px]:grid-cols-[280px_minmax(0,720px)_320px]' : ''}`}>
      <header className="min-w-0 min-[900px]:max-[1099px]:col-span-2">
        <h2 id={titleId} className="text-base font-semibold text-ink">{title}</h2>
        <p className="mt-2 text-sm leading-6 text-body">{description}</p>
      </header>
      <div className="min-w-0 [&>.glass]:shadow-sm">{children}</div>
      {aside && <aside aria-label={`Informasi ${title}`} className="hidden min-w-0 min-[900px]:max-[1099px]:block min-[1600px]:block [&>.glass]:shadow-sm">{aside}</aside>}
    </section>
  )
}
