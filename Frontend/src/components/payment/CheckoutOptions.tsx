import Input from '../ui/Input'
import { formatCurrency } from '../../lib/format'
import type { CheckoutAddOnSelection, EmbedPlanPreview } from '../../lib/api'

interface Props {
  preview: EmbedPlanPreview
  addOns: CheckoutAddOnSelection[]
  couponCode: string
  busy: boolean
  error: string | null
  onAddOns: (items: CheckoutAddOnSelection[]) => void
  onCoupon: (code: string) => void
}

export default function CheckoutOptions({ preview, addOns, couponCode, busy, error, onAddOns, onCoupon }: Props) {
  return <div className="space-y-4 rounded-2xl border border-border bg-white/50 p-4 dark:bg-slate-900/50">
    {preview.add_ons.length > 0 && <fieldset className="space-y-3">
      <legend className="mb-3 text-sm font-semibold">Add-on (opsional)</legend>
      {preview.add_ons.map((addon) => {
        const selected = addOns.find((item) => item.add_on_id === addon.id)
        return <div key={addon.id} className="flex flex-wrap items-center justify-between gap-3">
          <label className="flex min-w-0 flex-1 items-center gap-3 text-sm">
            <input type="checkbox" checked={Boolean(selected)} onChange={(event) => onAddOns(event.target.checked ? [...addOns, { add_on_id: addon.id, quantity: 1 }] : addOns.filter((item) => item.add_on_id !== addon.id))} className="size-4 accent-brand" />
            <span>{addon.name}<span className="block text-xs text-body">{formatCurrency(addon.price)} / {preview.billing_cycle === 'monthly' ? 'bulan' : 'tahun'}</span></span>
          </label>
          {selected && <input aria-label={`Jumlah ${addon.name}`} type="number" min={1} max={100} value={selected.quantity} onChange={(event) => onAddOns(addOns.map((item) => item.add_on_id === addon.id ? { ...item, quantity: Math.max(1, Math.min(100, Number(event.target.value) || 1)) } : item))} className="w-20 rounded-lg border border-border bg-transparent px-3 py-2 text-sm focus-visible:outline-2 focus-visible:outline-brand" />}
        </div>
      })}
    </fieldset>}
    <Input id="checkout-coupon" label="Kode kupon (opsional)" value={couponCode} maxLength={50} onChange={(event) => onCoupon(event.target.value)} />
    <p role={error ? 'alert' : 'status'} aria-live="polite" className={`text-xs ${error ? 'text-red-600 dark:text-red-300' : 'text-body'}`}>{busy ? 'Memeriksa harga...' : error || (preview.coupon_code ? 'Kupon diterapkan. Total dihitung oleh PayLanggan.' : 'Harga diperbarui otomatis saat pilihan berubah.')}</p>
    {error && addOns.length > 0 && <button type="button" onClick={() => onAddOns([])} className="rounded text-xs font-semibold text-brand underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand dark:text-blue-300">Reset pilihan add-on</button>}
  </div>
}
