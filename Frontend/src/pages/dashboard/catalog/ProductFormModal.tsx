import { useEffect, useState, type FormEvent } from 'react'
import Modal from '../../../components/ui/Modal'
import Input from '../../../components/ui/Input'
import Select from '../../../components/ui/Select'
import SubmitButton from '../../../components/ui/SubmitButton'
import type { Product, Service } from '../../../lib/api'

interface ProductFormValues {
  service_id: string
  name: string
  description: string
  status: string
}

interface ProductFormModalProps {
  open: boolean
  onClose: () => void
  onSubmit: (values: ProductFormValues) => Promise<void>
  product: Product | null
  services: Service[]
}

export default function ProductFormModal({
  open,
  onClose,
  onSubmit,
  product,
  services,
}: ProductFormModalProps) {
  const [values, setValues] = useState<ProductFormValues>({
    service_id: '',
    name: '',
    description: '',
    status: 'active',
  })
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      setValues({
        service_id: product ? String(product.service_id) : String(services[0]?.id ?? ''),
        name: product?.name ?? '',
        description: product?.description ?? '',
        status: product?.status ?? 'active',
      })
      setError(null)
    }
  }, [open, product, services])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setError(null)

    if (!values.service_id) {
      setError('Pilih layanan terlebih dahulu.')
      return
    }

    setIsSubmitting(true)
    try {
      await onSubmit(values)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Terjadi kesalahan, coba lagi.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={product ? 'Edit Produk' : 'Tambah Produk'}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Select
          id="product_service"
          label="Layanan"
          required
          value={values.service_id}
          onChange={(event) => setValues((v) => ({ ...v, service_id: event.target.value }))}
        >
          <option value="" disabled>
            Pilih layanan
          </option>
          {services.map((service) => (
            <option key={service.id} value={service.id}>
              {service.name}
            </option>
          ))}
        </Select>
        <Input
          id="product_name"
          label="Nama Produk"
          required
          value={values.name}
          onChange={(event) => setValues((v) => ({ ...v, name: event.target.value }))}
        />
        <Input
          id="product_description"
          label="Deskripsi"
          value={values.description}
          onChange={(event) => setValues((v) => ({ ...v, description: event.target.value }))}
        />
        <Select
          id="product_status"
          label="Status"
          value={values.status}
          onChange={(event) => setValues((v) => ({ ...v, status: event.target.value }))}
        >
          <option value="active">Aktif</option>
          <option value="inactive">Nonaktif</option>
        </Select>

        {error && <p className="text-[13px] text-red-600">{error}</p>}

        <SubmitButton type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? 'Menyimpan...' : 'Simpan'}
        </SubmitButton>
      </form>
    </Modal>
  )
}
