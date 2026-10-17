# Penggabungan langganan dan tagihan bisnis

Kode lengkap komponen baru/berubah, routing, dan sidebar untuk perubahan ini.

Halaman PlatformBillingPage.tsx yang lama dihapus. Route tagihan-platform dan tagihan mengarah ke pengaturan#langganan-tagihan, dengan query pembayaran tetap dipertahankan.

Semua endpoint pembayaran tetap memakai API yang tersedia. API invoice platform menyediakan id numerik; tabel menampilkan #id. Aktivasi ulang pembayaran otomatis belum tersedia dan ditandai TODO di SubscriptionCard.tsx.

Pemeriksaan: TypeScript, lint, build, browser pada 360/768/1440/2560px, dark mode, redirect, consent, fokus dialog, dan pembatalan melalui API tiruan. Transaksi gateway nyata tidak dilakukan saat pengujian ini.

## Frontend/src/pages/dashboard/PengaturanPage.tsx

```tsx
import { useEffect, useRef, useState, type FormEvent } from 'react'
import { ArrowRight, FileText } from 'lucide-react'
import { Link, useLocation } from 'react-router-dom'
import PlatformBillingSection from '../../components/billing/PlatformBillingSection'
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
  const location = useLocation()
  const billingRef = useRef<HTMLElement>(null)
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
    <div className="business-settings-page mx-auto w-full max-w-[1480px]">
      <DashboardPageHeader
        title="Pengaturan Bisnis"
        description="Atur informasi dan preferensi bisnis Anda."
      />

      {error && !businessModalOpen && <p className="mt-4 text-[13px] text-red-600">{error}</p>}

      <div className="mt-6 grid w-full max-w-none gap-6 items-start min-[1100px]:grid-cols-2">
        {isLoading ? (
          <FormSkeleton fields={['Nama Bisnis', 'Email Admin', 'Zona Waktu']} />
        ) : (
          <GlassPanel className="flex min-w-0 max-w-none flex-col">
            <div className="mb-5 flex flex-wrap items-start justify-between gap-4">
              <div>
                <p className="text-[14px] font-semibold text-ink">Identitas Bisnis</p>
                <p className="mt-1 text-[13px] text-body">Informasi utama yang digunakan untuk bisnis Anda.</p>
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
            <div className="flex flex-col">
              <div className="space-y-3">
                <div className="rounded-lg bg-white/40 px-4 py-3">
                  <p className="text-[11px] font-semibold uppercase tracking-wide text-body">Nama Bisnis</p>
                  <p className="mt-1 truncate text-[14px] font-semibold text-ink">{name || '—'}</p>
                </div>
                <div className="rounded-lg bg-white/40 px-4 py-3">
                  <p className="text-[11px] font-semibold uppercase tracking-wide text-body">Email Admin</p>
                  <p className="mt-1 truncate text-[14px] text-body">{auth?.user.email || '—'}</p>
                </div>
                <div className="flex items-center justify-between rounded-lg bg-white/40 px-4 py-3">
                  <div>
                    <p className="text-[11px] font-semibold uppercase tracking-wide text-body">Zona Waktu</p>
                    <p className="mt-1 text-[14px] text-body">{timezone}</p>
                  </div>
                  <p className="text-[12px] text-body">
                    <span className="rounded-full bg-emerald-100 px-3 py-1 font-semibold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200">{business?.status === 'active' ? 'Aktif' : business?.status === 'inactive' ? 'Tidak aktif' : business?.status === 'suspended' ? 'Ditangguhkan' : 'Belum aktif'}</span>
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
          <GlassPanel className="flex min-w-0 max-w-none flex-col">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div>
                <p className="text-[14px] font-semibold text-ink">Tim Bisnis</p>
                <p className="mt-1 text-[13px] text-body">Anggota yang terhubung ke bisnis ini.</p>
              </div>
              <SubmitButton type="button" onClick={() => {
                setStaffError(null)
                setStaffSuccess(null)
                setStaffModalOpen(true)
              }} className="shrink-0 px-4 py-2 text-[12px]">
                + Tambah Anggota
              </SubmitButton>
            </div>

            <div className="mt-4 space-y-2">
              {visibleUsers.map((user) => (
                <div key={user.id} className="flex h-14 items-center justify-between gap-3 rounded-lg bg-white/40 px-4">
                  <div className="min-w-0">
                    <p className="truncate text-[13px] font-semibold text-ink">{user.name}</p>
                    <p className="truncate text-[12px] text-body">{user.email}</p>
                  </div>
                  <span className="shrink-0 rounded-md bg-brand/10 px-2.5 py-1 text-[11px] font-semibold text-brand">
                    {roleLabel[user.role]}
                  </span>
                </div>
              ))}
            </div>

            <div className="mt-4 flex min-h-8 items-center justify-between gap-3">
              <p className="text-[12px] text-body">{users.length} anggota</p>
              {totalTeamPages > 1 && (
                <div className="flex items-center gap-1">
                  {Array.from({ length: totalTeamPages }, (_, index) => index + 1).map((pageNumber) => (
                    <button
                      key={pageNumber}
                      type="button"
                      onClick={() => setTeamPage(pageNumber)}
                      className={`h-7 min-w-7 rounded-md px-2 text-[12px] font-semibold transition-colors ${
                        pageNumber === teamPage ? 'bg-brand text-white' : 'text-body hover:bg-brand/10 hover:text-brand'
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
          className="glass mt-6 flex w-full max-w-none items-center justify-between gap-6 rounded-2xl px-6 py-4 transition-all duration-300 hover:-translate-y-0.5 hover:bg-white/60 dark:hover:bg-white/10"
        >
          <div className="flex min-w-0 items-center gap-4">
            <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-brand/10 text-brand">
              <FileText size={21} />
            </span>
            <div className="min-w-0">
              <p className="text-[14px] font-semibold text-ink">Tampilan Invoice</p>
              <p className="mt-1 text-[13px] text-body">
                Atur logo, warna, alamat, dan catatan pada invoice bisnis Anda.
              </p>
            </div>
          </div>
          <div className="hidden shrink-0 items-center gap-5 min-[1100px]:flex">
            <div className="flex items-center gap-2 rounded-lg border border-white/50 bg-white/35 px-3 py-2">
              <span className="h-3 w-3 rounded-sm bg-brand" />
              <span className="h-3 w-3 rounded-sm bg-amber" />
              <span className="h-3 w-3 rounded-sm bg-violet" />
              <span className="ml-1 text-[11px] font-medium text-body">Preview identitas</span>
            </div>
            <span className="flex items-center gap-2 text-[13px] font-semibold text-brand">
              Kustomisasi <ArrowRight size={16} />
            </span>
          </div>
          <ArrowRight className="shrink-0 text-brand min-[1100px]:hidden" size={18} />
        </Link>
      )}

      <section ref={billingRef} id="langganan-tagihan" aria-labelledby="billing-heading" className="scroll-mt-6">
        <hr className="my-10 border-t border-slate-200 dark:border-slate-700" />
        <h2 id="billing-heading" className="text-xl font-bold text-ink">Langganan &amp; Tagihan</h2>
        <p className="mb-6 mt-2 text-sm text-body">Langganan bisnis Anda untuk menggunakan PayLanggan.</p>
        <PlatformBillingSection />
      </section>

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

