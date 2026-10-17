# Layout Settings PayLanggan

Kode lengkap komponen baru atau berubah untuk layout rata kiri. Tidak ada dependency atau endpoint baru.

Pengaturan Bisnis: label 280px, gap 48px, kartu utama fleksibel. Pada viewport mulai 1600px muncul kartu pendamping 360px; di bawah 900px label berada di atas kartu. Langganan dan riwayat tagihan ditumpuk dalam kolom utama.

Integrasi Website: langkah instalasi di kiri, panel sticky 400px di kanan, Cabut ID integrasi di bawah langkah instalasi menggunakan ConfirmDialog. Di bawah 1100px menjadi satu kolom.

Padding kedua halaman: 24px di bawah 900px, 40px mulai 900px. Container memenuhi area konten dan tidak menggunakan mx-auto atau max-width.

TODO: verifikasi instalasi belum memiliki API; UI selalu menjelaskan bahwa status belum diperiksa dan tidak mengklaim instalasi berhasil. Tautan dokumentasi dan dukungan masih placeholder. Aktivasi ulang pembayaran otomatis masih menunggu API yang sudah ditandai TODO pada SubscriptionCard.

Pemeriksaan memakai API tiruan: layout 360/768/1440/1920/2560px, dark mode, salin kode, plan reaktif, fokus dialog, pencabutan ID, redirect dan parameter kembali Xendit, consent, pembatalan langganan. TypeScript, lint komponen yang berubah dan build berhasil. Transaksi gateway nyata tidak dilakukan dalam pemeriksaan layout.

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
    <section aria-labelledby={titleId} className={`settings-section grid items-start gap-6 min-[900px]:grid-cols-[280px_minmax(0,1fr)] min-[900px]:gap-12 ${aside ? 'min-[1600px]:grid-cols-[280px_minmax(0,1fr)_360px]' : ''}`}>
      <header className="min-w-0">
        <h2 id={titleId} className="text-base font-semibold text-ink">{title}</h2>
        <p className="mt-2 text-sm leading-6 text-body">{description}</p>
      </header>
      <div className="min-w-0 space-y-6">{children}</div>
      {aside && <aside aria-label={`Informasi ${title}`} className="hidden min-w-0 min-[1600px]:block">{aside}</aside>}
    </section>
  )
}
```

## Frontend/src/components/dashboard/SettingsRow.tsx

```tsx
import type { ReactNode } from 'react'

interface SettingsRowProps {
  title: ReactNode
  description?: ReactNode
  children: ReactNode
}

export default function SettingsRow({ title, description, children }: SettingsRowProps) {
  return (
    <div className="grid min-w-0 items-center gap-4 border-b border-slate-200 py-4 first:pt-0 last:border-b-0 last:pb-0 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] dark:border-slate-700">
      <div className="min-w-0"><div className="text-sm font-semibold text-ink">{title}</div>{description && <div className="mt-1 break-words text-sm leading-6 text-body">{description}</div>}</div>
      <div className="min-w-0 break-words text-sm text-ink sm:text-right">{children}</div>
    </div>
  )
}
```

## Frontend/src/components/dashboard/InstallationHelp.tsx

```tsx
import { useState, type FormEvent } from 'react'
import CodeBlock from '../ui/CodeBlock'
import GlassPanel from '../ui/GlassPanel'
import Input from '../ui/Input'
import SubmitButton from '../ui/SubmitButton'

interface InstallationHelpProps {
  headerCode: string
  checkoutCode: string
}

export default function InstallationHelp({ headerCode, checkoutCode }: InstallationHelpProps) {
  const [url, setUrl] = useState('')
  const [notice, setNotice] = useState<string | null>(null)
  function verify(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    // TODO: Connect to an installation verification API when one is available.
    // A valid URL alone is not evidence that the script has been installed.
    setNotice('Verifikasi otomatis belum tersedia. Buka halaman penjualan Anda dan coba tombol checkout untuk memeriksa pemasangan.')
  }
  const example = `<!doctype html>
<html lang="id">
<head>
  <!-- Skrip header dipasang di sini -->
${headerCode.replace(/\n\s*/g, ' ') || '  <!-- Salin skrip header dari Langkah 1 -->'}
</head>
<body>
  <!-- Tombol dipasang di halaman penjualan -->
${checkoutCode.replace(/\n\s*/g, ' ') || '  <!-- Salin tombol dari Langkah 2 -->'}
</body>
</html>`
  return <>
    <GlassPanel>
      <div className="flex flex-wrap items-center justify-between gap-3"><h2 className="text-base font-semibold text-ink">Status instalasi</h2><span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-700 dark:bg-slate-800 dark:text-slate-200">Belum terdeteksi</span></div>
      <p className="mt-3 text-sm leading-6 text-body">Status belum diperiksa. Verifikasi otomatis belum tersedia.</p>
      <form onSubmit={verify} className="mt-4 space-y-4">
        <Input id="installation-url" label="URL halaman penjualan" type="url" required placeholder="https://website-anda.com/langganan" value={url} onChange={event => { setUrl(event.target.value); setNotice(null) }} />
        <SubmitButton type="submit" disabled={!url.trim()} className="text-sm">Verifikasi</SubmitButton>
      </form>
      {notice && <p role="status" aria-live="polite" className="mt-4 text-sm leading-6 text-body">{notice}</p>}
    </GlassPanel>
    <GlassPanel>
      <h2 className="text-base font-semibold text-ink">Contoh penempatan</h2>
      <p className="mb-4 mt-3 text-sm leading-6 text-body">Letakkan skrip di <code>&lt;head&gt;</code> dan tombol di <code>&lt;body&gt;</code> halaman penjualan.</p>
      <CodeBlock code={example} label="contoh penempatan" />
    </GlassPanel>
  </>
}
```

## Frontend/src/pages/dashboard/PengaturanPage.tsx

```tsx
import { useEffect, useRef, useState, type FormEvent } from 'react'
import { ArrowRight, FileText } from 'lucide-react'
import { Link, useLocation } from 'react-router-dom'
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
const TEAM_PAGE_SIZE = 4

