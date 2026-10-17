export const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export interface WebhookLogDTO {
  id: number; provider: string; event_type: string; reference_id: string
  status: 'accepted' | 'duplicate' | 'failed'; http_status: number; error_code: string
  duration_ms: number; received_at: string; processed: boolean
}
export interface WebhookLogPageDTO { items: WebhookLogDTO[]; total: number; page: number; limit: number }
export async function getWebhookLogs(token: string, page: number, provider: string, status: string, signal?: AbortSignal) {
  const { data } = await request<WebhookLogPageDTO>('/api/v1/settings/webhook-logs', { token, signal, query: { page, provider, status } })
  return data
}

export interface WebhookUnreadDTO { count: number; last_seen_at: string; checked_at: string }
export async function getNewWebhookCount(token: string, signal?: AbortSignal) {
  const { data } = await request<WebhookUnreadDTO>('/api/v1/settings/webhook-logs/new-count', { token, signal })
  return data
}
export async function markWebhookLogsSeen(token: string, through: string, signal?: AbortSignal) {
  const { data } = await request<WebhookUnreadDTO>('/api/v1/settings/webhook-logs/seen', { token, signal, method: 'POST', body: { seen_through: through } })
  return data
}

export class ApiError extends Error {}

export type CustomerEmailKind = 'invoice_issued' | 'payment_succeeded' | 'trial_started' | 'trial_ending' | 'invoice_due'
export type CustomerEmailStatus = 'queued' | 'sending' | 'sent' | 'previewed' | 'failed' | 'cancelled'
export interface CustomerEmailLog {
  id: number; source: 'invoice' | 'checkout'; kind: CustomerEmailKind; invoice_id: number | null
  customer_name: string; recipient: string; status: CustomerEmailStatus; attempts: number
  last_error?: string; created_at: string; sent_at?: string
}
export interface CustomerEmailPage { items: CustomerEmailLog[]; total: number; page: number; limit: number; mode: string }
export interface EmailReminderPreferences { enabled: boolean; enabled_at?: string }
export async function getCustomerEmailHistory(token: string, page: number, status: string, kind: string, search: string, signal?: AbortSignal) {
  const { data } = await request<CustomerEmailPage>('/api/v1/email-history', { token, signal, query: { page, status, kind, search } }); return data
}
export async function retryCustomerEmail(token: string, item: CustomerEmailLog) {
  await request<null>(`/api/v1/email-history/${item.source}/${item.id}/retry`, { token, method: 'POST' })
}
export async function getCustomerEmailPreview(token: string, item: CustomerEmailLog, signal?: AbortSignal) {
  const { data } = await request<{ subject: string; html: string }>(`/api/v1/email-history/${item.source}/${item.id}/preview`, { token, signal }); return data
}
export async function getEmailReminderPreferences(token: string, signal?: AbortSignal) {
  const { data } = await request<EmailReminderPreferences>('/api/v1/email-reminders', { token, signal }); return data
}
export async function setEmailReminderPreferences(token: string, enabled: boolean) {
  const { data } = await request<EmailReminderPreferences>('/api/v1/email-reminders', { token, method: 'PUT', body: { enabled } }); return data
}

export async function requestPasswordReset(email: string, signal?: AbortSignal) {
  await request<null>('/api/v1/forgot-password', { method: 'POST', body: { email }, signal })
}

export async function resetPassword(token: string, password: string, confirmation: string, signal?: AbortSignal) {
  await request<null>('/api/v1/reset-password', { method: 'POST', body: { token, password, password_confirmation: confirmation }, signal })
}

interface ApiEnvelope<T> {
  success: boolean
  message: string
  data?: T
  error?: string
  meta?: { page: number; limit: number; total: number }
}

interface RequestOptions {
  signal?: AbortSignal
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  token?: string
  body?: unknown
  query?: Record<string, string | number | undefined>
}

