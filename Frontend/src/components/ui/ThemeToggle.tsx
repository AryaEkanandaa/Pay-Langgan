import { Moon, Sun } from 'lucide-react'
import { useTheme } from '../../context/ThemeContext'

interface ThemeToggleProps {
  compact?: boolean
}

export default function ThemeToggle({ compact = false }: ThemeToggleProps) {
  const { theme, toggleTheme } = useTheme()
  const isDark = theme === 'dark'

  return (
    <button
      type="button"
      onClick={toggleTheme}
      aria-label={isDark ? 'Aktifkan mode terang' : 'Aktifkan mode gelap'}
      title={isDark ? 'Mode terang' : 'Mode gelap'}
      className={`group inline-flex items-center gap-2 border transition-all duration-300 focus:outline-none focus:ring-2 focus:ring-brand/40 ${
        compact
          ? 'h-10 rounded-lg border-white/10 bg-white/10 px-3 text-white/75 hover:bg-white/15 hover:text-white'
          : 'rounded-xl border-border/70 bg-white/55 px-3 py-2 text-body shadow-sm hover:bg-white/75 hover:text-ink dark:border-white/10 dark:bg-white/10 dark:text-white/75 dark:hover:bg-white/15 dark:hover:text-white'
      }`}
    >
      <span
        className={`relative flex h-5 w-9 items-center rounded-full p-0.5 transition-colors duration-300 ${
          isDark ? 'bg-brand' : 'bg-ink/20'
        }`}
      >
        <span
          className={`flex h-4 w-4 items-center justify-center rounded-full bg-white shadow-sm transition-transform duration-300 ${
            isDark ? 'translate-x-4' : 'translate-x-0'
          }`}
        >
          {isDark ? <Moon size={10} className="text-brand" /> : <Sun size={10} className="text-amber" />}
        </span>
      </span>
      {!compact && <span className="text-[12px] font-semibold">{isDark ? 'Mode gelap' : 'Mode terang'}</span>}
    </button>
  )
}
