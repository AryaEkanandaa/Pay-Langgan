const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export class ApiError extends Error {}

interface ApiEnvelope<T> {
  success: boolean
  message: string
  data?: T
  error?: string
  meta?: { page: number; limit: number; total: number }
}

interface RequestOptions {
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

export interface UserDTO {
  id: number
  business_id: string
  name: string
  email: string
  role: string
}

export interface BusinessDTO {
  id: string
  name: string
  status: string
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

export interface Customer {
  id: number
  business_id: string
  name: string
  email: string | null
  contact: string | null
  created_at: string
}

export interface CustomerListParams {
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
  id: number
  product_id: number
  name: string
  price: number
  billing_cycle: string
  meta: JSONMap | null
  created_at: string
}

export interface AddOnPayload {
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
