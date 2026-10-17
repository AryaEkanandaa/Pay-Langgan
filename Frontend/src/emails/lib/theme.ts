import type { CSSProperties } from 'react'
import { safeBrandColor } from './format'

export const colors = {
  page: '#f0f3fb', card: '#ffffff', ink: '#16213a', muted: '#50617f',
  line: '#dbe3ef', table: '#eef3f9', green: '#087a55', amber: '#805500',
  red: '#b42318', blue: '#2454a6',
} as const

export const emailCss = `
  @media only screen and (max-width: 480px) {
    .email-card { width: 100% !important; }
    .email-padding { padding: 24px 16px !important; }
    .email-info-column { display: block !important; width: 100% !important; padding: 0 0 16px !important; }
    .email-unit-price { display: none !important; }
    .email-cta { display: block !important; width: 100% !important; box-sizing: border-box !important; text-align: center !important; }
    .email-header-right { text-align: right !important; }
  }
  @media (prefers-color-scheme: dark) {
    .email-page { background-color: #0f172a !important; }
    .email-card { background-color: #182339 !important; }
    .email-ink { color: #f3f6fc !important; }
    .email-muted { color: #becbe0 !important; }
    .email-table-head { background-color: #263651 !important; }
    .email-line { border-color: #354562 !important; }
    .email-notice { background-color: #263651 !important; }
  }
`

export function theme(brandColor?: string) {
  const accent = safeBrandColor(brandColor)
  const text: CSSProperties = { color: colors.ink, fontFamily: 'Arial, Helvetica, sans-serif', fontSize: 14, lineHeight: '22px' }
  const muted: CSSProperties = { ...text, color: colors.muted, fontSize: 12 }
  return { accent, text, muted }
}
