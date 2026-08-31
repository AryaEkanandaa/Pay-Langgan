import { useCallback, useEffect, useState } from 'react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import Select from '../../components/ui/Select'
import TableSkeleton from '../../components/ui/skeletons/TableSkeleton'
import { useAuth } from '../../context/AuthContext'
import {
  ApiError,
  cancelSubscription,
  listCustomers,
  listPlans,
  listSubscriptions,
  pauseSubscription,
  resumeSubscription,
  type Customer,
  type Plan,
  type Subscription,
  type SubscriptionStatus,
} from '../../lib/api'
import { formatCurrency, subscriptionStatusLabel } from '../../lib/format'

const PAGE_SIZE = 10

const statusClass: Record<SubscriptionStatus, string> = {
  trial: 'bg-violet-100 text-violet-700',
  active: 'bg-emerald-100 text-emerald-700',
  paused: 'bg-amber-100 text-amber-700',
  cancelled: 'bg-red-100 text-red-700',
}

function formatDate(value: string | null) {
  return value ? new Date(value).toLocaleDateString('id-ID') : '—'
}

export default function LanggananPage() {
  const { auth } = useAuth()
  const token = auth?.token

  const [subscriptions, setSubscriptions] = useState<Subscription[]>([])
  const [customers, setCustomers] = useState<Customer[]>([])
  const [plans, setPlans] = useState<Plan[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<SubscriptionStatus | undefined>()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionID, setActionID] = useState<number | null>(null)

  const loadSubscriptions = useCallback(async () => {
    if (!token) return
    setIsLoading(true)
    setError(null)
    try {
      const [subscriptionResult, customerResult, planResult] = await Promise.all([
        listSubscriptions(token, {
          page,
          limit: PAGE_SIZE,
          search: search || undefined,
          status,
        }),
        listCustomers(token, { limit: 100 }),
        listPlans(token, { limit: 100 }),
      ])
      setSubscriptions(subscriptionResult.items)
      setTotal(subscriptionResult.total)
      setCustomers(customerResult.items)
      setPlans(planResult.items)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memuat data langganan.')
    } finally {
      setIsLoading(false)
    }
  }, [token, page, search, status])

  useEffect(() => {
    loadSubscriptions()
  }, [loadSubscriptions])

  function customerName(customerID: number) {
    return customers.find((customer) => customer.id === customerID)?.name ?? `Pelanggan #${customerID}`
  }

  function planName(planID: number) {
    return plans.find((plan) => plan.id === planID)?.name ?? `Plan #${planID}`
  }

  async function handleAction(subscription: Subscription, action: 'pause' | 'resume' | 'cancel') {
    if (!token) return

    let reason = ''
    if (action === 'cancel' || action === 'pause') {
      const label = action === 'cancel' ? 'pembatalan' : 'penundaan'
      const value = window.prompt(`Masukkan alasan ${label} (opsional):`)
      if (value === null) return
      reason = value
    }

    setActionID(subscription.id)
    setError(null)
    try {
      if (action === 'cancel') {
        await cancelSubscription(token, subscription.id, reason)
      } else if (action === 'pause') {
        await pauseSubscription(token, subscription.id, reason)
      } else {
        await resumeSubscription(token, subscription.id)
      }
      await loadSubscriptions()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memperbarui status langganan.')
    } finally {
      setActionID(null)
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  if (!auth) return null

  return (
    <div>
      <div className="mb-8">
        <DashboardPageHeader
          title="Langganan"
          description="Pantau status dan siklus hidup langganan pelanggan."
        />
      </div>

      <div className="mb-4 flex flex-wrap items-center gap-3">
        <input
          type="search"
          placeholder="Cari pelanggan atau email..."
          value={search}
          onChange={(event) => {
            setPage(1)
            setSearch(event.target.value)
          }}
          className="w-full max-w-xs rounded-xl border border-white/60 bg-white/50 px-4 py-2.5 text-[14px] text-ink outline-none backdrop-blur-xl transition-colors placeholder:text-muted focus:border-brand/50 focus:bg-white/70 focus:ring-2 focus:ring-brand/30"
        />
        <Select
          aria-label="Filter status langganan"
          label="Filter status langganan"
          hideLabel
          value={status ?? ''}
          onChange={(event) => {
            setPage(1)
            setStatus((event.target.value || undefined) as SubscriptionStatus | undefined)
          }}
          className="w-full rounded-xl border border-white/60 bg-white/50 px-4 py-2.5 text-[14px] text-ink outline-none backdrop-blur-xl focus:border-brand/50 focus:ring-2 focus:ring-brand/30 sm:w-[180px]"
        >
          <option value="">Semua status</option>
          <option value="trial">Trial</option>
          <option value="active">Aktif</option>
          <option value="paused">Dijeda</option>
          <option value="cancelled">Dibatalkan</option>
        </Select>
      </div>

      {error && <p className="mb-4 text-[13px] text-red-600">{error}</p>}

      {isLoading ? (
        <TableSkeleton columns={['Pelanggan', 'Plan', 'Status', 'Mulai', 'Tagihan Berikutnya', '']} />
      ) : subscriptions.length === 0 ? (
        <div className="glass flex min-h-[200px] items-center justify-center rounded-2xl text-center">
          <p className="text-[14px] font-medium text-body">
            {search || status ? 'Tidak ada langganan yang cocok.' : 'Belum ada langganan.'}
          </p>
        </div>
      ) : (
        <div className="glass overflow-hidden rounded-2xl">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[960px] text-left text-[13px]">
              <thead>
                <tr className="border-b border-white/50">
                  {['Pelanggan', 'Plan', 'Status', 'Mulai', 'Tagihan Berikutnya', 'Aksi'].map((heading) => (
                    <th
                      key={heading}
                      className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted"
                    >
                      {heading}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {subscriptions.map((subscription) => (
                  <tr key={subscription.id} className="border-b border-white/30 last:border-0">
                    <td className="px-5 py-4 font-medium text-ink">
                      {customerName(subscription.customer_id)}
                    </td>
                    <td className="px-5 py-4 text-body">
                      <p>{planName(subscription.plan_id)}</p>
                      {plans.find((plan) => plan.id === subscription.plan_id) && (
                        <p className="mt-1 text-[12px] text-muted">
                          {formatCurrency(plans.find((plan) => plan.id === subscription.plan_id)?.price ?? 0)}
                        </p>
                      )}
                    </td>
                    <td className="px-5 py-4">
                      <span
                        className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${statusClass[subscription.status]}`}
                      >
                        {subscriptionStatusLabel[subscription.status]}
                      </span>
                    </td>
                    <td className="px-5 py-4 text-body">{formatDate(subscription.start_date)}</td>
                    <td className="px-5 py-4 text-body">
                      {formatDate(subscription.next_billing_date)}
                    </td>
                    <td className="px-5 py-4">
                      <div className="flex items-center justify-end gap-3">
                        {subscription.status === 'paused' && (
                          <button
                            type="button"
                            onClick={() => handleAction(subscription, 'resume')}
                            disabled={actionID === subscription.id}
                            className="text-[13px] font-semibold text-brand hover:underline disabled:opacity-50"
                          >
                            Lanjutkan
                          </button>
                        )}
                        {(subscription.status === 'active' || subscription.status === 'trial') && (
                          <button
                            type="button"
                            onClick={() => handleAction(subscription, 'pause')}
                            disabled={actionID === subscription.id}
                            className="text-[13px] font-semibold text-amber-700 hover:underline disabled:opacity-50"
                          >
                            Jeda
                          </button>
                        )}
                        {(subscription.status === 'active' || subscription.status === 'trial' || subscription.status === 'paused') && (
                          <button
                            type="button"
                            onClick={() => handleAction(subscription, 'cancel')}
                            disabled={actionID === subscription.id}
                            className="text-[13px] font-semibold text-red-600 hover:underline disabled:opacity-50"
                          >
                            Batalkan
                          </button>
                        )}
                        {actionID === subscription.id && <span className="text-muted">Memproses...</span>}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {!isLoading && subscriptions.length > 0 && totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between text-[13px] text-body">
          <span>
            Halaman {page} dari {totalPages} ({total} langganan)
          </span>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setPage((current) => Math.max(1, current - 1))}
              disabled={page <= 1}
              className="glass rounded-full px-4 py-1.5 font-medium text-ink disabled:cursor-not-allowed disabled:opacity-50"
            >
              Sebelumnya
            </button>
            <button
              type="button"
              onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
              disabled={page >= totalPages}
              className="glass rounded-full px-4 py-1.5 font-medium text-ink disabled:cursor-not-allowed disabled:opacity-50"
            >
              Berikutnya
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
