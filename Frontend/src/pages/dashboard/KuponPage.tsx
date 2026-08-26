import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import TableSkeleton from '../../components/ui/skeletons/TableSkeleton'

export default function KuponPage() {
  return (
    <div>
      <DashboardPageHeader title="Kupon" description="Buat dan kelola kode kupon diskon." />
      <TableSkeleton columns={['Kode', 'Diskon', 'Kuota', 'Status']} />
      <p className="mt-4 text-[12px] text-muted">
        Data kupon akan tampil di sini setelah halaman ini terhubung ke API.
      </p>
    </div>
  )
}
