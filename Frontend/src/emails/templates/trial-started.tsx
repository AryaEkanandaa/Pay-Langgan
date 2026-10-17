import { Column, Img, Row, Section, Text } from '@react-email/components'
import type { TrialStartedProps } from '../types'
import { copy } from '../lib/i18n'
import { formatDate, formatRupiah, safeLogoUrl } from '../lib/format'
import { colors } from '../lib/theme'
import { EmailLayout } from '../components/EmailLayout'
import { Banner } from '../components/Banner'
import { ItemsTable } from '../components/ItemsTable'

function TrialStartedEmail({ brand, customer, trial, footerNote }: TrialStartedProps) {
  const strings = copy.trialStarted
  const endsAt = formatDate(trial.endsAt) ?? 'sesuai tanggal pada halaman langganan'
  const cadence = trial.billingCycle === 'yearly' ? strings.yearly : strings.monthly
  const logo = safeLogoUrl(brand.logoUrl)
  const invoice = {
    number: 'TRIAL', status: 'pending' as const, issuedAt: trial.endsAt, items: trial.items.map((item) => ({
      name: item.name, quantity: item.quantity,
      unitPrice: item.quantity > 0 ? item.subtotal / item.quantity : item.subtotal,
      subtotal: item.subtotal,
    })), subtotal: trial.amount, total: trial.amount, currency: 'IDR' as const, url: '',
  }

  return <EmailLayout brand={brand} preheader={strings.preheader(trial.plan, brand.name)} closing={footerNote}>
    <Section style={{ borderBottom: `4px solid ${brand.brandColor ?? colors.blue}`, paddingBottom: 18, marginBottom: 22 }}>
      <Row>
        <Column style={{ verticalAlign: 'middle' }}>
          {logo && <Img src={logo} alt={`Logo ${brand.name}`} width="90" height="60" style={{ display: 'block', maxWidth: 90, maxHeight: 60, objectFit: 'contain' }} />}
          <Text className="email-ink" style={{ margin: logo ? '8px 0 0' : 0, color: colors.ink, fontSize: 18, lineHeight: '24px', fontWeight: 700 }}>{brand.name}</Text>
        </Column>
        <Column style={{ textAlign: 'right', verticalAlign: 'middle' }}>
          <Text style={{ margin: 0, color: brand.brandColor ?? colors.blue, fontSize: 18, fontWeight: 700 }}>{strings.document}</Text>
        </Column>
      </Row>
    </Section>
    <Text className="email-ink" style={{ margin: '0 0 12px', color: colors.ink, fontSize: 14 }}>{copy.shared.greeting(customer.name)}</Text>
    <Banner tone="success" title={strings.title}>{trial.autoDebitEnabled
      ? strings.automaticDebit(endsAt, formatRupiah(trial.amount), cadence)
      : strings.manualInvoice(endsAt)}</Banner>
    <Section style={{ marginBottom: 18, padding: '14px 16px', border: `1px solid ${colors.line}`, borderRadius: 12 }}>
      <Row>
        <Column style={{ width: '50%' }}>
          <Text className="email-muted" style={{ margin: '0 0 4px', color: colors.muted, fontSize: 12 }}>{strings.plan}</Text>
          <Text className="email-ink" style={{ margin: 0, color: colors.ink, fontSize: 14, fontWeight: 700 }}>{trial.plan}</Text>
        </Column>
        <Column style={{ width: '50%', textAlign: 'right' }}>
          <Text className="email-muted" style={{ margin: '0 0 4px', color: colors.muted, fontSize: 12 }}>{strings.trialEnds}</Text>
          <Text className="email-ink" style={{ margin: 0, color: colors.ink, fontSize: 14, fontWeight: 700 }}>{endsAt}</Text>
        </Column>
      </Row>
      <Text className="email-muted" style={{ margin: '12px 0 4px', color: colors.muted, fontSize: 12 }}>{strings.estimate}</Text>
      <Text style={{ margin: 0, color: brand.brandColor ?? colors.blue, fontSize: 21, fontWeight: 700 }}>{formatRupiah(trial.amount)} / {cadence}</Text>
      {trial.autoDebitEnabled && trial.paymentMethod && <Text className="email-muted" style={{ margin: '6px 0 0', color: colors.muted, fontSize: 12 }}>Metode pembayaran: kartu {trial.paymentMethod}</Text>}
    </Section>
    <Text className="email-ink" style={{ margin: '0 0 10px', color: colors.ink, fontSize: 14, fontWeight: 700 }}>{strings.cta}</Text>
    <ItemsTable invoice={invoice} brandColor={brand.brandColor} />
  </EmailLayout>
}

TrialStartedEmail.PreviewProps = {
  brand: { name: 'Kelana Workspace', brandColor: '#3B4EFF', supportEmail: 'bantuan@kelana.example' },
  customer: { name: 'Arya Ekananda', email: 'arya@example.com' },
  trial: {
    plan: 'Profesional Bulanan', endsAt: '2026-10-18T09:00:00+08:00', billingCycle: 'monthly', amount: 200000,
    items: [{ name: 'Profesional Bulanan', quantity: 1, subtotal: 200000 }], autoDebitEnabled: true, paymentMethod: 'Xendit',
  },
} satisfies TrialStartedProps

export default TrialStartedEmail
