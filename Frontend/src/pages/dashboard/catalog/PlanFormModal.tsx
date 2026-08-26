import { useEffect, useState, type FormEvent } from 'react'
import Modal from '../../../components/ui/Modal'
import Input from '../../../components/ui/Input'
import Select from '../../../components/ui/Select'
import SubmitButton from '../../../components/ui/SubmitButton'
import type { Plan, Product } from '../../../lib/api'

interface PlanFormValues {
  product_id: string
  name: string
  price: string
  billing_cycle: string
  trial_days: string
}

interface PlanFormModalProps {
  open: boolean
  onClose: () => void
  onSubmit: (values: PlanFormValues) => Promise<void>
  plan: Plan | null
  products: Product[]
}

export default function PlanFormModal({ open, onClose, onSubmit, plan, products }: PlanFormModalProps) {
  const [values, setValues] = useState<PlanFormValues>({
    product_id: '',
    name: '',
    price: '',
    billing_cycle: 'monthly',
    trial_days: '0',
  })
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      setValues({
        product_id: plan ? String(plan.product_id) : String(products[0]?.id ?? ''),
        name: plan?.name ?? '',
        price: plan ? String(plan.price) : '',
        billing_cycle: plan?.billing_cycle ?? 'monthly',
        trial_days: plan ? String(plan.trial_days) : '0',
      })
      setError(null)
    }
  }, [open, plan, products])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setError(null)

    if (!values.product_id) {
      setError('Pilih produk terlebih dahulu.')
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
    <Modal open={open} onClose={onClose} title={plan ? 'Edit Plan' : 'Tambah Plan'}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Select
          id="plan_product"
          label="Produk"
          required
          value={values.product_id}
          onChange={(event) => setValues((v) => ({ ...v, product_id: event.target.value }))}
        >
          <option value="" disabled>
            Pilih produk
          </option>
          {products.map((product) => (
            <option key={product.id} value={product.id}>
              {product.name}
            </option>
          ))}
        </Select>
        <Input
          id="plan_name"
          label="Nama Plan"
          required
          value={values.name}
          onChange={(event) => setValues((v) => ({ ...v, name: event.target.value }))}
        />
        <Input
          id="plan_price"
          label="Harga (Rp)"
          type="number"
          min="0"
          required
          value={values.price}
          onChange={(event) => setValues((v) => ({ ...v, price: event.target.value }))}
        />
        <Select
          id="plan_billing_cycle"
          label="Siklus Tagihan"
          value={values.billing_cycle}
          onChange={(event) => setValues((v) => ({ ...v, billing_cycle: event.target.value }))}
        >
          <option value="monthly">Bulanan</option>
          <option value="yearly">Tahunan</option>
          <option value="weekly">Mingguan</option>
          <option value="daily">Harian</option>
        </Select>
        <Input
          id="plan_trial_days"
          label="Masa Trial (hari)"
          type="number"
          min="0"
          value={values.trial_days}
          onChange={(event) => setValues((v) => ({ ...v, trial_days: event.target.value }))}
        />

        {error && <p className="text-[13px] text-red-600">{error}</p>}

        <SubmitButton type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? 'Menyimpan...' : 'Simpan'}
        </SubmitButton>
      </form>
    </Modal>
  )
}