const roleLabel: Record<UserDTO['role'], string> = {
  super_admin: 'Super Admin',
  admin: 'Admin',
  sales: 'Sales',
  finance: 'Finance',
}

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
    <div className="business-settings-page w-full min-w-0">
      <DashboardPageHeader
        title="Pengaturan Bisnis"
        description="Atur informasi dan preferensi bisnis Anda."
      />

      {error && !businessModalOpen && <p className="mt-4 text-[13px] text-red-600">{error}</p>}

      <div className="mt-10 space-y-10 [&>.settings-section+*]:border-t [&>.settings-section+*]:border-slate-200 [&>.settings-section+*]:pt-10 dark:[&>.settings-section+*]:border-slate-700">
        <SettingsSection title="Identitas Bisnis" description="Informasi utama yang digunakan untuk bisnis Anda." aside={<GlassPanel><h3 className="text-sm font-semibold text-ink">Identitas bisnis Anda</h3><p className="mt-3 text-sm leading-6 text-body">Nama dan zona waktu digunakan untuk pengaturan bisnis. Email admin mengikuti akun yang sedang masuk.</p></GlassPanel>}>
          {isLoading ? <FormSkeleton fields={['Nama Bisnis', 'Email Admin', 'Zona Waktu']} /> : <GlassPanel>
            <div className="mb-6 flex justify-end"><SubmitButton type="button" onClick={() => { setError(null); setSuccess(null); setBusinessModalOpen(true) }} className="px-4 py-2 text-xs">Edit Identitas</SubmitButton></div>
            <SettingsRow title="Nama Bisnis"><span className="font-semibold">{name || '—'}</span></SettingsRow>
            <SettingsRow title="Email Admin">{auth?.user.email || '—'}</SettingsRow>
            <SettingsRow title="Zona Waktu"><div className="flex flex-wrap items-center gap-3 sm:justify-end"><span>{timezone}</span><span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200">{business?.status === 'active' ? 'Aktif' : business?.status === 'inactive' ? 'Tidak aktif' : business?.status === 'suspended' ? 'Ditangguhkan' : 'Belum aktif'}</span></div></SettingsRow>
            {success && <p role="status" className="mt-4 text-sm text-emerald-700 dark:text-emerald-200">{success}</p>}
          </GlassPanel>}
        </SettingsSection>

        {auth?.user.role === 'admin' && !isLoading && <SettingsSection title="Tim Bisnis" description="Anggota yang terhubung ke bisnis ini." aside={<GlassPanel><h3 className="text-sm font-semibold text-ink">Akses tim</h3><p className="mt-3 text-sm leading-6 text-body">Tambahkan anggota dengan peran Sales atau Finance sesuai tanggung jawabnya. Akun staf dapat digunakan untuk login setelah dibuat.</p></GlassPanel>}>
          <GlassPanel>
            <div className="mb-6 flex flex-wrap items-center justify-between gap-4"><p className="text-sm font-semibold text-ink">{users.length} anggota</p><SubmitButton type="button" onClick={() => { setStaffError(null); setStaffSuccess(null); setStaffModalOpen(true) }} className="px-4 py-2 text-xs">+ Tambah Anggota</SubmitButton></div>
            {visibleUsers.map(user => <SettingsRow key={user.id} title={<div className="flex items-center gap-3"><span aria-hidden="true" className="flex size-10 shrink-0 items-center justify-center rounded-full bg-brand/10 text-sm font-semibold text-brand dark:text-blue-300">{user.name.trim().split(/\s+/).slice(0, 2).map(part => part[0]).join('').toUpperCase()}</span><span className="min-w-0 break-words">{user.name}</span></div>} description={user.email}><span className="inline-block rounded-full bg-brand/10 px-3 py-1 text-xs font-semibold text-brand dark:text-blue-300">{roleLabel[user.role]}</span></SettingsRow>)}
            {totalTeamPages > 1 && <nav aria-label="Halaman anggota tim" className="mt-6 flex flex-wrap gap-2">{Array.from({ length: totalTeamPages }, (_, index) => index + 1).map(pageNumber => <button key={pageNumber} type="button" onClick={() => setTeamPage(pageNumber)} aria-label={`Halaman anggota ${pageNumber}`} aria-current={pageNumber === teamPage ? 'page' : undefined} className={`size-8 rounded-lg text-xs font-semibold focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand ${pageNumber === teamPage ? 'bg-brand text-white' : 'text-body hover:bg-brand/10'}`}>{pageNumber}</button>)}</nav>}
            {staffSuccess && <p role="status" className="mt-4 text-sm text-emerald-700 dark:text-emerald-200">{staffSuccess}</p>}
          </GlassPanel>
        </SettingsSection>}

        {!isLoading && <SettingsSection title="Tampilan Invoice" description="Atur logo, warna, alamat, dan catatan pada invoice bisnis Anda." aside={<GlassPanel><h3 className="text-sm font-semibold text-ink">Identitas pada invoice</h3><p className="mt-3 text-sm leading-6 text-body">Sesuaikan tampilan invoice dengan identitas bisnis Anda melalui halaman kustomisasi.</p></GlassPanel>}>
          <Link to="/dashboard/pengaturan/invoice" className="glass flex w-full flex-wrap items-center justify-between gap-4 rounded-2xl px-6 py-4 transition-colors hover:bg-white/60 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand dark:hover:bg-white/10">
            <div className="flex min-w-0 flex-1 items-center gap-4"><span className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-brand/10 text-brand dark:text-blue-300"><FileText size={21} aria-hidden="true" /></span><div className="min-w-0"><p className="text-sm font-semibold text-ink">Tampilan Invoice</p><p className="mt-1 text-sm leading-6 text-body">Atur logo, warna, alamat, dan catatan pada invoice bisnis Anda.</p></div></div>
            <div className="flex items-center gap-4"><div aria-label="Preview identitas" className="hidden items-center gap-2 sm:flex"><span className="size-3 rounded-full bg-brand" /><span className="size-3 rounded-full bg-amber" /><span className="size-3 rounded-full bg-violet" /><span className="sr-only">Preview identitas</span></div><span className="hidden text-sm font-semibold text-brand dark:text-blue-300 sm:inline">Kustomisasi</span><ArrowRight className="text-brand dark:text-blue-300" size={18} aria-hidden="true" /></div>
          </Link>
        </SettingsSection>}

        <div ref={billingRef} id="langganan-tagihan" className="scroll-mt-6 border-t border-slate-200 pt-10 dark:border-slate-700">
          <SettingsSection title="Langganan & Tagihan" description="Langganan bisnis Anda untuk menggunakan PayLanggan." aside={<GlassPanel><h3 className="text-sm font-semibold text-ink">Perpanjangan langganan</h3><p className="mt-3 text-sm leading-6 text-body">Periksa status pembayaran otomatis dan riwayat tagihan di sini. Menghentikan perpanjangan tetap mempertahankan periode yang sudah dibayar hingga tanggal berakhir.</p></GlassPanel>}><PlatformBillingSection /></SettingsSection>
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
    <div className="grid items-start gap-6">
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

