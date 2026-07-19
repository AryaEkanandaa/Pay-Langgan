import type { InputHTMLAttributes } from 'react'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string
  variant?: 'light' | 'dark'
}

const variantClass: Record<'light' | 'dark', { label: string; field: string }> = {
  light: {
    label: 'text-body',
    field:
      'border-white/60 bg-white/50 text-ink placeholder:text-muted focus:border-brand/50 focus:bg-white/70 focus:ring-brand/30',
  },
  dark: {
    label: 'text-neutral/70',
    field:
      'border-white/10 bg-black/25 text-neutral placeholder:text-neutral/40 focus:border-brand/50 focus:bg-black/35 focus:ring-brand/40',
  },
}

export default function Input({
  label,
  id,
  variant = 'light',
  className = '',
  ...props
}: InputProps) {
  const styles = variantClass[variant]

  return (
    <label className="block text-left" htmlFor={id}>
      <span className={`text-[13px] font-medium ${styles.label}`}>{label}</span>
      <input
        id={id}
        className={`mt-1.5 w-full rounded-xl border px-4 py-2.5 text-[14px] outline-none backdrop-blur-xl transition-colors focus:ring-2 ${styles.field} ${className}`}
        {...props}
      />
    </label>
  )
}
