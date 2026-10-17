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
          {isLoading ? <FormSkeleton fields={['Nama Bisnis', 'Email Admin', 'Zona Waktu']} /> : <GlassPanel className="flex flex-col">
            <div>
            <SettingsRow title="Nama Bisnis"><span className="font-semibold">{name || '—'}</span></SettingsRow>
            <SettingsRow title="Email Admin">{auth?.user.email || '—'}</SettingsRow>
            <SettingsRow title="Zona Waktu"><div className="flex flex-wrap items-center gap-3 sm:justify-end"><span>{timezone}</span><span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200">{business?.status === 'active' ? 'Aktif' : business?.status === 'inactive' ? 'Tidak aktif' : business?.status === 'suspended' ? 'Ditangguhkan' : 'Belum aktif'}</span></div></SettingsRow>
            </div>
            <footer className="pt-6"><div className="flex justify-end border-t border-slate-200 pt-4 dark:border-slate-700"><SubmitButton type="button" onClick={() => { setError(null); setSuccess(null); setBusinessModalOpen(true) }} className="px-4 py-2 text-xs">Edit Identitas</SubmitButton></div></footer>
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
          <SettingsSection title="Langganan &amp; Tagihan" description="Langganan bisnis Anda untuk menggunakan PayLanggan." aside={<GlassPanel><h3 className="text-sm font-semibold text-ink">Perpanjangan langganan</h3><p className="mt-3 text-sm leading-6 text-body">Periksa status pembayaran otomatis dan riwayat tagihan di sini. Menghentikan perpanjangan tetap mempertahankan periode yang sudah dibayar hingga tanggal berakhir.</p></GlassPanel>}>
            <PlatformBillingSection />
          </SettingsSection>
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
