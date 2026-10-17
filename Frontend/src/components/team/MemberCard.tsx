import type { UserDTO } from '../../lib/api'

const roles: Record<UserDTO['role'], string> = { super_admin: 'Super Admin', admin: 'Admin', sales: 'Sales', finance: 'Finance' }

export default function MemberCard({ member }: { member: UserDTO }) {
  const initials = member.name.trim().split(/\s+/).slice(0, 2).map(part => part[0]).join('').toUpperCase()
  return <article className="min-w-0 rounded-2xl border border-slate-200 bg-white/40 p-4 dark:border-slate-700 dark:bg-white/5">
    <div className="flex items-center gap-3"><span aria-hidden="true" className="flex size-10 shrink-0 items-center justify-center rounded-full bg-brand/10 text-sm font-semibold text-brand dark:text-blue-300">{initials || '?'}</span><div className="min-w-0 flex-1"><p title={member.name} className="truncate text-sm font-semibold text-ink">{member.name}</p><p title={member.email} className="mt-1 truncate text-xs text-body">{member.email}</p></div></div>
    <span className="mt-3 inline-block rounded-full bg-brand/10 px-3 py-1 text-xs font-semibold text-brand dark:text-blue-300">{roles[member.role]}</span>
  </article>
}
