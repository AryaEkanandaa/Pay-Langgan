import { useEffect, useState } from 'react'
import type { UserDTO } from '../../lib/api'
import Modal from '../ui/Modal'
import Input from '../ui/Input'
import SubmitButton from '../ui/SubmitButton'
import MemberCard from './MemberCard'

interface TeamModalProps { open: boolean; onClose: () => void; onAdd: () => void; members: UserDTO[] }

export default function TeamModal({ open, onClose, onAdd, members }: TeamModalProps) {
  const [query, setQuery] = useState('')
  useEffect(() => {
    if (!open) return
    const previous = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => { document.body.style.overflow = previous }
  }, [open])
  const normalized = query.trim().toLocaleLowerCase('id-ID')
  const filtered = members.filter(member => `${member.name} ${member.email}`.toLocaleLowerCase('id-ID').includes(normalized))
  return <Modal open={open} onClose={onClose} title="Tim Bisnis" size="team" trapFocus>
    <p className="mb-4 text-sm text-body">{members.length} anggota</p>
    <Input id="team-search" label="Cari anggota" placeholder="Cari nama atau email" value={query} onChange={event => setQuery(event.target.value)} />
    <div role="region" aria-label="Hasil pencarian anggota" tabIndex={0} className="mt-6 min-h-0 flex-1 overflow-y-auto pr-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand">
      {filtered.length ? <div className="grid grid-cols-[repeat(auto-fill,minmax(min(260px,100%),1fr))] gap-4">{filtered.map(member => <MemberCard key={member.id} member={member} />)}</div> : <p role="status" className="py-8 text-center text-sm text-body">{members.length ? 'Tidak ada anggota yang cocok dengan pencarian.' : 'Belum ada anggota tim.'}</p>}
    </div>
    <footer className="mt-6 flex justify-end border-t border-slate-200 pt-4 dark:border-slate-700"><SubmitButton type="button" onClick={onAdd} className="text-sm">Tambah Anggota</SubmitButton></footer>
  </Modal>
}
