import type { CSSProperties, ReactNode } from 'react'
import { Body, Container, Head, Html, Link, Preview, Section, Text } from '@react-email/components'
import type { BusinessBrand } from '../types'
import { colors, emailCss } from '../lib/theme'
import { copy } from '../lib/i18n'
import { safeHttpsUrl } from '../lib/format'

interface EmailLayoutProps {
  brand: BusinessBrand
  preheader: string
  children: ReactNode
  closing?: string
}

const body: CSSProperties = { margin: 0, padding: '24px 10px', backgroundColor: colors.page, fontFamily: 'Arial, Helvetica, sans-serif' }
const card: CSSProperties = { width: '100%', maxWidth: 640, margin: '0 auto', backgroundColor: colors.card }

export function EmailLayout({ brand, preheader, children, closing = copy.shared.thankYou }: EmailLayoutProps) {
  const supportUrl = safeHttpsUrl(brand.supportUrl)
  return (
    <Html lang="id">
      <Head>
        <meta name="color-scheme" content="light dark" />
        <meta name="supported-color-schemes" content="light dark" />
        <style>{emailCss}</style>
      </Head>
      <Preview>{preheader}</Preview>
      <Body className="email-page" style={body}>
        <Container className="email-card" style={card}>
          <Section className="email-padding" style={{ padding: '36px 40px 32px' }}>
            {children}
            <Section style={{ borderTop: `1px solid ${colors.line}`, marginTop: 28, paddingTop: 18 }}>
              <Text className="email-ink" style={{ color: colors.ink, fontSize: 13, lineHeight: '20px', margin: '0 0 4px' }}>{closing}</Text>
              {brand.supportEmail && <Text className="email-muted" style={{ color: colors.muted, fontSize: 12, margin: '0 0 4px' }}>
                {copy.shared.help(brand.supportEmail)}
              </Text>}
              {supportUrl && <Text style={{ margin: '0 0 4px', fontSize: 12 }}><Link href={supportUrl} style={{ color: '#3B4EFF' }}>{copy.shared.helpCenter}</Link></Text>}
              <Text className="email-muted" style={{ color: colors.muted, fontSize: 11, margin: '10px 0 0' }}>{copy.shared.sentVia(brand.name)}</Text>
            </Section>
          </Section>
        </Container>
      </Body>
    </Html>
  )
}
