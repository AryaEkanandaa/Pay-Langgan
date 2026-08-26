import { useCallback, useEffect, useState } from 'react'
import TableSkeleton from '../../../components/ui/skeletons/TableSkeleton'
import SubmitButton from '../../../components/ui/SubmitButton'
import { useAuth } from '../../../context/AuthContext'
import {
  listAddOns,
  createAddOn,
  updateAddOn,
  deleteAddOn,
  listProducts,
  ApiError,
  type AddOn,
  type Product,
} from '../../../lib/api'
import { formatCurrency, billingCycleLabel } from '../../../lib/format'
import AddOnFormModal from './AddOnFormModal'

const PAGE_SIZE = 10

export default function AddOnsTab() {
  const { auth } = useAuth()
  const token = auth?.token

  const [addOns, setAddOns] = useState<AddOn[]>([])
  const [products, setProducts] = useState<Product[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [modalOpen, setModalOpen] = useState(false)
  const [editingAddOn, setEditingAddOn] = useState<AddOn | null>(null)

  const loadAddOns = useCallback(async () => {
    if (!token) return
    setIsLoading(true)
    setError(null)
    try {
      const result = await listAddOns(token, { page, limit: PAGE_SIZE, search: search || undefined })
      setAddOns(result.items)
      setTotal(result.total)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memuat data add-on.')
    } finally {
      setIsLoading(false)
    }
  }, [token, page, search])

  useEffect(() => {
    loadAddOns()
  }, [loadAddOns])

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
    setEditingAddOn(null)
    setModalOpen(true)
  }

  function openEditModal(addOn: AddOn) {
    setEditingAddOn(addOn)
    setModalOpen(true)
  }

  async function handleSubmit(values: {
    product_id: string
    name: string
    price: string
    billing_cycle: string
  }) {
    if (!token) return
    const payload = {
      product_id: Number(values.product_id),
      name: values.name,
      price: Number(values.price),
      billing_cycle: values.billing_cycle,
    }

    if (editingAddOn) {
      await updateAddOn(token, editingAddOn.id, payload)
    } else {
      await createAddOn(token, payload)
    }

    await loadAddOns()
  }

  async function handleDelete(addOn: AddOn) {
    if (!token) return
    if (!window.confirm(`Hapus add-on "${addOn.name}"?`)) return

    try {
      await deleteAddOn(token, addOn.id)
      await loadAddOns()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal menghapus add-on.')
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  if (!auth) return null

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-4">
        <input
          type="search"
          placeholder="Cari add-on..."
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
          + Tambah Add-on
        </SubmitButton>
      </div>

      {products.length === 0 && !isLoading && (
        <p className="mb-4 text-[13px] text-muted">
          Tambahkan produk terlebih dahulu sebelum membuat add-on.
        </p>
      )}

      {error && <p className="mb-4 text-[13px] text-red-600">{error}</p>}

      {isLoading ? (
        <TableSkeleton columns={['Nama', 'Produk', 'Harga', 'Siklus', '']} />
      ) : addOns.length === 0 ? (
        <div className="glass flex min-h-[200px] flex-col items-center justify-center rounded-2xl text-center">
          <p className="text-[14px] font-medium text-body">
            {search ? 'Tidak ada add-on yang cocok.' : 'Belum ada add-on.'}
          </p>
          {!search && products.length > 0 && (
            <button
              type="button"
              onClick={openCreateModal}
              className="mt-2 text-[13px] font-semibold text-brand hover:underline"
            >
              Tambah add-on pertama
            </button>
          )}
        </div>
      ) : (
        <div className="glass overflow-hidden rounded-2xl">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[640px] text-left text-[13px]">
              <thead>
                <tr className="border-b border-white/50">
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Nama</th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Produk</th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Harga</th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Siklus</th>
                  <th className="px-5 py-3" />
                </tr>
              </thead>
              <tbody>
                {addOns.map((addOn) => (
                  <tr key={addOn.id} className="border-b border-white/30 last:border-0">
                    <td className="px-5 py-4 font-medium text-ink">{addOn.name}</td>
                    <td className="px-5 py-4 text-body">{productName(addOn.product_id)}</td>
                    <td className="px-5 py-4 text-body">{formatCurrency(addOn.price)}</td>
                    <td className="px-5 py-4 text-body">
                      {billingCycleLabel[addOn.billing_cycle] ?? addOn.billing_cycle}
                    </td>
                    <td className="px-5 py-4 text-right">
                      <button
                        type="button"
                        onClick={() => openEditModal(addOn)}
                        className="mr-3 text-[13px] font-semibold text-brand hover:underline"
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDelete(addOn)}
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

      {!isLoading && addOns.length > 0 && totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between text-[13px] text-body">
          <span>
            Halaman {page} dari {totalPages} ({total} add-on)
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

      <AddOnFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSubmit={handleSubmit}
        addOn={editingAddOn}
        products={products}
      />
    </div>
  )
}
