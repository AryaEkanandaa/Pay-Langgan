import { Section, Text } from '@react-email/components'
import type { InvoiceIssuedProps } from '../types'
import { copy } from '../lib/i18n'
import { formatRupiah } from '../lib/format'
import { colors } from '../lib/theme'
import { EmailLayout } from '../components/EmailLayout'
import { BrandHeader } from '../components/BrandHeader'
import { Banner } from '../components/Banner'
import { InfoBlock } from '../components/InfoBlock'
import { ItemsTable } from '../components/ItemsTable'
import { Notice } from '../components/Notice'
import { CtaButton } from '../components/CtaButton'

function InvoiceIssuedEmail({ brand, customer, invoice, paymentInstructions, messageOverride, footerNote, publicActions }: InvoiceIssuedProps) {
  const strings = copy.invoiceIssued
  return <EmailLayout brand={brand} preheader={strings.preheader(invoice.number, formatRupiah(invoice.total))} closing={footerNote}>
    <BrandHeader brand={brand} documentTitle={strings.document} number={invoice.number} status={invoice.status} />
    {!messageOverride && <Text className="email-ink" style={{ margin: '0 0 14px', color: colors.ink, fontSize: 14 }}>{copy.shared.greeting(customer.name)}</Text>}
    <Banner tone="info" title={strings.title}>{messageOverride || strings.intro(invoice.number)}</Banner>
    <InfoBlock customer={customer} invoice={invoice} />
    <ItemsTable invoice={invoice} brandColor={brand.brandColor} />
    {paymentInstructions && <Notice title={strings.instructions}>{paymentInstructions}</Notice>}
    {publicActions ? <Section style={{ marginTop: 26 }}>
      {publicActions.checkoutPath && <form method="post" action={publicActions.checkoutPath} style={{ margin: '0 0 12px' }}>
        <button type="submit" style={{ display: 'inline-block', padding: '13px 20px', border: 0, borderRadius: 9, backgroundColor: brand.brandColor ?? '#3B4EFF', color: '#ffffff', fontSize: 13, lineHeight: '18px', fontWeight: 700, cursor: 'pointer' }}>Lanjut ke pembayaran</button>
      </form>}
      <a href={publicActions.downloadPath} style={{ display: 'inline-block', padding: '12px 20px', border: '1px solid #cbd5e1', borderRadius: 9, color: '#2442a3', fontSize: 13, lineHeight: '18px', fontWeight: 700, textDecoration: 'none' }}>Unduh PDF</a>
    </Section> : <CtaButton href={invoice.url} brandColor={brand.brandColor}>{strings.cta}</CtaButton>}
  </EmailLayout>
}

InvoiceIssuedEmail.PreviewProps = {
  brand: { name: 'Kelana Workspace', brandColor: '#3B4EFF' },
  customer: { name: 'Arya Ekananda', email: 'arya@example.com' },
  invoice: {
    number: 'INV-20261015-001', status: 'pending', issuedAt: '2026-10-15T00:00:00+08:00', dueDate: '2026-10-22T00:00:00+08:00',
    items: [{ name: 'Profesional Bulanan', quantity: 1, unitPrice: 100000, subtotal: 100000 }],
    subtotal: 100000, total: 100000, currency: 'IDR', url: 'https://paylanggan.example/invoice/preview',
  },
  paymentInstructions: copy.invoiceIssued.previewInstructions,
} satisfies InvoiceIssuedProps

export default InvoiceIssuedEmail
