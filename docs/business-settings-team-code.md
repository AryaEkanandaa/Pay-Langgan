# Revisi Pengaturan Bisnis dan Tim

Kode lengkap komponen baru atau berubah dalam revisi ini. Tidak menambah dependency atau endpoint.

## Perubahan dan asumsi

- Konten tetap rata kiri. Pada >=1600px: label 280px, kolom utama maksimal 720px (dapat mengecil untuk menjaga kolom pendamping minimal 320px), gap 48px. Kartu utama dan pendamping sama tinggi.
- Pada 1100-1599px: label kiri dan kartu kanan; pendamping disembunyikan. Pada 900-1099px: label di atas, kartu utama dan pendamping dua kolom. Di bawah 900px satu kolom, pendamping disembunyikan.
- Langganan dan riwayat memakai lebar penuh dan dua kolom mulai 1100px. Tinggi riwayat mengikuti kartu langganan melalui ResizeObserver; mobile memiliki batas tinggi 480px. Isi riwayat bergulir secara internal.
- Tim memakai semua anggota dari API yang sama, menggantikan pagination lama dengan grid maksimal dua baris. Email/nama terpotong memiliki atribut title. Panah, keyboard, swipe dan trackpad mendukung geser horizontal.
- Modal tim menggunakan modal dasar, tidak memakai Fullscreen API. Lebar maksimal 1100px/92vw, tinggi maksimal 85dvh, pencarian lokal nama/email, fokus terkunci, body scroll dikunci, Esc/backdrop menutup, fokus kembali ke pemicu. Reduced motion mematikan animasi masuk dan scroll halus.
- Edit Identitas berada di footer kartu. Tambah Anggota tetap menggunakan form dan API yang sudah ada.

## API dan verifikasi

Tidak ada sambungan API baru yang diperlukan untuk fitur tim ini. Semua data anggota, perubahan identitas dan tagihan memakai fungsi API lama. TODO lama aktivasi ulang pembayaran otomatis tetap menunggu API yang tersedia.

TypeScript, lint dan build berhasil; build tetap mengeluarkan peringatan bundle besar. Browser dengan API tiruan memeriksa 0/1/2/4/7/20 anggota di 360/768/1440/1920/2560px, serta breakpoint tambahan 1024/1600px; dark mode; kesamaan tinggi; scroll internal; keyboard; reduced motion; pencarian/fokus/body lock modal; tambah anggota; redirect/pembatalan tagihan; salin dan pencabutan integrasi. Tidak melakukan transaksi gateway nyata.

## Frontend/src/pages/dashboard/PengaturanPage.tsx

