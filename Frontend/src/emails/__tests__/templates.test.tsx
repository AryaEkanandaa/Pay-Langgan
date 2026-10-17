import { describe, expect, it } from 'vitest'
import { emailRegistry } from '../registry'
import { renderEmail } from '../render'
import type { EmailType } from '../types'

const expectedTypes = ['payment-success', 'invoice-issued', 'payment-failed'] as const satisfies readonly EmailType[]

describe('transactional email registry', () => {
  it('registers every template in this stage', () => {
    expect(Object.keys(emailRegistry).sort()).toEqual([...expectedTypes].sort())
  })

  it('renders payment success HTML and plain text', async () => {
    const result = await renderEmail('payment-success', emailRegistry['payment-success'].previewProps)
    expect(result.subject).toContain('INV-20261008-tlwp6pgc')
    expect(result.subject).toContain('Rp 218.500')
    expect(result.html).toContain('Dikirim melalui PayLanggan untuk Kelana Workspace')
    expect(result.text).toContain('Arya Ekananda')
    expect(result.html).toMatchSnapshot()
  })

  it('renders invoice issued HTML and plain text', async () => {
    const result = await renderEmail('invoice-issued', emailRegistry['invoice-issued'].previewProps)
    expect(result.subject).toContain('INV-20261015-001')
    expect(result.html).toContain('Bayar sekarang')
    expect(result.text).toContain('Profesional Bulanan')
    expect(result.html).toMatchSnapshot()
  })

  it('renders payment failure HTML and plain text', async () => {
    const result = await renderEmail('payment-failed', emailRegistry['payment-failed'].previewProps)
    expect(result.subject).toContain('INV-20261015-002')
    expect(result.html).toContain('Bank penerbit menolak pembayaran kartu')
    expect(result.text).toContain('Perbarui metode pembayaran')
    expect(result.html).toMatchSnapshot()
  })
})