## Frontend/src/components/billing/PlatformBillingSection.tsx

```tsx
import { useCallback, useEffect, useRef, useState } from 'react'
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
    <div className="grid items-start gap-6 min-[1100px]:grid-cols-2">
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
    <InvoiceHistory invoices={invoices} resumableInvoiceId={subscription?.checkout_available ? subscription.pending_invoice?.id : undefined} busy={busy} onResume={() => { void resumeCheckout() }} />
    </div>
    <ConfirmDialog open={cancelOpen} onClose={() => setCancelOpen(false)} onConfirm={() => { void cancel() }} busy={busy} error={error} title="Hentikan perpanjangan otomatis?" description="Periode yang sudah dibayar tetap tercatat hingga tanggal berakhir." confirmLabel="Hentikan perpanjangan otomatis" />
  </div>
}
```

## Frontend/src/components/billing/SubscriptionCard.tsx

```tsx
import type { ReactNode } from 'react'
import type { PlatformPlan, PlatformSubscription } from '../../lib/api'
import { formatCurrency } from '../../lib/format'
import GlassPanel from '../ui/GlassPanel'
import { billingDate } from './format'

const statuses: Record<PlatformSubscription['status'], string> = {
  not_subscribed: 'Belum berlangganan', pending: 'Menunggu pembayaran', active: 'Aktif',
  past_due: 'Pembayaran belum selesai', cancelled: 'Perpanjangan dihentikan',
}

interface SubscriptionCardProps {
  plan: PlatformPlan | null
  subscription: PlatformSubscription | null
  busy: boolean
  onCancel: () => void
  children?: ReactNode
}

export default function SubscriptionCard({ plan, subscription, busy, onCancel, children }: SubscriptionCardProps) {
  const paidThrough = billingDate(subscription?.paid_through)
  const nextBilling = billingDate(subscription?.next_billing_date)
  return (
    <GlassPanel className="min-w-0">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h3 className="text-base font-semibold text-ink">{plan?.name ?? 'PayLanggan Bulanan'}</h3>
        <span className={`rounded-full px-3 py-1 text-xs font-semibold ${subscription?.status === 'active' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200' : 'bg-amber-100 text-amber-900 dark:bg-amber-950 dark:text-amber-200'}`}>{subscription ? statuses[subscription.status] : '—'}</span>
      </div>
      <p className="mt-4 text-3xl font-bold text-ink">{formatCurrency(plan?.amount ?? 300000)} <span className="text-sm font-normal text-body">/ bulan</span></p>
      <dl className="mt-6 grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)] gap-x-4 gap-y-4 text-sm">
        {subscription?.paid_through && <><dt className="text-body">Berlaku sampai</dt><dd className="font-medium text-ink">{paidThrough}</dd></>}
        {subscription?.next_billing_date && nextBilling !== paidThrough && <><dt className="text-body">Periode tagihan berikutnya</dt><dd className="font-medium text-ink">{nextBilling}</dd></>}
        <dt className="text-body">Pembayaran otomatis</dt>
        <dd><span className={`inline-block rounded-full px-3 py-1 text-xs font-semibold ${subscription?.auto_renew ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200' : 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-200'}`}>{subscription?.auto_renew ? 'Aktif' : 'Belum aktif'}</span></dd>
      </dl>
      {!subscription?.auto_renew && <p className="mt-4 text-sm text-body">Pembayaran otomatis belum aktif</p>}
      {/* TODO: Render an activation action when an API for reactivating auto-renew exists. */}
      {children}
      {subscription?.auto_renew && <button type="button" disabled={busy} onClick={onCancel} className="mt-6 rounded-xl border border-red-300 px-4 py-3 text-sm font-semibold text-red-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-red-600 disabled:opacity-50 dark:border-red-800 dark:text-red-300">Hentikan perpanjangan otomatis</button>}
    </GlassPanel>
  )
}
```

## Frontend/src/components/billing/InvoiceHistory.tsx

```tsx
import type { PlatformInvoice } from '../../lib/api'
import { formatCurrency } from '../../lib/format'
import GlassPanel from '../ui/GlassPanel'
import { billingDate } from './format'