```tsx
import { useEffect, useRef, useState, type FormEvent } from 'react'
import { ArrowRight, FileText } from 'lucide-react'
import { Link, useLocation } from 'react-router-dom'
import TeamMembersCard from '../../components/team/TeamMembersCard'
import PlatformBillingSection from '../../components/billing/PlatformBillingSection'
import SettingsSection from '../../components/dashboard/SettingsSection'
import SettingsRow from '../../components/dashboard/SettingsRow'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import FormSkeleton from '../../components/ui/skeletons/FormSkeleton'
import GlassPanel from '../../components/ui/GlassPanel'
import Input from '../../components/ui/Input'
import Modal from '../../components/ui/Modal'
import Select from '../../components/ui/Select'
import SubmitButton from '../../components/ui/SubmitButton'
import { useAuth } from '../../context/AuthContext'
import {
  ApiError,
  createStaffUser,
  getMyBusiness,
  listBusinessUsers,
  updateMyBusiness,
  type BusinessDTO,
  type UserDTO,
} from '../../lib/api'

const timezones = ['Asia/Makassar', 'Asia/Jakarta', 'Asia/Jayapura', 'UTC']
export default function PengaturanPage() {
  const location = useLocation()
  const billingRef = useRef<HTMLDivElement>(null)
  const { auth } = useAuth()
  const token = auth?.token
  const [business, setBusiness] = useState<BusinessDTO | null>(null)
  const [name, setName] = useState('')
  const [timezone, setTimezone] = useState('Asia/Makassar')
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [users, setUsers] = useState<UserDTO[]>([])
  const [staffName, setStaffName] = useState('')
  const [staffEmail, setStaffEmail] = useState('')
  const [staffPassword, setStaffPassword] = useState('')
  const [staffRole, setStaffRole] = useState<'sales' | 'finance'>('sales')
  const [staffError, setStaffError] = useState<string | null>(null)
  const [staffSuccess, setStaffSuccess] = useState<string | null>(null)
  const [isCreatingStaff, setIsCreatingStaff] = useState(false)
  const [staffModalOpen, setStaffModalOpen] = useState(false)
  const [businessModalOpen, setBusinessModalOpen] = useState(false)

  useEffect(() => {
    if (location.hash !== '#langganan-tagihan') return
    const frame = window.requestAnimationFrame(() => billingRef.current?.scrollIntoView({ block: 'start' }))
    return () => window.cancelAnimationFrame(frame)
  }, [location.hash, isLoading])


  useEffect(() => {
    if (!token) return
    setIsLoading(true)
    setError(null)
    Promise.all([getMyBusiness(token), auth?.user.role === 'admin' ? listBusinessUsers(token) : Promise.resolve([])])
      .then(([result, businessUsers]) => {
        setBusiness(result)
        setName(result.name)
        setUsers(businessUsers)
        const storedTimezone = result.meta?.timezone
        if (typeof storedTimezone === 'string' && storedTimezone) {
          setTimezone(storedTimezone)
        }
      })
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : 'Gagal memuat pengaturan bisnis.')
      })
      .finally(() => setIsLoading(false))
  }, [token, auth?.user.role])

  async function handleCreateStaff(event: FormEvent) {
    event.preventDefault()
    if (!token) return

    setIsCreatingStaff(true)
    setStaffError(null)
    setStaffSuccess(null)
    try {
      const user = await createStaffUser(token, {
        name: staffName,
        email: staffEmail,
        password: staffPassword,
        role: staffRole,
      })
      setUsers((current) => [...current, user])
      setStaffName('')
      setStaffEmail('')
      setStaffPassword('')
      setStaffSuccess('Akun staf berhasil dibuat dan dapat langsung digunakan untuk login.')
      setStaffModalOpen(false)
    } catch (err) {
      setStaffError(err instanceof ApiError ? err.message : 'Gagal membuat akun staf.')
    } finally {
      setIsCreatingStaff(false)
    }
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!token || !business) return

    const trimmedName = name.trim()
    if (!trimmedName) {
      setError('Nama bisnis wajib diisi.')
      return
    }

    setIsSaving(true)
    setError(null)
    setSuccess(null)
    try {
      const updated = await updateMyBusiness(token, {
        name: trimmedName,
        meta: { ...(business.meta ?? {}), timezone },
      })
      setBusiness(updated)
      setName(updated.name)
      setSuccess('Pengaturan bisnis berhasil disimpan.')
      setBusinessModalOpen(false)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal menyimpan pengaturan bisnis.')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className="business-settings-page w-full min-w-0">
      <DashboardPageHeader
        title="Pengaturan Bisnis"
        description="Atur informasi dan preferensi bisnis Anda."
      />

      {error && !businessModalOpen && <p className="mt-4 text-[13px] text-red-600">{error}</p>}

      <div className="mt-10 space-y-10 [&>.settings-section+*]:border-t [&>.settings-section+*]:border-slate-200 [&>.settings-section+*]:pt-10 dark:[&>.settings-section+*]:border-slate-700">
        <SettingsSection title="Identitas Bisnis" description="Informasi utama yang digunakan untuk bisnis Anda." aside={<GlassPanel><h3 className="text-sm font-semibold text-ink">Identitas bisnis Anda</h3><p className="mt-3 text-sm leading-6 text-body">Nama dan zona waktu digunakan untuk pengaturan bisnis. Email admin mengikuti akun yang sedang masuk.</p></GlassPanel>}>
          {isLoading ? <FormSkeleton fields={['Nama Bisnis', 'Email Admin', 'Zona Waktu']} /> : <GlassPanel className="flex h-full flex-col">
            <div>
            <SettingsRow title="Nama Bisnis"><span className="font-semibold">{name || '—'}</span></SettingsRow>
            <SettingsRow title="Email Admin">{auth?.user.email || '—'}</SettingsRow>
            <SettingsRow title="Zona Waktu"><div className="flex flex-wrap items-center gap-3 sm:justify-end"><span>{timezone}</span><span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200">{business?.status === 'active' ? 'Aktif' : business?.status === 'inactive' ? 'Tidak aktif' : business?.status === 'suspended' ? 'Ditangguhkan' : 'Belum aktif'}</span></div></SettingsRow>
            </div>
            <footer className="mt-auto pt-6"><div className="flex justify-end border-t border-slate-200 pt-4 dark:border-slate-700"><SubmitButton type="button" onClick={() => { setError(null); setSuccess(null); setBusinessModalOpen(true) }} className="px-4 py-2 text-xs">Edit Identitas</SubmitButton></div></footer>
            {success && <p role="status" className="mt-4 text-sm text-emerald-700 dark:text-emerald-200">{success}</p>}
          </GlassPanel>}
        </SettingsSection>

        {auth?.user.role === 'admin' && !isLoading && <SettingsSection title="Tim Bisnis" description="Anggota yang terhubung ke bisnis ini." aside={<GlassPanel><h3 className="text-sm font-semibold text-ink">Akses tim</h3><p className="mt-3 text-sm leading-6 text-body">Tambahkan anggota dengan peran Sales atau Finance sesuai tanggung jawabnya. Akun staf dapat digunakan untuk login setelah dibuat.</p></GlassPanel>}>
          <TeamMembersCard members={users} notice={staffSuccess} onAdd={() => { setStaffError(null); setStaffSuccess(null); setStaffModalOpen(true) }} />
        </SettingsSection>}

        {!isLoading && <SettingsSection title="Tampilan Invoice" description="Atur logo, warna, alamat, dan catatan pada invoice bisnis Anda." aside={<GlassPanel><h3 className="text-sm font-semibold text-ink">Identitas pada invoice</h3><p className="mt-3 text-sm leading-6 text-body">Sesuaikan tampilan invoice dengan identitas bisnis Anda melalui halaman kustomisasi.</p></GlassPanel>}>
          <Link to="/dashboard/pengaturan/invoice" className="glass flex w-full flex-wrap items-center justify-between gap-4 rounded-2xl px-6 py-4 transition-colors hover:bg-white/60 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand dark:hover:bg-white/10">
            <div className="flex min-w-0 flex-1 items-center gap-4"><span className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-brand/10 text-brand dark:text-blue-300"><FileText size={21} aria-hidden="true" /></span><div className="min-w-0"><p className="text-sm font-semibold text-ink">Tampilan Invoice</p><p className="mt-1 text-sm leading-6 text-body">Atur logo, warna, alamat, dan catatan pada invoice bisnis Anda.</p></div></div>
            <div className="flex items-center gap-4"><div aria-label="Preview identitas" className="hidden items-center gap-2 sm:flex"><span className="size-3 rounded-full bg-brand" /><span className="size-3 rounded-full bg-amber" /><span className="size-3 rounded-full bg-violet" /><span className="sr-only">Preview identitas</span></div><span className="hidden text-sm font-semibold text-brand dark:text-blue-300 sm:inline">Kustomisasi</span><ArrowRight className="text-brand dark:text-blue-300" size={18} aria-hidden="true" /></div>
          </Link>
        </SettingsSection>}

        <div ref={billingRef} id="langganan-tagihan" className="scroll-mt-6 border-t border-slate-200 pt-10 dark:border-slate-700">
          <h2 className="text-xl font-semibold text-ink">Langganan &amp; Tagihan</h2>
          <p className="mb-6 mt-2 text-sm text-body">Langganan bisnis Anda untuk menggunakan PayLanggan.</p>
          <PlatformBillingSection />
        </div>
      </div>

      <Modal open={businessModalOpen} onClose={() => setBusinessModalOpen(false)} title="Edit Identitas Bisnis">
        <form onSubmit={handleSubmit} className="space-y-5">
          <Input
            id="business_name"
            label="Nama Bisnis"
            required
            maxLength={100}
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
          <Input
            id="business_email"
            label="Email Admin"
            type="email"
            value={auth?.user.email ?? ''}
            readOnly
            className="cursor-not-allowed opacity-70"
          />
          <Select
            id="business_timezone"
            label="Zona Waktu"
            value={timezone}
            onChange={(event) => setTimezone(event.target.value)}
          >
            {timezones.map((value) => (
              <option key={value} value={value}>
                {value}
              </option>
            ))}
          </Select>

          {error && <p className="text-[13px] text-red-600">{error}</p>}
          <SubmitButton type="submit" className="w-full" disabled={isSaving}>
            {isSaving ? 'Menyimpan...' : 'Simpan Perubahan'}
          </SubmitButton>
        </form>
      </Modal>

      <Modal open={staffModalOpen} onClose={() => setStaffModalOpen(false)} title="Tambah Anggota Tim">
        <form onSubmit={handleCreateStaff} className="space-y-4">
          <Input
            id="staff_name"
            label="Nama Anggota"
            required
            value={staffName}
            onChange={(event) => setStaffName(event.target.value)}
          />
          <Input
            id="staff_email"
            label="Email Anggota"
            type="email"
            required
            value={staffEmail}
            onChange={(event) => setStaffEmail(event.target.value)}
          />
          <Input
            id="staff_password"
            label="Password Awal"
            type="password"
            minLength={6}
            required
            value={staffPassword}
            onChange={(event) => setStaffPassword(event.target.value)}
          />
          <Select
            id="staff_role"
            label="Role"
            value={staffRole}
            onChange={(event) => setStaffRole(event.target.value as 'sales' | 'finance')}
          >
            <option value="sales">Sales</option>
            <option value="finance">Finance</option>
          </Select>
          <SubmitButton type="submit" className="w-full" disabled={isCreatingStaff}>
            {isCreatingStaff ? 'Membuat...' : 'Buat Akun Anggota'}
          </SubmitButton>
          {staffError && <p className="text-[13px] text-red-600">{staffError}</p>}
        </form>
      </Modal>
    </div>
  )
}
```

