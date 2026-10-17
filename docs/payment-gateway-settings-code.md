# Revisi Payment Gateway

Halaman memakai API konfigurasi yang sudah ada. Satu gateway aktif per bisnis. Pergantian berlaku setelah konfigurasi baru berhasil disimpan; kredensial lama ditimpa dan ID/URL webhook tetap sama. Tidak ada endpoint baru atau bagian API yang belum disambungkan.

Verifikasi: TypeScript, build frontend, dan browser dengan API mock (tanpa gateway, Midtrans, Xendit; ukuran 360 sampai 2560px; dark mode; fokus dialog; salin URL; payload penyimpanan). Pengujian ini tidak menjalankan transaksi gateway nyata.

## Kode lengkap file baru atau berubah

### Frontend/src/pages/dashboard/PaymentGatewayPage.tsx

```tsx
﻿import { useEffect, useRef, useState, type ChangeEvent, type FormEvent } from 'react'
import { ArrowRight, CheckCircle2, CreditCard, ExternalLink, KeyRound, ShieldCheck } from 'lucide-react'
import { Link } from 'react-router-dom'
import xenditLogo from '../../assets/gateways/xendit.png'
import midtransLogo from '../../assets/gateways/midtrans.svg'
import SettingsSection from '../../components/dashboard/SettingsSection'
import SettingsRow from '../../components/dashboard/SettingsRow'
import StatusBadge from '../../components/ui/StatusBadge'
import CopyButton from '../../components/ui/CopyButton'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import GlassPanel from '../../components/ui/GlassPanel'
import Input from '../../components/ui/Input'
import Modal from '../../components/ui/Modal'
import Select from '../../components/ui/Select'
import SubmitButton from '../../components/ui/SubmitButton'
import { useAuth } from '../../context/AuthContext'
import { ApiError, getBusinessPaymentGateway, saveBusinessPaymentGateway, type BusinessPaymentGatewayDTO, type BusinessPaymentGatewayInput } from '../../lib/api'

type Provider = BusinessPaymentGatewayInput['provider']
const emptyValues = { xenditSecret: '', xenditCallback: '', midtransMID: '', midtransClient: '', midtransServer: '' }
const gatewayInfo = {
  xendit: { name: 'Xendit', description: 'Terima pembayaran kartu pelanggan melalui akun Xendit bisnis Anda.', documentation: 'https://docs.xendit.co/docs/payments-integration-setup' },
  midtrans: { name: 'Midtrans', description: 'Hubungkan akun Midtrans untuk menerima pembayaran kartu pelanggan.', documentation: 'https://docs.midtrans.com/docs/access-keys' },
}
function GatewayMark({ provider, compact = false }: { provider: Provider; compact?: boolean }) {
  return <span className={`flex h-10 ${compact ? 'w-10 px-1' : 'w-20 px-3'} shrink-0 items-center justify-center rounded-xl ${provider === 'midtrans' ? 'bg-[#002f50]' : 'bg-white'}`}><img src={provider === 'xendit' ? xenditLogo : midtransLogo} alt={`Logo ${gatewayInfo[provider].name}`} className="max-h-7 w-full object-contain" /></span>
}

export default function PaymentGatewayPage() {
  const { auth } = useAuth()
  const token = auth?.token
  const [provider, setProvider] = useState<Provider>('xendit')
  const [environment, setEnvironment] = useState<'sandbox' | 'production'>('sandbox')
  const [values, setValues] = useState(emptyValues)
  const [saved, setSaved] = useState<BusinessPaymentGatewayDTO | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [formError, setFormError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const [editing, setEditing] = useState(false)
  const [showSecrets, setShowSecrets] = useState(false)
  const [switchTarget, setSwitchTarget] = useState<Provider | null>(null)
  const radioRefs = useRef<Array<HTMLButtonElement | null>>([])
  const [radioFocus, setRadioFocus] = useState(0)
  function requestGateway(next: Provider) {
    if (saved && saved.provider !== next) setSwitchTarget(next)
    else openGateway(next)
  }

  useEffect(() => {
    if (!token) return
    getBusinessPaymentGateway(token).then(value => { setSaved(value); setRadioFocus(value?.provider === 'midtrans' ? 1 : 0) })
      .catch((err) => setError(err instanceof ApiError ? err.message : 'Gagal memuat konfigurasi gateway.'))
      .finally(() => setLoading(false))
  }, [token])

  const set = (key: keyof typeof values) => (event: ChangeEvent<HTMLInputElement>) => setValues((current) => ({ ...current, [key]: event.target.value }))
  function openGateway(next: Provider) {
    setProvider(next)
    setEnvironment(saved?.provider === next ? saved.environment : 'sandbox')
    setValues(emptyValues)
    setFormError(null)
    setShowSecrets(false)
    setEditing(true)
  }
  function closeGateway() {
    if (saving) return
    setEditing(false)
    setValues(emptyValues)
    setFormError(null)
    setShowSecrets(false)
  }
  async function submit(event: FormEvent) {
    event.preventDefault()
    if (!token || saving) return
    setFormError(null)
    setNotice(null)
    setSaving(true)
    const payload: BusinessPaymentGatewayInput = { provider, environment }
    if (provider === 'xendit') {
      payload.xendit_secret_key = values.xenditSecret
      payload.xendit_callback_token = values.xenditCallback
    } else {
      payload.midtrans_merchant_id = values.midtransMID
      payload.midtrans_client_key = values.midtransClient
      payload.midtrans_server_key = values.midtransServer
    }
    try {
      setSaved(await saveBusinessPaymentGateway(token, payload))
      setValues(emptyValues)
      setEditing(false)
      setShowSecrets(false)
      setRadioFocus(provider === 'midtrans' ? 1 : 0)
      setNotice('Konfigurasi gateway disimpan. Daftarkan URL webhook di dashboard gateway untuk menerima status pembayaran.')
    } catch (err) {
      setFormError(err instanceof ApiError ? err.message : 'Gagal menyimpan konfigurasi gateway.')
    } finally { setSaving(false) }
  }
  const replacing = saved && saved.provider !== provider
  const fieldType = showSecrets ? 'text' : 'password'

  return (
    <div className="payment-gateway-page business-settings-page w-full min-w-0">
      <DashboardPageHeader title="Payment Gateway" description="Pilih satu gateway untuk menerima pembayaran pelanggan." />
      {error && <p role="alert" className="mt-4 text-sm text-red-700 dark:text-red-300">{error}</p>}
      {notice && <p role="status" className="mt-4 rounded-xl bg-emerald-50 p-4 text-sm text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200">{notice}</p>}

      <div className="mt-10 space-y-10">
        <SettingsSection stretch title="Gateway aktif" description="Satu gateway digunakan untuk semua pembayaran bisnis Anda." aside={<GlassPanel>
          <h3 className="mb-6 text-base font-semibold text-ink">Ringkasan konfigurasi</h3>
          <SettingsRow title="Gateway">{loading ? 'Memuat...' : saved ? gatewayInfo[saved.provider].name : 'Belum terhubung'}</SettingsRow>
          <SettingsRow title="Mode akun">{saved ? saved.environment === 'sandbox' ? 'Sandbox' : 'Live' : '—'}</SettingsRow>
          <SettingsRow title="Metode pembayaran">Kartu kredit</SettingsRow>
          <SettingsRow title="Status" stacked><StatusBadge tone={saved ? 'success' : 'neutral'}>{saved ? saved.status === 'verified' ? 'Pembayaran telah terverifikasi' : 'Konfigurasi tersimpan' : 'Belum aktif'}</StatusBadge></SettingsRow>
        </GlassPanel>}>
          <GlassPanel className="flex min-w-0 flex-col">
            {loading ? <p role="status" className="text-sm text-body">Memuat pengaturan pembayaran...</p> : saved ? <>
              <div className="grid grid-cols-[80px_minmax(0,1fr)] items-start gap-4 sm:flex sm:flex-wrap sm:items-center"><GatewayMark provider={saved.provider} /><div className="min-w-0 flex-1"><div className="flex flex-wrap items-center gap-2"><h3 className="text-base font-semibold text-ink">{gatewayInfo[saved.provider].name}</h3><StatusBadge tone={saved.environment === 'sandbox' ? 'warning' : 'success'}>{saved.environment === 'sandbox' ? 'Sandbox' : 'Live'}</StatusBadge></div><p className="mt-2 flex items-start gap-2 text-sm text-body"><CheckCircle2 size={16} aria-hidden="true" className="mt-0.5 shrink-0 text-emerald-700 dark:text-emerald-300" /><span>{saved.status === 'verified' ? 'Pembayaran telah terverifikasi' : 'Konfigurasi tersimpan'}{saved.account_label ? ` - ${saved.account_label}` : ''}</span></p></div><SubmitButton type="button" variant="secondary" onClick={() => openGateway(saved.provider)} className="col-span-2 justify-self-end select-none text-sm">Kelola</SubmitButton></div>
              <div className="mt-6 border-t border-slate-200 pt-6 dark:border-slate-700"><h3 className="text-sm font-semibold text-ink">URL webhook pembayaran</h3><div className="mt-4 flex min-w-0 flex-wrap items-center gap-3"><pre tabIndex={0} aria-label="URL webhook pembayaran" className="min-w-0 flex-1 overflow-x-auto whitespace-pre rounded-xl bg-white/40 p-4 font-mono text-xs leading-6 text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand dark:bg-white/5"><code>{saved.webhook_url}</code></pre><CopyButton text={saved.webhook_url} label="Salin URL" buttonText="Salin URL" /></div><p className="mt-3 text-sm leading-6 text-body">{saved.provider === 'midtrans' ? 'Daftarkan di Midtrans sebagai Payment Notification URL pada mode akun yang sama.' : 'Daftarkan di Xendit untuk event Payment Session Completed dan Expired pada mode akun yang sama.'}</p></div>
            </> : <div className="flex flex-1 flex-col justify-center py-6"><CreditCard size={24} aria-hidden="true" className="text-body" /><h3 className="mt-4 text-base font-semibold text-ink">Belum ada gateway aktif</h3><p className="mt-2 text-sm leading-6 text-body">Pilih Xendit atau Midtrans di bawah untuk menghubungkan akun bisnis Anda.</p><a href="#pilih-gateway" className="mt-4 w-fit select-none rounded-lg text-sm font-semibold text-brand focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand dark:text-blue-300">Pilih gateway <ArrowRight size={16} aria-hidden="true" className="inline" /></a></div>}
          </GlassPanel>
        </SettingsSection>

        <div id="pilih-gateway" className="scroll-mt-6 border-t border-slate-200 pt-10 dark:border-slate-700">
          <SettingsSection stretch title="Pilih gateway" description="Hanya satu gateway yang dapat aktif. Mengganti gateway akan menonaktifkan yang sekarang." aside={<GlassPanel>
            <h3 className="text-base font-semibold text-ink">Langkah berikutnya</h3>
            <ol className="mt-6 space-y-6">{[['Hubungkan akun', 'Pilih gateway dan masukkan kredensial akun bisnis Anda.'], ['Daftarkan webhook', 'Salin URL di halaman ini ke dashboard gateway pada mode akun yang sama.'], ['Periksa pembayaran', 'Coba transaksi, lalu periksa invoice dan notifikasi yang diterima.']].map(([title, description], index) => <li key={title} className="flex gap-4"><span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-brand/10 text-xs font-semibold text-brand dark:text-blue-300">{index + 1}</span><div><p className="text-sm font-semibold text-ink">{title}</p><p className="mt-2 text-sm leading-6 text-body">{description}</p></div></li>)}</ol>
            <div className="mt-auto pt-6"><Link to="/dashboard/pengaturan/webhook-logs" className="inline-flex select-none items-center gap-2 rounded-lg text-sm font-semibold text-brand focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand dark:text-blue-300">Lihat Webhook Logs <ArrowRight size={16} aria-hidden="true" /></Link></div>
          </GlassPanel>}>
            <GlassPanel>
              <div role="radiogroup" aria-label="Payment gateway bisnis" aria-describedby="gateway-choice-help" className="space-y-4">
                {(['xendit', 'midtrans'] as const).map((item, index) => <div key={item} className={`select-none rounded-2xl border p-4 ${saved?.provider === item ? 'border-brand/50 bg-brand/5' : 'border-slate-200 bg-white/30 dark:border-slate-700 dark:bg-white/5'}`}>
                  <button ref={node => { radioRefs.current[index] = node }} type="button" role="radio" aria-checked={saved?.provider === item} aria-label={gatewayInfo[item].name} tabIndex={radioFocus === index ? 0 : -1} disabled={loading} onFocus={() => setRadioFocus(index)} onClick={() => requestGateway(item)} onKeyDown={event => { const direction = ['ArrowRight', 'ArrowDown'].includes(event.key) ? 1 : ['ArrowLeft', 'ArrowUp'].includes(event.key) ? -1 : 0; if (!direction && event.key !== 'Home' && event.key !== 'End') return; event.preventDefault(); const next = event.key === 'Home' ? 0 : event.key === 'End' ? 1 : (index + direction + 2) % 2; setRadioFocus(next); radioRefs.current[next]?.focus() }} className="flex w-full select-none items-start gap-4 rounded-xl text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-brand disabled:opacity-50">
                    <GatewayMark provider={item} compact /><span className="min-w-0 flex-1"><span className="flex flex-wrap items-center gap-2"><span className="text-base font-semibold text-ink">{gatewayInfo[item].name}</span>{saved?.provider === item && <StatusBadge tone="success">Aktif</StatusBadge>}</span><span className="mt-2 block text-sm leading-6 text-body">{gatewayInfo[item].description}</span></span><span aria-hidden="true" className={`mt-2 flex size-5 shrink-0 items-center justify-center rounded-full border ${saved?.provider === item ? 'border-brand' : 'border-slate-400'}`}>{saved?.provider === item && <span className="size-2.5 rounded-full bg-brand" />}</span>
                  </button>
                  <div className="mt-4 flex justify-end"><SubmitButton type="button" variant="secondary" disabled={loading} onClick={() => requestGateway(item)} className="select-none text-sm">{saved?.provider === item ? 'Kelola konfigurasi' : saved ? `Ganti ke ${gatewayInfo[item].name}` : `Hubungkan ${gatewayInfo[item].name}`}</SubmitButton></div>
                </div>)}
              </div>
              <p id="gateway-choice-help" className="mt-4 text-sm leading-6 text-body">Gateway aktif berubah setelah konfigurasi baru berhasil disimpan.</p>
            </GlassPanel>
          </SettingsSection>
        </div>
      </div>

      <ConfirmDialog open={switchTarget !== null} onClose={() => setSwitchTarget(null)} onConfirm={() => { if (switchTarget) { const next = switchTarget; setSwitchTarget(null); openGateway(next) } }} title={`Ganti ke ${switchTarget ? gatewayInfo[switchTarget].name : 'gateway baru'}?`} confirmLabel="Lanjutkan ganti gateway" description={<div><p>Perubahan berlaku setelah konfigurasi baru berhasil disimpan.</p><ul className="mt-4 list-disc space-y-2 pl-5"><li>{saved ? gatewayInfo[saved.provider].name : 'Gateway saat ini'} akan dinonaktifkan sebagai gateway aktif bisnis.</li><li>Daftarkan URL webhook pembayaran di {switchTarget ? gatewayInfo[switchTarget].name : 'gateway baru'} pada mode akun yang sama. Alamat webhook saat ini tetap digunakan.</li><li>Pembayaran yang sedang berjalan di {saved ? gatewayInfo[saved.provider].name : 'gateway lama'} dapat terdampak.</li><li>Kredensial gateway lama akan ditimpa oleh kredensial baru.</li></ul></div>} />

      <Modal open={editing} onClose={closeGateway} title={`${saved?.provider === provider ? 'Kelola' : 'Hubungkan'} ${gatewayInfo[provider].name}`} size="wide" trapFocus>
        <div className="grid gap-7 md:grid-cols-[minmax(0,1.15fr)_minmax(0,1fr)]">
          <form onSubmit={submit} className="min-w-0 space-y-4 md:border-r md:border-border md:pr-7">
            <div className="mb-5 flex items-center gap-3"><GatewayMark provider={provider} /><p className="text-[12px] leading-relaxed text-body">Masukkan kredensial akun {gatewayInfo[provider].name} milik bisnis Anda.</p></div>
            {replacing && <p className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-[12px] text-amber-900">Menyimpan {gatewayInfo[provider].name} akan menggantikan {gatewayInfo[saved.provider].name} sebagai gateway aktif bisnis.</p>}
            <Select id="gateway-environment" label="Mode akun" value={environment} disabled={saving} onChange={(event) => setEnvironment(event.target.value as 'sandbox' | 'production')}><option value="sandbox">Sandbox (pengujian)</option><option value="production">Production (transaksi nyata)</option></Select>
            {provider === 'xendit' ? <>
              <Input id="xendit-secret" label="Secret API Key" type={fieldType} autoComplete="new-password" value={values.xenditSecret} onChange={set('xenditSecret')} disabled={saving} required />
              <p className="!mt-1 text-[11px] leading-relaxed text-body">Ambil dari Settings / Developers / API Keys di dashboard Xendit.</p>
              <Input id="xendit-callback" label="Webhook Verification Token" type={fieldType} autoComplete="new-password" value={values.xenditCallback} onChange={set('xenditCallback')} disabled={saving} required />
              <p className="!mt-1 text-[11px] leading-relaxed text-body">Gunakan verification token dari pengaturan webhook akun Xendit.</p>
            </> : <>
              <Input id="midtrans-mid" label="Merchant ID" autoComplete="off" value={values.midtransMID} onChange={set('midtransMID')} disabled={saving} required />
              <Input id="midtrans-client" label="Client Key" autoComplete="off" value={values.midtransClient} onChange={set('midtransClient')} disabled={saving} required />
              <Input id="midtrans-server" label="Server Key" type={fieldType} autoComplete="new-password" value={values.midtransServer} onChange={set('midtransServer')} disabled={saving} required />
              <p className="!mt-1 text-[11px] leading-relaxed text-body">Ketiga nilai tersedia pada Settings / Access Keys di dashboard Midtrans.</p>
            </>}
            <label className="flex items-center gap-2 text-[12px] text-body"><input type="checkbox" checked={showSecrets} onChange={(event) => setShowSecrets(event.target.checked)} disabled={saving} className="accent-brand" />Tampilkan nilai yang sedang diisi</label>
            <div className="flex items-start gap-2 rounded-xl bg-emerald-50 p-3 text-[11px] leading-relaxed text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200"><ShieldCheck size={16} className="mt-0.5 shrink-0" /><p>Kredensial rahasia disimpan terenkripsi dan tidak ditampilkan kembali setelah disimpan.</p></div>
            {formError && <p role="alert" className="text-[13px] text-red-600">{formError}</p>}
            <div className="flex justify-end gap-2 border-t border-border pt-4"><SubmitButton type="button" variant="secondary" onClick={closeGateway} disabled={saving}>Batal</SubmitButton><SubmitButton type="submit" disabled={saving}>{saving ? 'Menyimpan...' : saved?.provider === provider ? 'Simpan konfigurasi' : 'Hubungkan'}</SubmitButton></div>
          </form>
          <aside className="min-w-0">
            <div className="rounded-2xl bg-brand/5 p-5"><span className="mb-4 flex size-10 items-center justify-center rounded-xl bg-brand/10 text-brand"><KeyRound size={20} /></span><h3 className="text-[15px] font-semibold text-ink">Panduan menghubungkan akun</h3><p className="mt-2 text-[12px] leading-relaxed text-body">Siapkan akun {gatewayInfo[provider].name} dan gunakan kredensial dari mode yang sama.</p>
              <ol className="mt-5 space-y-5">{[
                ['Pilih mode akun', 'Gunakan Sandbox untuk mencoba alur pembayaran sebelum transaksi nyata.'],
                ['Salin kredensial', provider === 'xendit' ? 'Siapkan Secret API Key dan Webhook Verification Token dari dashboard Xendit.' : 'Siapkan Merchant ID, Client Key, dan Server Key dari Access Keys Midtrans.'],
                ['Daftarkan webhook', 'Setelah menyimpan, salin URL webhook yang muncul di halaman ini ke dashboard gateway.'],
                ['Coba pembayaran kartu', 'Lakukan transaksi pengujian dan periksa pembaruan invoice serta Webhook Logs.'],
              ].map(([title, description], index) => <li key={title} className="flex gap-3"><span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-brand/10 text-[11px] font-bold text-brand">{index + 1}</span><div><p className="text-[12px] font-semibold text-ink">{title}</p><p className="mt-1 text-[11px] leading-relaxed text-body">{description}</p></div></li>)}</ol>
            </div>
            <a href={gatewayInfo[provider].documentation} target="_blank" rel="noreferrer" className="mt-5 inline-flex items-center gap-2 text-[12px] font-semibold text-brand hover:underline">Dokumentasi {gatewayInfo[provider].name}<ExternalLink size={14} /></a>
          </aside>
        </div>
      </Modal>
    </div>
  )
}

```

