import { Text } from '@react-email/components'
import type { PaymentFailedProps } from '../types'
import { copy } from '../lib/i18n'
import { formatDate } from '../lib/format'
import { colors } from '../lib/theme'
import { EmailLayout } from '../components/EmailLayout'
import { BrandHeader } from '../components/BrandHeader'
import { Banner } from '../components/Banner'
import { InfoBlock } from '../components/InfoBlock'
import { ItemsTable } from '../components/ItemsTable'
import { Notice } from '../components/Notice'
import { CtaButton } from '../components/CtaButton'

function PaymentFailedEmail({ brand, customer, invoice, failureReason, retryAt, actionUrl, messageOverride, footerNote }: PaymentFailedProps) {
  const strings = copy.paymentFailed
  const retryDate = formatDate(retryAt)
  return <EmailLayout brand={brand} preheader={strings.preheader(invoice.number)} closing={footerNote}>
    <BrandHeader brand={brand} documentTitle={strings.document} number={invoice.number} status="failed" />
    {!messageOverride && <Text className="email-ink" style={{ margin: '0 0 14px', color: colors.ink, fontSize: 14 }}>{copy.shared.greeting(customer.name)}</Text>}
    <Banner tone="danger" title={strings.title}>{messageOverride || strings.intro(invoice.number)}</Banner>
    <InfoBlock customer={customer} invoice={invoice} />
    <ItemsTable invoice={invoice} brandColor={brand.brandColor} />
    <Notice title={strings.reason}>{failureReason}</Notice>
    {retryDate && <Text className="email-muted" style={{ margin: '12px 0 0', color: colors.muted, fontSize: 12 }}>{strings.retryAt(retryDate)}</Text>}
    <CtaButton href={actionUrl} brandColor={brand.brandColor}>{strings.cta}</CtaButton>
  </EmailLayout>
}

PaymentFailedEmail.PreviewProps = {
  brand: { name: 'Kelana Workspace', brandColor: '#3B4EFF', supportEmail: 'bantuan@kelana.example' },
  customer: { name: 'Arya Ekananda', email: 'arya@example.com' },
  invoice: {
    number: 'INV-20261015-002', status: 'failed', issuedAt: '2026-10-15T00:00:00+08:00', dueDate: '2026-10-15T00:00:00+08:00',
    items: [{ name: 'Enterprise Bulanan', quantity: 1, unitPrice: 200000, subtotal: 200000 }],
    subtotal: 200000, total: 200000, currency: 'IDR', url: 'https://paylanggan.example/invoice/preview',
  },
  failureReason: 'Bank penerbit menolak pembayaran kartu. Periksa kartu Anda atau gunakan kartu lain.',
  retryAt: '2026-10-17T00:00:00+08:00', actionUrl: 'https://paylanggan.example/portal/preview',
} satisfies PaymentFailedProps

export default PaymentFailedEmail
