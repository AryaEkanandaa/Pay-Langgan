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
