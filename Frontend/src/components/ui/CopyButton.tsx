import { useEffect, useState } from 'react'
import { Check, Copy } from 'lucide-react'
import { copyText } from '../../lib/clipboard'

interface CopyButtonProps {
  text: string
  label?: string
  iconOnly?: boolean
  buttonText?: string
}

export default function CopyButton({ text, label = 'Salin kode', iconOnly = false, buttonText = 'Salin' }: CopyButtonProps) {
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState(false)
  useEffect(() => { setCopied(false); setError(false) }, [text])
  useEffect(() => {
    if (!copied) return
    const timer = window.setTimeout(() => setCopied(false), 2000)
    return () => window.clearTimeout(timer)
  }, [copied])
  async function copy() {
    try { await copyText(text); setCopied(true); setError(false) }
    catch { setError(true) }
  }
  return <span className="inline-flex shrink-0 flex-col items-end gap-1">
    <button type="button" aria-label={copied ? `${label}: tersalin` : label} onClick={() => void copy()} className="inline-flex select-none min-h-10 items-center justify-center gap-2 rounded-lg border border-current/20 px-3 text-[12px] font-semibold transition hover:bg-current/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand dark:text-slate-200">
      {copied ? <Check size={16} aria-hidden="true" /> : <Copy size={16} aria-hidden="true" />}
      {!iconOnly && (copied ? 'Tersalin ✓' : buttonText)}
    </button>
    <span aria-live="polite" className={error ? 'max-w-60 text-[12px] text-red-600 dark:text-red-300' : 'sr-only'}>{error ? 'Gagal menyalin. Pilih kode dan salin secara manual.' : copied ? 'Tersalin' : ''}</span>
  </span>
}
