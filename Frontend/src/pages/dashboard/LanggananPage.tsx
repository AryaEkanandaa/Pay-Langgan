import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import TableSkeleton from '../../components/ui/skeletons/TableSkeleton'

export default function LanggananPage() {
  return (
    <div>
      <DashboardPageHeader
        title="Langganan"
        description="Pantau status dan siklus hidup langganan pelanggan."
      />
      <TableSkeleton columns={['Pelanggan', 'Plan', 'Status', 'Tagihan Berikutnya']} />
      <p className="mt-4 text-[12px] text-muted">
        Data langganan akan tampil di sini setelah halaman ini terhubung ke API.
      </p>
    </div>
  )
}
