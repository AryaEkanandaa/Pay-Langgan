import { useEffect, useState, type FormEvent } from 'react'
import Modal from '../../../components/ui/Modal'
import Input from '../../../components/ui/Input'
import SubmitButton from '../../../components/ui/SubmitButton'
import type { Service } from '../../../lib/api'

interface ServiceFormValues {
  name: string
  description: string
}

interface ServiceFormModalProps {
  open: boolean
  onClose: () => void
  onSubmit: (values: ServiceFormValues) => Promise<void>
  service: Service | null
}

export default function ServiceFormModal({ open, onClose, onSubmit, service }: ServiceFormModalProps) {
  const [values, setValues] = useState<ServiceFormValues>({ name: '', description: '' })
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      setValues({ name: service?.name ?? '', description: service?.description ?? '' })
      setError(null)
    }
  }, [open, service])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setError(null)
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
    <Modal open={open} onClose={onClose} title={service ? 'Edit Layanan' : 'Tambah Layanan'}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          id="service_name"
          label="Nama Layanan"
          required
          value={values.name}
          onChange={(event) => setValues((v) => ({ ...v, name: event.target.value }))}
        />
        <Input
          id="service_description"
          label="Deskripsi"
          value={values.description}
          onChange={(event) => setValues((v) => ({ ...v, description: event.target.value }))}
        />

        {error && <p className="text-[13px] text-red-600">{error}</p>}

        <SubmitButton type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? 'Menyimpan...' : 'Simpan'}
        </SubmitButton>
      </form>
    </Modal>
  )
}
