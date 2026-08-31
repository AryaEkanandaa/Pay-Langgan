import { useCallback, useEffect, useState } from 'react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import TableSkeleton from '../../components/ui/skeletons/TableSkeleton'
import SubmitButton from '../../components/ui/SubmitButton'
import { useAuth } from '../../context/AuthContext'
import {
  ApiError,
  createCoupon,
  deleteCoupon,
  listCoupons,
  updateCoupon,
  type Coupon,
  type CouponPayload,
} from '../../lib/api'
import { formatCurrency } from '../../lib/format'
import CouponFormModal from './CouponFormModal'

const PAGE_SIZE = 10

function couponStatus(coupon: Coupon) {
  if (coupon.expires_at && new Date(coupon.expires_at).getTime() < Date.now()) {
    return { label: 'Kedaluwarsa', className: 'bg-red-100 text-red-700' }
  }
  if (coupon.max_usage !== null && coupon.used_count >= coupon.max_usage) {
    return { label: 'Habis', className: 'bg-amber-100 text-amber-700' }
  }
  return { label: 'Aktif', className: 'bg-emerald-100 text-emerald-700' }
}

export default function KuponPage() {
  const { auth } = useAuth()
  const token = auth?.token

  const [coupons, setCoupons] = useState<Coupon[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingCoupon, setEditingCoupon] = useState<Coupon | null>(null)

  const loadCoupons = useCallback(async () => {
    if (!token) return
    setIsLoading(true)
    setError(null)
    try {
      const result = await listCoupons(token, {
        page,
        limit: PAGE_SIZE,
        search: search || undefined,
      })
      setCoupons(result.items)
      setTotal(result.total)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memuat data kupon.')
    } finally {
      setIsLoading(false)
    }
  }, [token, page, search])

  useEffect(() => {
    loadCoupons()
  }, [loadCoupons])

  function openCreateModal() {
    setEditingCoupon(null)
    setModalOpen(true)
  }

  function openEditModal(coupon: Coupon) {
    setEditingCoupon(coupon)
    setModalOpen(true)
  }

  async function handleSubmit(payload: CouponPayload) {
    if (!token) return
    if (editingCoupon) {
      await updateCoupon(token, editingCoupon.id, payload)
    } else {
      await createCoupon(token, payload)
    }
    await loadCoupons()
  }

  async function handleDelete(coupon: Coupon) {
    if (!token || !window.confirm(`Hapus kupon "${coupon.code}"?`)) return
    try {
      await deleteCoupon(token, coupon.id)
      await loadCoupons()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal menghapus kupon.')
    }
  }

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  if (!auth) return null

  return (
    <div>
      <div className="mb-8 flex flex-wrap items-start justify-between gap-4">
        <DashboardPageHeader title="Kupon" description="Buat dan kelola kode kupon diskon." />
        <SubmitButton type="button" onClick={openCreateModal} className="shrink-0">
          + Tambah Kupon
        </SubmitButton>
      </div>

      <input
        type="search"
        placeholder="Cari kode kupon..."
        value={search}
        onChange={(event) => {
          setPage(1)
          setSearch(event.target.value)
        }}
        className="mb-4 w-full max-w-xs rounded-xl border border-white/60 bg-white/50 px-4 py-2.5 text-[14px] text-ink outline-none backdrop-blur-xl transition-colors placeholder:text-muted focus:border-brand/50 focus:bg-white/70 focus:ring-2 focus:ring-brand/30"
      />

      {error && <p className="mb-4 text-[13px] text-red-600">{error}</p>}

      {isLoading ? (
        <TableSkeleton columns={['Kode', 'Diskon', 'Kuota', 'Berlaku sampai', 'Status', '']} />
      ) : coupons.length === 0 ? (
        <div className="glass flex min-h-[200px] flex-col items-center justify-center rounded-2xl text-center">
          <p className="text-[14px] font-medium text-body">
            {search ? 'Tidak ada kupon yang cocok.' : 'Belum ada kupon.'}
          </p>
          {!search && (
            <button
              type="button"
              onClick={openCreateModal}
              className="mt-2 text-[13px] font-semibold text-brand hover:underline"
            >
              Tambah kupon pertama
            </button>
          )}
        </div>
      ) : (
        <div className="glass overflow-hidden rounded-2xl">
          <div className="overflow-x-auto">
            <table className="w-full min-w-[760px] text-left text-[13px]">
              <thead>
                <tr className="border-b border-white/50">
                  {['Kode', 'Diskon', 'Kuota', 'Berlaku sampai', 'Status', ''].map((heading) => (
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
                {coupons.map((coupon) => {
                  const status = couponStatus(coupon)
                  return (
                    <tr key={coupon.id} className="border-b border-white/30 last:border-0">
                      <td className="px-5 py-4 font-semibold tracking-wide text-ink">{coupon.code}</td>
                      <td className="px-5 py-4 text-body">
                        {coupon.discount_type === 'percentage'
                          ? `${coupon.discount_value}%`
                          : formatCurrency(coupon.discount_value)}
                      </td>
                      <td className="px-5 py-4 text-body">
                        {coupon.max_usage === null
                          ? `${coupon.used_count} / tidak terbatas`
                          : `${coupon.used_count} / ${coupon.max_usage}`}
                      </td>
                      <td className="px-5 py-4 text-body">
                        {coupon.expires_at
                          ? new Date(coupon.expires_at).toLocaleDateString('id-ID')
                          : 'Tidak ada'}
                      </td>
                      <td className="px-5 py-4">
                        <span className={`rounded-full px-2.5 py-1 text-[11px] font-semibold ${status.className}`}>
                          {status.label}
                        </span>
                      </td>
                      <td className="px-5 py-4 text-right">
                        <button
                          type="button"
                          onClick={() => openEditModal(coupon)}
                          className="mr-3 text-[13px] font-semibold text-brand hover:underline"
                        >
                          Edit
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDelete(coupon)}
                          className="text-[13px] font-semibold text-red-600 hover:underline"
                        >
                          Hapus
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {!isLoading && coupons.length > 0 && totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between text-[13px] text-body">
          <span>
            Halaman {page} dari {totalPages} ({total} kupon)
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

      <CouponFormModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSubmit={handleSubmit}
        coupon={editingCoupon}
      />
    </div>
  )
}
