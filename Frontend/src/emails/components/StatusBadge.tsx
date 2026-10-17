import { Text } from '@react-email/components'
import type { InvoiceStatus } from '../types'
import { copy } from '../lib/i18n'

const palette: Record<InvoiceStatus, { background: string; color: string }> = {
  paid: { background: '#e5f7ee', color: '#087a55' },
  pending: { background: '#fff3cf', color: '#805500' },
  overdue: { background: '#ffe4df', color: '#a5291c' },
  failed: { background: '#ffe4df', color: '#a5291c' },
  cancelled: { background: '#e8edf4', color: '#40516a' },
  refunded: { background: '#e6efff', color: '#2454a6' },
}

export function StatusBadge({ status }: { status: InvoiceStatus }) {
  const color = palette[status]
  return <Text style={{ display: 'inline-block', margin: '7px 0 0', padding: '4px 10px', borderRadius: 12, backgroundColor: color.background, color: color.color, fontSize: 11, fontWeight: 700 }}>{copy.shared.status[status]}</Text>
}
