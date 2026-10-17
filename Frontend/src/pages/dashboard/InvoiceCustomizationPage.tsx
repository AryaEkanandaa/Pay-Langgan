import { useEffect, useState, type ChangeEvent, type FormEvent } from 'react'
import { ArrowLeft, ImagePlus, Palette, Upload } from 'lucide-react'
import { Link } from 'react-router-dom'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import GlassPanel from '../../components/ui/GlassPanel'
import Input from '../../components/ui/Input'
import SubmitButton from '../../components/ui/SubmitButton'
import Select from '../../components/ui/Select'
import { useAuth } from '../../context/AuthContext'
import { ApiError, getMyBusiness, updateMyBusiness, previewInvoiceAppearance, type InvoiceAppearancePreview, type BusinessDTO, type JSONMap } from '../../lib/api'

interface InvoiceAppearance {
  logo_url: string
  accent_color: string
  address: string
  footer_note: string
  email_subject: string
  email_message: string
  reply_to: string
  auto_send: boolean
}

const defaultAppearance: InvoiceAppearance = {
  email_subject: 'Invoice {{invoice_number}} - {{business_name}}',
  email_message: 'Halo {{customer_name}}, berikut invoice langganan Anda. Total tagihan {{total}} dengan jatuh tempo {{due_date}}.',
  reply_to: '',
  auto_send: false,
  logo_url: '',
  accent_color: '#3654ff',
  address: '',
  footer_note: 'Terima kasih telah menggunakan layanan kami.',
}

function readAppearance(meta?: JSONMap | null): InvoiceAppearance {
  const raw = meta?.invoice
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return defaultAppearance

  const saved = raw as Record<string, unknown>
  return {
    email_subject: typeof saved.email_subject === 'string' ? saved.email_subject : defaultAppearance.email_subject,
    email_message: typeof saved.email_message === 'string' ? saved.email_message : defaultAppearance.email_message,
    reply_to: typeof saved.reply_to === 'string' ? saved.reply_to : '',
    auto_send: saved.auto_send === true,
    logo_url: typeof saved.logo_url === 'string' ? saved.logo_url : defaultAppearance.logo_url,
    accent_color: typeof saved.accent_color === 'string' ? saved.accent_color : defaultAppearance.accent_color,
    address: typeof saved.address === 'string' ? saved.address : defaultAppearance.address,
    footer_note: typeof saved.footer_note === 'string' ? saved.footer_note : defaultAppearance.footer_note,
  }
}

