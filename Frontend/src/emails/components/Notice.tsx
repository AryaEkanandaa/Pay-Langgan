import { Section, Text } from '@react-email/components'
import { colors } from '../lib/theme'

export function Notice({ title, children }: { title?: string; children: string }) {
  return <Section className="email-notice" style={{ marginTop: 22, padding: '12px 14px', backgroundColor: colors.table, borderLeft: `3px solid ${colors.line}` }}>
    {title && <Text className="email-ink" style={{ margin: '0 0 4px', color: colors.ink, fontSize: 12, fontWeight: 700 }}>{title}</Text>}
    <Text className="email-muted" style={{ margin: 0, color: colors.muted, fontSize: 12, lineHeight: '19px' }}>{children}</Text>
  </Section>
}
