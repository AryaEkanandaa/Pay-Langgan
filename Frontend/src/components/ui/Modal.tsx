import { useEffect, useId, useRef, type ReactNode, type RefObject } from 'react'
import { X } from 'lucide-react'

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
  size?: 'default' | 'wide' | 'team'
  trapFocus?: boolean
  initialFocusRef?: RefObject<HTMLElement | null>
}

export default function Modal({ open, onClose, title, children, size = 'default', trapFocus = false, initialFocusRef }: ModalProps) {
  const titleId = useId()
  const dialogRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open || !trapFocus) return
    const previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null
    const dialog = dialogRef.current
    const focusables = () => dialog?.querySelectorAll<HTMLElement>('button:not(:disabled), a[href], input:not(:disabled), select:not(:disabled), textarea:not(:disabled), [tabindex="0"]')
    ;(initialFocusRef?.current ?? focusables()?.[0] ?? dialog)?.focus()
    function handleTab(event: KeyboardEvent) {
      if (event.key !== 'Tab') return
      const elements = focusables()
      const first = elements?.[0]
      const last = elements?.[elements.length - 1]
      if (!first || !last) { event.preventDefault(); dialog?.focus(); return }
      if (event.shiftKey && (document.activeElement === first || !dialog?.contains(document.activeElement))) { event.preventDefault(); last.focus() }
      else if (!event.shiftKey && (document.activeElement === last || !dialog?.contains(document.activeElement))) { event.preventDefault(); first.focus() }
    }
    document.addEventListener('keydown', handleTab)
    return () => { document.removeEventListener('keydown', handleTab); previousFocus?.focus({ preventScroll: true }) }
  }, [open, trapFocus, initialFocusRef])
  useEffect(() => {
    if (!open) return
    const handleKey = (event: KeyboardEvent) => { if (event.key === 'Escape') onClose() }
    window.addEventListener('keydown', handleKey)
    return () => window.removeEventListener('keydown', handleKey)
  }, [open, onClose])
  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center px-4">
      <div className="absolute inset-0 bg-ink/50 backdrop-blur-sm" onClick={onClose} />

      <div ref={dialogRef} tabIndex={-1} role="dialog" aria-modal="true" aria-labelledby={titleId} className={`glass relative ${size === 'team' ? 'team-modal flex max-h-[85dvh] w-[92vw] max-w-[1100px] flex-col overflow-hidden' : `max-h-[90dvh] w-full ${size === 'wide' ? 'max-w-4xl' : 'max-w-md'} overflow-y-auto`} rounded-2xl bg-white/95 p-6 shadow-2xl`}>
        <div className="flex shrink-0 items-center justify-between">
          <h2 id={titleId} className="font-display text-[18px] font-bold text-ink">{title}</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup"
            className="flex h-8 w-8 items-center justify-center rounded-full text-body transition-colors hover:bg-white/60 hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand"
          >
            <X size={18} />
          </button>
        </div>

        <div className={size === 'team' ? 'mt-4 flex min-h-0 flex-1 flex-col overflow-hidden' : 'mt-4'}>{children}</div>
      </div>
    </div>
  )
}