async function request<T>(
  path: string,
  options: RequestOptions = {},
): Promise<{ data: T; meta?: ApiEnvelope<T>['meta'] }> {
  const { method = 'GET', token, body, query } = options

  const url = new URL(`${API_BASE_URL}${path}`)
  if (query) {
    Object.entries(query).forEach(([key, value]) => {
      if (value !== undefined && value !== '') url.searchParams.set(key, String(value))
    })
  }

  const headers: Record<string, string> = {}
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(url.toString(), {
    signal: options.signal,
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  const json = (await res.json().catch(() => null)) as ApiEnvelope<T> | null

  if (!res.ok || !json?.success) {
    throw new ApiError(json?.error || json?.message || 'Terjadi kesalahan, coba lagi.')
  }

  return { data: json.data as T, meta: json.meta }
}

export type UserRole = 'super_admin' | 'admin' | 'sales' | 'finance'

export interface UserDTO {
  id: number
  business_id: string
  name: string
  email: string
  role: UserRole
}

export interface BusinessDTO {
  id: string
  name: string
  status: string
  meta?: JSONMap | null
}

export interface LoginPayload {
  email: string
  password: string
}

export interface LoginResult {
  token: string
  user: UserDTO
}

export interface SignupPayload {
  business_name: string
  name: string
  email: string
  password: string
}

export interface SignupResult {
  token: string
  user: UserDTO
  business: BusinessDTO
}

export async function login(payload: LoginPayload) {
  const { data } = await request<LoginResult>('/api/v1/login', { method: 'POST', body: payload })
  return data
}

export async function signup(payload: SignupPayload) {
  const { data } = await request<SignupResult>('/api/v1/signup', {
    method: 'POST',
    body: payload,
  })
  return data
}

export interface BusinessUpdatePayload {
  name?: string
  meta?: JSONMap | null
}

export async function getMyBusiness(token: string) {
  const { data } = await request<BusinessDTO>('/api/v1/businesses/me', { token })
  return data
}

export async function updateMyBusiness(token: string, payload: BusinessUpdatePayload) {
  await request<null>('/api/v1/businesses/me', { method: 'PUT', token, body: payload })
  return getMyBusiness(token)
}

export interface BusinessPaymentGatewayDTO {
  provider: 'xendit' | 'midtrans'
  environment: 'sandbox' | 'production'
  configured_at: string
  updated_at: string
  verified_at: string | null
  account_label?: string
  webhook_url: string
  status: 'configured' | 'verified'
}

export interface BusinessPaymentGatewayInput {
  provider: 'xendit' | 'midtrans'
  environment: 'sandbox' | 'production'
  xendit_secret_key?: string
  xendit_callback_token?: string
  midtrans_merchant_id?: string
  midtrans_client_key?: string
  midtrans_server_key?: string
}

export interface IntegrationKey {
  id: number
  name: string
  created_at: string
  revoked_at?: string | null
}

export interface CreatedIntegrationKey extends IntegrationKey {
  secret_key: string
}

export async function listIntegrationKeys(token: string) {
  const { data } = await request<IntegrationKey[]>('/api/v1/integrations/keys', { token })
  return data ?? []
}

export async function createIntegrationKey(token: string, name: string) {
  const { data } = await request<CreatedIntegrationKey>('/api/v1/integrations/keys', {
    method: 'POST', token, body: { name },
  })
  return data
}

export async function revokeIntegrationKey(token: string, id: number) {
  await request<{ revoked: boolean }>(`/api/v1/integrations/keys/${id}`, { method: 'DELETE', token })
}

export interface EmbedSite {
  id: number
  public_id: string
  created_at: string
  revoked_at?: string | null
}

export async function listEmbedSites(token: string) {
  return (await request<EmbedSite[]>('/api/v1/integrations/embed-sites', { token })).data ?? []
}

export async function createEmbedSite(token: string, payload: { name?: string; origin?: string } = {}) {
  return (await request<EmbedSite>('/api/v1/integrations/embed-sites', { method: 'POST', token, body: payload })).data
}

export async function revokeEmbedSite(token: string, id: number) {
  await request<{ revoked: boolean }>(`/api/v1/integrations/embed-sites/${id}`, { method: 'DELETE', token })
}

export interface CreateStaffPayload {
  name: string
  email: string
  password: string
  role: 'sales' | 'finance'
}

export async function listBusinessUsers(token: string) {
  const { data } = await request<UserDTO[]>('/api/v1/users', { token })
  return data ?? []
}

export async function createStaffUser(token: string, payload: CreateStaffPayload) {
  const { data } = await request<UserDTO>('/api/v1/users', {
    method: 'POST',
    token,
    body: payload,
  })
  return data
}

export type CustomerStatus = 'active' | 'trial' | 'prospect' | 'inactive'

export interface Customer {
  status?: CustomerStatus
  id: number
  business_id: string
  name: string
  email: string | null
  contact: string | null
  created_at: string
}

export interface CustomerListParams {
  status?: CustomerStatus
  page?: number
  limit?: number
  search?: string
}

export interface CustomerListResult {
  items: Customer[]
  page: number
  limit: number
  total: number
}

export interface CustomerPayload {
  name: string
  email?: string | null
  contact?: string | null
}

export async function listCustomers(
  token: string,
  params: CustomerListParams = {},
): Promise<CustomerListResult> {
  const { data, meta } = await request<Customer[] | null>('/api/v1/customers', {
    token,
    query: { page: params.page, limit: params.limit, search: params.search, status: params.status },
  })
  const items = data ?? []

  return {
    items,
    page: meta?.page ?? params.page ?? 1,
    limit: meta?.limit ?? params.limit ?? 10,
    total: meta?.total ?? items.length,
  }
}

export async function createCustomer(token: string, payload: CustomerPayload) {
  const { data } = await request<Customer>('/api/v1/customers', {
    method: 'POST',
    token,
    body: payload,
  })
  return data
}

export async function updateCustomer(
  token: string,
  id: number,
  payload: Partial<CustomerPayload>,
) {
  const { data } = await request<Customer>(`/api/v1/customers/${id}`, {
    method: 'PUT',
    token,
    body: payload,
  })
  return data
}

export async function deleteCustomer(token: string, id: number) {
  await request<null>(`/api/v1/customers/${id}`, { method: 'DELETE', token })
}

export interface JSONMap {
  [key: string]: unknown
}

export interface ListParams {
  page?: number
  limit?: number
  search?: string
}

export interface ListResult<T> {
  items: T[]
  page: number
  limit: number
  total: number
}

export interface Service {
  id: number
  business_id: string
  name: string
  description: string
  meta: JSONMap | null
  created_at: string
}

export interface ServicePayload {
  name: string
  description?: string
  meta?: JSONMap | null
}

export async function listServices(token: string, params: ListParams = {}): Promise<ListResult<Service>> {
  const { data, meta } = await request<Service[] | null>('/api/v1/services', {
    token,
    query: { page: params.page, limit: params.limit, search: params.search },
  })
  const items = data ?? []
  return { items, page: meta?.page ?? params.page ?? 1, limit: meta?.limit ?? params.limit ?? 10, total: meta?.total ?? items.length }
}

export async function createService(token: string, payload: ServicePayload) {
  const { data } = await request<Service>('/api/v1/services', { method: 'POST', token, body: payload })
  return data
}

export async function updateService(token: string, id: number, payload: Partial<ServicePayload>) {
  const { data } = await request<Service>(`/api/v1/services/${id}`, { method: 'PUT', token, body: payload })
  return data
}

export async function deleteService(token: string, id: number) {
  await request<null>(`/api/v1/services/${id}`, { method: 'DELETE', token })
}

export interface Product {
  id: number
  service_id: number
  name: string
  description: string
  status: string
  meta: JSONMap | null
  created_at: string
}

export interface ProductPayload {
  service_id: number
  name: string
  description?: string
  status?: string
  meta?: JSONMap | null
}

export async function listProducts(token: string, params: ListParams = {}): Promise<ListResult<Product>> {
  const { data, meta } = await request<Product[] | null>('/api/v1/products', {
    token,
    query: { page: params.page, limit: params.limit, search: params.search },
  })
  const items = data ?? []
  return { items, page: meta?.page ?? params.page ?? 1, limit: meta?.limit ?? params.limit ?? 10, total: meta?.total ?? items.length }
}

export async function createProduct(token: string, payload: ProductPayload) {
  const { data } = await request<Product>('/api/v1/products', { method: 'POST', token, body: payload })
  return data
}

export async function updateProduct(token: string, id: number, payload: Partial<ProductPayload>) {
  const { data } = await request<Product>(`/api/v1/products/${id}`, { method: 'PUT', token, body: payload })
  return data
}

export async function deleteProduct(token: string, id: number) {
  await request<null>(`/api/v1/products/${id}`, { method: 'DELETE', token })
}

export interface Plan {
  id: number
  public_id: string
  product_id: number
  name: string
  price: number
  billing_cycle: string
  trial_days: number
  meta: JSONMap | null
  created_at: string
}

export interface PlanPayload {
  product_id: number
  name: string
  price: number
  billing_cycle: string
  trial_days?: number
  meta?: JSONMap | null
}

export async function listPlans(token: string, params: ListParams = {}): Promise<ListResult<Plan>> {
  const { data, meta } = await request<Plan[] | null>('/api/v1/plans', {
    token,
    query: { page: params.page, limit: params.limit, search: params.search },
  })
  const items = data ?? []
  return { items, page: meta?.page ?? params.page ?? 1, limit: meta?.limit ?? params.limit ?? 10, total: meta?.total ?? items.length }
}

export async function createPlan(token: string, payload: PlanPayload) {
  const { data } = await request<Plan>('/api/v1/plans', { method: 'POST', token, body: payload })
  return data
}

export async function updatePlan(token: string, id: number, payload: Partial<PlanPayload>) {
  const { data } = await request<Plan>(`/api/v1/plans/${id}`, { method: 'PUT', token, body: payload })
  return data
}

export async function deletePlan(token: string, id: number) {
  await request<null>(`/api/v1/plans/${id}`, { method: 'DELETE', token })
}

export interface AddOn {
  visible_in_checkout: boolean
  id: number
  product_id: number
  name: string
  price: number
  billing_cycle: string
  meta: JSONMap | null
  created_at: string
}

export interface AddOnPayload {
  visible_in_checkout?: boolean
  product_id: number
  name: string
  price: number
  billing_cycle: string
  meta?: JSONMap | null
}

export async function listAddOns(token: string, params: ListParams = {}): Promise<ListResult<AddOn>> {
  const { data, meta } = await request<AddOn[] | null>('/api/v1/add-ons', {
    token,
    query: { page: params.page, limit: params.limit, search: params.search },
  })
  const items = data ?? []
  return { items, page: meta?.page ?? params.page ?? 1, limit: meta?.limit ?? params.limit ?? 10, total: meta?.total ?? items.length }
}

export async function createAddOn(token: string, payload: AddOnPayload) {
  const { data } = await request<AddOn>('/api/v1/add-ons', { method: 'POST', token, body: payload })
  return data
}

export async function updateAddOn(token: string, id: number, payload: Partial<AddOnPayload>) {
  const { data } = await request<AddOn>(`/api/v1/add-ons/${id}`, { method: 'PUT', token, body: payload })
  return data
}

export async function deleteAddOn(token: string, id: number) {
  await request<null>(`/api/v1/add-ons/${id}`, { method: 'DELETE', token })
}

export interface DashboardSummary {
  range_stats?: DashboardRangeStats
  active_customers: number
  active_subscriptions: number
  mrr: number
  due_invoices: number
  revenue_this_month: number
  outstanding_amount: number
  overdue_amount: number
  payments_succeeded: number
  payments_failed: number
  autopay_succeeded: number
  autopay_failed: number
  new_subscriptions: number
  cancelled_subscriptions: number
  revenue_by_plan: Array<{ plan_id: number; name: string; amount: number }>
  monthly_performance: Array<{ month: string; value: number }>
  yearly_breakdown: Array<{ year: string; income: number }>
}

export interface DashboardRangePoint {
  date: string; revenue: number; new_subscriptions: number; cancelled_subscriptions: number
  payments_succeeded: number; payments_failed: number; autopay_succeeded: number; autopay_failed: number
}
export interface DashboardRangeStats {
  from: string; to: string; timezone: string; granularity: 'day' | 'week' | 'month'
  series: DashboardRangePoint[]; revenue_by_plan: DashboardSummary['revenue_by_plan']
}
export async function getDashboardSummary(token: string, period?: { from: string; to: string; granularity: 'day' | 'week' | 'month' }) {
  const { data } = await request<DashboardSummary>('/api/v1/dashboard/summary', { token, query: period })
  return data
}

export interface PlatformDashboardSummary {
  active_businesses: number
  past_due_businesses: number
  revenue_this_month: number
  outstanding_amount: number
  failed_autopay_this_month: number
}

export async function getPlatformDashboardSummary(token: string) {
  const { data } = await request<PlatformDashboardSummary>('/api/v1/platform/dashboard/summary', { token })
  return data
}

export type InvoiceStatus = 'pending' | 'paid' | 'failed' | 'cancelled'

export interface Invoice {
  checkout_status?: string
  id: number
  subscription_id: number
  invoice_number: string
  amount: number
  status: InvoiceStatus
  due_date: string | null
  paid_at: string | null
  created_at: string
  customer_id: number
  customer_name: string
  customer_email: string | null
  plan_name: string
}

export interface InvoiceItem {
  id: number
  invoice_id: number
  item_type: string
  description: string
  quantity: number
  unit_price: number
  subtotal: number
  meta: JSONMap
  created_at: string
}

export interface InvoicePayment {
  id: number
  invoice_id: number
  provider: string
  provider_transaction_id: string | null
  amount: number
  status: string
  paid_at: string | null
  created_at: string
}

export interface InvoiceDetail extends Invoice {
  items: InvoiceItem[]
  payment?: InvoicePayment
}

export interface InvoiceListParams extends ListParams {
  status?: InvoiceStatus
}

export async function listInvoices(
  token: string,
  params: InvoiceListParams = {},
): Promise<ListResult<Invoice>> {
  const { data, meta } = await request<Invoice[] | null>('/api/v1/invoices', {
    token,
    query: { page: params.page, limit: params.limit, search: params.search, status: params.status },
  })
  const items = data ?? []
  return {
    items,
    page: meta?.page ?? params.page ?? 1,
    limit: meta?.limit ?? params.limit ?? 10,
    total: meta?.total ?? items.length,
  }
}

export async function getInvoice(token: string, id: number, signal?: AbortSignal) {
  const { data } = await request<InvoiceDetail>(`/api/v1/invoices/${id}`, { token, signal })
  return data
}

export async function markInvoicePaid(token: string, id: number) {
  const { data } = await request<InvoiceDetail>(`/api/v1/invoices/${id}/mark-paid`, {
    method: 'POST',
    token,
  })
  return data
}

export interface CheckoutLink {
  token: string
  checkout_url: string
  expires_at: string
}

export interface PublicCheckout {
	card_registration?: boolean
	payment_provider?: 'xendit' | 'midtrans'
	billing_cycle?: 'monthly' | 'yearly'
	trial_days?: number
	trial_ends_at?: string
  items?: { type: string; description: string; quantity: number; subtotal: number }[]
  invoice_number: string
  amount: number
  invoice_status: InvoiceStatus | 'not_issued'
  checkout_status: 'created' | 'creating' | 'active' | 'completed' | 'expired' | 'failed' | 'cancelled'
  customer_name: string
  customer_email?: string
  customer_mobile?: string
  business_name: string
  plan_name: string
  expires_at: string
}

export interface StartCheckoutPayload {
	card_consent?: boolean
  name: string
  email: string
  mobile: string
  parent_origin?: string
}

export interface StartedCheckout {
	card_registration?: boolean
  provider: 'xendit' | 'midtrans'
  payment_session_id: string
  mode: 'COMPONENTS' | 'SNAP' | 'CORE'
  client_key?: string
  environment?: 'sandbox' | 'production'
  authentication_url?: string
  components_sdk_key?: string
  snap_token?: string
  snap_client_key?: string
  snap_script_url?: string
  status: 'active'
  expires_at: string
}

export async function createCheckoutLink(token: string, invoiceID: number) {
  const { data } = await request<CheckoutLink>(`/api/v1/invoices/${invoiceID}/checkout-links`, {
    method: 'POST',
    token,
  })
  return data
}

export async function getPublicCheckout(token: string) {
  const { data } = await request<PublicCheckout>(`/api/v1/checkout/${token}`)
  return data
}

export async function cancelPublicCheckout(token: string) {
  const { data } = await request<PublicCheckout>(`/api/v1/checkout/${token}/cancel`, { method: 'POST' })
  return data
}

export async function registerMidtransTrialCard(token: string, receipt: { saved_token_id: string; transaction_id: string }) {
  const { data } = await request<PublicCheckout>(`/api/v1/checkout/${token}/card-registration`, { method: 'POST', body: receipt })
  return data
}

export async function verifyMidtransCard(token: string) {
  const { data } = await request<PublicCheckout>(`/api/v1/checkout/${token}/card-payment/verify`, { method: 'POST' })
  return data
}

export async function startPublicCheckout(token: string, payload: StartCheckoutPayload) {
  const { data } = await request<StartedCheckout>(`/api/v1/checkout/${token}/session`, {
    method: 'POST',
    body: payload,
  })
  return data
}

export async function chargeMidtransCard(token: string, payload: { token_id: string; name: string; email: string; mobile: string }) {
  const { data } = await request<{ authentication_url?: string; status: string }>(`/api/v1/checkout/${token}/card-payment`, {
    method: 'POST', body: payload,
  })
  return data
}

export interface Coupon {
  id: number
  business_id: string
  code: string
  discount_type: 'percentage' | 'fixed'
  discount_value: number
  max_usage: number | null
  used_count: number
  expires_at: string | null
  created_at: string
}

export interface CouponPayload {
  code: string
  discount_type: 'percentage' | 'fixed'
  discount_value: number
  max_usage?: number | null
  expires_at?: string | null
}

export async function listCoupons(token: string, params: ListParams = {}): Promise<ListResult<Coupon>> {
  const { data, meta } = await request<Coupon[] | null>('/api/v1/coupons', {
    token,
    query: { page: params.page, limit: params.limit, search: params.search },
  })
  const items = data ?? []
  return {
    items,
    page: meta?.page ?? params.page ?? 1,
    limit: meta?.limit ?? params.limit ?? 10,
    total: meta?.total ?? items.length,
  }
}

export async function createCoupon(token: string, payload: CouponPayload) {
  const { data } = await request<Coupon>('/api/v1/coupons', {
    method: 'POST',
    token,
    body: payload,
  })
  return data
}

export async function updateCoupon(token: string, id: number, payload: Partial<CouponPayload>) {
  const { data } = await request<Coupon>(`/api/v1/coupons/${id}`, {
    method: 'PUT',
    token,
    body: payload,
  })
  return data
}

export async function deleteCoupon(token: string, id: number) {
  await request<null>(`/api/v1/coupons/${id}`, { method: 'DELETE', token })
}

export type SubscriptionStatus = 'pending' | 'trial' | 'active' | 'paused' | 'cancelled'

export interface Subscription {
  checkout_status?: string
  id: number
  customer_id: number
  plan_id: number
  status: SubscriptionStatus
  start_date: string
  next_billing_date: string | null
  end_date: string | null
  trial_ends_at: string | null
  meta: JSONMap | null
  created_at: string
}

export interface SubscriptionListParams extends ListParams {
  status?: SubscriptionStatus
}

export interface SubscriptionDetail extends Subscription {
  customer?: { id: number; name: string; email: string | null }
  plan?: {
    id: number
    name: string
    price: number
    billing_cycle: string
    trial_days: number
  }
  product?: { id: number; name: string }
  service?: { id: number; name: string }
  add_ons?: Array<{
    id: number
    add_on_id: number
    name: string
    price: number
    quantity: number
  }>
  coupons?: Array<{
    id: number
    coupon_id: number
    code: string
    discount_type: string
    discount_value: number
  }>
}

export async function listSubscriptions(
  token: string,
  params: SubscriptionListParams = {},
): Promise<ListResult<Subscription>> {
  const { data, meta } = await request<Subscription[] | null>('/api/v1/subscriptions', {
    token,
    query: {
      page: params.page,
      limit: params.limit,
      search: params.search,
      status: params.status,
    },
  })
  const items = data ?? []
  return {
    items,
    page: meta?.page ?? params.page ?? 1,
    limit: meta?.limit ?? params.limit ?? 10,
    total: meta?.total ?? items.length,
  }
}

export async function getSubscription(token: string, id: number) {
  const { data } = await request<SubscriptionDetail>(`/api/v1/subscriptions/${id}`, { token })
  return data
}

export async function cancelSubscription(token: string, id: number, reason = '') {
  const { data } = await request<SubscriptionDetail>(`/api/v1/subscriptions/${id}/cancel`, {
    method: 'POST',
    token,
    body: { reason },
  })
  return data
}

export async function pauseSubscription(token: string, id: number, reason = '') {
  const { data } = await request<SubscriptionDetail>(`/api/v1/subscriptions/${id}/pause`, {
    method: 'POST',
    token,
    body: { reason },
  })
  return data
}

export async function resumeSubscription(token: string, id: number) {
  const { data } = await request<SubscriptionDetail>(`/api/v1/subscriptions/${id}/resume`, {
    method: 'POST',
    token,
  })
  return data
}

export interface PlatformPlan {
  code: string
  name: string
  amount: number
  currency: string
  billing_cycle: string
}

export interface PlatformInvoice {
  id: number
  amount: number
  status: 'pending' | 'paid'
  period_start: string
  period_end: string
  due_date: string
  paid_at: string | null
  created_at: string
}

export interface PlatformSubscription {
  business_id: string
  status: 'not_subscribed' | 'pending' | 'active' | 'past_due' | 'cancelled'
  paid_through: string | null
  next_billing_date: string | null
  auto_renew: boolean
  consent_at: string | null
  pending_invoice?: PlatformInvoice
  checkout_available: boolean
  legacy_checkout_active: boolean
}

export interface PlatformCheckout {
  invoice_id: number
  amount: number
  payment_session_id: string
  mode: 'COMPONENTS'
  components_sdk_key: string
  status: string
}

export async function getPlatformPlan(token: string) {
  return (await request<PlatformPlan>('/api/v1/platform/plan', { token })).data
}

export async function getPlatformSubscription(token: string) {
  return (await request<PlatformSubscription>('/api/v1/platform/subscription', { token })).data
}

export async function getPlatformInvoices(token: string) {
  return (await request<PlatformInvoice[]>('/api/v1/platform/invoices', { token })).data
}

export async function startPlatformCheckout(token: string) {
  return (await request<PlatformCheckout>(
    '/api/v1/platform/subscription/checkout',
    { method: 'POST', token, body: { auto_renew_consent: true } },
  )).data
}

export async function resumePlatformCheckout(token: string) {
  return (await request<PlatformCheckout>('/api/v1/platform/subscription/checkout', { token })).data
}

export async function cancelPlatformSubscription(token: string) {
  return request<null>('/api/v1/platform/subscription/cancel', { method: 'POST', token })
}

export async function getBusinessPaymentGateway(token: string) {
  return (await request<BusinessPaymentGatewayDTO | null>('/api/v1/settings/payment-gateway', { token })).data
}

export async function saveBusinessPaymentGateway(token: string, payload: BusinessPaymentGatewayInput) {
  return (await request<BusinessPaymentGatewayDTO>('/api/v1/settings/payment-gateway', {
    method: 'PUT', token, body: payload,
  })).data
}

export async function deleteBusinessPaymentGateway(token: string) {
  await request<{ removed: boolean }>('/api/v1/settings/payment-gateway', { method: 'DELETE', token })
}

export interface CheckoutAddOnSelection { add_on_id: number; quantity: number }
export interface EmbedPlanPreview {
  business_name: string; plan_name: string; billing_cycle: string; amount: number
  due_now: number; trial_days: number; coupon_code?: string
  add_ons: Array<{ id: number; name: string; price: number }>
  items: Array<{ type: string; name: string; quantity: number; subtotal: number }>
}
export interface EmbedCheckoutOrder {
  checkout_url?: string; external_order_id: string; invoice_id: number; invoice_status: string
  subscription_id: number; subscription_status: string; trial_ends_at?: string
}
export async function getEmbedPlan(site: string, plan: string, addOns: CheckoutAddOnSelection[] = [], couponCode = '', signal?: AbortSignal) {
  const query = new URLSearchParams({ add_ons: JSON.stringify(addOns), coupon_code: couponCode })
  return (await request<EmbedPlanPreview>(`/api/v1/embed/sites/${encodeURIComponent(site)}/plans/${encodeURIComponent(plan)}?${query}`, { signal })).data
}
export async function createEmbedCheckout(site: string, plan: string, customer: { name: string; email: string; mobile: string }, options: { add_ons?: CheckoutAddOnSelection[]; coupon_code?: string; request_id?: string } = {}) {
  return (await request<EmbedCheckoutOrder>(`/api/v1/embed/sites/${encodeURIComponent(site)}/orders`, {
    method: 'POST', body: { ...(plan.startsWith('pl_plan_') ? { plan_public_id: plan } : { plan_id: Number(plan) }), customer, ...options },
  })).data
}


export interface InvoiceEmailState {
  mode: 'disabled' | 'preview' | 'smtp'
  delivery: { status: 'queued' | 'sending' | 'sent' | 'previewed' | 'failed'; recipient: string; attempts: number; last_error: string | null; sent_at: string | null } | null
}
export async function getInvoiceEmail(token: string, id: number, signal?: AbortSignal) {
  return (await request<InvoiceEmailState>(`/api/v1/invoices/${id}/email`, { token, signal })).data
}
export async function queueInvoiceEmail(token: string, id: number, signal?: AbortSignal) {
  return (await request<InvoiceEmailState>(`/api/v1/invoices/${id}/email`, { token, method: 'POST', signal })).data
}
export async function previewInvoiceEmail(token: string, id: number) {
  return (await request<{subject: string; html: string; logo_warning?: string}>(`/api/v1/invoices/${id}/email-preview`, { token })).data
}

export interface InvoiceAppearancePreview { subject: string; html: string; logo_warning: string }
export async function previewInvoiceAppearance(token: string, settings: JSONMap, kind: 'invoice_issued' | 'payment_succeeded', signal?: AbortSignal) {
  return (await request<InvoiceAppearancePreview>('/api/v1/invoice-appearance/preview', { token, method: 'POST', body: { settings, kind }, signal })).data
}
