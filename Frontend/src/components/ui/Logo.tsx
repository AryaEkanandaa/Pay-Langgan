import { Link } from 'react-router-dom'
import logo from '../../assets/logo_paylanggan.png'

interface LogoProps {
  className?: string
  variant?: 'dark' | 'light'
}

export default function Logo({ className = '', variant = 'dark' }: LogoProps) {
  return (
    <Link to="/" className={`flex items-center gap-2 ${className}`}>
      <img src={logo} alt="PayLanggan" className="h-8 w-8 object-contain" />
      <span
        className={`font-display text-[18px] font-bold tracking-tight ${
          variant === 'light' ? 'text-neutral' : 'text-ink'
        }`}
      >
        PayLanggan
      </span>
    </Link>
  )
}