## Frontend/src/components/dashboard/SettingsSection.tsx

```tsx
import { useId, type ReactNode } from 'react'

interface SettingsSectionProps {
  title: string
  description: string
  children: ReactNode
  aside?: ReactNode
}

export default function SettingsSection({ title, description, children, aside }: SettingsSectionProps) {
  const titleId = useId()
  return (
    <section aria-labelledby={titleId} className={`settings-section grid items-stretch gap-6 min-[900px]:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] min-[1100px]:grid-cols-[280px_minmax(0,1fr)] min-[1100px]:gap-8 min-[1600px]:gap-12 ${aside ? 'min-[1600px]:grid-cols-[280px_minmax(0,720px)_minmax(320px,1fr)]' : ''}`}>
      <header className="min-w-0 min-[900px]:max-[1099px]:col-span-2">
        <h2 id={titleId} className="text-base font-semibold text-ink">{title}</h2>
        <p className="mt-2 text-sm leading-6 text-body">{description}</p>
      </header>
      <div className="min-w-0 [&>.glass]:h-full [&>.glass]:shadow-sm">{children}</div>
      {aside && <aside aria-label={`Informasi ${title}`} className="hidden min-w-0 min-[900px]:max-[1099px]:block min-[1600px]:block [&>.glass]:flex [&>.glass]:h-full [&>.glass]:flex-col [&>.glass]:shadow-sm [&>.glass>p:last-child]:mt-auto [&>.glass>p:last-child]:pt-4">{aside}</aside>}
    </section>
  )
}
```

