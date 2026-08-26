import { useCallback, useEffect, useState } from 'react'
import TableSkeleton from '../../../components/ui/skeletons/TableSkeleton'
import SubmitButton from '../../../components/ui/SubmitButton'
import { useAuth } from '../../../context/AuthContext'
import {
  listServices,
  createService,
  updateService,
  deleteService,
  ApiError,
  type Service,
} from '../../../lib/api'
import ServiceFormModal from './ServiceFormModal'

const PAGE_SIZE = 10

export default function ServicesTab() {
  const { auth } = useAuth()
  const token = auth?.token

  const [services, setServices] = useState<Service[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [modalOpen, setModalOpen] = useState(false)
  const [editingService, setEditingService] = useState<Service | null>(null)

  const loadServices = useCallback(async () => {
    if (!token) return
    setIsLoading(true)
    setError(null)
    try {
      const result = await listServices(token, { page, limit: PAGE_SIZE, search: search || undefined })
      setServices(result.items)
      setTotal(result.total)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memuat data layanan.')
    } finally {
      setIsLoading(false)
    }
  }, [token, page, search])

  useEffect(() => {
    loadServices()
  }, [loadServices])

  function openCreateModal() {
    setEditingService(null)
    setModalOpen(true)
  }

  function openEditModal(service: Service) {
    setEditingService(service)
    setModalOpen(true)
  }

  async function handleSubmit(values: { name: string; description: string }) {
    if (!token) return
    const payload = { name: values.name, description: values.description }

    if (editingService) {
      await updateService(token, editingService.id, payload)
    } else {
      await createService(token, payload)
    }

    await loadServices()
  }

  async function handleDelete(service: Service) {
    if (!token) return
    if (!window.confirm(`Hapus layanan "${service.name}"?`)) return

    try {
      await deleteService(token, service.id)
      await loadServices()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal menghapus layanan.')
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  if (!auth) return null

  return (
    <div>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-4">
        <input
          type="search"
          placeholder="Cari layanan..."
          value={search}
          onChange={(event) => {
            setPage(1)
            setSearch(event.target.value)
          }}
          className="w-full max-w-xs rounded-xl border border-white/60 bg-white/50 px-4 py-2.5 text-[14px] text-ink outline-none backdrop-blur-xl transition-colors placeholder:text-muted focus:border-brand/50 focus:bg-white/70 focus:ring-2 focus:ring-brand/30"
        />
        <SubmitButton type="button" onClick={openCreateModal} className="shrink-0">
          + Tambah Layanan
        </SubmitButton>
      </div>

      {error && <p className="mb-4 text-[13px] text-red-600">{error}</p>}

      {isLoading ? (
        <TableSkeleton columns={['Nama', 'Deskripsi', 'Dibuat', '']} />
      ) : services.length === 0 ? (
        <div className="glass flex min-h-[200px] flex-col items-center justify-center rounded-2xl text-center">
          <p className="text-[14px] font-medium text-body">
            {search ? 'Tidak ada layanan yang cocok.' : 'Belum ada layanan.'}
          </p>
          {!search && (
            <button
              type="button"
              onClick={openCreateModal}
              className="mt-2 text-[13px] font-semibold text-brand hover:underline"
            >
              Tambah layanan pertama
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
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Deskripsi</th>
                  <th className="px-5 py-3 text-[11px] font-semibold uppercase tracking-wide text-muted">Dibuat</th>
                  <th className="px-5 py-3" />
                </tr>
              </thead>
              <tbody>
                {services.map((service) => (
                  <tr key={service.id} className="border-b border-white/30 last:border-0">
                    <td className="px-5 py-4 font-medium text-ink">{service.name}</td>
                    <td className="px-5 py-4 text-body">{service.description || '—'}</td>
                    <td className="px-5 py-4 text-body">
                      {new Date(service.created_at).toLocaleDateString('id-ID')}
                    </td>
                    <td className="px-5 py-4 text-right">
                      <button
                        type="button"
                        onClick={() => openEditModal(service)}
                        className="mr-3 text-[13px] font-semibold text-brand hover:underline"
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDelete(service)}
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

      {!isLoading && services.length > 0 && totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between text-[13px] text-body">
          <span>
            Halaman {page} dari {totalPages} ({total} layanan)
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

      <ServiceFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSubmit={handleSubmit}
        service={editingService}
      />
    </div>
  )
}
