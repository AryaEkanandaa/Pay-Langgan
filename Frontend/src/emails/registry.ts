import type { ComponentType } from 'react'
import type { EmailPropsByType, EmailType, TrialStartedProps } from './types'
import { formatRupiah } from './lib/format'
import { copy } from './lib/i18n'
import PaymentSuccessEmail from './templates/payment-success'
import InvoiceIssuedEmail from './templates/invoice-issued'
import PaymentFailedEmail from './templates/payment-failed'
import TrialStartedEmail from './templates/trial-started'

export interface EmailRegistryEntry<P> {
  component: ComponentType<P>
  subject: (props: P) => string
  previewProps: P
}

export const emailRegistry: { [K in EmailType]: EmailRegistryEntry<EmailPropsByType[K]> } = {
  'payment-success': {
    component: PaymentSuccessEmail,
    subject: ({ brand, invoice, subjectOverride }) => subjectOverride || copy.paymentSuccess.subject(invoice.number, formatRupiah(invoice.total), brand.name),
    previewProps: PaymentSuccessEmail.PreviewProps,
  },
  'invoice-issued': {
    component: InvoiceIssuedEmail,
    subject: ({ brand, invoice, subjectOverride }) => subjectOverride || copy.invoiceIssued.subject(invoice.number, formatRupiah(invoice.total), brand.name),
    previewProps: InvoiceIssuedEmail.PreviewProps,
  },
  'payment-failed': {
    component: PaymentFailedEmail,
    subject: ({ brand, invoice, subjectOverride }) => subjectOverride || copy.paymentFailed.subject(invoice.number, formatRupiah(invoice.total), brand.name),
    previewProps: PaymentFailedEmail.PreviewProps,
  },
  'trial-started': {
    component: TrialStartedEmail,
    subject: ({ brand }) => copy.trialStarted.subject(brand.name),
    previewProps: TrialStartedEmail.PreviewProps as TrialStartedProps,
  },
}