## Frontend/src/pages/dashboard/IntegrationKeysPage.tsx

```tsx
import { useEffect, useState, type FormEvent } from 'react'
import { Check, Copy, KeyRound, Plus, RotateCw, ShieldCheck, Trash2 } from 'lucide-react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import GlassPanel from '../../components/ui/GlassPanel'
import Input from '../../components/ui/Input'
import Modal from '../../components/ui/Modal'
import SubmitButton from '../../components/ui/SubmitButton'
import { useAuth } from '../../context/AuthContext'
import EmbedSitesPanel from './EmbedSitesPanel'
import {
  ApiError,
  createIntegrationKey,
  listIntegrationKeys,
  revokeIntegrationKey,
  type CreatedIntegrationKey,
  type IntegrationKey,
} from '../../lib/api'

// API credentials remain in the codebase but are temporarily hidden from the dashboard.
const API_INTEGRATION_ENABLED = false

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime())
    ? 'Tanggal tidak tersedia'
    : date.toLocaleString('id-ID', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

export default function IntegrationKeysPage() {
  const { auth } = useAuth()
  const token = auth?.token
  const [keys, setKeys] = useState<IntegrationKey[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [creating, setCreating] = useState(false)
  const [revokingId, setRevokingId] = useState<number | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [keyName, setKeyName] = useState('')
  const [newKey, setNewKey] = useState<CreatedIntegrationKey | null>(null)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const [mode, setMode] = useState<'dropin' | 'api'>('dropin')

  useEffect(() => {
    if (!token) return
    let active = true
    listIntegrationKeys(token)
      .then((result) => { if (active) setKeys(result) })
      .catch((err) => { if (active) setError(err instanceof ApiError ? err.message : 'Gagal memuat kunci integrasi.') })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [token])

  async function refresh() {
    if (!token) return
    setRefreshing(true)
    setError(null)
    try {
      setKeys(await listIntegrationKeys(token))
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memperbarui daftar kunci.')
    } finally {
      setRefreshing(false)
    }
  }

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!token || creating) return
    const name = keyName.trim()
    if (!name) {
      setError('Nama kunci wajib diisi.')
      return
    }
    setCreating(true)
    setError(null)
    setNotice(null)
    try {
      const created = await createIntegrationKey(token, name)
      const keyInfo: IntegrationKey = {
        id: created.id,
        name: created.name,
        created_at: created.created_at,
        revoked_at: created.revoked_at,
      }
      setKeys((current) => [keyInfo, ...current])
      setNewKey(created)
      setCopied(false)
      setKeyName('')
      setCreateOpen(false)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal membuat kunci integrasi.')
    } finally {
      setCreating(false)
    }
  }

  async function copySecret() {
    if (!newKey) return
    try {
      await navigator.clipboard.writeText(newKey.secret_key)
      setCopied(true)
      setError(null)
    } catch {
      setError('Gagal menyalin otomatis. Pilih dan salin kunci secara manual.')
    }
  }

  async function revoke(key: IntegrationKey) {
    if (!token || revokingId !== null) return
    if (!window.confirm(`Cabut kunci "${key.name}"? Website yang memakai kunci ini tidak bisa lagi membuat atau membaca pesanan.`)) return
    setRevokingId(key.id)
    setError(null)
    setNotice(null)
    try {
      await revokeIntegrationKey(token, key.id)
      if (newKey?.id === key.id) setNewKey(null)
      setKeys((current) => current.map((item) => item.id === key.id
        ? { ...item, revoked_at: new Date().toISOString() }
        : item))
      setNotice(`Kunci "${key.name}" berhasil dicabut.`)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal mencabut kunci integrasi.')
    } finally {
      setRevokingId(null)
    }
  }

  const activeKeys = keys.filter((key) => !key.revoked_at)

  return (
    <div className="integration-page w-full min-w-0">
      <DashboardPageHeader
        title="Integrasi Website"
        description="Pasang skrip checkout PayLanggan pada website Anda."
      />

      {API_INTEGRATION_ENABLED && <div role="tablist" aria-label="Cara integrasi" className="mb-6 flex gap-4">
        <button type="button" role="tab" aria-selected={mode === 'dropin'} onClick={() => setMode('dropin')}>Via skrip drop-in</button>
        <button type="button" role="tab" aria-selected={mode === 'api'} onClick={() => setMode('api')}>Via API</button>
      </div>}
      <div className="mb-6">
          {mode === 'dropin' ? <EmbedSitesPanel /> : (
            <div className="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,2fr)]">
              <div>
                <h3 className="font-display text-[16px] font-bold text-ink">Integrasi melalui server</h3>
                <p className="mt-2 text-[13px] leading-relaxed text-body">Buat pesanan dan periksa status pembayaran dari server website Anda. Simpan secret key hanya di server.</p>
              </div>
              <div className="rounded-xl border border-slate-200 bg-slate-50 p-4 text-[13px] text-body">
                <p className="font-semibold text-ink">Langkah singkat</p>
                <ol className="mt-2 list-inside list-decimal space-y-1">
                  <li>Buat kunci integrasi di bawah.</li>
                  <li>Simpan kunci pada server website bisnis.</li>
                  <li>Gunakan API PayLanggan untuk membuat pesanan dan membaca statusnya.</li>
                </ol>
                <p className="mt-3 text-[12px] text-muted">Gunakan endpoint API integrasi yang didokumentasikan bersama proyek.</p>
              </div>
            </div>
          )}
      </div>

      {mode === 'api' && <>

      <div className="mb-6 grid gap-4 sm:grid-cols-2">
        <GlassPanel className="flex items-start gap-4">
          <span className="rounded-xl bg-brand/10 p-3 text-brand"><KeyRound size={22} /></span>
          <div>
            <p className="text-[13px] text-muted">Kunci aktif</p>
            <p className="mt-1 font-display text-[28px] font-bold text-ink">{activeKeys.length}</p>
            <p className="text-[12px] text-body">Kunci berlaku sampai Anda mencabutnya.</p>
          </div>
        </GlassPanel>
        <GlassPanel className="flex items-start gap-4">
          <span className="rounded-xl bg-emerald-100 p-3 text-emerald-700"><ShieldCheck size={22} /></span>
          <div>
            <p className="text-[14px] font-semibold text-ink">Simpan kunci di server bisnis</p>
            <p className="mt-1 text-[13px] leading-relaxed text-body">Kunci ini digunakan oleh server website untuk membuat dan memeriksa pesanan. Jangan masukkan ke kode browser.</p>
          </div>
        </GlassPanel>
      </div>

      {error && <p role="alert" className="mb-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-[13px] text-red-700">{error}</p>}
      {notice && <p role="status" className="mb-4 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-[13px] text-emerald-800">{notice}</p>}

      {newKey && (
        <section aria-labelledby="new-key-title" className="mb-6 rounded-2xl border border-amber-300 bg-amber-50 p-5 shadow-sm sm:p-6">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <h2 id="new-key-title" className="font-display text-[17px] font-bold text-ink">Kunci baru: {newKey.name}</h2>
              <p className="mt-1 text-[13px] leading-relaxed text-amber-900">Salin dan simpan sekarang. Kunci rahasia ini tidak akan ditampilkan lagi setelah panel ditutup atau halaman dimuat ulang.</p>
            </div>
            <span className="rounded-full bg-amber-100 px-3 py-1 text-[11px] font-semibold text-amber-900">Ditampilkan sekali</span>
          </div>
          <label htmlFor="integration-secret" className="mt-5 block text-[12px] font-semibold text-ink">Secret key</label>
          <div className="mt-2 flex flex-col gap-2 sm:flex-row">
            <input id="integration-secret" readOnly value={newKey.secret_key} onFocus={(event) => event.currentTarget.select()}
              className="min-w-0 flex-1 rounded-xl border border-amber-300 bg-white px-4 py-2.5 font-mono text-[13px] text-ink" />
            <SubmitButton type="button" onClick={copySecret} className="justify-center gap-2">
              {copied ? <Check size={16} /> : <Copy size={16} />}
              {copied ? 'Tersalin' : 'Salin kunci'}
            </SubmitButton>
          </div>
          <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
            <p className="text-[12px] text-amber-900">Simpan sebagai <code className="font-semibold">PAYLANGGAN_SECRET_KEY</code> di server website bisnis.</p>
            <button type="button" onClick={() => { setNewKey(null); setCopied(false) }}
              className="rounded-lg border border-amber-500 px-3 py-2 text-[12px] font-semibold text-amber-950 hover:bg-amber-100">
              Saya sudah menyimpan kunci
            </button>
          </div>
        </section>
      )}

      <GlassPanel>
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h2 className="font-display text-[18px] font-bold text-ink">Kunci integrasi</h2>
            <p className="mt-1 text-[13px] text-body">Gunakan nama yang menunjukkan website atau sistem pemakainya.</p>
          </div>
          <div className="flex gap-2">
            <button type="button" onClick={refresh} disabled={refreshing || loading}
              className="flex h-10 items-center gap-2 rounded-xl border border-white/70 bg-white/50 px-3 text-[13px] font-semibold text-body hover:bg-white disabled:opacity-50">
              <RotateCw size={16} className={refreshing ? 'animate-spin' : ''} /> Perbarui
            </button>
            <SubmitButton type="button" disabled={Boolean(newKey)} onClick={() => { setError(null); setCreateOpen(true) }}
              className="gap-2 px-4 py-2 text-[13px]"><Plus size={16} /> Buat kunci</SubmitButton>
          </div>
        </div>

        {loading ? (
          <p role="status" className="mt-8 text-[13px] text-muted">Memuat kunci integrasi...</p>
        ) : keys.length === 0 ? (
          <div className="mt-7 rounded-xl border border-dashed border-white/70 bg-white/40 px-5 py-10 text-center">
            <KeyRound className="mx-auto text-muted" size={28} />
            <p className="mt-3 font-semibold text-ink">Belum ada kunci integrasi</p>
            <p className="mt-1 text-[13px] text-body">Buat kunci pertama untuk menghubungkan server website bisnis.</p>
          </div>
        ) : (
          <div className="mt-6 space-y-3">
            {keys.map((key) => (
              <div key={key.id} className="flex flex-col gap-4 rounded-xl border border-white/60 bg-white/40 p-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="font-semibold text-ink">{key.name}</p>
                    <span className={`rounded-full px-2.5 py-0.5 text-[11px] font-semibold ${key.revoked_at ? 'bg-slate-200 text-slate-600' : 'bg-emerald-100 text-emerald-800'}`}>
                      {key.revoked_at ? 'Dicabut' : 'Aktif'}
                    </span>
                  </div>
                  <p className="mt-1 text-[12px] text-muted">Dibuat {formatDate(key.created_at)}{key.revoked_at ? ` · Dicabut ${formatDate(key.revoked_at)}` : ''}</p>
                </div>
                {!key.revoked_at && (
                  <button type="button" disabled={revokingId !== null} onClick={() => revoke(key)}
                    className="inline-flex items-center justify-center gap-2 rounded-lg border border-red-200 px-3 py-2 text-[12px] font-semibold text-red-700 hover:bg-red-50 disabled:opacity-50">
                    <Trash2 size={14} /> {revokingId === key.id ? 'Mencabut...' : 'Cabut kunci'}
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </GlassPanel>

      </>}

      <Modal open={createOpen} onClose={() => { if (!creating) setCreateOpen(false) }} title="Buat kunci integrasi">
        <form onSubmit={create} className="space-y-5">
          <p className="text-[13px] leading-relaxed text-body">Beri nama sesuai website atau aplikasi yang akan menggunakan kunci ini. Secret key akan ditampilkan satu kali setelah dibuat.</p>
          <Input id="integration-name" label="Nama kunci" value={keyName} onChange={(event) => setKeyName(event.target.value)}
            placeholder="Website Kelana Workspace" maxLength={100} required autoFocus />
          {error && <p role="alert" className="text-[13px] text-red-700">{error}</p>}
          <div className="flex justify-end gap-2">
            <button type="button" disabled={creating} onClick={() => setCreateOpen(false)}
              className="rounded-xl border border-white/70 px-4 py-2 text-[13px] font-semibold text-body hover:bg-white/50">Batal</button>
            <SubmitButton type="submit" disabled={creating} className="px-4 py-2 text-[13px]">
              {creating ? 'Membuat...' : 'Buat kunci'}
            </SubmitButton>
          </div>
        </form>
      </Modal>
    </div>
  )
}
```

