import { useCallback, useEffect, useState } from 'react'
import TableSkeleton from '../../../components/ui/skeletons/TableSkeleton'
import SubmitButton from '../../../components/ui/SubmitButton'
import { useAuth } from '../../../context/AuthContext'
import {
  listPlans,
  createPlan,
  updatePlan,
  deletePlan,
  listProducts,
  ApiError,
  type Plan,
  type Product,
} from '../../../lib/api'
import { formatCurrency, billingCycleLabel } from '../../../lib/format'
import PlanFormModal from './PlanFormModal'

const PAGE_SIZE = 10

export default function PlansTab() {
  const { auth } = useAuth()
  const token = auth?.token

  const [plans, setPlans] = useState<Plan[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [modalOpen, setModalOpen] = useState(false)
  const [editingPlan, setEditingPlan] = useState<Plan | null>(null)

  const loadPlans = useCallback(async () => {
    if (!token) return
    setIsLoading(true)
    setError(null)
    try {
      const result = await listPlans(token, { page, limit: PAGE_SIZE, search: search || undefined })
      setPlans(result.items)
      setTotal(result.total)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memuat data plan.')
    } finally {
      setIsLoading(false)
    }
  }, [token, page, search])

  useEffect(() => {
    loadPlans()
  }, [loadPlans])

  useEffect(() => {
    if (!token) return
    listProducts(token, { limit: 100 })
      .then((result) => setProducts(result.items))
      .catch(() => setProducts([]))
  }, [token])

  function productName(productId: number) {
    return products.find((product) => product.id === productId)?.name ?? '—'
  }

  function openCreateModal() {
    setEditingPlan(null)
    setModalOpen(true)
  }

  function openEditModal(plan: Plan) {
    setEditingPlan(plan)
    setModalOpen(true)
  }

  async function handleSubmit(values: {
    product_id: string
    name: string
    price: string
    billing_cycle: string
    trial_days: string
  }) {
    if (!token) return
    const payload = {
      product_id: Number(values.product_id),
      name: values.name,
      price: Number(values.price),
      billing_cycle: values.billing_cycle,
      trial_days: Number(values.trial_days) || 0,
    }

    if (editingPlan) {
      await updatePlan(token, editingPlan.id, payload)
    } else {
      await createPlan(token, payload)
    }

    await loadPlans()
  }

  async function handleDelete(plan: Plan) {
    if (!token) return
    if (!window.confirm(`Hapus plan "${plan.name}"?`)) return

    try {
      await deletePlan(token, plan.id)
      await loadPlans()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal menghapus plan.')
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  if (!auth) return null

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-4">
        <input
          type="search"
          placeholder="Cari plan..."
          value={search}
          onChange={(event) => {
            setPage(1)
            setSearch(event.target.value)
          }}
          className="w-full max-w-xs rounded-xl border border-white/60 bg-white/50 px-4 py-2.5 text-[14px] text-ink outline-none backdrop-blur-xl transition-colors placeholder:text-muted focus:border-brand/50 focus:bg-white/70 focus:ring-2 focus:ring-brand/30"
        />
        <SubmitButton
          type="button"
          onClick={openCreateModal}
          className="shrink-0"
          disabled={products.length === 0}
        >
          + Tambah Plan
        </SubmitButton>
      </div>

      {products.length === 0 && !isLoading && (
        <p className="mb-4 text-[13px] text-muted">
          Tambahkan produk terlebih dahulu sebelum membuat plan.
        </p>
      )}

      {error && <p className="mb-4 text-[13px] text-red-600">{error}</p>}

      {isLoading ? (
        <TableSkeleton columns={['Nama', 'Produk', 'Harga', 'Siklus', 'Trial', '']} />
      ) : plans.length === 0 ? (
        <div className="glass flex min-h-[200px] flex-col items-center justify-center rounded-2xl text-center">
          <p className="text-[14px] font-medium text-body">
            {search ? 'Tidak ada plan yang cocok.' : 'Belum ada plan.'}
          </p>
          {!search && products.length > 0 && (
            <button
              type="button"
              onClick={openCreateModal}
              className="mt-2 text-[13px] font-semibold text-brand hover:underline"
            >
              Tambah plan pertama
            </button>
          )}
        </div>
      ) : (
        <div className="glass overflow-hidden rounded-2xl">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[680px] text-left text-[13px]">
              <thead>
                <tr className="border-b border-white/50">
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Nama</th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Produk</th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Harga</th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Siklus</th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Trial</th>
                  <th className="px-5 py-3" />
                </tr>
              </thead>
              <tbody>
                {plans.map((plan) => (
                  <tr key={plan.id} className="border-b border-white/30 last:border-0">
                    <td className="px-5 py-4 font-medium text-ink">{plan.name}</td>
                    <td className="px-5 py-4 text-body">{productName(plan.product_id)}</td>
                    <td className="px-5 py-4 text-body">{formatCurrency(plan.price)}</td>
                    <td className="px-5 py-4 text-body">
                      {billingCycleLabel[plan.billing_cycle] ?? plan.billing_cycle}
                    </td>
                    <td className="px-5 py-4 text-body">
                      {plan.trial_days > 0 ? `${plan.trial_days} hari` : '—'}
                    </td>
                    <td className="px-5 py-4 text-right">
                      <button
                        type="button"
                        onClick={() => openEditModal(plan)}
                        className="mr-3 text-[13px] font-semibold text-brand hover:underline"
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDelete(plan)}
                        className="text-[13px] font-semibold text-red-600 hover:underline"
                      >
                        Hapus
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {!isLoading && plans.length > 0 && totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between text-[13px] text-body">
          <span>
            Halaman {page} dari {totalPages} ({total} plan)
          </span>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="glass rounded-full px-4 py-1.5 font-medium text-ink disabled:cursor-not-allowed disabled:opacity-50"
            >
              Sebelumnya
            </button>
            <button
              type="button"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="glass rounded-full px-4 py-1.5 font-medium text-ink disabled:cursor-not-allowed disabled:opacity-50"
            >
              Berikutnya
            </button>
          </div>
        </div>
      )}

      <PlanFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSubmit={handleSubmit}
        plan={editingPlan}
        products={products}
      />
    </div>
  )
}
