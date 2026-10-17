import { useEffect, useState, type FormEvent } from 'react'
import Modal from '../../../components/ui/Modal'
import Input from '../../../components/ui/Input'
import Select from '../../../components/ui/Select'
import SubmitButton from '../../../components/ui/SubmitButton'
import type { AddOn, Product } from '../../../lib/api'

interface AddOnFormValues {
  visible_in_checkout: boolean
  product_id: string
  name: string
  price: string
  billing_cycle: string
}

interface AddOnFormModalProps {
  open: boolean
  onClose: () => void
  onSubmit: (values: AddOnFormValues) => Promise<void>
  addOn: AddOn | null
  products: Product[]
}

export default function AddOnFormModal({ open, onClose, onSubmit, addOn, products }: AddOnFormModalProps) {
  const [values, setValues] = useState<AddOnFormValues>({
    visible_in_checkout: true,
    product_id: '',
    name: '',
    price: '',
    billing_cycle: 'monthly',
  })
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (open) {
      setValues({
        visible_in_checkout: addOn?.visible_in_checkout ?? true,
        product_id: addOn ? String(addOn.product_id) : String(products[0]?.id ?? ''),
        name: addOn?.name ?? '',
        price: addOn ? String(addOn.price) : '',
        billing_cycle: addOn?.billing_cycle ?? 'monthly',
      })
      setError(null)
    }
  }, [open, addOn, products])

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
    <Modal open={open} onClose={onClose} title={addOn ? 'Edit Add-on' : 'Tambah Add-on'}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Select
          id="addon_product"
          label="Produk"
          disabled={Boolean(addOn)}
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
        <p className="text-sm text-body">Add-on tersedia untuk plan pada produk dan siklus tagihan yang sama. Produk tidak dapat diubah setelah add-on dibuat.</p>
        <Input
          id="addon_name"
          label="Nama Add-on"
          required
          value={values.name}
          onChange={(event) => setValues((v) => ({ ...v, name: event.target.value }))}
        />
        <Input
          id="addon_price"
          label="Harga (Rp)"
          type="number"
          min="0"
          required
          value={values.price}
          onChange={(event) => setValues((v) => ({ ...v, price: event.target.value }))}
        />
        <Select
          id="addon_billing_cycle"
          label="Siklus Tagihan"
          value={values.billing_cycle}
          onChange={(event) => setValues((v) => ({ ...v, billing_cycle: event.target.value }))}
        >
          <option value="monthly">Bulanan</option>
          <option value="yearly">Tahunan</option>
        </Select>

        <fieldset className="rounded-xl border border-border p-4">
          <legend className="px-2 text-sm font-semibold text-ink">Tampilan pelanggan</legend>
          <label className="flex cursor-pointer items-start gap-3 text-sm text-ink">
            <input
              type="checkbox"
              checked={values.visible_in_checkout}
              onChange={(event) => setValues((v) => ({ ...v, visible_in_checkout: event.target.checked }))}
              aria-describedby="addon_visibility_help"
              className="mt-0.5 size-4 shrink-0 accent-brand focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
            />
            <span className="font-medium">Tampilkan di checkout</span>
          </label>
          <p id="addon_visibility_help" className="mt-2 text-sm text-body">Pelanggan dapat melihat dan memilih add-on ini pada langkah Paket. Jika dimatikan, add-on tetap tersedia untuk admin dan langganan yang sudah menggunakannya.</p>
        </fieldset>

        {error && <p className="text-[13px] text-red-600">{error}</p>}

        <SubmitButton type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? 'Menyimpan...' : 'Simpan'}
        </SubmitButton>
      </form>
    </Modal>
  )
}