interface InvoiceHistoryProps {
  invoices: PlatformInvoice[]
  resumableInvoiceId?: number
  busy: boolean
  onResume: () => void
}

function Status({ status }: { status: PlatformInvoice['status'] }) {
  return <span className={`inline-block rounded-full px-3 py-1 text-xs font-semibold ${status === 'paid' ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200' : 'bg-amber-100 text-amber-900 dark:bg-amber-950 dark:text-amber-200'}`}>{status === 'paid' ? 'Lunas' : 'Menunggu pembayaran'}</span>
}

export default function InvoiceHistory({ invoices, resumableInvoiceId, busy, onResume }: InvoiceHistoryProps) {
  function action(invoice: PlatformInvoice) {
    return invoice.status === 'pending' && invoice.id === resumableInvoiceId
      ? <button type="button" disabled={busy} onClick={onResume} aria-label={`Lanjutkan pembayaran invoice ${invoice.id}`} className="rounded-lg px-2 py-2 font-semibold text-brand focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand disabled:opacity-50">Lanjutkan pembayaran</button>
      : <span className="text-body">—</span>
  }
  return (
    <GlassPanel className="min-w-0">
      <h3 className="text-base font-semibold text-ink">Riwayat tagihan</h3>
      {invoices.length === 0 ? <div className="py-8 text-center"><p className="font-semibold text-ink">Belum ada tagihan</p><p className="mt-2 text-sm text-body">Tagihan langganan PayLanggan akan ditampilkan di sini setelah dibuat.</p></div> : <>
        <div className="mt-4 hidden overflow-x-auto md:block">
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
        <ul className="mt-4 space-y-4 md:hidden">{invoices.map(invoice => <li key={invoice.id} className="rounded-xl border border-slate-200 p-4 dark:border-slate-700">
          <div className="flex flex-wrap items-center justify-between gap-2"><span className="font-mono font-semibold text-ink">#{invoice.id}</span><Status status={invoice.status} /></div>
          <p className="mt-4 font-semibold text-ink">{formatCurrency(invoice.amount)}</p>
          <dl className="mt-3 space-y-2 text-sm text-body"><div><dt className="inline">Tanggal: </dt><dd className="inline">{billingDate(invoice.created_at)}</dd></div><div><dt className="inline">Periode: </dt><dd className="inline">{billingDate(invoice.period_start)} – {billingDate(invoice.period_end)}</dd></div><div><dt className="inline">Dibayar: </dt><dd className="inline">{billingDate(invoice.paid_at)}</dd></div></dl>
          {invoice.id === resumableInvoiceId && <div className="mt-3">{action(invoice)}</div>}
        </li>)}</ul>
      </>}
    </GlassPanel>
  )
}
```

## Frontend/src/components/billing/format.ts

```tsx
export function billingDate(value: string | null | undefined) {
  return value ? new Date(value).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) : '—'
}
```

## Frontend/src/components/ui/ConfirmDialog.tsx

```tsx
import { useRef } from 'react'
import Modal from './Modal'
import SubmitButton from './SubmitButton'

