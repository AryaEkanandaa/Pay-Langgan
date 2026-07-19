const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export class ApiError extends Error {}

interface ApiEnvelope<T> {
  success: boolean
  message: string
  data?: T
  error?: string
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })

  const json = (await res.json().catch(() => null)) as ApiEnvelope<T> | null

  if (!res.ok || !json?.success) {
    throw new ApiError(json?.error || json?.message || 'Terjadi kesalahan, coba lagi.')
  }

  return json.data as T
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

export function login(payload: LoginPayload) {
  return post<LoginResult>('/api/v1/login', payload)
}

export function signup(payload: SignupPayload) {
  return post<SignupResult>('/api/v1/signup', payload)
}
