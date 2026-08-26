import type { SelectHTMLAttributes } from 'react'

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string
}

export default function Select({ label, id, className = '', children, ...props }: SelectProps) {
  return (
    <label className="block text-left" htmlFor={id}>
      <span className="text-[13px] font-medium text-body">{label}</span>
      <select
        id={id}
        className={`mt-1.5 w-full rounded-xl border border-white/60 bg-white/50 px-4 py-2.5 text-[14px] text-ink outline-none backdrop-blur-xl transition-colors focus:border-brand/50 focus:bg-white/70 focus:ring-2 focus:ring-brand/30 ${className}`}
        {...props}
      >
        {children}
      </select>
    </label>
  )
}
