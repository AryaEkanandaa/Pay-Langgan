import { useEffect, useMemo, useState } from 'react'
import { Eye, LifeBuoy, Trash2 } from 'lucide-react'
import CodeBlock from '../../components/ui/CodeBlock'
import CopyButton from '../../components/ui/CopyButton'
import GlassPanel from '../../components/ui/GlassPanel'
import ConfirmDialog from '../../components/ui/ConfirmDialog'
import InstallationHelp from '../../components/dashboard/InstallationHelp'
import SubmitButton from '../../components/ui/SubmitButton'
import {
  API_BASE_URL, createEmbedSite, listEmbedSites, listPlans, listProducts, listAddOns, revokeEmbedSite,
  type EmbedSite, type Plan, type Product, type AddOn, type CheckoutAddOnSelection,
} from '../../lib/api'
import { formatCurrency } from '../../lib/format'
import { useAuth } from '../../context/AuthContext'
import Select from '../../components/ui/Select'
import Input from '../../components/ui/Input'

// Bagian tambahan disembunyikan sementara; aktifkan kembali saat diperlukan.
const SHOW_SUPPLEMENTARY_INTEGRATION_SECTIONS = false

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
  const [addOnCatalog, setAddOnCatalog] = useState<AddOn[]>([])
  const [buttonAddOns, setButtonAddOns] = useState<CheckoutAddOnSelection[]>([])
  const [buttonCoupon, setButtonCoupon] = useState('')
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
    async function loadAddOnCatalog() {
      const first = await listAddOns(token!, { page: 1, limit: 100 })
      const items = [...first.items]
      for (let page = 2; items.length < first.total; page++) {
        const next = await listAddOns(token!, { page, limit: 100 }); if (!next.items.length) break; items.push(...next.items)
      }
      return items
    }
    void Promise.allSettled([listEmbedSites(token), loadPlans(token), loadProducts(token), loadAddOnCatalog()]).then(([siteResult, planResult, productResult, addOnResult]) => {
      if (!active) return
      if (siteResult.status === 'fulfilled') {
        setSites(siteResult.value)
        setSiteId(siteResult.value.find((site) => !site.revoked_at)?.id ?? null)
      }
      if (planResult.status === 'fulfilled') setPlans(planResult.value)
      if (productResult.status === 'fulfilled') setProducts(productResult.value)
      if (addOnResult.status === 'fulfilled') setAddOnCatalog(addOnResult.value)
      setPlansLoaded(true)
      const failures = [
        addOnResult.status === 'rejected' && 'Gagal memuat katalog add-on',
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
  const availablePlans = activeSite ? plans.filter((plan) => plan.trial_days >= 0 && plan.price >= 5000 && plan.price <= 200000000 && Number.isInteger(plan.price) && ['monthly', 'yearly'].includes(plan.billing_cycle) && products.some((product) => product.id === plan.product_id && product.status === 'active')) : []
  const allowedProducts = products.filter((product) => availablePlans.some((plan) => plan.product_id === product.id))
  const selectedProductId = allowedProducts.some((product) => product.id === productId) ? productId : allowedProducts[0]?.id
  const productPlans = availablePlans.filter((plan) => plan.product_id === selectedProductId)
  const selectedPlanId = productPlans.some((plan) => plan.id === buttonPlanIds[activeSite?.id ?? 0])
    ? buttonPlanIds[activeSite?.id ?? 0]
    : productPlans[0]?.id
  const selectedPlan = plans.find((plan) => plan.id === selectedPlanId)
  const compatibleAddOns = addOnCatalog.filter((item) => item.visible_in_checkout && item.product_id === selectedPlan?.product_id && item.billing_cycle === selectedPlan?.billing_cycle)
  const selectedAddOns = useMemo(() => buttonAddOns.filter((item) => addOnCatalog.some((addon) => addon.visible_in_checkout && addon.id === item.add_on_id && addon.product_id === selectedPlan?.product_id && addon.billing_cycle === selectedPlan?.billing_cycle)), [buttonAddOns, addOnCatalog, selectedPlan])

  const headerCode = useMemo(() => activeSite ? `<script
  src="${window.location.origin}/paylanggan.js"
  data-pl-site="${escapeHtml(activeSite.public_id)}"
  data-pl-api-url="${escapeHtml(API_BASE_URL)}">
</script>` : '', [activeSite])
  const checkoutCode = useMemo(() => selectedPlan?.public_id ? `<button
  data-pl-checkout
  data-pl-plan="${escapeHtml(selectedPlan.public_id)}"${selectedAddOns.length ? `\n  data-pl-add-ons="${escapeHtml(JSON.stringify(selectedAddOns))}"` : ''}${buttonCoupon.trim() ? `\n  data-pl-coupon="${escapeHtml(buttonCoupon.trim())}"` : ''}>
  Langganan ${escapeHtml(selectedPlan.name)}
</button>` : '', [selectedPlan, selectedAddOns, buttonCoupon])
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
            {selectedPlan && <div className="mt-4 space-y-4">
              {selectedPlan.trial_days > 0 && <p className="rounded-xl bg-brand/10 p-3 text-sm text-ink">Trial gratis {selectedPlan.trial_days} hari. Pelanggan tidak ditagih saat mendaftar.</p>}
              {compatibleAddOns.length > 0 && <fieldset className="space-y-3"><legend className="mb-2 text-sm font-semibold text-ink">Add-on awal (opsional)</legend>{compatibleAddOns.map((addon) => {
                const chosen = selectedAddOns.find((item) => item.add_on_id === addon.id)
                return <div key={addon.id} className="flex items-center gap-3"><label className="flex min-w-0 flex-1 items-center gap-3 text-sm text-ink"><input type="checkbox" checked={Boolean(chosen)} onChange={(event) => setButtonAddOns((items) => event.target.checked ? [...items.filter((item) => item.add_on_id !== addon.id), { add_on_id: addon.id, quantity: 1 }] : items.filter((item) => item.add_on_id !== addon.id))} className="size-4 accent-brand" />{addon.name} · {formatCurrency(addon.price)}</label>{chosen && <input type="number" min={1} max={100} aria-label={`Jumlah ${addon.name}`} value={chosen.quantity} onChange={(event) => setButtonAddOns((items) => items.map((item) => item.add_on_id === addon.id ? { ...item, quantity: Math.max(1, Math.min(100, Number(event.target.value) || 1)) } : item))} className="w-20 rounded-lg border border-border bg-transparent px-3 py-2 text-sm text-ink" />}</div>
              })}</fieldset>}
              <Input id="embed-coupon" label="Kode kupon awal (opsional)" value={buttonCoupon} maxLength={50} onChange={(event) => setButtonCoupon(event.target.value)} />
              <p className="text-xs text-body">Pelanggan dapat menyesuaikan add-on dan kupon di langkah Paket. Total akhir diperiksa oleh backend.</p>
            </div>}
            <dl className="my-4 flex flex-wrap gap-3 text-[12px] text-body">{[['Jenis checkout', 'Langganan'], ['Mata uang', 'IDR'], ['Periode', selectedPlan ? period : '-']].map(([label, value]) => <div key={label} className="rounded-lg border border-border bg-white/40 px-3 py-2"><dt className="text-[11px]">{label}</dt><dd className="mt-1 font-semibold text-ink">{value}</dd></div>)}</dl>
            {selectedPlan && <p className="mb-4 text-[13px] text-body">Harga: <strong className="font-display text-[22px] font-bold text-ink">{formatCurrency(selectedPlan.price)}</strong><span className="ml-1">/ {selectedPlan.billing_cycle === 'monthly' ? 'bulan' : 'tahun'}</span></p>}
            {checkoutCode ? <CodeBlock code={checkoutCode} label="kode tombol checkout" /> : <p className="rounded-xl bg-white/40 p-4 text-[13px] leading-6 text-body">Belum ada plan yang mendukung checkout kartu. Buat plan aktif di Katalog terlebih dahulu.</p>}
          </GlassPanel>
        </div>
        {SHOW_SUPPLEMENTARY_INTEGRATION_SECTIONS && <GlassPanel>
          <h2 className="text-[16px] font-semibold text-ink">Portal pelanggan</h2>
          <p className="my-3 text-[13px] leading-6 text-body">Pelanggan masuk melalui tautan email untuk melihat langganan, invoice, dan melanjutkan pembayaran. Gunakan skrip header yang sama, lalu pasang tombol ini di website Anda.</p>
          <CodeBlock code={'<button data-pl-portal>Kelola langganan saya</button>'} label="tombol portal pelanggan" />
          <a className="mt-4 inline-flex rounded text-[13px] font-semibold text-brand focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand dark:text-blue-300" href={`/portal/${encodeURIComponent(activeSite.public_id)}`} target="_blank" rel="noopener noreferrer">Buka portal pelanggan →</a>
        </GlassPanel>}
        </> : <GlassPanel><h2 className="text-[16px] font-semibold text-ink">Skrip drop-in</h2><p className="mt-2 text-[13px] leading-6 text-body">Buat ID integrasi untuk menampilkan skrip header dan kode tombol checkout.</p></GlassPanel>}
          {activeSite && (
      <GlassPanel className="space-y-4">
        <div><h2 className="flex items-center gap-2 text-[16px] font-semibold text-ink"><Trash2 size={18} aria-hidden="true" />Cabut ID integrasi</h2><p className="mt-3 text-[13px] leading-6 text-body">Semua tombol checkout yang memakai ID ini akan berhenti membuat pesanan.</p></div>
        <button type="button" onClick={() => setRevokeTarget(activeSite)} disabled={busy} className="inline-flex min-h-10 items-center gap-2 rounded-lg border border-red-300 px-4 py-2 text-[13px] font-semibold text-red-700 hover:bg-red-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600 dark:text-red-300 dark:hover:bg-red-950/40"><Trash2 size={16} aria-hidden="true" />Cabut</button>
      </GlassPanel>
          )}
        </div>
        <aside className="min-w-0 space-y-6 min-[1100px]:sticky min-[1100px]:top-6" aria-label="ID integrasi dan preview">
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
          {SHOW_SUPPLEMENTARY_INTEGRATION_SECTIONS && <>
          <InstallationHelp headerCode={headerCode} checkoutCode={checkoutCode} />
          <GlassPanel>
            <h2 className="flex items-center gap-2 text-[16px] font-semibold text-ink"><LifeBuoy size={18} aria-hidden="true" />Butuh bantuan?</h2>
            <p className="mt-3 text-[13px] leading-6 text-body">Harga dan informasi plan mengikuti data katalog PayLanggan. Tambahkan tombol lain untuk setiap plan yang ingin dijual.</p>
            {/* TODO: Replace placeholders with published documentation and support URLs. */}
            <div className="mt-4 flex flex-wrap gap-4 text-[13px] font-semibold text-brand dark:text-blue-300"><a href="#dokumentasi-drop-in" onClick={(event) => event.preventDefault()} className="rounded focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand">Dokumentasi (segera)</a><a href="#dukungan-paylanggan" onClick={(event) => event.preventDefault()} className="rounded focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand">Kontak dukungan (segera)</a></div>
          </GlassPanel>
          </>}
        </aside>
      </div>
    <ConfirmDialog open={revokeTarget !== null} onClose={() => setRevokeTarget(null)} onConfirm={() => { if (revokeTarget) void revoke(revokeTarget) }} title="Cabut ID integrasi ini?" description={`Semua tombol checkout yang memakai ID ini akan berhenti membuat pesanan. ${revokeTarget?.public_id ?? ''}`} confirmLabel="Cabut ID integrasi" busy={busy} error={error} />
  </section>
}
