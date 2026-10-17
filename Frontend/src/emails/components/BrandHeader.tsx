import { Column, Img, Row, Section, Text } from '@react-email/components'
import type { BusinessBrand, InvoiceStatus } from '../types'
import { safeLogoUrl } from '../lib/format'
import { colors, theme } from '../lib/theme'
import { StatusBadge } from './StatusBadge'

interface BrandHeaderProps {
  brand: BusinessBrand
  documentTitle: string
  number: string
  status: InvoiceStatus
}

export function BrandHeader({ brand, documentTitle, number, status }: BrandHeaderProps) {
  const logo = safeLogoUrl(brand.logoUrl)
  const { accent } = theme(brand.brandColor)
  return <Section style={{ borderBottom: `4px solid ${accent}`, paddingBottom: 22, marginBottom: 24 }}>
    <Row>
      <Column style={{ width: '52%', verticalAlign: 'top' }}>
        {logo && <Img src={logo} alt={`Logo ${brand.name}`} width="90" height="60" style={{ display: 'block', maxWidth: 90, maxHeight: 60, objectFit: 'contain' }} />}
        <Text className="email-ink" style={{ margin: logo ? '12px 0 0' : '0', color: colors.ink, fontSize: 18, fontWeight: 700, lineHeight: '24px' }}>{brand.name}</Text>
        {brand.address && <Text className="email-muted" style={{ margin: '5px 0 0', color: colors.muted, fontSize: 12, lineHeight: '18px', whiteSpace: 'pre-line' }}>{brand.address}</Text>}
      </Column>
      <Column className="email-header-right" style={{ width: '48%', textAlign: 'right', verticalAlign: 'top' }}>
        <Text style={{ margin: 0, color: accent, fontSize: 21, fontWeight: 700, lineHeight: '26px' }}>{documentTitle}</Text>
        <Text className="email-ink" style={{ margin: '5px 0 0', color: colors.ink, fontSize: 11, fontWeight: 700, overflowWrap: 'anywhere' }}>{number}</Text>
        <StatusBadge status={status} />
      </Column>
    </Row>
  </Section>
}
