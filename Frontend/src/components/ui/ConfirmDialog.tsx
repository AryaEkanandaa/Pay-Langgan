import { useRef, type ReactNode } from 'react'
import Modal from './Modal'
import SubmitButton from './SubmitButton'

interface ConfirmDialogProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  description: ReactNode
  confirmLabel: string
  busy?: boolean
  error?: string | null
}

export default function ConfirmDialog({ open, onClose, onConfirm, title, description, confirmLabel, busy = false, error }: ConfirmDialogProps) {
  const cancelRef = useRef<HTMLButtonElement>(null)
  return (
    <Modal open={open} onClose={() => { if (!busy) onClose() }} title={title} trapFocus initialFocusRef={cancelRef}>
      <div className="mt-4 text-sm leading-6 text-body">{description}</div>
      {error && <p role="alert" className="mt-4 text-sm text-red-700 dark:text-red-300">{error}</p>}
      <div className="mt-6 flex flex-wrap justify-end gap-3">
        <button ref={cancelRef} type="button" disabled={busy} onClick={onClose} className="rounded-xl border border-slate-300 px-4 py-2 text-sm font-semibold text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand disabled:opacity-50 dark:border-slate-600">Batal</button>
        <SubmitButton type="button" disabled={busy} onClick={onConfirm} className="bg-red-700 text-sm hover:bg-red-800">{busy ? 'Memproses...' : confirmLabel}</SubmitButton>
      </div>
    </Modal>
  )
}
