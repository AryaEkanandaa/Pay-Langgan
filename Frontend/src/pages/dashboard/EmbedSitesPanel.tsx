import { useEffect, useState } from 'react'
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
  const [copied, setCopied] = useState<string | null>(null)
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

  function headerScript(site: EmbedSite) {
    return `<script src="${window.location.origin}/paylanggan.js" data-pl-site="${site.public_id}" data-pl-api-url="${API_BASE_URL}"></script>`
  }

  function buttonCode(planId: number) {
    const plan = plans.find((item) => item.id === planId)
    if (!plan?.public_id) return ''
    return `<button data-pl-checkout data-pl-plan="${plan.public_id}">Langganan ${escapeHtml(plan.name)}</button>`
  }

  async function copy(value: string, key: string) {
    try { await navigator.clipboard.writeText(value); setCopied(key); setError(null) }
    catch { setError('Salin kode secara manual dari kotak kode di bawah.') }
  }

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
    if (!token || busy || !window.confirm('Cabut ID integrasi ini? Semua tombol checkout yang memakai ID ini akan berhenti membuat pesanan.')) return
    setBusy(true); setError(null)
    try {
      await revokeEmbedSite(token, site.id)
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

  return <section>
    <div className="flex flex-wrap items-start justify-between gap-4">
      <div><h2 className="font-display text-[20px] font-bold text-ink">Skrip drop-in</h2>
        <p className="mt-2 text-[13px] leading-relaxed text-body">Gunakan satu ID integrasi untuk memasang tombol checkout pada halaman penjualan Anda. Tidak perlu mendaftarkan nama atau URL website.</p>
      </div>
      {!activeSite && <button type="button" onClick={() => void create()} disabled={busy} className="rounded-lg bg-brand px-4 py-2 text-[13px] font-semibold text-white disabled:opacity-50">{busy ? 'Menyiapkan...' : 'Buat ID integrasi'}</button>}
    </div>
    {error && <p role="alert" className="mt-4 rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</p>}
    {plansLoaded && activeSites.length === 0 && <p className="mt-3 text-[13px] text-body">Buat ID integrasi untuk menampilkan skrip header dan kode tombol checkout.</p>}
    <div className="mt-5 space-y-4">
      {activeSite && [activeSite].map((site) => {
        return <div key={site.id} className="rounded-xl border border-white/70 bg-white/60 p-4">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div><h3 className="font-semibold text-ink">ID integrasi bisnis</h3><p className="mt-1 text-[12px] text-muted">{site.public_id}</p></div>
            {!site.revoked_at && <div className="flex gap-2">
              <button type="button" onClick={() => revoke(site)} disabled={busy} className="rounded-lg border border-red-200 px-3 py-2 text-[12px] font-semibold text-red-700">Cabut</button>
            </div>}
          </div>
          {!site.revoked_at && <div className="mt-4 space-y-4">
            <div className="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,2fr)]">
              <div><h4 className="text-[13px] font-semibold text-ink">1. Skrip header</h4><p className="mt-1 text-[12px] leading-relaxed text-body">Salin ke bagian <code>&lt;head&gt;</code> halaman yang menjual langganan. Cukup pasang satu kali untuk semua plan bisnis.</p></div>
              <div><pre className="overflow-x-auto whitespace-pre-wrap break-all rounded-lg bg-slate-900 p-4 text-[12px] text-slate-100"><code>{headerScript(site)}</code></pre>
                <button type="button" onClick={() => copy(headerScript(site), `script-${site.id}`)} className="mt-2 text-[12px] font-semibold text-brand">{copied === `script-${site.id}` ? 'Tersalin' : 'Salin skrip header'}</button>
              </div>
            </div>
            <div className="grid gap-4 border-t border-slate-200 pt-5 md:grid-cols-[minmax(0,1fr)_minmax(0,2fr)]">
              <div><h4 className="text-[13px] font-semibold text-ink">2. Kode tombol checkout</h4><p className="mt-1 text-[12px] leading-relaxed text-body">Pilih produk dan plan. Salin tombol ke halaman penjualan. Skrip header tetap sama saat plan baru ditambahkan.</p></div>
              <div>
                <div className="grid gap-3 sm:grid-cols-2">
                  <Select id={`embed-product-${site.id}`} label="Produk" value={selectedProductId ?? ''} onChange={(event) => setProductId(Number(event.target.value))} disabled={allowedProducts.length === 0}>
                      {allowedProducts.length === 0 && <option value="">Belum ada produk</option>}
                      {allowedProducts.map((product) => <option key={product.id} value={product.id}>{product.name}</option>)}
                  </Select>
                  <label className="text-[12px] font-semibold text-ink">Jenis checkout
                    <input readOnly value="Langganan" className="mt-1 block w-full rounded-lg border border-slate-300 bg-slate-100 px-3 py-2 text-[13px] font-normal text-body" />
                  </label>
                  <Select id={`embed-plan-${site.id}`} label="Plan" value={selectedPlanId ?? ''} onChange={(event) => setButtonPlanIds((current) => ({ ...current, [site.id]: Number(event.target.value) }))} disabled={productPlans.length === 0}>
                      {productPlans.length === 0 && <option value="">Belum ada plan</option>}
                      {productPlans.map((plan) => <option key={plan.id} value={plan.id}>{plan.name}</option>)}
                  </Select>
                  <label className="text-[12px] font-semibold text-ink">Mata uang dan periode
                    <input readOnly value={selectedPlan ? `IDR · ${selectedPlan.billing_cycle === 'monthly' ? 'Bulanan' : selectedPlan.billing_cycle === 'yearly' ? 'Tahunan' : selectedPlan.billing_cycle}` : '—'} className="mt-1 block w-full rounded-lg border border-slate-300 bg-slate-100 px-3 py-2 text-[13px] font-normal text-body" />
                  </label>
                </div>
                {selectedPlan && <p className="mt-3 text-[13px] text-body">Harga: <strong className="text-ink">{formatCurrency(selectedPlan.price)}</strong></p>}
                {selectedPlanId ? <>
                  <pre className="mt-3 overflow-x-auto whitespace-pre-wrap break-all rounded-lg bg-slate-900 p-4 text-[12px] text-slate-100"><code>{buttonCode(selectedPlanId)}</code></pre>
                  <button type="button" onClick={() => copy(buttonCode(selectedPlanId), `button-${site.id}-${selectedPlanId}`)} className="mt-2 text-[12px] font-semibold text-brand">{copied === `button-${site.id}-${selectedPlanId}` ? 'Tersalin' : 'Salin kode tombol'}</button>
                </> : <p className="mt-3 text-[12px] text-body">Belum ada plan yang mendukung checkout kartu. Buat plan aktif di Katalog terlebih dahulu.</p>}
              </div>
            </div>
          </div>}
        </div>
      })}
    </div>
  </section>
}
