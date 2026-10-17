export type EmailDate = Date | string

export interface BusinessBrand {
  name: string
  logoUrl?: string
  brandColor?: string
  address?: string
  supportEmail?: string
  supportUrl?: string
}

export interface Customer {
  name: string
  email: string
}

export interface InvoiceItem {
  name: string
  quantity: number
  unitPrice: number
  subtotal: number
}

export type InvoiceStatus = 'paid' | 'pending' | 'overdue' | 'failed' | 'cancelled' | 'refunded'

export interface Invoice {
  number: string
  status: InvoiceStatus
  issuedAt: EmailDate
  dueDate?: EmailDate
  paidAt?: EmailDate
  periodStart?: EmailDate
  periodEnd?: EmailDate
  paymentMethod?: string
  items: InvoiceItem[]
  subtotal: number
  discountTotal?: number
  tax?: number
  adminFee?: number
  total: number
  currency: 'IDR'
  url: string
}

export interface BaseEmailProps {
  brand: BusinessBrand
  customer: Customer
  invoice: Invoice
  subjectOverride?: string
  messageOverride?: string
  footerNote?: string
}

export interface PaymentSuccessProps extends BaseEmailProps {
  receiptUrl?: string
}

export interface InvoiceIssuedProps extends BaseEmailProps {
  paymentInstructions?: string
  publicActions?: {
    downloadPath: string
    checkoutPath?: string
  }
}

export interface PaymentFailedProps extends BaseEmailProps {
  failureReason: string
  retryAt?: EmailDate
  actionUrl: string
}

export interface TrialStartedProps {
  brand: BusinessBrand
  customer: Customer
  trial: {
    plan: string
    endsAt: EmailDate
    billingCycle: 'monthly' | 'yearly'
    amount: number
    items: Array<{ name: string; quantity: number; subtotal: number }>
    autoDebitEnabled: boolean
    paymentMethod?: string
  }
  footerNote?: string
}

export interface EmailPropsByType {
  'payment-success': PaymentSuccessProps
  'invoice-issued': InvoiceIssuedProps
  'payment-failed': PaymentFailedProps
  'trial-started': TrialStartedProps
}

export type EmailType = keyof EmailPropsByType

export type EmailPayload = {
  [K in EmailType]: { type: K; props: EmailPropsByType[K] }
}[EmailType]