### Frontend/src/components/dashboard/SettingsSection.tsx

```tsx
import { useId, type ReactNode } from 'react'

interface SettingsSectionProps {
  title: string
  description: string
  children: ReactNode
  aside?: ReactNode
  stretch?: boolean
}

export default function SettingsSection({ title, description, children, aside, stretch = false }: SettingsSectionProps) {
  const titleId = useId()
  return (
    <section aria-labelledby={titleId} className={`settings-section grid ${stretch ? 'items-stretch [&>div>.glass]:h-full [&>aside>.glass]:h-full [&>aside>.glass]:flex [&>aside>.glass]:flex-col' : 'items-start'} gap-6 min-[900px]:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] min-[1100px]:grid-cols-[280px_minmax(0,1fr)] min-[1100px]:gap-8 min-[1600px]:gap-12 ${aside ? stretch ? 'min-[1600px]:grid-cols-[280px_minmax(0,720px)_minmax(320px,360px)]' : 'min-[1600px]:grid-cols-[280px_minmax(0,720px)_320px]' : ''}`}>
      <header className="min-w-0 min-[900px]:max-[1099px]:col-span-2">
        <h2 id={titleId} className="text-base font-semibold text-ink">{title}</h2>
        <p className="mt-2 text-sm leading-6 text-body">{description}</p>
      </header>
      <div className="min-w-0 [&>.glass]:shadow-sm">{children}</div>
      {aside && <aside aria-label={`Informasi ${title}`} className="hidden min-w-0 min-[900px]:max-[1099px]:block min-[1600px]:block [&>.glass]:shadow-sm">{aside}</aside>}
    </section>
  )
}

```

