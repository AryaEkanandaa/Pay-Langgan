import { useEffect, useState, type ChangeEvent, type FormEvent } from 'react'
import { ArrowLeft, ImagePlus, Palette, Upload } from 'lucide-react'
import { Link } from 'react-router-dom'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import GlassPanel from '../../components/ui/GlassPanel'
import Input from '../../components/ui/Input'
import SubmitButton from '../../components/ui/SubmitButton'
import { useAuth } from '../../context/AuthContext'
import { ApiError, getMyBusiness, updateMyBusiness, type BusinessDTO, type JSONMap } from '../../lib/api'
import { formatCurrency } from '../../lib/format'

interface InvoiceAppearance {
  logo_url: string
  accent_color: string
  address: string
  footer_note: string
}

const defaultAppearance: InvoiceAppearance = {
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

  function updateAppearance<K extends keyof InvoiceAppearance>(key: K, value: InvoiceAppearance[K]) {
    setAppearance((current) => ({ ...current, [key]: value }))
    setSuccess(null)
  }

  function handleLogoUpload(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    if (!file) return
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

  return (
    <div>
      <Link to="/dashboard/pengaturan" className="mb-5 inline-flex items-center gap-2 text-[13px] font-semibold text-brand hover:underline">
        <ArrowLeft size={16} /> Kembali ke Pengaturan Bisnis
      </Link>
      <DashboardPageHeader
        title="Tampilan Invoice"
        description="Sesuaikan identitas visual yang tampil pada invoice bisnis Anda."
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
                    {appearance.logo_url ? <img src={appearance.logo_url} alt="Preview logo bisnis" className="h-full w-full object-contain p-2" /> : <ImagePlus size={22} />}
                  </div>
                  <label className="glass inline-flex cursor-pointer items-center gap-2 rounded-lg px-3 py-2 text-[12px] font-semibold text-ink transition-colors hover:bg-white/70">
                    <Upload size={15} /> Unggah Logo
                    <input type="file" accept="image/png,image/jpeg,image/webp,image/svg+xml" onChange={handleLogoUpload} className="sr-only" />
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
                <p className="mt-1 text-[11px] text-muted">Format gambar PNG, JPG, WEBP, atau SVG. Maksimal 1 MB.</p>
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

              <SubmitButton type="submit" className="w-full" disabled={isSaving}>
                {isSaving ? 'Menyimpan...' : 'Simpan Tampilan Invoice'}
              </SubmitButton>
            </form>
          </GlassPanel>

          <GlassPanel className="self-start overflow-hidden p-0">
            <div className="border-b border-white/50 px-6 py-5">
              <p className="text-[14px] font-semibold text-ink">Preview Invoice</p>
              <p className="mt-1 text-[13px] text-muted">Contoh tampilan invoice pelanggan.</p>
            </div>
            <div className="bg-white/55 p-6 text-ink dark:bg-white/5">
              <div className="flex items-start justify-between gap-5 border-b border-slate-200 pb-5">
                <div className="flex items-center gap-3">
                  {appearance.logo_url ? <img src={appearance.logo_url} alt="Logo bisnis" className="h-10 w-10 object-contain" /> : <span className="flex h-10 w-10 items-center justify-center rounded-lg text-white" style={{ backgroundColor: appearance.accent_color }}>P</span>}
                  <div>
                    <p className="font-display text-[16px] font-bold">{business?.name || 'Nama Bisnis'}</p>
                    <p className="mt-1 max-w-[190px] whitespace-pre-line text-[11px] text-slate-500">{appearance.address || 'Alamat bisnis Anda'}</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-[11px] font-semibold uppercase tracking-wide" style={{ color: appearance.accent_color }}>Invoice</p>
                  <p className="mt-1 text-[12px] font-semibold">INV-202608-0001</p>
                </div>
              </div>
              <div className="flex items-start justify-between gap-5 py-5">
                <div><p className="text-[10px] uppercase tracking-wide text-slate-500">Ditagihkan kepada</p><p className="mt-1 text-[13px] font-semibold">Arya Corp</p><p className="text-[11px] text-slate-500">customer@example.com</p></div>
                <span className="rounded-full px-2.5 py-1 text-[10px] font-semibold text-white" style={{ backgroundColor: appearance.accent_color }}>Lunas</span>
              </div>
              <div className="rounded-lg border border-slate-200 bg-white/70 p-3 text-[12px]">
                <div className="flex justify-between"><span className="text-slate-500">Paket Premium</span><span className="font-semibold">{formatCurrency(150000)}</span></div>
                <div className="mt-2 flex justify-between"><span className="text-slate-500">Diskon</span><span className="font-semibold text-emerald-600">- {formatCurrency(25000)}</span></div>
                <div className="mt-3 flex justify-between border-t border-slate-200 pt-3 text-[14px] font-bold"><span>Total</span><span style={{ color: appearance.accent_color }}>{formatCurrency(125000)}</span></div>
              </div>
              <p className="mt-6 text-center text-[11px] text-slate-500">{appearance.footer_note || '—'}</p>
            </div>
          </GlassPanel>
        </div>
      )}
    </div>
  )
}
