import type { EmailDate } from '../types'

export function formatRupiah(amount: number): string {
  if (!Number.isFinite(amount)) return 'Rp 0'
  const absolute = new Intl.NumberFormat('id-ID', { maximumFractionDigits: 0 }).format(Math.abs(amount))
  return `${amount < 0 ? '-' : ''}Rp ${absolute}`
}

export function formatDate(value?: EmailDate): string | undefined {
  if (!value) return undefined
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime())) return undefined
  return new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'short', year: 'numeric', timeZone: 'Asia/Makassar' }).format(date)
}

export function safeHttpsUrl(value?: string): string | undefined {
  if (!value) return undefined
  try {
    const url = new URL(value)
    if (url.protocol !== 'https:' || url.username || url.password) return undefined
    return url.toString()
  } catch { return undefined }
}

export function safeLogoUrl(value?: string): string | undefined {
  const url = safeHttpsUrl(value)
  if (!url) return undefined
  return /\.(png|jpe?g)$/i.test(new URL(url).pathname) ? url : undefined
}

export function safeBrandColor(value?: string): string {
  return value && /^#[0-9a-fA-F]{6}$/.test(value) ? value : '#3B4EFF'
}
