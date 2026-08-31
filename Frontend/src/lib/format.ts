export function formatCurrency(value: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(value)
}

export const billingCycleLabel: Record<string, string> = {
  monthly: 'Bulanan',
  yearly: 'Tahunan',
  weekly: 'Mingguan',
  daily: 'Harian',
}

export const subscriptionStatusLabel: Record<string, string> = {
  trial: 'Trial',
  active: 'Aktif',
  paused: 'Dijeda',
  cancelled: 'Dibatalkan',
}
