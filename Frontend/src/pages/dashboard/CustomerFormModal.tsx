import { useEffect, useState, type FormEvent } from 'react'
import Modal from '../../components/ui/Modal'
import Input from '../../components/ui/Input'
import SubmitButton from '../../components/ui/SubmitButton'
import type { Customer } from '../../lib/api'

interface CustomerFormValues {
  name: string
  email: string
  contact: string
}

interface CustomerFormModalProps {
  open: boolean
  onClose: () => void
  onSubmit: (values: CustomerFormValues) => Promise<void>
  customer: Customer | null
}

export default function CustomerFormModal({
  open,
  onClose,
  onSubmit,
  customer,
}: CustomerFormModalProps) {
  const [values, setValues] = useState<CustomerFormValues>({ name: '', email: '', contact: '' })
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      setValues({
        name: customer?.name ?? '',
        email: customer?.email ?? '',
        contact: customer?.contact ?? '',
      })
      setError(null)
    }
  }, [open, customer])

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
    <Modal open={open} onClose={onClose} title={customer ? 'Edit Pelanggan' : 'Tambah Pelanggan'}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          id="customer_name"
          label="Nama"
          required
          value={values.name}
          onChange={(event) => setValues((v) => ({ ...v, name: event.target.value }))}
        />
        <Input
          id="customer_email"
          label="Email"
          type="email"
          value={values.email}
          onChange={(event) => setValues((v) => ({ ...v, email: event.target.value }))}
        />
        <Input
          id="customer_contact"
          label="Kontak"
          value={values.contact}
          onChange={(event) => setValues((v) => ({ ...v, contact: event.target.value }))}
        />

        {error && <p className="text-[13px] text-red-600">{error}</p>}

        <SubmitButton type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? 'Menyimpan...' : 'Simpan'}
        </SubmitButton>
      </form>
    </Modal>
  )
}
