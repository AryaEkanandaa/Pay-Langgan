import { Column, Img, Row, Section, Text } from '@react-email/components'
import { safeLogoUrl } from '../lib/format'

type Tone = 'success' | 'info' | 'warning' | 'danger'
const tones: Record<Tone, { background: string; border: string; title: string }> = {
  success: { background: '#eaf8f1', border: '#0d9869', title: '#086e4f' },
  info: { background: '#edf2ff', border: '#3B4EFF', title: '#2545b2' },
  warning: { background: '#fff5da', border: '#b77a00', title: '#805500' },
  danger: { background: '#fff0ed', border: '#d33b2e', title: '#a5291c' },
}

interface BannerProps {
  tone: Tone
  title: string
  children: string
  iconUrl?: string
}

export function Banner({ tone, title, children, iconUrl }: BannerProps) {
  const style = tones[tone]
  const icon = safeLogoUrl(iconUrl)
  return <Section style={{ backgroundColor: style.background, borderLeft: `4px solid ${style.border}`, padding: '14px 16px', marginBottom: 24 }}>
    <Row>
      {icon && <Column style={{ width: 34, verticalAlign: 'top' }}><Img src={icon} alt="" width="24" height="24" style={{ display: 'block' }} /></Column>}
      <Column style={{ verticalAlign: 'top' }}>
        <Text style={{ margin: 0, color: style.title, fontSize: 15, lineHeight: '21px', fontWeight: 700 }}>{title}</Text>
        <Text style={{ margin: '5px 0 0', color: '#26364f', fontSize: 13, lineHeight: '20px' }}>{children}</Text>
      </Column>
    </Row>
  </Section>
}
