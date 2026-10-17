import { Button, Section } from '@react-email/components'
import { safeHttpsUrl } from '../lib/format'
import { theme } from '../lib/theme'

interface CtaButtonProps {
  href: string
  children: string
  brandColor?: string
  variant?: 'primary' | 'secondary'
}

export function CtaButton({ href, children, brandColor, variant = 'primary' }: CtaButtonProps) {
  const safeUrl = safeHttpsUrl(href)
  if (!safeUrl) return null
  const { accent } = theme(brandColor)
  return <Section style={{ marginTop: 26 }}>
    <Button className="email-cta" href={safeUrl} style={{ display: 'inline-block', padding: '13px 20px', borderRadius: 9, backgroundColor: variant === 'primary' ? accent : '#e9eeff', color: variant === 'primary' ? '#ffffff' : '#2442a3', fontSize: 13, lineHeight: '18px', fontWeight: 700, textDecoration: 'none' }}>{children}</Button>
  </Section>
}
