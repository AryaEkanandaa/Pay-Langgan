import { describe, expect, it } from 'vitest'
import { formatDate, formatRupiah, safeBrandColor, safeLogoUrl } from '../lib/format'

describe('email formatters', () => {
  it('formats rupiah with a minus sign for discounts', () => {
    expect(formatRupiah(218500)).toBe('Rp 218.500')
    expect(formatRupiah(-11500)).toBe('-Rp 11.500')
  })

  it('formats dates in the business timezone', () => {
    expect(formatDate('2026-10-08T16:30:00Z')).toBe('9 Okt 2026')
    expect(formatDate('invalid')).toBeUndefined()
  })

  it('accepts only public HTTPS PNG or JPEG logos', () => {
    expect(safeLogoUrl('https://assets.example.com/logo.png?version=1')).toBe('https://assets.example.com/logo.png?version=1')
    expect(safeLogoUrl('http://localhost/logo.png')).toBeUndefined()
    expect(safeLogoUrl('https://assets.example.com/logo.svg')).toBeUndefined()
    expect(safeBrandColor('not-a-color')).toBe('#3B4EFF')
  })
})
