import { useCallback, useEffect, useState } from 'react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import TableSkeleton from '../../components/ui/skeletons/TableSkeleton'
import SubmitButton from '../../components/ui/SubmitButton'
import { useAuth } from '../../context/AuthContext'
import {
  listCustomers,
  createCustomer,
  updateCustomer,
  deleteCustomer,
  ApiError,
  type Customer,
} from '../../lib/api'
import CustomerFormModal from './CustomerFormModal'

const PAGE_SIZE = 10

export default function PelangganPage() {
  const { auth } = useAuth()
  const token = auth?.token

  const [customers, setCustomers] = useState<Customer[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [modalOpen, setModalOpen] = useState(false)
  const [editingCustomer, setEditingCustomer] = useState<Customer | null>(null)

  const loadCustomers = useCallback(async () => {
    if (!token) return
    setIsLoading(true)
    setError(null)
    try {
      const result = await listCustomers(token, {
        page,
        limit: PAGE_SIZE,
        search: search || undefined,
      })
      setCustomers(result.items)
      setTotal(result.total)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memuat data pelanggan.')
    } finally {
      setIsLoading(false)
    }
  }, [token, page, search])

  useEffect(() => {
    loadCustomers()
  }, [loadCustomers])

  function openCreateModal() {
    setEditingCustomer(null)
    setModalOpen(true)
  }

  function openEditModal(customer: Customer) {
    setEditingCustomer(customer)
    setModalOpen(true)
  }

  async function handleSubmit(values: { name: string; email: string; contact: string }) {
    if (!token) return
    const payload = {
      name: values.name,
      email: values.email || null,
      contact: values.contact || null,
    }

    if (editingCustomer) {
      await updateCustomer(token, editingCustomer.id, payload)
    } else {
      await createCustomer(token, payload)
    }

    await loadCustomers()
  }

  async function handleDelete(customer: Customer) {
    if (!token) return
    if (!window.confirm(`Hapus pelanggan "${customer.name}"?`)) return

    try {
      await deleteCustomer(token, customer.id)
      await loadCustomers()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal menghapus pelanggan.')
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  if (!auth) return null

  return (
    <div>
      <div className="mb-8 flex flex-wrap items-start justify-between gap-4">
        <DashboardPageHeader title="Pelanggan" description="Kelola data pelanggan bisnis Anda." />
        <SubmitButton type="button" onClick={openCreateModal} className="shrink-0">
          + Tambah Pelanggan
        </SubmitButton>
      </div>

      <input
        type="search"
        placeholder="Cari nama atau email..."
        value={search}
        onChange={(event) => {
          setPage(1)
          setSearch(event.target.value)
        }}
        className="mb-4 w-full max-w-xs rounded-xl border border-white/60 bg-white/50 px-4 py-2.5 text-[14px] text-ink outline-none backdrop-blur-xl transition-colors placeholder:text-muted focus:border-brand/50 focus:bg-white/70 focus:ring-2 focus:ring-brand/30"
      />

      {error && <p className="mb-4 text-[13px] text-red-600">{error}</p>}

      {isLoading ? (
        <TableSkeleton columns={['Nama', 'Email', 'Kontak', 'Terdaftar', '']} />
      ) : customers.length === 0 ? (
        <div className="glass flex min-h-[200px] flex-col items-center justify-center rounded-2xl text-center">
          <p className="text-[14px] font-medium text-body">
            {search ? 'Tidak ada pelanggan yang cocok.' : 'Belum ada pelanggan.'}
          </p>
          {!search && (
            <button
              type="button"
              onClick={openCreateModal}
              className="mt-2 text-[13px] font-semibold text-brand hover:underline"
            >
              Tambah pelanggan pertama
            </button>
          )}
        </div>
      ) : (
        <div className="glass overflow-hidden rounded-2xl">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[640px] text-left text-[13px]">
              <thead>
                <tr className="border-b border-white/50">
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">
                    Nama
                  </th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">
                    Email
                  </th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">
                    Kontak
                  </th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">
                    Terdaftar
                  </th>
                  <th className="px-5 py-3" />
                </tr>
              </thead>
              <tbody>
                {customers.map((customer) => (
                  <tr key={customer.id} className="border-b border-white/30 last:border-0">
                    <td className="px-5 py-4 font-medium text-ink">{customer.name}</td>
                    <td className="px-5 py-4 text-body">{customer.email || '—'}</td>
                    <td className="px-5 py-4 text-body">{customer.contact || '—'}</td>
                    <td className="px-5 py-4 text-body">
                      {new Date(customer.created_at).toLocaleDateString('id-ID')}
                    </td>
                    <td className="px-5 py-4 text-right">
                      <button
                        type="button"
                        onClick={() => openEditModal(customer)}
                        className="mr-3 text-[13px] font-semibold text-brand hover:underline"
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDelete(customer)}
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

      {!isLoading && customers.length > 0 && totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between text-[13px] text-body">
          <span>
            Halaman {page} dari {totalPages} ({total} pelanggan)
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

      <CustomerFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSubmit={handleSubmit}
        customer={editingCustomer}
      />
    </div>
  )
}
