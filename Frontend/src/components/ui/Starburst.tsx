interface StarburstProps {
  className?: string
}

export default function Starburst({ className = '' }: StarburstProps) {
  return (
    <svg viewBox="0 0 200 200" fill="none" className={className} aria-hidden="true">
      <path
        d="M100 8 L110 90 L192 100 L110 110 L100 192 L90 110 L8 100 L90 90 Z"
        fill="currentColor"
        opacity="0.18"
      />
      <g stroke="currentColor" strokeWidth="1" opacity="0.35">
        <line x1="100" y1="0" x2="100" y2="200" />
        <line x1="0" y1="100" x2="200" y2="100" />
        <line x1="35" y1="35" x2="165" y2="165" />
        <line x1="165" y1="35" x2="35" y2="165" />
      </g>
      <circle cx="100" cy="100" r="88" stroke="currentColor" strokeWidth="0.75" opacity="0.25" />
    </svg>
  )
}
