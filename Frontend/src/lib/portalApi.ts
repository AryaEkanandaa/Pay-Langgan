import { API_BASE_URL } from './api'

export class PortalError extends Error {
  readonly status: number
  constructor(message: string, status: number) { super(message); this.status = status }
}
export interface PortalInvoice {
  id: number; invoice_number: string; amount: number; status: string; plan_name: string
  due_date: string | null; paid_at: string | null; created_at: string
}
export interface PortalSubscription {
  id: number; plan: string; price: number; status: string; cycle: string
  start_date: string; next_billing_date: string | null; trial_ends_at: string | null
}
export interface PortalOverview {
  business_name: string; email: string; subscriptions: PortalSubscription[]; invoices: PortalInvoice[]
  subscription_total: number; invoice_total: number
}
export interface PortalInvoiceDetail {
  invoice: PortalInvoice
  items: { id: number; description: string; quantity: number; subtotal: number }[]
}
export async function portalRequest<T>(site: string, path: string, options: { token?: string; body?: unknown; signal?: AbortSignal } = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}/api/v1/portal/${encodeURIComponent(site)}${path}`, {
    method: options.body === undefined ? 'GET' : 'POST',
    headers: { 'Content-Type': 'application/json', ...(options.token ? { Authorization: `Bearer ${options.token}` } : {}) },
    credentials: 'omit', cache: 'no-store', signal: options.signal,
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  })
  const result = await response.json().catch(() => null) as { success?: boolean; data?: T; message?: string; error?: string } | null
  if (!response.ok || !result?.success) throw new PortalError(result?.error || result?.message || 'Portal belum dapat dimuat. Coba lagi.', response.status)
  return result.data as T
}
