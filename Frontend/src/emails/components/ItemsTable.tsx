import { Column, Row, Section, Text } from '@react-email/components'
import type { Invoice, InvoiceItem } from '../types'
import { colors, theme } from '../lib/theme'
import { copy } from '../lib/i18n'
import { formatRupiah } from '../lib/format'

const cell = { fontSize: 12, lineHeight: '18px', margin: 0, color: colors.ink }

function ItemRow({ item }: { item: InvoiceItem }) {
  const discount = item.subtotal < 0
  return <Row style={{ borderBottom: `1px solid ${colors.line}` }}>
    <Column style={{ width: '42%', padding: '12px 8px', verticalAlign: 'top' }}><Text className="email-ink" style={{ ...cell, overflowWrap: 'anywhere' }}>{item.name}</Text></Column>
    <Column style={{ width: '12%', padding: '12px 4px', textAlign: 'center', verticalAlign: 'top' }}><Text className="email-ink" style={cell}>{item.quantity}</Text></Column>
    <Column className="email-unit-price" style={{ width: '22%', padding: '12px 4px', textAlign: 'right', verticalAlign: 'top' }}><Text className="email-ink" style={{ ...cell, color: discount ? colors.green : colors.ink }}>{formatRupiah(item.unitPrice)}</Text></Column>
    <Column style={{ width: '24%', padding: '12px 8px', textAlign: 'right', verticalAlign: 'top' }}><Text className="email-ink" style={{ ...cell, color: discount ? colors.green : colors.ink }}>{formatRupiah(item.subtotal)}</Text></Column>
  </Row>
}

function SummaryRow({ label, amount, discount = false }: { label: string; amount: number; discount?: boolean }) {
  return <Row>
    <Column style={{ width: '60%', padding: '5px 8px' }}><Text className="email-muted" style={{ ...cell, color: colors.muted }}>{label}</Text></Column>
    <Column style={{ width: '40%', padding: '5px 8px', textAlign: 'right' }}><Text className={discount ? undefined : 'email-ink'} style={{ ...cell, color: discount ? colors.green : colors.ink }}>{formatRupiah(amount)}</Text></Column>
  </Row>
}

export function ItemsTable({ invoice, brandColor }: { invoice: Invoice; brandColor?: string }) {
  const { accent } = theme(brandColor)
  const hasDiscountItem = invoice.items.some((item) => item.subtotal < 0)
  const showDiscount = !hasDiscountItem && Boolean(invoice.discountTotal)
  const showSubtotal = !hasDiscountItem && (showDiscount || Boolean(invoice.tax) || Boolean(invoice.adminFee))
  return <Section>
    <Row className="email-table-head" style={{ backgroundColor: colors.table }}>
      <Column style={{ width: '42%', padding: '10px 8px' }}><Text className="email-ink" style={{ ...cell, fontSize: 10, fontWeight: 700, textTransform: 'uppercase' }}>{copy.shared.item}</Text></Column>
      <Column style={{ width: '12%', padding: '10px 4px', textAlign: 'center' }}><Text className="email-ink" style={{ ...cell, fontSize: 10, fontWeight: 700, textTransform: 'uppercase' }}>{copy.shared.quantity}</Text></Column>
      <Column className="email-unit-price" style={{ width: '22%', padding: '10px 4px', textAlign: 'right' }}><Text className="email-ink" style={{ ...cell, fontSize: 10, fontWeight: 700, textTransform: 'uppercase' }}>{copy.shared.unitPrice}</Text></Column>
      <Column style={{ width: '24%', padding: '10px 8px', textAlign: 'right' }}><Text className="email-ink" style={{ ...cell, fontSize: 10, fontWeight: 700, textTransform: 'uppercase' }}>{copy.shared.subtotal}</Text></Column>
    </Row>
    {invoice.items.map((item, index) => <ItemRow key={`${item.name}-${index}`} item={item} />)}
    {invoice.items.length === 0 && <Text className="email-muted" style={{ color: colors.muted, fontSize: 12, margin: '12px 8px' }}>{copy.shared.noItems}</Text>}
    {showSubtotal && <SummaryRow label={copy.shared.subtotal} amount={invoice.subtotal} />}
    {showDiscount && <SummaryRow label={copy.shared.discount} amount={-Math.abs(invoice.discountTotal ?? 0)} discount />}
    {Boolean(invoice.tax) && <SummaryRow label={copy.shared.tax} amount={invoice.tax ?? 0} />}
    {Boolean(invoice.adminFee) && <SummaryRow label={copy.shared.adminFee} amount={invoice.adminFee ?? 0} />}
    <Row style={{ borderTop: `1px solid ${colors.line}`, marginTop: 4 }}>
      <Column style={{ width: '50%', paddingTop: 18 }}><Text className="email-ink" style={{ margin: 0, color: colors.ink, fontSize: 15, fontWeight: 700 }}>{copy.shared.total}</Text></Column>
      <Column style={{ width: '50%', paddingTop: 18, textAlign: 'right' }}><Text style={{ margin: 0, color: accent, fontSize: 21, lineHeight: '26px', fontWeight: 700 }}>{formatRupiah(invoice.total)}</Text></Column>
    </Row>
  </Section>
}
