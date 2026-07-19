import type { AnchorHTMLAttributes, ReactNode } from 'react'
import { buttonBaseClass, buttonVariantClass, type ButtonVariant } from './buttonStyles'

interface ButtonProps extends AnchorHTMLAttributes<HTMLAnchorElement> {
  variant?: ButtonVariant
  children: ReactNode
}

export default function Button({
  variant = 'primary',
  className = '',
  children,
  ...props
}: ButtonProps) {
  return (
    <a
      className={`${buttonBaseClass} ${buttonVariantClass[variant]} ${className}`}
      {...props}
    >
      {children}
    </a>
  )
}
