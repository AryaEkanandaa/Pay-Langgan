import { useEffect, useState, type FormEvent } from 'react'
import Modal from '../../components/ui/Modal'
import Input from '../../components/ui/Input'
import Select from '../../components/ui/Select'
import SubmitButton from '../../components/ui/SubmitButton'
import type { Coupon, CouponPayload } from '../../lib/api'

interface CouponFormModalProps {
  open: boolean
  onClose: () => void
  onSubmit: (payload: CouponPayload) => Promise<void>
  coupon: Coupon | null
}

interface CouponFormValues {
  code: string
  discount_type: 'percentage' | 'fixed'
  discount_value: string
  max_usage: string
  expires_at: string
}

const emptyValues: CouponFormValues = {
  code: '',
  discount_type: 'percentage',
  discount_value: '',
  max_usage: '',
  expires_at: '',
}

function dateInputValue(value: string | null) {
  return value ? value.slice(0, 10) : ''
}

export default function CouponFormModal({
  open,
  onClose,
  onSubmit,
  coupon,
}: CouponFormModalProps) {
  const [values, setValues] = useState<CouponFormValues>(emptyValues)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (!open) return
    setValues({
      code: coupon?.code ?? '',
      discount_type: coupon?.discount_type ?? 'percentage',
      discount_value: coupon ? String(coupon.discount_value) : '',
      max_usage: coupon?.max_usage == null ? '' : String(coupon.max_usage),
      expires_at: dateInputValue(coupon?.expires_at ?? null),
    })
    setError(null)
  }, [open, coupon])

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setError(null)

    const discountValue = Number(values.discount_value)
    const maxUsage = values.max_usage === '' ? null : Number(values.max_usage)
    if (!values.code.trim() || !Number.isFinite(discountValue) || discountValue <= 0) {
      setError('Kode dan nilai diskon wajib diisi dengan benar.')
      return
    }
    if (maxUsage !== null && (!Number.isInteger(maxUsage) || maxUsage < 1)) {
      setError('Batas penggunaan harus berupa angka bulat minimal 1.')
      return
    }

    setIsSubmitting(true)
    try {
      await onSubmit({
        code: values.code.trim().toUpperCase(),
        discount_type: values.discount_type,
        discount_value: discountValue,
        max_usage: maxUsage,
        expires_at: values.expires_at ? `${values.expires_at}T23:59:59Z` : null,
      })
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Terjadi kesalahan, coba lagi.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={coupon ? 'Edit Kupon' : 'Tambah Kupon'}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          id="coupon_code"
          label="Kode Kupon"
          required
          maxLength={50}
          value={values.code}
          onChange={(event) => setValues((v) => ({ ...v, code: event.target.value }))}
        />
        <Select
          id="coupon_discount_type"
          label="Jenis Diskon"
          value={values.discount_type}
          onChange={(event) =>
            setValues((v) => ({ ...v, discount_type: event.target.value as CouponFormValues['discount_type'] }))
          }
        >
          <option value="percentage">Persentase</option>
          <option value="fixed">Nominal Tetap</option>
        </Select>
        <Input
          id="coupon_discount_value"
          label={values.discount_type === 'percentage' ? 'Diskon (%)' : 'Nilai Diskon (Rp)'}
          type="number"
          min="0.01"
          step="0.01"
          required
          value={values.discount_value}
          onChange={(event) => setValues((v) => ({ ...v, discount_value: event.target.value }))}
        />
        <Input
          id="coupon_max_usage"
          label="Batas Penggunaan (opsional)"
          type="number"
          min="1"
          step="1"
          value={values.max_usage}
          onChange={(event) => setValues((v) => ({ ...v, max_usage: event.target.value }))}
        />
        <Input
          id="coupon_expires_at"
          label="Tanggal Kedaluwarsa (opsional)"
          type="date"
          value={values.expires_at}
          onChange={(event) => setValues((v) => ({ ...v, expires_at: event.target.value }))}
        />

        {error && <p className="text-[13px] text-red-600">{error}</p>}

        <SubmitButton type="submit" className="w-full" disabled={isSubmitting}>
          {isSubmitting ? 'Menyimpan...' : 'Simpan'}
        </SubmitButton>
      </form>
    </Modal>
  )
}
