import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import type { UserRole } from '../lib/api'

interface RequireRoleProps {
  roles: UserRole[]
  fallback: string
}

export default function RequireRole({ roles, fallback }: RequireRoleProps) {
  const { auth } = useAuth()
  const location = useLocation()

  if (!auth) return <Navigate to="/login" replace state={{ from: location }} />
  if (!roles.includes(auth.user.role)) return <Navigate to={fallback} replace />

  return <Outlet />
}
