export interface StatusRow {
  label: string
  value: number
  className: string
}

export const statusRows: StatusRow[] = [
  { label: 'Trial', value: 18, className: 'bg-amber' },
  { label: 'Aktif', value: 64, className: 'bg-brand' },
  { label: 'Ditangguhkan', value: 9, className: 'bg-muted' },
  { label: 'Dibatalkan', value: 9, className: 'bg-ink/20' },
]

export const invoiceExample = {
  plan: 'Plan Basic · Bulanan',
  price: 'Rp100.000',
  coupon: 'Kupon DISC10',
  discount: '- Rp10.000',
}
