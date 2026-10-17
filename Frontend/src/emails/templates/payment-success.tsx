import { Text } from '@react-email/components'
import type { PaymentSuccessProps } from '../types'
import { copy } from '../lib/i18n'
import { colors } from '../lib/theme'
import { EmailLayout } from '../components/EmailLayout'
import { BrandHeader } from '../components/BrandHeader'
import { Banner } from '../components/Banner'
import { InfoBlock } from '../components/InfoBlock'
import { ItemsTable } from '../components/ItemsTable'
import { CtaButton } from '../components/CtaButton'

function PaymentSuccessEmail({ brand, customer, invoice, receiptUrl, messageOverride, footerNote }: PaymentSuccessProps) {
  const strings = copy.paymentSuccess
  return <EmailLayout brand={brand} preheader={strings.preheader(invoice.number, brand.name)} closing={footerNote}>
    <BrandHeader brand={brand} documentTitle={strings.document} number={invoice.number} status="paid" />
    {!messageOverride && <Text className="email-ink" style={{ margin: '0 0 14px', color: colors.ink, fontSize: 14 }}>{copy.shared.greeting(customer.name)}</Text>}
    <Banner tone="success" title={strings.title}>{messageOverride || strings.intro(invoice.number)}</Banner>
    <InfoBlock customer={customer} invoice={invoice} />
    <ItemsTable invoice={invoice} brandColor={brand.brandColor} />
    <CtaButton href={receiptUrl ?? invoice.url} brandColor={brand.brandColor}>{strings.cta}</CtaButton>
  </EmailLayout>
}

PaymentSuccessEmail.PreviewProps = {
  brand: { name: 'Kelana Workspace', brandColor: '#3B4EFF', supportEmail: 'bantuan@kelana.example' },
  customer: { name: 'Arya Ekananda', email: 'arya@example.com' },
  invoice: {
    number: 'INV-20261008-tlwp6pgc', status: 'paid', issuedAt: '2026-10-09T00:00:00+08:00',
    dueDate: '2026-10-08T00:00:00+08:00', paidAt: '2026-10-08T00:00:00+08:00', paymentMethod: 'Kartu',
    items: [
      { name: 'Enterprise Bulanan', quantity: 1, unitPrice: 200000, subtotal: 200000 },
      { name: 'Laporan Bisnis Lanjutan', quantity: 1, unitPrice: 30000, subtotal: 30000 },
      { name: 'Kupon 1STTIME', quantity: 1, unitPrice: -11500, subtotal: -11500 },
    ],
    subtotal: 230000, discountTotal: 11500, total: 218500, currency: 'IDR', url: 'https://paylanggan.example/invoice/preview',
  },
} satisfies PaymentSuccessProps

export default PaymentSuccessEmail
