import type { HTMLAttributes, ReactNode } from 'react'

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  children: ReactNode
}

export default function Badge({ children, className = '', ...props }: BadgeProps) {
  return (
    <span
      className={`glass inline-flex items-center gap-2 rounded-full px-3 py-1 text-[11px] font-semibold uppercase leading-[1.15] tracking-[0.08em] text-brand ${className}`}
      {...props}
    >
      {children}
    </span>
  )
}
