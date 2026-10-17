import { useEffect, useId, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { CircleHelp } from 'lucide-react'
import { statInfo, type StatInfoKey } from '../../data/statInfo'

export default function InfoPopover({ statistic, timezone }: { statistic: StatInfoKey; timezone: string }) {
  const info = statInfo[statistic], id = useId(), trigger = useRef<HTMLButtonElement>(null), popup = useRef<HTMLDivElement>(null)
  const [open, setOpen] = useState(false), [position, setPosition] = useState({ left: 0, top: 0, below: true })
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null)
  function show() { if (timer.current) clearTimeout(timer.current); setOpen(true) }
  function hideLater() { timer.current = setTimeout(() => setOpen(false), 100) }
  useEffect(() => () => { if (timer.current) clearTimeout(timer.current) }, [])
  useEffect(() => {
    if (!open) return
    const update = () => {
      const anchor = trigger.current?.getBoundingClientRect(), panel = popup.current
      if (!anchor || !panel) return
      const below = anchor.bottom + panel.offsetHeight + 16 < innerHeight
      setPosition({ left: Math.max(8, Math.min(anchor.right - panel.offsetWidth, innerWidth - panel.offsetWidth - 8)), top: below ? anchor.bottom + 8 : Math.max(8, anchor.top - panel.offsetHeight - 8), below })
    }
    const outside = (event: Event) => { if (event.target instanceof Node && !trigger.current?.contains(event.target) && !popup.current?.contains(event.target)) setOpen(false) }
    const escape = (event: KeyboardEvent) => { if (event.key === 'Escape') { event.stopPropagation(); setOpen(false) } }
    update(); window.addEventListener('resize', update); window.addEventListener('scroll', update, true)
    document.addEventListener('pointerdown', outside); document.addEventListener('focusin', outside); document.addEventListener('keydown', escape, true)
    return () => { window.removeEventListener('resize', update); window.removeEventListener('scroll', update, true); document.removeEventListener('pointerdown', outside); document.removeEventListener('focusin', outside); document.removeEventListener('keydown', escape, true) }
  }, [open])
  return <>
    <button ref={trigger} type="button" aria-label={`Info tentang ${info.title}`} aria-expanded={open} aria-describedby={open ? id : undefined} onMouseEnter={show} onMouseLeave={hideLater} onFocus={show} onClick={show} className="flex size-8 shrink-0 items-center justify-center rounded-full text-body hover:bg-brand/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"><CircleHelp size={17} aria-hidden="true" /></button>
    {open && createPortal(<div ref={popup} id={id} role="tooltip" onMouseEnter={show} onMouseLeave={hideLater} style={{ left: position.left, top: position.top }} className="stat-info-popup fixed z-[60] w-[280px] max-w-[calc(100vw-16px)] rounded-xl border border-slate-200 bg-white p-4 text-sm leading-6 text-slate-700 shadow-xl dark:border-slate-600 dark:bg-slate-900 dark:text-slate-200">
      <span aria-hidden="true" className={`absolute right-4 size-2 rotate-45 border-slate-200 bg-white dark:border-slate-600 dark:bg-slate-900 ${position.below ? '-top-1 border-l border-t' : '-bottom-1 border-b border-r'}`} />
      <p className="font-semibold text-ink">{info.title}</p><p className="mt-2">{info.definition}</p><div className="my-3 border-t border-slate-200 dark:border-slate-600" /><p className="font-semibold">Dihitung dari:</p><p>{info.calculation}</p><p className="mt-3 text-xs">Zona waktu: {statistic === 'yearly' ? 'UTC (tahun kalender)' : timezone}. {info.filtered ? 'Mengikuti filter rentang.' : 'Tidak terpengaruh filter rentang.'}</p>
    </div>, document.body)}
  </>
}