## Frontend/src/pages/dashboard/EmbedSitesPanel.tsx

```tsx
import { useEffect, useMemo, useState } from 'react'
import { Eye, LifeBuoy, Trash2 } from 'lucide-react'
import CodeBlock from '../../components/ui/CodeBlock'
import CopyButton from '../../components/ui/CopyButton'
import GlassPanel from '../../components/ui/GlassPanel'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import InstallationHelp from '../../components/dashboard/InstallationHelp'
import SubmitButton from '../../components/ui/SubmitButton'
import {
  API_BASE_URL, createEmbedSite, listEmbedSites, listPlans, listProducts, revokeEmbedSite,
  type EmbedSite, type Plan, type Product,
} from '../../lib/api'
import { formatCurrency } from '../../lib/format'
import { useAuth } from '../../context/AuthContext'
import Select from '../../components/ui/Select'

function escapeHtml(value: string) {
  return value.replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[char] ?? char)
}

async function loadPlans(token: string) {
  const first = await listPlans(token, { page: 1, limit: 100 })
  const items = [...first.items]
  for (let page = 2; items.length < first.total; page++) {
    const next = await listPlans(token, { page, limit: 100 })
    if (next.items.length === 0) break
    items.push(...next.items)
  }
  return items
}

async function loadProducts(token: string) {
  const first = await listProducts(token, { page: 1, limit: 100 })
  const items = [...first.items]
  for (let page = 2; items.length < first.total; page++) {
    const next = await listProducts(token, { page, limit: 100 })
    if (next.items.length === 0) break
    items.push(...next.items)
  }
  return items
}

export default function EmbedSitesPanel() {
  const { auth } = useAuth()
  const token = auth?.token
  const [sites, setSites] = useState<EmbedSite[]>([])
  const [plans, setPlans] = useState<Plan[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [plansLoaded, setPlansLoaded] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [revokeTarget, setRevokeTarget] = useState<EmbedSite | null>(null)
  const [buttonPlanIds, setButtonPlanIds] = useState<Record<number, number>>({})
  const [siteId, setSiteId] = useState<number | null>(null)
  const [productId, setProductId] = useState<number | null>(null)

  useEffect(() => {
    if (!token) return
    let active = true
    void Promise.allSettled([listEmbedSites(token), loadPlans(token), loadProducts(token)]).then(([siteResult, planResult, productResult]) => {
      if (!active) return
      if (siteResult.status === 'fulfilled') {
        setSites(siteResult.value)
        setSiteId(siteResult.value.find((site) => !site.revoked_at)?.id ?? null)
      }
      if (planResult.status === 'fulfilled') setPlans(planResult.value)
      if (productResult.status === 'fulfilled') setProducts(productResult.value)
      setPlansLoaded(true)
      const failures = [
        siteResult.status === 'rejected' && `Gagal memuat ID integrasi: ${siteResult.reason instanceof Error ? siteResult.reason.message : 'Kesalahan server'}`,
        planResult.status === 'rejected' && `Gagal memuat plan: ${planResult.reason instanceof Error ? planResult.reason.message : 'Kesalahan server'}`,
        productResult.status === 'rejected' && `Gagal memuat produk: ${productResult.reason instanceof Error ? productResult.reason.message : 'Kesalahan server'}`,
      ].filter(Boolean)
      if (failures.length) setError(`${failures.join('. ')}. Pastikan backend terbaru berjalan dan migrasi sudah diterapkan.`)
    })
    return () => { active = false }
  }, [token])

  async function create() {
    if (!token || busy) return
    setBusy(true); setError(null)
    try {
      const site = await createEmbedSite(token)
      setSites((current) => [site, ...current])
      setSiteId(site.id)
    } catch (err) { setError(err instanceof Error ? err.message : 'Gagal membuat ID integrasi.') }
    finally { setBusy(false) }
  }

  async function revoke(site: EmbedSite) {
    if (!token || busy) return
    setBusy(true); setError(null)
    try {
      await revokeEmbedSite(token, site.id)
      setRevokeTarget(null)
      setSites((current) => current.map((item) => item.id === site.id ? { ...item, revoked_at: new Date().toISOString() } : item))
    } catch (err) { setError(err instanceof Error ? err.message : 'Gagal mencabut ID integrasi.') }
    finally { setBusy(false) }
  }

  const activeSites = sites.filter((site) => !site.revoked_at)
  const activeSite = activeSites.find((site) => site.id === siteId) ?? activeSites[0]
  const availablePlans = activeSite ? plans.filter((plan) => plan.trial_days === 0 && plan.price >= 5000 && plan.price <= 200000000 && Number.isInteger(plan.price) && ['monthly', 'yearly'].includes(plan.billing_cycle) && products.some((product) => product.id === plan.product_id && product.status === 'active')) : []
  const allowedProducts = products.filter((product) => availablePlans.some((plan) => plan.product_id === product.id))
  const selectedProductId = allowedProducts.some((product) => product.id === productId) ? productId : allowedProducts[0]?.id
  const productPlans = availablePlans.filter((plan) => plan.product_id === selectedProductId)
  const selectedPlanId = productPlans.some((plan) => plan.id === buttonPlanIds[activeSite?.id ?? 0])
    ? buttonPlanIds[activeSite?.id ?? 0]
    : productPlans[0]?.id
  const selectedPlan = plans.find((plan) => plan.id === selectedPlanId)

  const headerCode = useMemo(() => activeSite ? `<script
  src="${window.location.origin}/paylanggan.js"
  data-pl-site="${escapeHtml(activeSite.public_id)}"
  data-pl-api-url="${escapeHtml(API_BASE_URL)}">