interface ConfirmDialogProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  description: string
  confirmLabel: string
  busy?: boolean
  error?: string | null
}

export default function ConfirmDialog({ open, onClose, onConfirm, title, description, confirmLabel, busy = false, error }: ConfirmDialogProps) {
  const cancelRef = useRef<HTMLButtonElement>(null)
  return (
    <Modal open={open} onClose={() => { if (!busy) onClose() }} title={title} trapFocus initialFocusRef={cancelRef}>
      <p className="mt-4 text-sm text-body">{description}</p>
      {error && <p role="alert" className="mt-4 text-sm text-red-700 dark:text-red-300">{error}</p>}
      <div className="mt-6 flex flex-wrap justify-end gap-3">
        <button ref={cancelRef} type="button" disabled={busy} onClick={onClose} className="rounded-xl border border-slate-300 px-4 py-2 text-sm font-semibold text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand disabled:opacity-50 dark:border-slate-600">Batal</button>
        <SubmitButton type="button" disabled={busy} onClick={onConfirm} className="bg-red-700 text-sm hover:bg-red-800">{busy ? 'Memproses...' : confirmLabel}</SubmitButton>
      </div>
    </Modal>
  )
}
```

## Frontend/src/routes/PlatformBillingRedirect.tsx

```tsx
import { Navigate, useLocation } from 'react-router-dom'