### Frontend/src/components/ui/StatusBadge.tsx

```tsx
import type { ReactNode } from 'react'

export default function StatusBadge({ children, tone = 'neutral' }: { children: ReactNode; tone?: 'success' | 'warning' | 'neutral' }) {
  const colors = { success: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-200', warning: 'bg-amber-100 text-amber-900 dark:bg-amber-950 dark:text-amber-200', neutral: 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-200' }
  return <span className={`inline-block rounded-full px-3 py-1 text-xs font-semibold ${colors[tone]}`}>{children}</span>
}

```

### Frontend/src/components/ui/CopyButton.tsx

```tsx
import { useEffect, useState } from 'react'
import { Check, Copy } from 'lucide-react'
import { copyText } from '../../lib/clipboard'

interface CopyButtonProps {
  text: string
  label?: string
  iconOnly?: boolean
  buttonText?: string
}

export default function CopyButton({ text, label = 'Salin kode', iconOnly = false, buttonText = 'Salin' }: CopyButtonProps) {
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState(false)
  useEffect(() => { setCopied(false); setError(false) }, [text])
  useEffect(() => {
    if (!copied) return
    const timer = window.setTimeout(() => setCopied(false), 2000)
    return () => window.clearTimeout(timer)
  }, [copied])
  async function copy() {
    try { await copyText(text); setCopied(true); setError(false) }
    catch { setError(true) }
  }
  return <span className="inline-flex shrink-0 flex-col items-end gap-1">
    <button type="button" aria-label={copied ? `${label}: tersalin` : label} onClick={() => void copy()} className="inline-flex select-none min-h-10 items-center justify-center gap-2 rounded-lg border border-current/20 px-3 text-[12px] font-semibold transition hover:bg-current/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand dark:text-slate-200">
      {copied ? <Check size={16} aria-hidden="true" /> : <Copy size={16} aria-hidden="true" />}
      {!iconOnly && (copied ? 'Tersalin ✓' : buttonText)}
    </button>
    <span aria-live="polite" className={error ? 'max-w-60 text-[12px] text-red-600 dark:text-red-300' : 'sr-only'}>{error ? 'Gagal menyalin. Pilih kode dan salin secara manual.' : copied ? 'Tersalin' : ''}</span>
  </span>
}

```