## Frontend/src/components/team/TeamMembersCard.tsx

```tsx
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
    <GlassPanel className="team-members-card flex h-full min-w-0 flex-col !shadow-sm">
      <header className="mb-6 flex items-start justify-between gap-4"><div><h3 className="text-base font-semibold text-ink">Tim Bisnis</h3><p className="mt-1 text-sm leading-6 text-body">Anggota yang terhubung ke bisnis ini.</p></div><button type="button" onClick={() => setExpanded(true)} aria-label="Layar besar Tim Bisnis" className="flex size-9 shrink-0 items-center justify-center rounded-full border border-slate-300 text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand dark:border-slate-600"><Maximize2 size={16} aria-hidden="true" /></button></header>
      <MemberGrid members={members} />
      {notice && <p role="status" className="mt-4 text-sm text-emerald-700 dark:text-emerald-200">{notice}</p>}
      <footer className="mt-auto pt-6"><div className="flex flex-wrap items-center justify-between gap-4 border-t border-slate-200 pt-4 dark:border-slate-700"><p className="text-sm text-body">{members.length} anggota</p><SubmitButton type="button" onClick={onAdd} className="px-4 py-2 text-xs">Tambah Anggota</SubmitButton></div></footer>
    </GlassPanel>
    {expanded && <TeamModal open onClose={() => setExpanded(false)} members={members} onAdd={() => { setExpanded(false); onAdd() }} />}
  </>
}
```

## Frontend/src/components/team/MemberGrid.tsx

```tsx
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
```

## Frontend/src/components/team/MemberCard.tsx

```tsx
import type { UserDTO } from '../../lib/api'

const roles: Record<UserDTO['role'], string> = { super_admin: 'Super Admin', admin: 'Admin', sales: 'Sales', finance: 'Finance' }

export default function MemberCard({ member }: { member: UserDTO }) {
  const initials = member.name.trim().split(/\s+/).slice(0, 2).map(part => part[0]).join('').toUpperCase()
  return <article className="min-w-0 rounded-2xl border border-slate-200 bg-white/40 p-4 dark:border-slate-700 dark:bg-white/5">
    <div className="flex items-center gap-3"><span aria-hidden="true" className="flex size-10 shrink-0 items-center justify-center rounded-full bg-brand/10 text-sm font-semibold text-brand dark:text-blue-300">{initials || '?'}</span><div className="min-w-0 flex-1"><p title={member.name} className="truncate text-sm font-semibold text-ink">{member.name}</p><p title={member.email} className="mt-1 truncate text-xs text-body">{member.email}</p></div></div>
    <span className="mt-3 inline-block rounded-full bg-brand/10 px-3 py-1 text-xs font-semibold text-brand dark:text-blue-300">{roles[member.role]}</span>
  </article>
}
```

