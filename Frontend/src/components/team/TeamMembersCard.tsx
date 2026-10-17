import { useState } from 'react'
import { Maximize2 } from 'lucide-react'
import type { UserDTO } from '../../lib/api'
import GlassPanel from '../ui/GlassPanel'
import SubmitButton from '../ui/SubmitButton'
import MemberGrid from './MemberGrid'
import TeamModal from './TeamModal'

interface TeamMembersCardProps { members: UserDTO[]; onAdd: () => void; notice?: string | null }

export default function TeamMembersCard({ members, onAdd, notice }: TeamMembersCardProps) {
  const [expanded, setExpanded] = useState(false)
  return <>
    <GlassPanel className="team-members-card flex min-w-0 flex-col !shadow-sm">
      <header className="mb-6 flex items-start justify-between gap-4"><div><h3 className="text-base font-semibold text-ink">Tim Bisnis</h3><p className="mt-1 text-sm leading-6 text-body">Anggota yang terhubung ke bisnis ini.</p></div><button type="button" onClick={() => setExpanded(true)} aria-label="Layar besar Tim Bisnis" className="flex size-9 shrink-0 items-center justify-center rounded-full border border-slate-300 text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand dark:border-slate-600"><Maximize2 size={16} aria-hidden="true" /></button></header>
      <MemberGrid members={members} />
      {notice && <p role="status" className="mt-4 text-sm text-emerald-700 dark:text-emerald-200">{notice}</p>}
      <footer className="pt-6"><div className="flex flex-wrap items-center justify-between gap-4 border-t border-slate-200 pt-4 dark:border-slate-700"><p className="text-sm text-body">{members.length} anggota</p><SubmitButton type="button" onClick={onAdd} className="px-4 py-2 text-xs">Tambah Anggota</SubmitButton></div></footer>
    </GlassPanel>
    {expanded && <TeamModal open onClose={() => setExpanded(false)} members={members} onAdd={() => { setExpanded(false); onAdd() }} />}
  </>
}