</script>` : '', [activeSite])
  const checkoutCode = useMemo(() => selectedPlan?.public_id ? `<button
  data-pl-checkout
  data-pl-plan="${escapeHtml(selectedPlan.public_id)}">
  Langganan ${escapeHtml(selectedPlan.name)}
</button>` : '', [selectedPlan])
  const period = selectedPlan?.billing_cycle === 'monthly' ? 'Bulanan' : 'Tahunan'

  return <section aria-label="Instalasi skrip drop-in" className="space-y-6">
    {error && <p role="alert" className="rounded-xl bg-red-50 p-4 text-[13px] text-red-700 dark:bg-red-950/40 dark:text-red-300">{error}</p>}
    <div className="grid items-start gap-6 min-[1100px]:grid-cols-[minmax(0,1fr)_400px]">
        <div className="min-w-0 space-y-6">
        {activeSite ? <>
        <div className="relative space-y-6">
          <div className="relative"><div aria-hidden="true" className="absolute left-8 top-8 h-[calc(100%+24px)] w-0.5 bg-brand/70 sm:left-10 sm:top-10" /><GlassPanel className="relative !p-4 sm:!p-6">
            <div className="mb-4 flex items-start gap-4"><span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-brand text-[14px] font-bold text-white">1</span><div><h2 className="text-[16px] font-semibold text-ink">Skrip header</h2><p className="mt-1 text-[13px] leading-6 text-body">Salin ke bagian <code>&lt;head&gt;</code> halaman yang menjual langganan. Cukup pasang satu kali untuk semua plan bisnis.</p></div></div>
            <CodeBlock code={headerCode} label="skrip header" />
          </GlassPanel></div>
          <GlassPanel className="relative !p-4 sm:!p-6">
            <div className="mb-4 flex items-start gap-4"><span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-brand text-[14px] font-bold text-white">2</span><div><h2 className="text-[16px] font-semibold text-ink">Kode tombol checkout</h2><p className="mt-1 text-[13px] leading-6 text-body">Pilih produk dan plan. Salin tombol ke halaman penjualan. Skrip header tetap sama saat plan baru ditambahkan.</p></div></div>
            <div className="grid gap-4 sm:grid-cols-2">
              <Select id={`embed-product-${activeSite.id}`} label="Produk" value={selectedProductId ?? ''} onChange={(event) => setProductId(Number(event.target.value))} disabled={allowedProducts.length === 0}>
                {allowedProducts.length === 0 && <option value="">Belum ada produk</option>}
                {allowedProducts.map((product) => <option key={product.id} value={product.id}>{product.name}</option>)}
              </Select>
              <Select id={`embed-plan-${activeSite.id}`} label="Plan" value={selectedPlanId ?? ''} onChange={(event) => setButtonPlanIds((current) => ({ ...current, [activeSite.id]: Number(event.target.value) }))} disabled={productPlans.length === 0}>
                {productPlans.length === 0 && <option value="">Belum ada plan</option>}
                {productPlans.map((plan) => <option key={plan.id} value={plan.id}>{plan.name}</option>)}
              </Select>
            </div>
            <dl className="my-4 flex flex-wrap gap-3 text-[12px] text-body">{[['Jenis checkout', 'Langganan'], ['Mata uang', 'IDR'], ['Periode', selectedPlan ? period : '-']].map(([label, value]) => <div key={label} className="rounded-lg border border-border bg-white/40 px-3 py-2"><dt className="text-[11px]">{label}</dt><dd className="mt-1 font-semibold text-ink">{value}</dd></div>)}</dl>
            {selectedPlan && <p className="mb-4 text-[13px] text-body">Harga: <strong className="font-display text-[22px] font-bold text-ink">{formatCurrency(selectedPlan.price)}</strong><span className="ml-1">/ {selectedPlan.billing_cycle === 'monthly' ? 'bulan' : 'tahun'}</span></p>}
            {checkoutCode ? <CodeBlock code={checkoutCode} label="kode tombol checkout" /> : <p className="rounded-xl bg-white/40 p-4 text-[13px] leading-6 text-body">Belum ada plan yang mendukung checkout kartu. Buat plan aktif di Katalog terlebih dahulu.</p>}
          </GlassPanel>
        </div>
        </> : <GlassPanel><h2 className="text-[16px] font-semibold text-ink">Skrip drop-in</h2><p className="mt-2 text-[13px] leading-6 text-body">Buat ID integrasi untuk menampilkan skrip header dan kode tombol checkout.</p></GlassPanel>}
          {activeSite && (
      <GlassPanel className="space-y-4">
        <div><h2 className="flex items-center gap-2 text-[16px] font-semibold text-ink"><Trash2 size={18} aria-hidden="true" />Cabut ID integrasi</h2><p className="mt-3 text-[13px] leading-6 text-body">Semua tombol checkout yang memakai ID ini akan berhenti membuat pesanan.</p></div>
        <button type="button" onClick={() => setRevokeTarget(activeSite)} disabled={busy} className="inline-flex min-h-10 items-center gap-2 rounded-lg border border-red-300 px-4 py-2 text-[13px] font-semibold text-red-700 hover:bg-red-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600 dark:text-red-300 dark:hover:bg-red-950/40"><Trash2 size={16} aria-hidden="true" />Cabut</button>
      </GlassPanel>
          )}
        </div>
        <aside className="min-w-0 space-y-6 min-[1100px]:sticky min-[1100px]:top-6" aria-label="ID integrasi, preview dan bantuan">
    <GlassPanel>
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-3"><p className="text-[13px] font-semibold text-ink">ID integrasi bisnis</p>{activeSite && <span className="rounded-full bg-emerald-100 px-3 py-1 text-[11px] font-semibold text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200">Aktif</span>}</div>
          {activeSite ? <div className="mt-2 flex items-center gap-3"><code className="min-w-0 flex-1 break-all font-mono text-[13px] text-ink">{activeSite.public_id}</code><CopyButton text={activeSite.public_id} label="Salin ID integrasi bisnis" iconOnly /></div> : <p role="status" className="mt-2 text-[13px] text-body">{plansLoaded ? 'Buat ID integrasi untuk menampilkan skrip header dan kode tombol checkout.' : 'Memuat ID integrasi...'}</p>}
        </div>
        {!activeSite && <SubmitButton type="button" onClick={() => void create()} disabled={busy || !plansLoaded}>{busy ? 'Menyiapkan...' : 'Buat ID integrasi'}</SubmitButton>}
      </div>
      <p className="mt-3 text-[12px] leading-5 text-body">Gunakan satu ID integrasi untuk memasang tombol checkout pada halaman penjualan Anda. Tidak perlu mendaftarkan nama atau URL website.</p>
    </GlassPanel>
          <GlassPanel>
            <h2 className="flex items-center gap-2 text-[16px] font-semibold text-ink"><Eye size={18} aria-hidden="true" />Preview</h2>
            <div className="mt-4 rounded-xl border border-border bg-white/40 p-6 text-center" aria-live="polite">
              <p className="text-[14px] font-semibold text-ink">{selectedPlan?.name ?? 'Pilih plan'}</p>
              {selectedPlan && <p className="mt-2 text-[13px] text-body">{formatCurrency(selectedPlan.price)} / {selectedPlan.billing_cycle === 'monthly' ? 'bulan' : 'tahun'}</p>}
              <button type="button" aria-disabled="true" className="mt-6 w-full rounded-xl bg-brand px-4 py-3 text-[13px] font-semibold text-white focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand">Langganan {selectedPlan?.name ?? 'plan'}</button>
              <p className="mt-3 text-[12px] leading-5 text-body">Preview tampilan tombol. Tidak membuat pesanan.</p>
            </div>
          </GlassPanel>
          <InstallationHelp headerCode={headerCode} checkoutCode={checkoutCode} />
          <GlassPanel>
            <h2 className="flex items-center gap-2 text-[16px] font-semibold text-ink"><LifeBuoy size={18} aria-hidden="true" />Butuh bantuan?</h2>
            <p className="mt-3 text-[13px] leading-6 text-body">Harga dan informasi plan mengikuti data katalog PayLanggan. Tambahkan tombol lain untuk setiap plan yang ingin dijual.</p>
            {/* TODO: Replace placeholders with published documentation and support URLs. */}
            <div className="mt-4 flex flex-wrap gap-4 text-[13px] font-semibold text-brand dark:text-blue-300"><a href="#dokumentasi-drop-in" onClick={(event) => event.preventDefault()} className="rounded focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand">Dokumentasi (segera)</a><a href="#dukungan-paylanggan" onClick={(event) => event.preventDefault()} className="rounded focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand">Kontak dukungan (segera)</a></div>
          </GlassPanel>
        </aside>
      </div>
    <ConfirmDialog open={revokeTarget !== null} onClose={() => setRevokeTarget(null)} onConfirm={() => { if (revokeTarget) void revoke(revokeTarget) }} title="Cabut ID integrasi ini?" description={`Semua tombol checkout yang memakai ID ini akan berhenti membuat pesanan. ${revokeTarget?.public_id ?? ''}`} confirmLabel="Cabut ID integrasi" busy={busy} error={error} />
  </section>
}
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

        <main className={`flex-1 py-8 md:py-10 ${['/dashboard/integrasi', '/dashboard/pengaturan'].includes(location.pathname) ? 'px-6 min-[900px]:px-10' : 'px-5 md:px-10'}`}>
          <Outlet />
        </main>
      </div>
    </div>
  )
}
```
