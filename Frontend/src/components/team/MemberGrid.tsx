import { ChevronLeft, ChevronRight } from 'lucide-react'
import type { UserDTO } from '../../lib/api'
import useHorizontalScroll from '../../hooks/useHorizontalScroll'
import MemberCard from './MemberCard'

export default function MemberGrid({ members }: { members: UserDTO[] }) {
  const { scrollRef, left, right, progress, scrollByPage } = useHorizontalScroll(members.length)
  return <div className="min-w-0">
    <div className="relative">
      <div ref={scrollRef} role="region" aria-label="Daftar anggota tim" tabIndex={0} onKeyDown={event => { if (event.key === 'ArrowLeft' || event.key === 'ArrowRight') { event.preventDefault(); scrollByPage(event.key === 'ArrowLeft' ? -1 : 1) } }} className="team-scroll min-w-0 overflow-x-auto rounded-xl focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand">
        {members.length ? <div className={`grid auto-cols-[minmax(220px,1fr)] grid-flow-col gap-3 ${members.length > 1 ? 'grid-rows-2' : 'grid-rows-1'}`}>{members.map(member => <div key={member.id} className="snap-start"><MemberCard member={member} /></div>)}</div> : <p className="rounded-xl bg-white/30 p-6 text-sm text-body dark:bg-white/5">Belum ada anggota tim.</p>}
      </div>
      {left && <span aria-hidden="true" className="pointer-events-none absolute inset-y-0 left-0 w-6 rounded-l-xl bg-gradient-to-r from-slate-100 to-transparent dark:from-slate-900" />}
      {right && <span aria-hidden="true" className="pointer-events-none absolute inset-y-0 right-0 w-6 rounded-r-xl bg-gradient-to-l from-slate-100 to-transparent dark:from-slate-900" />}
    </div>
    <div className="mt-4 flex items-center gap-4">
      <div role="progressbar" aria-label="Posisi daftar anggota" aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round(progress * 100)} className="h-1 flex-1 overflow-hidden rounded-full bg-slate-200 dark:bg-slate-700"><div className="h-full rounded-full bg-brand" style={{ width: `${left || right ? 20 + progress * 80 : 100}%` }} /></div>
      <div className="flex gap-2">{([-1, 1] as const).map(direction => <button key={direction} type="button" aria-label={direction === -1 ? 'Geser anggota ke kiri' : 'Geser anggota ke kanan'} disabled={direction === -1 ? !left : !right} onClick={() => scrollByPage(direction)} className="flex size-8 items-center justify-center rounded-full border border-slate-300 text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand disabled:opacity-40 dark:border-slate-600">{direction === -1 ? <ChevronLeft size={16} /> : <ChevronRight size={16} />}</button>)}</div>
    </div>
  </div>
}
