import { useCallback, useEffect, useState } from 'react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import { useAuth } from '../../context/AuthContext'
import {
  ApiError, cancelPlatformSubscription, getPlatformInvoices, getPlatformPlan,
  getPlatformSubscription, startPlatformCheckout,
  type PlatformInvoice, type PlatformPlan, type PlatformSubscription,
} from '../../lib/api'
import { formatCurrency } from '../../lib/format'

function date(value: string | null | undefined) {
  return value ? new Date(value).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }) : '—'
}

const subscriptionStatus: Record<PlatformSubscription['status'], string> = {
  not_subscribed: 'Belum berlangganan',
  pending: 'Menunggu pembayaran',
  active: 'Aktif',
  past_due: 'Pembayaran belum selesai',
  cancelled: 'Perpanjangan dihentikan',
}

export default function PlatformBillingPage() {
  const { auth } = useAuth()
  const token = auth?.token
  const [plan, setPlan] = useState<PlatformPlan | null>(null)
  const [subscription, setSubscription] = useState<PlatformSubscription | null>(null)
  const [invoices, setInvoices] = useState<PlatformInvoice[]>([])
  const [consent, setConsent] = useState(false)
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

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

  async function checkout() {
    if (!token || !consent) return
    setBusy(true); setError(null)
    try {
      const result = await startPlatformCheckout(token)
      const url = new URL(result.checkout_url)
      if (url.protocol !== 'https:') throw new Error('Alamat checkout tidak valid.')
      window.location.assign(url.toString())
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal membuka checkout.')
      setBusy(false)
    }
  }

  async function cancel() {
    if (!token || !window.confirm('Hentikan perpanjangan otomatis PayLanggan? Periode yang sudah dibayar tetap tercatat hingga tanggal berakhir.')) return
    setBusy(true); setError(null)
    try {
      await cancelPlatformSubscription(token)
      setNotice('Perpanjangan otomatis telah dihentikan.')
      await refresh()
    } catch (err) { setError(err instanceof ApiError ? err.message : 'Gagal membatalkan langganan.') }
    finally { setBusy(false) }
  }

  if (loading) return <p className="p-6">Memuat tagihan PayLanggan...</p>
  return <div className="space-y-6">
    <DashboardPageHeader title="Tagihan PayLanggan" description="Langganan bisnis Anda untuk menggunakan PayLanggan." />
    {error && <p role="alert" className="rounded-lg bg-red-50 p-4 text-red-700">{error}</p>}
    {notice && <p role="status" className="rounded-lg bg-green-50 p-4 text-green-700">{notice}</p>}
    <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-xl font-semibold">{plan?.name ?? 'PayLanggan Bulanan'}</h2>
      <p className="mt-2 text-2xl font-bold">{formatCurrency(plan?.amount ?? 300000)} <span className="text-base font-normal">/ bulan</span></p>
      <p className="mt-3">Status langganan: <strong>{subscription ? subscriptionStatus[subscription.status] : '—'}</strong></p>
      {subscription?.paid_through && <p>Berlaku sampai: {date(subscription.paid_through)}</p>}
      {subscription?.next_billing_date && <p>Periode tagihan berikutnya: {date(subscription.next_billing_date)}</p>}
      <p>Pembayaran otomatis: {subscription?.auto_renew ? 'Aktif' : 'Belum aktif'}</p>
      {subscription?.checkout_url && <a className="mt-4 inline-block text-blue-700 underline" href={subscription.checkout_url}>Lanjutkan pembayaran</a>}
      {(subscription?.status === 'not_subscribed' || subscription?.status === 'pending' || subscription?.status === 'past_due') && !subscription.checkout_url && <div className="mt-5 space-y-3">
        <label className="flex items-start gap-3"><input type="checkbox" checked={consent} onChange={e => setConsent(e.target.checked)} className="mt-1" />
          <span>Saya menyetujui pembayaran {formatCurrency(plan?.amount ?? 300000)} sekarang dan pendebitan kartu otomatis sebesar {formatCurrency(plan?.amount ?? 300000)} setiap bulan selama langganan aktif. Saya dapat membatalkan perpanjangan otomatis melalui halaman ini.</span>
        </label>
        <button type="button" disabled={!consent || busy} onClick={checkout} className="rounded-lg bg-blue-700 px-5 py-2 text-white disabled:opacity-50">{busy ? 'Memproses...' : 'Lanjut ke pembayaran'}</button>
      </div>}
      {(subscription?.status === 'active' || subscription?.status === 'past_due') && <button type="button" disabled={busy} onClick={cancel} className="mt-5 rounded-lg border border-red-300 px-4 py-2 text-red-700 disabled:opacity-50">Hentikan perpanjangan otomatis</button>}
    </section>
    <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="mb-4 text-xl font-semibold">Riwayat tagihan</h2>
      {invoices.length === 0 ? <p>Belum ada tagihan.</p> : <div className="overflow-x-auto"><table className="w-full text-left text-sm"><thead><tr><th className="p-2">Periode</th><th className="p-2">Jumlah</th><th className="p-2">Status</th><th className="p-2">Dibayar</th></tr></thead><tbody>{invoices.map(item => <tr key={item.id} className="border-t"><td className="p-2">{date(item.period_start)} – {date(item.period_end)}</td><td className="p-2">{formatCurrency(item.amount)}</td><td className="p-2">{item.status === 'paid' ? 'Lunas' : 'Menunggu pembayaran'}</td><td className="p-2">{date(item.paid_at)}</td></tr>)}</tbody></table></div>}
    </section>
  </div>
}
