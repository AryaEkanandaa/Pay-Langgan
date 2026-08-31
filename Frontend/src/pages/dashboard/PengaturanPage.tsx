import { useEffect, useState, type FormEvent } from 'react'
import { ArrowRight, FileText } from 'lucide-react'
import { Link } from 'react-router-dom'
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
const TEAM_PAGE_SIZE = 4

const roleLabel: Record<UserDTO['role'], string> = {
  super_admin: 'Super Admin',
  admin: 'Admin',
  sales: 'Sales',
  finance: 'Finance',
}

export default function PengaturanPage() {
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
  const [teamPage, setTeamPage] = useState(1)

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
  }, [token])

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
      setTeamPage(Math.ceil((users.length + 1) / TEAM_PAGE_SIZE))
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

  const totalTeamPages = Math.max(1, Math.ceil(users.length / TEAM_PAGE_SIZE))
  const visibleUsers = users.slice((teamPage - 1) * TEAM_PAGE_SIZE, teamPage * TEAM_PAGE_SIZE)

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
    <div>
      <DashboardPageHeader
        title="Pengaturan Bisnis"
        description="Atur informasi dan preferensi bisnis Anda."
      />

      {error && !businessModalOpen && <p className="mt-4 text-[13px] text-red-600">{error}</p>}

      <div className="mt-6 grid w-full max-w-none gap-6 xl:grid-cols-2 xl:items-start">
        {isLoading ? (
          <FormSkeleton fields={['Nama Bisnis', 'Email Admin', 'Zona Waktu']} />
        ) : (
          <GlassPanel className="flex h-[390px] max-w-none flex-col">
            <div className="mb-5 flex items-start justify-between gap-4">
              <div>
                <p className="text-[14px] font-semibold text-ink">Identitas Bisnis</p>
                <p className="mt-1 text-[13px] text-muted">Informasi utama yang digunakan untuk bisnis Anda.</p>
              </div>
              <SubmitButton
                type="button"
                onClick={() => {
                  setError(null)
                  setSuccess(null)
                  setBusinessModalOpen(true)
                }}
                className="shrink-0 px-4 py-2 text-[12px]"
              >
                Edit Identitas
              </SubmitButton>
            </div>
            <div className="flex flex-1 flex-col justify-between">
              <div className="space-y-3">
                <div className="rounded-lg bg-white/40 px-4 py-3">
                  <p className="text-[11px] font-semibold uppercase tracking-wide text-muted">Nama Bisnis</p>
                  <p className="mt-1 truncate text-[14px] font-semibold text-ink">{name || '—'}</p>
                </div>
                <div className="rounded-lg bg-white/40 px-4 py-3">
                  <p className="text-[11px] font-semibold uppercase tracking-wide text-muted">Email Admin</p>
                  <p className="mt-1 truncate text-[14px] text-body">{auth?.user.email || '—'}</p>
                </div>
                <div className="flex items-center justify-between rounded-lg bg-white/40 px-4 py-3">
                  <div>
                    <p className="text-[11px] font-semibold uppercase tracking-wide text-muted">Zona Waktu</p>
                    <p className="mt-1 text-[14px] text-body">{timezone}</p>
                  </div>
                  <p className="text-[12px] text-muted">
                    Status: <span className="font-semibold text-emerald-700">{business?.status}</span>
                  </p>
                </div>
              </div>

              <div className="pt-4">
                {success && <p className="truncate text-[12px] text-emerald-700">{success}</p>}
              </div>
            </div>
          </GlassPanel>
        )}

        {auth?.user.role === 'admin' && !isLoading && (
          <GlassPanel className="flex h-[390px] max-w-none flex-col">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-[14px] font-semibold text-ink">Tim Bisnis</p>
                <p className="mt-1 text-[13px] text-muted">Anggota yang terhubung ke bisnis ini.</p>
              </div>
              <SubmitButton type="button" onClick={() => {
                setStaffError(null)
                setStaffSuccess(null)
                setStaffModalOpen(true)
              }} className="shrink-0 px-4 py-2 text-[12px]">
                + Tambah Anggota
              </SubmitButton>
            </div>

            <div className="mt-4 min-h-0 flex-1 space-y-2 overflow-hidden">
              {visibleUsers.map((user) => (
                <div key={user.id} className="flex h-14 items-center justify-between gap-3 rounded-lg bg-white/40 px-4">
                  <div className="min-w-0">
                    <p className="truncate text-[13px] font-semibold text-ink">{user.name}</p>
                    <p className="truncate text-[12px] text-muted">{user.email}</p>
                  </div>
                  <span className="shrink-0 rounded-md bg-brand/10 px-2.5 py-1 text-[11px] font-semibold text-brand">
                    {roleLabel[user.role]}
                  </span>
                </div>
              ))}
            </div>

            <div className="mt-4 flex min-h-8 items-center justify-between gap-3">
              <p className="text-[12px] text-muted">{users.length} anggota</p>
              {totalTeamPages > 1 && (
                <div className="flex items-center gap-1">
                  {Array.from({ length: totalTeamPages }, (_, index) => index + 1).map((pageNumber) => (
                    <button
                      key={pageNumber}
                      type="button"
                      onClick={() => setTeamPage(pageNumber)}
                      className={`h-7 min-w-7 rounded-md px-2 text-[12px] font-semibold transition-colors ${
                        pageNumber === teamPage ? 'bg-brand text-white' : 'text-muted hover:bg-brand/10 hover:text-brand'
                      }`}
                    >
                      {pageNumber}
                    </button>
                  ))}
                </div>
              )}
            </div>

            {staffSuccess && <p className="mt-2 truncate text-[12px] text-emerald-700">{staffSuccess}</p>}
          </GlassPanel>
        )}
      </div>

      {!isLoading && (
        <Link
          to="/dashboard/pengaturan/invoice"
          className="glass mt-6 flex min-h-[150px] w-full max-w-none items-center justify-between gap-6 rounded-2xl p-6 transition-all duration-300 hover:-translate-y-0.5 hover:bg-white/60 dark:hover:bg-white/10"
        >
          <div className="flex min-w-0 items-center gap-4">
            <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-brand/10 text-brand">
              <FileText size={21} />
            </span>
            <div className="min-w-0">
              <p className="text-[14px] font-semibold text-ink">Tampilan Invoice</p>
              <p className="mt-1 truncate text-[13px] text-muted">
                Atur logo, warna, alamat, dan catatan pada invoice bisnis Anda.
              </p>
            </div>
          </div>
          <div className="hidden shrink-0 items-center gap-5 sm:flex">
            <div className="flex items-center gap-2 rounded-lg border border-white/50 bg-white/35 px-3 py-2">
              <span className="h-3 w-3 rounded-sm bg-brand" />
              <span className="h-3 w-3 rounded-sm bg-amber" />
              <span className="h-3 w-3 rounded-sm bg-violet" />
              <span className="ml-1 text-[11px] font-medium text-muted">Preview identitas</span>
            </div>
            <span className="flex items-center gap-2 text-[13px] font-semibold text-brand">
              Kustomisasi <ArrowRight size={16} />
            </span>
          </div>
          <ArrowRight className="shrink-0 text-brand sm:hidden" size={18} />
        </Link>
      )}

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
