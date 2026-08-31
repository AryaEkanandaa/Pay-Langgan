import { Navigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import DashboardHome from '../pages/dashboard/DashboardHome'

export default function DashboardIndex() {
  const { auth } = useAuth()

  if (auth?.user.role === 'super_admin') return <Navigate to="/dashboard/platform" replace />

  return <DashboardHome />
}