export default function PlatformBillingRedirect() {
  const location = useLocation()
  // Preserve Xendit return parameters so the existing checkout can resume.
  return <Navigate replace to={{ pathname: '/dashboard/pengaturan', search: location.search, hash: '#langganan-tagihan' }} />
}
```

## Frontend/src/App.tsx

```tsx
import { lazy, Suspense } from 'react'
import { Routes, Route } from 'react-router-dom'
import Home from './pages/Home'
import Login from './pages/Login'
import Register from './pages/Register'
import RequireAuth from './routes/RequireAuth'
import RequireGuest from './routes/RequireGuest'
import RequireRole from './routes/RequireRole'
import DashboardIndex from './routes/DashboardIndex'
import DashboardLayout from './components/layout/DashboardLayout'
import PelangganPage from './pages/dashboard/PelangganPage'
import LanggananPage from './pages/dashboard/LanggananPage'
import KatalogPage from './pages/dashboard/KatalogPage'
import KuponPage from './pages/dashboard/KuponPage'
import PengaturanPage from './pages/dashboard/PengaturanPage'
import PlatformAdminPage from './pages/dashboard/PlatformAdminPage'
import InvoicePage from './pages/dashboard/InvoicePage'
import InvoiceCustomizationPage from './pages/dashboard/InvoiceCustomizationPage'
import PlatformBillingRedirect from './routes/PlatformBillingRedirect'
import IntegrationKeysPage from './pages/dashboard/IntegrationKeysPage'
import PaymentGatewayPage from './pages/dashboard/PaymentGatewayPage'
import WebhookLogsPage from './pages/dashboard/WebhookLogsPage'

const CheckoutPage = lazy(() => import('./pages/CheckoutPage'))

function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/checkout/embed" element={<Suspense fallback={<p className="p-6 text-body">Memuat checkout...</p>}><CheckoutPage /></Suspense>} />
      <Route path="/checkout/:token" element={<Suspense fallback={<p className="p-6 text-body">Memuat checkout...</p>}><CheckoutPage /></Suspense>} />

      <Route element={<RequireGuest />}>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
      </Route>

      <Route element={<RequireAuth />}>
        <Route path="/dashboard" element={<DashboardLayout />}>
          <Route element={<RequireRole roles={['super_admin']} fallback="/dashboard" />}>
            <Route path="platform" element={<PlatformAdminPage />} />
          </Route>
          <Route index element={<DashboardIndex />} />
          <Route element={<RequireRole roles={['admin', 'sales']} fallback="/dashboard" />}>
            <Route path="pelanggan" element={<PelangganPage />} />
            <Route path="langganan" element={<LanggananPage />} />
            <Route path="katalog" element={<KatalogPage />} />
            <Route path="kupon" element={<KuponPage />} />
          </Route>
          <Route element={<RequireRole roles={['admin', 'sales', 'finance']} fallback="/dashboard" />}>
            <Route path="invoices" element={<InvoicePage />} />
          </Route>
          <Route element={<RequireRole roles={['admin']} fallback="/dashboard" />}>
			<Route path="tagihan-platform" element={<PlatformBillingRedirect />} />
            <Route path="tagihan" element={<PlatformBillingRedirect />} />
            <Route path="pengaturan" element={<PengaturanPage />} />
            <Route path="pengaturan/payment-gateway" element={<PaymentGatewayPage />} />
            <Route path="pengaturan/webhook-logs" element={<WebhookLogsPage />} />
            <Route path="integrasi" element={<IntegrationKeysPage />} />
            <Route path="pengaturan/invoice" element={<InvoiceCustomizationPage />} />
          </Route>
        </Route>
      </Route>
    </Routes>
  )
}

export default App
```

## Frontend/src/data/dashboardNav.ts

```tsx
import { FileText, LayoutDashboard, Users, Repeat, LayoutGrid, Ticket, Settings, type LucideIcon } from 'lucide-react'
import type { UserRole } from '../lib/api'

export interface DashboardNavItem {
  label: string
  to: string
  icon: LucideIcon
  allowedRoles?: UserRole[]
  children?: Array<{ label: string; to: string }>
}