## Frontend/src/components/team/TeamModal.tsx

```tsx
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
```

## Frontend/src/hooks/useHorizontalScroll.ts

```tsx
import { useCallback, useEffect, useRef, useState } from 'react'

export default function useHorizontalScroll(itemCount: number) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [position, setPosition] = useState({ left: false, right: false, progress: 0 })
  const update = useCallback(() => {
    const node = scrollRef.current
    if (!node) return
    const maximum = Math.max(0, node.scrollWidth - node.clientWidth)
    setPosition({ left: node.scrollLeft > 1, right: node.scrollLeft < maximum - 1, progress: maximum > 0 ? node.scrollLeft / maximum : 0 })
  }, [])
  useEffect(() => {
    const node = scrollRef.current
    if (!node) return
    node.addEventListener('scroll', update, { passive: true })
    const observer = new ResizeObserver(update)
    observer.observe(node)
    update()
    return () => { node.removeEventListener('scroll', update); observer.disconnect() }
  }, [itemCount, update])
  const scrollByPage = useCallback((direction: -1 | 1) => {
    const node = scrollRef.current
    if (!node) return
    node.scrollBy({ left: direction * node.clientWidth, behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth' })
  }, [])
  return { scrollRef, ...position, scrollByPage }
}
```

## Frontend/src/components/ui/Modal.tsx

```tsx
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
```

## Frontend/src/components/billing/PlatformBillingSection.tsx

