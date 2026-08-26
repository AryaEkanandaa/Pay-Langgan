import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import FormSkeleton from '../../components/ui/skeletons/FormSkeleton'

export default function PengaturanPage() {
  return (
    <div>
      <DashboardPageHeader
        title="Pengaturan Bisnis"
        description="Atur informasi dan preferensi bisnis Anda."
      />
      <FormSkeleton fields={['Nama Bisnis', 'Email Bisnis', 'Zona Waktu']} />
      <p className="mt-4 text-[12px] text-muted">
        Formulir ini akan aktif setelah halaman ini terhubung ke API.
      </p>
    </div>
  )
}