export const dashboardNavItems: DashboardNavItem[] = [
  { label: 'Ringkasan', to: '/dashboard', icon: LayoutDashboard, allowedRoles: ['admin', 'sales', 'finance'] },
  { label: 'Pelanggan', to: '/dashboard/pelanggan', icon: Users, allowedRoles: ['admin', 'sales'] },
  { label: 'Langganan', to: '/dashboard/langganan', icon: Repeat, allowedRoles: ['admin', 'sales'] },
  { label: 'Invoice', to: '/dashboard/invoices', icon: FileText, allowedRoles: ['admin', 'sales', 'finance'] },
  {
    label: 'Katalog',
    to: '/dashboard/katalog',
    icon: LayoutGrid,
    allowedRoles: ['admin', 'sales'],
    children: [
      { label: 'Layanan', to: '/dashboard/katalog#layanan' },
      { label: 'Produk', to: '/dashboard/katalog#produk' },
      { label: 'Plan', to: '/dashboard/katalog#plan' },
      { label: 'Add-on', to: '/dashboard/katalog#addon' },
    ],
  },
  { label: 'Kupon', to: '/dashboard/kupon', icon: Ticket, allowedRoles: ['admin', 'sales'] },
  {
    label: 'Pengaturan',
    to: '/dashboard/pengaturan',
    icon: Settings,
    allowedRoles: ['admin'],
    children: [
      { label: 'Bisnis', to: '/dashboard/pengaturan' },
      { label: 'Payment Gateway', to: '/dashboard/pengaturan/payment-gateway' },
      { label: 'Webhook Logs', to: '/dashboard/pengaturan/webhook-logs' },
      { label: 'Integrasi Website', to: '/dashboard/integrasi' },
    ],
  },
]
```

## Frontend/src/components/layout/DashboardLayout.tsx

```tsx
import { Fragment, useEffect, useState } from 'react'
import { Link, Outlet, useLocation } from 'react-router-dom'
import { ChevronDown, LogOut } from 'lucide-react'
import Logo from '../ui/Logo'
import ThemeToggle from '../ui/ThemeToggle'
import { dashboardNavItems } from '../../data/dashboardNav'
import { useAuth } from '../../context/AuthContext'

