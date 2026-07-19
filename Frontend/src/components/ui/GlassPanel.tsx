import type { HTMLAttributes, ReactNode } from 'react'

interface GlassPanelProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode
}

export default function GlassPanel({ children, className = '', ...props }: GlassPanelProps) {
  return (
    <div
      className={`glass rounded-2xl p-6 text-ink shadow-[0_8px_32px_-8px_rgba(16,16,20,0.18)] ${className}`}
      {...props}
    >
      {children}
    </div>
  )
}
