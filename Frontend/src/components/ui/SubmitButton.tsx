import type { ButtonHTMLAttributes } from 'react'
import { buttonBaseClass, buttonVariantClass, type ButtonVariant } from './buttonStyles'

interface SubmitButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
}

export default function SubmitButton({
  variant = 'primary',
  className = '',
  ...props
}: SubmitButtonProps) {
  return (
    <button
      className={`${buttonBaseClass} ${buttonVariantClass[variant]} ${className}`}
      {...props}
    />
  )
}