export default function DashboardLayout() {
  const { auth, logout } = useAuth()
  const location = useLocation()
  const [openMenu, setOpenMenu] = useState<string | null>(null)
  const visibleNavItems = dashboardNavItems.filter(
    (item) => !item.allowedRoles || item.allowedRoles.includes(auth?.user.role ?? 'admin'),
  )
  const isPathActive = (to: string) =>
    location.pathname === to || (to !== '/dashboard' && location.pathname.startsWith(`${to}/`))
  const isNavItemActive = (item: (typeof dashboardNavItems)[number]) =>
    isPathActive(item.to) || Boolean(item.children?.some((child) => isPathActive(child.to.split('#')[0])))

  useEffect(() => {
    const activeItem = visibleNavItems.find((item) => isNavItemActive(item) && item.children)
    if (activeItem) setOpenMenu(activeItem.to)
  }, [location.pathname, auth?.user.role])

  return (
    <div className="flex min-h-screen">
      <aside className="sticky top-0 hidden h-screen max-h-screen w-64 shrink-0 flex-col overflow-y-auto bg-gradient-to-b from-ink to-deep px-5 py-6 md:flex">
        <div className="mb-8 flex items-center justify-between gap-2">
          <Logo variant="light" />
          <ThemeToggle compact />
        </div>

        <nav className="flex flex-1 flex-col gap-1">
          {visibleNavItems.map((item) => {
            const isActive = isNavItemActive(item)
            const Icon = item.icon
            return (
              <div key={item.to}>
                <div className="flex items-center">
                  <Link
                    to={item.to}
                    className={`flex min-w-0 flex-1 items-center gap-3 rounded-xl px-3 py-2.5 text-[14px] font-medium transition-colors ${
                      isActive ? 'bg-brand text-white' : 'text-white/60 hover:bg-white/10 hover:text-white'
                    }`}
                  >
                    <Icon size={18} strokeWidth={2} />
                    <span className="truncate">{item.label}</span>
                  </Link>
                  {item.children && (
                    <button
                      type="button"
                      aria-label={`${openMenu === item.to ? 'Tutup' : 'Buka'} submenu ${item.label}`}
                      aria-expanded={openMenu === item.to}
                      onClick={() => setOpenMenu((current) => (current === item.to ? null : item.to))}
                      className={`ml-1 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-white/50 transition-colors hover:bg-white/10 hover:text-white ${
                        isActive ? 'text-white/80' : ''
                      }`}
                    >
                      <ChevronDown size={16} className={openMenu === item.to ? 'rotate-180' : ''} />
                    </button>
                  )}
                </div>
                {item.children && openMenu === item.to && (
                  <div className="ml-9 mt-1 space-y-0.5 border-l border-white/10 pl-3">
                    {item.children.map((child) => {
                      const [childPath, childHash] = child.to.split('#')
                      const childActive = isPathActive(childPath) && (!childHash || location.hash === `#${childHash}`)
                      return (
                        <Link
                          key={child.to}
                          to={child.to}
                          className={`block rounded-lg px-3 py-2 text-[13px] transition-colors ${
                            childActive ? 'bg-white/15 text-white' : 'text-white/50 hover:bg-white/10 hover:text-white'
                          }`}
                        >
                          {child.label}
                        </Link>
                      )
                    })}
                  </div>
                )}
              </div>
            )
          })}
        </nav>

        <div className="mt-6 border-t border-white/10 pt-4">
          <p className="truncate text-[13px] font-semibold text-white">{auth?.user.name}</p>
          <p className="truncate text-[12px] text-white/50">{auth?.user.email}</p>
          <button
            onClick={logout}
            className="mt-3 flex w-full items-center gap-2 rounded-xl px-3 py-2 text-left text-[13px] font-medium text-white/60 transition-colors hover:bg-white/10 hover:text-white"
          >
            <LogOut size={16} strokeWidth={2} />
            Keluar
          </button>
        </div>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex items-center justify-between bg-gradient-to-b from-ink to-deep px-5 py-4 md:hidden">
          <Logo variant="light" />
          <div className="flex items-center gap-2">
            <ThemeToggle compact />
            <button onClick={logout} className="text-[13px] font-semibold text-white">
              Keluar
            </button>
          </div>
        </header>

        <nav className="flex gap-1 overflow-x-auto bg-gradient-to-b from-ink to-deep px-4 py-2 md:hidden">
          {visibleNavItems.map((item) => {
            const isActive = isNavItemActive(item)
            return (
              <Fragment key={item.to}>
                <Link
                  to={item.to}
                  className={`whitespace-nowrap rounded-full px-3 py-1.5 text-[13px] font-medium transition-colors ${
                    isActive ? 'bg-brand text-white' : 'text-white/60 hover:bg-white/10 hover:text-white'
                  }`}
                >
                  {item.label}
                </Link>
                {isActive && item.children?.map((child) => {
                  const [childPath, childHash] = child.to.split('#')
                  const childActive = isPathActive(childPath) && (!childHash || location.hash === `#${childHash}`)
                  return (
                    <Link
                      key={child.to}
                      to={child.to}
                      className={`whitespace-nowrap rounded-full px-3 py-1.5 text-[13px] font-medium ${
                        childActive ? 'bg-white/20 text-white' : 'bg-white/10 text-white/70'
                      }`}
                    >
                      {child.label}
                    </Link>
                  )
                })}
              </Fragment>
            )
          })}
        </nav>

        <main className={`flex-1 py-8 md:px-10 md:py-10 ${['/dashboard/integrasi', '/dashboard/pengaturan'].includes(location.pathname) ? 'px-6' : 'px-5'}`}>
          <Outlet />
        </main>
      </div>
    </div>
  )
}
```

## Frontend/src/index.css

Selector dark mode kartu diperluas:

```css
  html.dark .integration-page .glass,
  html.dark .business-settings-page .glass {
    background-color: rgba(22, 29, 48, 0.9);
    border-color: rgba(255, 255, 255, 0.12);
  }
```