export default function InvoiceCustomizationPage() {
  const { auth } = useAuth()
  const token = auth?.token
  const [business, setBusiness] = useState<BusinessDTO | null>(null)
  const [appearance, setAppearance] = useState<InvoiceAppearance>(defaultAppearance)
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [preview, setPreview] = useState<InvoiceAppearancePreview | null>(null)
  const [previewKind, setPreviewKind] = useState<'invoice_issued' | 'payment_succeeded'>('invoice_issued')
  const [previewBusy, setPreviewBusy] = useState(false)
  const [previewError, setPreviewError] = useState<string | null>(null)
  const [previewHeight, setPreviewHeight] = useState(720)
  const [previewRetry, setPreviewRetry] = useState(0)
  const [logoFailed, setLogoFailed] = useState(false)

  useEffect(() => {
    if (!token) return
    getMyBusiness(token)
      .then((result) => {
        setBusiness(result)
        setAppearance(readAppearance(result.meta))
      })
      .catch((err) => setError(err instanceof ApiError ? err.message : 'Gagal memuat tampilan invoice.'))
      .finally(() => setIsLoading(false))
  }, [token])

  useEffect(() => {
    if (!token || !business) return
    const controller = new AbortController()
    setPreviewBusy(true)
    setPreviewError(null)
    const timer = window.setTimeout(() => {
      previewInvoiceAppearance(token, { ...appearance }, previewKind, controller.signal)
        .then((result) => { if (!controller.signal.aborted) setPreview(result) })
        .catch((err: unknown) => {
          if (!controller.signal.aborted) setPreviewError(err instanceof ApiError ? err.message : 'Gagal memuat preview email. Silakan coba lagi.')
        })
        .finally(() => { if (!controller.signal.aborted) setPreviewBusy(false) })
    }, 350)
    return () => { window.clearTimeout(timer); controller.abort() }
  }, [token, business, appearance, previewKind, previewRetry])

  function updateAppearance<K extends keyof InvoiceAppearance>(key: K, value: InvoiceAppearance[K]) {
    if (key === 'logo_url') setLogoFailed(false)
    setAppearance((current) => ({ ...current, [key]: value }))
    setSuccess(null)
  }

  function handleLogoUpload(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    if (!file) return
    if (!['image/png', 'image/jpeg'].includes(file.type)) {
      setError('Gunakan logo PNG atau JPG agar kompatibel dengan email.')
      return
    }
    if (file.size > 1024 * 1024) {
      setError('Ukuran logo maksimal 1 MB.')
      return
    }

    const reader = new FileReader()
    reader.onload = () => {
      if (typeof reader.result === 'string') updateAppearance('logo_url', reader.result)
    }
    reader.readAsDataURL(file)
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!token || !business) return

    setIsSaving(true)
    setError(null)
    setSuccess(null)
    try {
      const updated = await updateMyBusiness(token, {
        meta: { ...(business.meta ?? {}), invoice: appearance },
      })
      setBusiness(updated)
      setAppearance(readAppearance(updated.meta))
      setSuccess('Tampilan invoice berhasil disimpan.')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal menyimpan tampilan invoice.')
    } finally {
      setIsSaving(false)
    }
  }

  if (!auth) return null
  const hasUnsavedChanges = business !== null && JSON.stringify(appearance) !== JSON.stringify(readAppearance(business.meta))

  return (
    <div>
      <Link to="/dashboard/pengaturan" className="mb-5 inline-flex items-center gap-2 text-[13px] font-semibold text-brand hover:underline">
        <ArrowLeft size={16} /> Kembali ke Pengaturan Bisnis
      </Link>
      <DashboardPageHeader
        title="Tampilan Invoice"
        description="Sesuaikan invoice dan email yang dikirim PayLanggan kepada pelanggan bisnis Anda."
      />

      {error && <p className="mb-4 text-[13px] text-red-600">{error}</p>}
      {success && <p className="mb-4 text-[13px] text-emerald-700">{success}</p>}

      {isLoading ? (
        <GlassPanel className="max-w-6xl"><p className="text-[13px] text-body">Memuat pengaturan invoice...</p></GlassPanel>
      ) : (
        <div className="grid max-w-6xl gap-6 xl:grid-cols-[minmax(0,0.9fr)_minmax(420px,1.1fr)]">
          <GlassPanel>
            <div className="mb-6 flex items-center gap-3">
              <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-brand/10 text-brand"><Palette size={19} /></span>
              <div>
                <p className="text-[14px] font-semibold text-ink">Identitas Invoice</p>
                <p className="mt-1 text-[13px] text-muted">Pengaturan yang akan digunakan pada dokumen tagihan.</p>
              </div>
            </div>

            <form onSubmit={handleSubmit} className="space-y-5">
              <div>
                <p className="text-[13px] font-medium text-body">Logo bisnis</p>
                <div className="mt-2 flex items-center gap-3">
                  <div className="flex h-16 w-16 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-white/60 bg-white/40 text-muted">
                    {appearance.logo_url && !logoFailed ? <img src={appearance.logo_url} alt="Preview logo bisnis" onError={() => setLogoFailed(true)} className="h-full w-full object-contain p-2" /> : <ImagePlus size={22} />}
                  </div>
                  <label className="glass inline-flex cursor-pointer items-center gap-2 rounded-lg px-3 py-2 text-[12px] font-semibold text-ink transition-colors hover:bg-white/70">
                    <Upload size={15} /> Unggah Logo
                    <input type="file" accept="image/png,image/jpeg" onChange={handleLogoUpload} className="sr-only" />
                  </label>
                </div>
                <Input
                  id="invoice_logo_url"
                  label="atau URL Logo"
                  type="url"
                  value={appearance.logo_url.startsWith('data:') ? '' : appearance.logo_url}
                  onChange={(event) => updateAppearance('logo_url', event.target.value)}
                  placeholder="https://contoh.com/logo.png"
                  className="mt-3"
                />
                <p className="mt-1 text-[11px] text-muted">Format PNG atau JPG, maksimal 1 MB. Logo email memerlukan alamat gambar HTTPS permanen yang dapat dibuka tanpa login.</p>
              </div>

              <label className="block text-left" htmlFor="invoice_accent_color">
                <span className="text-[13px] font-medium text-body">Warna aksen</span>
                <span className="mt-1.5 flex h-11 items-center gap-3 rounded-xl border border-white/60 bg-white/50 px-3">
                  <input id="invoice_accent_color" type="color" value={appearance.accent_color} onChange={(event) => updateAppearance('accent_color', event.target.value)} className="h-7 w-9 cursor-pointer rounded border-0 bg-transparent p-0" />
                  <span className="text-[13px] font-medium uppercase text-ink">{appearance.accent_color}</span>
                </span>
              </label>

              <label className="block text-left" htmlFor="invoice_address">
                <span className="text-[13px] font-medium text-body">Alamat bisnis</span>
                <textarea id="invoice_address" value={appearance.address} onChange={(event) => updateAppearance('address', event.target.value)} rows={3} placeholder="Alamat yang tampil pada invoice" className="mt-1.5 w-full resize-none rounded-xl border border-white/60 bg-white/50 px-4 py-2.5 text-[14px] text-ink outline-none backdrop-blur-xl transition-colors placeholder:text-muted focus:border-brand/50 focus:bg-white/70 focus:ring-2 focus:ring-brand/30" />
              </label>

              <label className="block text-left" htmlFor="invoice_footer_note">
                <span className="text-[13px] font-medium text-body">Catatan footer invoice</span>
                <textarea id="invoice_footer_note" value={appearance.footer_note} onChange={(event) => updateAppearance('footer_note', event.target.value)} rows={2} placeholder="Catatan atau ucapan terima kasih" className="mt-1.5 w-full resize-none rounded-xl border border-white/60 bg-white/50 px-4 py-2.5 text-[14px] text-ink outline-none backdrop-blur-xl transition-colors placeholder:text-muted focus:border-brand/50 focus:bg-white/70 focus:ring-2 focus:ring-brand/30" />
              </label>

              <div className="border-t border-border pt-5">
                <h2 className="text-[14px] font-semibold text-ink">Email invoice</h2>
                <p className="mt-1 text-[12px] text-muted">Pengirim menggunakan alamat PayLanggan. Saat mode preview, email disiapkan secara lokal tanpa dikirim.</p>
              </div>
              <Input id="invoice_email_subject" label="Subjek email" value={appearance.email_subject} maxLength={200} onChange={(event) => updateAppearance('email_subject', event.target.value)} />
              <p className="text-[12px] text-body">Subjek dan pesan ini digunakan untuk email invoice. Pesan konfirmasi pembayaran dibuat otomatis setelah pembayaran berhasil.</p>
              <label className="block text-[13px] font-medium text-body" htmlFor="invoice_email_message">Pesan email
                <textarea id="invoice_email_message" rows={4} maxLength={3000} value={appearance.email_message} onChange={(event) => updateAppearance('email_message', event.target.value)} className="mt-1.5 w-full rounded-xl border border-white/60 bg-white/50 px-4 py-2.5 text-ink outline-none focus:ring-2 focus:ring-brand/30" />
              </label>
              <p className="text-[11px] text-muted">Variabel: {'{{business_name}}, {{customer_name}}, {{invoice_number}}, {{total}}, {{due_date}}'}.</p>
              <Input id="invoice_reply_to" label="Alamat email balasan bisnis (opsional)" type="email" value={appearance.reply_to} onChange={(event) => updateAppearance('reply_to', event.target.value)} />
              <label className="flex items-start gap-3 text-[13px] text-body"><input type="checkbox" checked={appearance.auto_send} onChange={(event) => updateAppearance('auto_send', event.target.checked)} className="mt-1 accent-brand" /><span>Siapkan email otomatis ketika invoice baru diterbitkan.<span className="mt-1 block text-[11px] text-muted">Invoice lama tidak dikirim otomatis. Pengiriman sungguhan mengikuti konfigurasi SMTP PayLanggan.</span></span></label>
              <SubmitButton type="submit" className="w-full" disabled={isSaving}>
                {isSaving ? 'Menyimpan...' : 'Simpan Tampilan Invoice'}
              </SubmitButton>
            </form>
          </GlassPanel>

          <GlassPanel className="self-start overflow-hidden p-0" aria-busy={previewBusy}>
            <div className="border-b border-white/50 px-6 py-5">
              <h2 className="text-[14px] font-semibold text-ink">Preview email</h2>
              <p className="mt-1 text-[13px] text-body">Menggunakan template yang sama dengan email yang dikirim. Data tagihan di bawah adalah contoh. {hasUnsavedChanges ? 'Simpan perubahan agar digunakan pada email baru.' : 'Pengaturan tersimpan digunakan untuk email baru.'}</p>
              <div className="mt-4"><Select label="Jenis email" value={previewKind} onChange={(event) => setPreviewKind(event.target.value as 'invoice_issued' | 'payment_succeeded')}>
                <option value="invoice_issued">Email invoice</option>
                <option value="payment_succeeded">Konfirmasi pembayaran</option>
              </Select></div>
              <p className="mt-3 text-xs text-body" aria-live="polite">{previewBusy ? 'Memperbarui preview...' : previewError ? 'Preview belum diperbarui.' : 'Preview mengikuti pengaturan saat ini.'}</p>
              {previewError && <div role="alert" className="mt-3 text-sm text-red-700 dark:text-red-300">{previewError}<button type="button" className="ml-2 rounded font-semibold underline focus-visible:outline-2 focus-visible:outline-brand" onClick={() => setPreviewRetry((value) => value + 1)}>Coba lagi</button></div>}
              {preview?.logo_warning && <p className="mt-3 rounded-xl bg-amber-50 p-3 text-xs text-amber-900 dark:bg-amber-950 dark:text-amber-200">{preview.logo_warning}</p>}
              {preview && <p className="mt-3 break-words text-[13px] text-body"><span className="font-semibold">Subjek: </span>{preview.subject}</p>}
            </div>
            {preview && <iframe
              title="Preview email invoice pelanggan"
              sandbox="allow-same-origin"
              referrerPolicy="no-referrer"
              srcDoc={preview.html}
              className="block w-full border-0 bg-white"
              style={{ height: previewHeight }}
              onLoad={(event) => {
                const body = event.currentTarget.contentDocument?.body
                if (body) setPreviewHeight(Math.max(440, body.scrollHeight + 48))
              }}
            />}
          </GlassPanel>
        </div>
      )}
    </div>
  )
}