```tsx
import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { XenditComponents, type XenditSubmissionEndEvent } from 'xendit-components-web'
import SubscriptionCard from './SubscriptionCard'
import InvoiceHistory from './InvoiceHistory'
import ConfirmDialog from '../ui/ConfirmDialog'
import { useAuth } from '../../context/AuthContext'
import {
  ApiError, cancelPlatformSubscription, getPlatformInvoices, getPlatformPlan,
  getPlatformSubscription, resumePlatformCheckout, startPlatformCheckout,
  type PlatformCheckout, type PlatformInvoice, type PlatformPlan, type PlatformSubscription,
} from '../../lib/api'
import { formatCurrency } from '../../lib/format'

export default function PlatformBillingSection() {
  const [searchParams] = useSearchParams()
  const returningFromXendit = searchParams.has('token_request_id')
  const { auth } = useAuth()
  const token = auth?.token
  const [plan, setPlan] = useState<PlatformPlan | null>(null)
  const [subscription, setSubscription] = useState<PlatformSubscription | null>(null)
  const [invoices, setInvoices] = useState<PlatformInvoice[]>([])
  const [consent, setConsent] = useState(false)
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [cancelOpen, setCancelOpen] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const [session, setSession] = useState<PlatformCheckout | null>(null)
  const [cardReady, setCardReady] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [verifying, setVerifying] = useState(returningFromXendit)
  const cardContainer = useRef<HTMLDivElement>(null)
  const componentsRef = useRef<XenditComponents | null>(null)
  const resumed = useRef(false)
  const subscriptionCardRef = useRef<HTMLDivElement>(null)
  const [historyHeight, setHistoryHeight] = useState<number>()
  useLayoutEffect(() => {
    const node = subscriptionCardRef.current
    if (!node || loading) return
    const update = () => setHistoryHeight(window.matchMedia('(min-width: 1100px)').matches ? node.getBoundingClientRect().height : undefined)
    const observer = new ResizeObserver(update)
    observer.observe(node)
    window.addEventListener('resize', update)
    update()
    return () => { observer.disconnect(); window.removeEventListener('resize', update) }
  }, [loading])

  const refresh = useCallback(async () => {
    if (!token) return
    try {
      const [p, s, i] = await Promise.all([
        getPlatformPlan(token), getPlatformSubscription(token), getPlatformInvoices(token),
      ])
      setPlan(p); setSubscription(s); setInvoices(i)
      setError(null)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memuat tagihan PayLanggan.')
    } finally { setLoading(false) }
  }, [token])

  useEffect(() => { void refresh() }, [refresh])

  useEffect(() => {
    if (!token || !subscription?.checkout_available || resumed.current) return
    resumed.current = true
    void resumePlatformCheckout(token).then((value) => {
      if (value.mode === 'COMPONENTS' && value.components_sdk_key) {
        setSession(value)
        setVerifying(false)
      }
    }).catch(() => { setVerifying(true) })
  }, [token, subscription?.checkout_available])

  useEffect(() => {
    if (!token || (!session && !verifying)) return
    const timer = window.setInterval(() => {
      void Promise.all([getPlatformSubscription(token), getPlatformInvoices(token)]).then(([next, items]) => {
        setSubscription(next)
        setInvoices(items)
        if (next.status === 'active' && !next.pending_invoice) {
          setSession(null)
          setVerifying(false)
          setNotice('Pembayaran berhasil. Langganan PayLanggan aktif.')
        } else if (verifying && !next.checkout_available) {
          setVerifying(false)
          setSession(null)
        }
      }).catch(() => undefined)
    }, 3000)
    return () => window.clearInterval(timer)
  }, [token, session, verifying])

  useEffect(() => {
    if (!session || verifying || !cardContainer.current) return
    const target = cardContainer.current
    const components = new XenditComponents({ componentsSdkKey: session.components_sdk_key, resume: returningFromXendit })
    componentsRef.current = components
    let cardElement: HTMLElement | null = null
    const onInit = () => {
      const card = components.getActiveChannels({ filter: 'CARDS' })[0]
      if (!card) { setError('Pembayaran kartu sedang tidak tersedia.'); return }
      cardElement = components.createChannelComponent(card)
      target.replaceChildren(cardElement)
    }
    const onReady = () => setCardReady(true)
    const onNotReady = () => setCardReady(false)
    const onBegin = () => { setSubmitting(true); setError(null) }
    const onEnd = (event: XenditSubmissionEndEvent) => {
      setSubmitting(false)
      if (event.developerErrorMessage?.type === 'NETWORK_ERROR') {
        setError('Status pembayaran belum dapat dipastikan. Periksa kembali status tagihan sebelum mencoba lagi.')
      } else if (event.developerErrorMessage || event.userErrorMessage?.length) {
        setError('Pembayaran belum berhasil. Periksa data kartu dan coba lagi.')
      }
    }
    const onComplete = () => { setSubmitting(false); setVerifying(true); void refresh() }
    const onExpired = () => { setSubmitting(false); setSession(null); setVerifying(true); void refresh() }
    const onFatal = () => { setSubmitting(false); setError('Formulir kartu belum dapat ditampilkan. Muat ulang halaman dan coba lagi.') }
    components.addEventListener('init', onInit)
    components.addEventListener('submission-ready', onReady)
    components.addEventListener('submission-not-ready', onNotReady)
    components.addEventListener('submission-begin', onBegin)
    components.addEventListener('submission-end', onEnd)
    components.addEventListener('session-complete', onComplete)
    components.addEventListener('session-expired-or-canceled', onExpired)
    components.addEventListener('fatal-error', onFatal)
    return () => {
      components.removeEventListener('init', onInit)
      components.removeEventListener('submission-ready', onReady)
      components.removeEventListener('submission-not-ready', onNotReady)
      components.removeEventListener('submission-begin', onBegin)
      components.removeEventListener('submission-end', onEnd)
      components.removeEventListener('session-complete', onComplete)
      components.removeEventListener('session-expired-or-canceled', onExpired)
      components.removeEventListener('fatal-error', onFatal)
      if (componentsRef.current === components) componentsRef.current = null
      if (cardElement) components.destroyComponent(cardElement)
      target.replaceChildren()
    }
  }, [session, returningFromXendit, verifying, refresh])

  async function checkout() {
    if (!token || !consent) return
    setBusy(true); setError(null)
    try {
      const result = await startPlatformCheckout(token)
      if (result.mode !== 'COMPONENTS' || !result.components_sdk_key) throw new Error('Formulir pembayaran belum tersedia.')
      resumed.current = true
      setSession(result)
      setVerifying(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal membuka checkout.')
    } finally { setBusy(false) }
  }

  async function resumeCheckout() {
    if (!token) return
    setBusy(true); setError(null)
    try {
      const result = await resumePlatformCheckout(token)
      resumed.current = true
      setSession(result)
      setVerifying(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Sesi pembayaran belum dapat dilanjutkan.')
      await refresh()
    } finally { setBusy(false) }
  }

  function pay() {
    if (!componentsRef.current || !cardReady || submitting) return
    try { componentsRef.current.submit() }
    catch { setError('Periksa kembali data kartu Anda.') }
  }

  async function cancel() {
    if (!token || !subscription?.auto_renew) return
    setBusy(true); setError(null)
    try {
      await cancelPlatformSubscription(token)
      setCancelOpen(false)
      setNotice('Perpanjangan otomatis telah dihentikan.')
      await refresh()
    } catch (err) { setError(err instanceof ApiError ? err.message : 'Gagal membatalkan langganan.') }
    finally { setBusy(false) }
  }

  if (loading) return <p className="p-6">Memuat tagihan PayLanggan...</p>
  return <div className="space-y-6">
    {error && <p role="alert" className="rounded-lg bg-red-50 p-4 text-red-700 dark:bg-red-950 dark:text-red-200">{error}</p>}
    {notice && <p role="status" className="rounded-lg bg-green-50 p-4 text-green-700 dark:bg-emerald-950 dark:text-emerald-200">{notice}</p>}
    <div className="grid items-stretch gap-6 min-[1100px]:grid-cols-2">
    <div ref={subscriptionCardRef} className="min-w-0 self-start [&>.glass]:h-full [&>.glass]:shadow-sm">
    <SubscriptionCard plan={plan} subscription={subscription} busy={busy} onCancel={() => setCancelOpen(true)}>
      {subscription?.legacy_checkout_active && <p className="mt-4 text-sm text-amber-800 dark:text-amber-200">Sesi pembayaran lama masih aktif. Tunggu sesi berakhir sebelum membuat pembayaran baru di halaman ini.</p>}
      {subscription?.checkout_available && !session && !verifying && <button type="button" disabled={busy} onClick={resumeCheckout} className="mt-4 rounded-lg bg-blue-700 px-5 py-2 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand disabled:opacity-50">Lanjutkan pembayaran</button>}
      {verifying && <p role="status" className="mt-4 text-sm text-blue-700 dark:text-blue-200">Kami sedang memastikan pembayaran Anda. Status tagihan akan diperbarui otomatis.</p>}
      {session && !verifying && <div className="mt-5 max-w-xl space-y-4">
        <p className="text-sm text-body">Masukkan detail kartu pada formulir pembayaran aman berikut.</p>
        <div ref={cardContainer} className="checkout-card-fields" aria-label="Formulir pembayaran kartu" />
        <button type="button" disabled={!cardReady || submitting} onClick={pay} className="rounded-lg bg-blue-700 px-5 py-2 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand disabled:opacity-50">{submitting ? 'Memproses pembayaran...' : `Bayar ${formatCurrency(session.amount)}`}</button>
      </div>}
      {(subscription?.status === 'not_subscribed' || subscription?.status === 'pending' || subscription?.status === 'past_due') && !subscription.checkout_available && !subscription.legacy_checkout_active && !session && !verifying && <div className="mt-5 space-y-3">
        <label className="flex items-start gap-3 text-sm text-body"><input type="checkbox" checked={consent} onChange={e => setConsent(e.target.checked)} className="mt-1 accent-brand focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand" />
          <span>Saya menyetujui pembayaran {formatCurrency(plan?.amount ?? 300000)} sekarang dan pendebitan kartu otomatis sebesar {formatCurrency(plan?.amount ?? 300000)} setiap bulan selama langganan aktif. Saya dapat membatalkan perpanjangan otomatis melalui halaman ini.</span>
        </label>
        <button type="button" disabled={!consent || busy} onClick={checkout} className="rounded-lg bg-blue-700 px-5 py-2 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand disabled:opacity-50">{busy ? 'Memproses...' : 'Lanjut ke pembayaran'}</button>
      </div>}
    </SubscriptionCard>
    </div>
    <InvoiceHistory height={historyHeight} invoices={invoices} resumableInvoiceId={subscription?.checkout_available ? subscription.pending_invoice?.id : undefined} busy={busy} onResume={() => { void resumeCheckout() }} />
    </div>
    <ConfirmDialog open={cancelOpen} onClose={() => setCancelOpen(false)} onConfirm={() => { void cancel() }} busy={busy} error={error} title="Hentikan perpanjangan otomatis?" description="Periode yang sudah dibayar tetap tercatat hingga tanggal berakhir." confirmLabel="Hentikan perpanjangan otomatis" />
  </div>
}
```

