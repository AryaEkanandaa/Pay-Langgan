import { createContext, useContext, useState, type ReactNode } from 'react'
import type { UserDTO } from '../lib/api'

interface AuthState {
  token: string
  user: UserDTO
}

interface AuthContextValue {
  auth: AuthState | null
  login: (state: AuthState, remember?: boolean) => void
  logout: () => void
}

const STORAGE_KEY = 'paylanggan_auth'

const AuthContext = createContext<AuthContextValue | null>(null)

function readStoredAuth(): AuthState | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY) ?? sessionStorage.getItem(STORAGE_KEY)
    return raw ? (JSON.parse(raw) as AuthState) : null
  } catch {
    return null
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [auth, setAuth] = useState<AuthState | null>(() => readStoredAuth())

  function login(state: AuthState, remember = true) {
    setAuth(state)
    const target = remember ? localStorage : sessionStorage
    const other = remember ? sessionStorage : localStorage
    other.removeItem(STORAGE_KEY)
    target.setItem(STORAGE_KEY, JSON.stringify(state))
  }

  function logout() {
    setAuth(null)
    localStorage.removeItem(STORAGE_KEY)
    sessionStorage.removeItem(STORAGE_KEY)
  }

  return <AuthContext.Provider value={{ auth, login, logout }}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return ctx
}
