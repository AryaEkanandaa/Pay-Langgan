import { useCallback, useEffect, useState } from 'react'
import TableSkeleton from '../../../components/ui/skeletons/TableSkeleton'
import Badge from '../../../components/ui/Badge'
import SubmitButton from '../../../components/ui/SubmitButton'
import { useAuth } from '../../../context/AuthContext'
import {
  listProducts,
  createProduct,
  updateProduct,
  deleteProduct,
  listServices,
  ApiError,
  type Product,
  type Service,
} from '../../../lib/api'
import ProductFormModal from './ProductFormModal'

const PAGE_SIZE = 10

export default function ProductsTab() {
  const { auth } = useAuth()
  const token = auth?.token

  const [products, setProducts] = useState<Product[]>([])
  const [services, setServices] = useState<Service[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [modalOpen, setModalOpen] = useState(false)
  const [editingProduct, setEditingProduct] = useState<Product | null>(null)

  const loadProducts = useCallback(async () => {
    if (!token) return
    setIsLoading(true)
    setError(null)
    try {
      const result = await listProducts(token, { page, limit: PAGE_SIZE, search: search || undefined })
      setProducts(result.items)
      setTotal(result.total)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memuat data produk.')
    } finally {
      setIsLoading(false)
    }
  }, [token, page, search])

  useEffect(() => {
    loadProducts()
  }, [loadProducts])

  useEffect(() => {
    if (!token) return
    listServices(token, { limit: 100 })
      .then((result) => setServices(result.items))
      .catch(() => setServices([]))
  }, [token])

  function serviceName(serviceId: number) {
    return services.find((service) => service.id === serviceId)?.name ?? '—'
  }

  function openCreateModal() {
    setEditingProduct(null)
    setModalOpen(true)
  }

  function openEditModal(product: Product) {
    setEditingProduct(product)
    setModalOpen(true)
  }

  async function handleSubmit(values: {
    service_id: string
    name: string
    description: string
    status: string
  }) {
    if (!token) return
    const payload = {
      service_id: Number(values.service_id),
      name: values.name,
      description: values.description,
      status: values.status,
    }

    if (editingProduct) {
      await updateProduct(token, editingProduct.id, payload)
    } else {
      await createProduct(token, payload)
    }

    await loadProducts()
  }

  async function handleDelete(product: Product) {
    if (!token) return
    if (!window.confirm(`Hapus produk "${product.name}"?`)) return

    try {
      await deleteProduct(token, product.id)
      await loadProducts()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal menghapus produk.')
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  if (!auth) return null

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-4">
        <input
          type="search"
          placeholder="Cari produk..."
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
          disabled={services.length === 0}
        >
          + Tambah Produk
        </SubmitButton>
      </div>

      {services.length === 0 && !isLoading && (
        <p className="mb-4 text-[13px] text-muted">
          Tambahkan layanan terlebih dahulu sebelum membuat produk.
        </p>
      )}

      {error && <p className="mb-4 text-[13px] text-red-600">{error}</p>}

      {isLoading ? (
        <TableSkeleton columns={['Nama', 'Layanan', 'Status', 'Dibuat', '']} />
      ) : products.length === 0 ? (
        <div className="glass flex min-h-[200px] flex-col items-center justify-center rounded-2xl text-center">
          <p className="text-[14px] font-medium text-body">
            {search ? 'Tidak ada produk yang cocok.' : 'Belum ada produk.'}
          </p>
          {!search && services.length > 0 && (
            <button
              type="button"
              onClick={openCreateModal}
              className="mt-2 text-[13px] font-semibold text-brand hover:underline"
            >
              Tambah produk pertama
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
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Layanan</th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Status</th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Dibuat</th>
                  <th className="px-5 py-3" />
                </tr>
              </thead>
              <tbody>
                {products.map((product) => (
                  <tr key={product.id} className="border-b border-white/30 last:border-0">
                    <td className="px-5 py-4 font-medium text-ink">{product.name}</td>
                    <td className="px-5 py-4 text-body">{serviceName(product.service_id)}</td>
                    <td className="px-5 py-4">
                      <Badge className={product.status !== 'active' ? 'text-muted' : undefined}>
                        {product.status === 'active' ? 'Aktif' : 'Nonaktif'}
                      </Badge>
                    </td>
                    <td className="px-5 py-4 text-body">
                      {new Date(product.created_at).toLocaleDateString('id-ID')}
                    </td>
                    <td className="px-5 py-4 text-right">
                      <button
                        type="button"
                        onClick={() => openEditModal(product)}
                        className="mr-3 text-[13px] font-semibold text-brand hover:underline"
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDelete(product)}
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

      {!isLoading && products.length > 0 && totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between text-[13px] text-body">
          <span>
            Halaman {page} dari {totalPages} ({total} produk)
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

      <ProductFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSubmit={handleSubmit}
        product={editingProduct}
        services={services}
      />
    </div>
  )
}