### Frontend/src/components/ui/ConfirmDialog.tsx

```tsx
import { useRef, type ReactNode } from 'react'
import Modal from './Modal'
import SubmitButton from './SubmitButton'

interface ConfirmDialogProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  description: ReactNode
  confirmLabel: string
  busy?: boolean
  error?: string | null
}

export default function ConfirmDialog({ open, onClose, onConfirm, title, description, confirmLabel, busy = false, error }: ConfirmDialogProps) {
  const cancelRef = useRef<HTMLButtonElement>(null)
  return (
    <Modal open={open} onClose={() => { if (!busy) onClose() }} title={title} trapFocus initialFocusRef={cancelRef}>
      <div className="mt-4 text-sm leading-6 text-body">{description}</div>
      {error && <p role="alert" className="mt-4 text-sm text-red-700 dark:text-red-300">{error}</p>}
      <div className="mt-6 flex flex-wrap justify-end gap-3">
        <button ref={cancelRef} type="button" disabled={busy} onClick={onClose} className="rounded-xl border border-slate-300 px-4 py-2 text-sm font-semibold text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-brand disabled:opacity-50 dark:border-slate-600">Batal</button>
        <SubmitButton type="button" disabled={busy} onClick={onConfirm} className="bg-red-700 text-sm hover:bg-red-800">{busy ? 'Memproses...' : confirmLabel}</SubmitButton>
      </div>
    </Modal>
  )
}

```

