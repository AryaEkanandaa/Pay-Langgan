import { Navigate, useLocation } from 'react-router-dom'

export default function PlatformBillingRedirect() {
  const location = useLocation()
  // Preserve Xendit return parameters so the existing checkout can resume.
  return <Navigate replace to={{ pathname: '/dashboard/pengaturan', search: location.search, hash: '#langganan-tagihan' }} />
}
