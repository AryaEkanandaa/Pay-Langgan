import { Column, Row, Section, Text } from '@react-email/components'
import type { Customer, Invoice } from '../types'
import { copy } from '../lib/i18n'
import { formatDate } from '../lib/format'
import { colors } from '../lib/theme'

function Line({ label, value }: { label: string; value?: string }) {
  return value ? <Text className="email-muted" style={{ margin: '0 0 5px', color: colors.muted, fontSize: 12, lineHeight: '18px' }}>{label}: <span className="email-ink" style={{ color: colors.ink }}>{value}</span></Text> : null
}

export function InfoBlock({ customer, invoice }: { customer: Customer; invoice: Invoice }) {
  const period = invoice.periodStart && invoice.periodEnd ? `${formatDate(invoice.periodStart)} – ${formatDate(invoice.periodEnd)}` : undefined
  return <Section style={{ marginBottom: 26 }}><Row>
    <Column className="email-info-column" style={{ width: '50%', verticalAlign: 'top', paddingRight: 12 }}>
      <Text className="email-muted" style={{ margin: '0 0 5px', color: colors.muted, fontSize: 12 }}>{copy.shared.billedTo}</Text>
      <Text className="email-ink" style={{ margin: 0, color: colors.ink, fontSize: 16, fontWeight: 700 }}>{customer.name}</Text>
      <Text className="email-muted" style={{ margin: '3px 0 0', color: colors.muted, fontSize: 12, overflowWrap: 'anywhere' }}>{customer.email}</Text>
    </Column>
    <Column className="email-info-column" style={{ width: '50%', textAlign: 'right', verticalAlign: 'top' }}>
      <Line label={copy.shared.issuedAt} value={formatDate(invoice.issuedAt)} />
      <Line label={copy.shared.dueDate} value={formatDate(invoice.dueDate)} />
      <Line label={copy.shared.paidAt} value={formatDate(invoice.paidAt)} />
      <Line label={copy.shared.method} value={invoice.paymentMethod} />
      <Line label={copy.shared.period} value={period} />
    </Column>
  </Row></Section>
}