### Frontend/src/components/layout/DashboardLayout.tsx

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

        <main className={`flex-1 py-8 md:py-10 ${['/dashboard/integrasi', '/dashboard/pengaturan', '/dashboard/pengaturan/payment-gateway'].includes(location.pathname) ? 'px-6 min-[900px]:px-10' : 'px-5 md:px-10'}`}>
          <Outlet />
        </main>
      </div>
    </div>
  )
}

```

### Frontend/src/components/dashboard/SettingsRow.tsx

```tsx
import type { ReactNode } from 'react'

interface SettingsRowProps {
  title: ReactNode
  description?: ReactNode
  children: ReactNode
  stacked?: boolean
}

export default function SettingsRow({ title, description, children, stacked = false }: SettingsRowProps) {
  return (
    <div className={`grid min-w-0 items-center gap-4 border-b border-slate-200 py-4 first:pt-0 last:border-b-0 last:pb-0 ${stacked ? '' : 'sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]'} dark:border-slate-700`}>
      <div className="min-w-0"><div className="text-sm font-semibold text-ink">{title}</div>{description && <div className="mt-1 break-words text-sm leading-6 text-body">{description}</div>}</div>
      <div className={`min-w-0 break-words text-sm text-ink ${stacked ? '[&>span]:text-center' : 'sm:text-right'}`}>{children}</div>
    </div>
  )
}

```
