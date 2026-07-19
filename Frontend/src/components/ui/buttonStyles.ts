export type ButtonVariant = 'primary' | 'secondary'

export const buttonBaseClass =
  'inline-flex h-11 items-center justify-center rounded-full px-6 text-[14px] font-semibold leading-[1.2] transition-all active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-60'

export const buttonVariantClass: Record<ButtonVariant, string> = {
  primary:
    'bg-brand text-neutral shadow-[0_10px_30px_-10px_rgba(54,84,255,0.55)] hover:bg-brand-dark hover:shadow-[0_14px_34px_-10px_rgba(54,84,255,0.65)]',
  secondary: 'glass text-ink hover:bg-white/60',
}