## Frontend/src/components/billing/InvoiceHistory.tsx

```tsx
import type { PlatformInvoice } from '../../lib/api'
import { formatCurrency } from '../../lib/format'
import GlassPanel from '../ui/GlassPanel'
import { billingDate } from './format'

interface InvoiceHistoryProps {
  height?: number
  invoices: PlatformInvoice[]
  resumableInvoiceId?: number
  busy: boolean
  onResume: () => void
}

function Status({ status }: { status: PlatformInvoice['status'] }) {
  return <span className={`inline-block rounded-full px-3 py-1 text-xs font-semibold ${status === 'paid' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200' : 'bg-amber-100 text-amber-900 dark:bg-amber-950 dark:text-amber-200'}`}>{status === 'paid' ? 'Lunas' : 'Menunggu pembayaran'}</span>
}

export default function InvoiceHistory({ height, invoices, resumableInvoiceId, busy, onResume }: InvoiceHistoryProps) {
  function action(invoice: PlatformInvoice) {
    return invoice.status === 'pending' && invoice.id === resumableInvoiceId
      ? <button type="button" disabled={busy} onClick={onResume} aria-label={`Lanjutkan pembayaran invoice ${invoice.id}`} className="rounded-lg px-2 py-2 font-semibold text-brand focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand disabled:opacity-50">Lanjutkan pembayaran</button>
      : <span className="text-body">—</span>
  }
  return (
    <GlassPanel className="flex min-h-0 min-w-0 flex-col !shadow-sm" style={{ height, maxHeight: height ?? 480 }}>
      <h3 className="shrink-0 text-base font-semibold text-ink">Riwayat tagihan</h3>
      {invoices.length === 0 ? <div className="py-8 text-center"><p className="font-semibold text-ink">Belum ada tagihan</p><p className="mt-2 text-sm text-body">Tagihan langganan PayLanggan akan ditampilkan di sini setelah dibuat.</p></div> : <div tabIndex={0} role="region" aria-label="Riwayat tagihan langganan" className="mt-4 min-h-0 flex-1 overflow-y-auto focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand">
        <div className="hidden overflow-x-auto md:block">
          <table className="w-full text-left text-sm">
            <caption className="sr-only">Riwayat tagihan langganan PayLanggan</caption>
            <thead className="text-body"><tr>{['Tanggal', 'Nomor invoice', 'Jumlah', 'Status', 'Aksi'].map(label => <th key={label} scope="col" className="px-3 py-3 font-semibold">{label}</th>)}</tr></thead>
            <tbody>{invoices.map(invoice => <tr key={invoice.id} className="border-t border-slate-200 align-top dark:border-slate-700">
              <td className="px-3 py-4 text-body">{billingDate(invoice.created_at)}<p className="mt-2 text-xs">Periode: {billingDate(invoice.period_start)} – {billingDate(invoice.period_end)}</p><p className="mt-2 text-xs">Dibayar: {billingDate(invoice.paid_at)}</p></td>
              <td className="px-3 py-4 font-mono text-ink">#{invoice.id}</td>
              <td className="whitespace-nowrap px-3 py-4 font-semibold text-ink">{formatCurrency(invoice.amount)}</td>
              <td className="px-3 py-4"><Status status={invoice.status} /></td><td className="px-3 py-4">{action(invoice)}</td>
            </tr>)}</tbody>
          </table>
        </div>
        <ul className="space-y-4 md:hidden">{invoices.map(invoice => <li key={invoice.id} className="rounded-xl border border-slate-200 p-4 dark:border-slate-700">
          <div className="flex flex-wrap items-center justify-between gap-2"><span className="font-mono font-semibold text-ink">#{invoice.id}</span><Status status={invoice.status} /></div>
          <p className="mt-4 font-semibold text-ink">{formatCurrency(invoice.amount)}</p>
          <dl className="mt-3 space-y-2 text-sm text-body"><div><dt className="inline">Tanggal: </dt><dd className="inline">{billingDate(invoice.created_at)}</dd></div><div><dt className="inline">Periode: </dt><dd className="inline">{billingDate(invoice.period_start)} – {billingDate(invoice.period_end)}</dd></div><div><dt className="inline">Dibayar: </dt><dd className="inline">{billingDate(invoice.paid_at)}</dd></div></dl>
          {invoice.id === resumableInvoiceId && <div className="mt-3">{action(invoice)}</div>}
        </li>)}</ul>
      </div>}
    </GlassPanel>
  )
}
```

## Tambahan Frontend/src/index.css

```css
.team-scroll { scrollbar-width: none; scroll-snap-type: x proximity; }
.team-scroll::-webkit-scrollbar { display: none; }
@keyframes team-dialog-enter { from { opacity: 0; transform: scale(0.98); } to { opacity: 1; transform: scale(1); } }
.team-modal { animation: team-dialog-enter 180ms ease-out; }
@media (prefers-reduced-motion: reduce) { .team-modal { animation: none; } }
```
