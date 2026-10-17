import { lazy, Suspense } from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import Home from './pages/Home'
import Login from './pages/Login'
import Register from './pages/Register'
import ForgotPassword from './pages/ForgotPassword'
import ResetPassword from './pages/ResetPassword'
import RequireAuth from './routes/RequireAuth'
import RequireGuest from './routes/RequireGuest'
import RequireRole from './routes/RequireRole'
import DashboardIndex from './routes/DashboardIndex'
import DashboardLayout from './components/layout/DashboardLayout'
import PelangganPage from './pages/dashboard/PelangganPage'
import LanggananPage from './pages/dashboard/LanggananPage'
import KatalogPage from './pages/dashboard/KatalogPage'
import KuponPage from './pages/dashboard/KuponPage'
import PengaturanPage from './pages/dashboard/PengaturanPage'
import PlatformAdminPage from './pages/dashboard/PlatformAdminPage'
import InvoicePage from './pages/dashboard/InvoicePage'
import InvoiceCustomizationPage from './pages/dashboard/InvoiceCustomizationPage'
import PlatformBillingRedirect from './routes/PlatformBillingRedirect'
import IntegrationKeysPage from './pages/dashboard/IntegrationKeysPage'
import PaymentGatewayPage from './pages/dashboard/PaymentGatewayPage'
import WebhookLogsPage from './pages/dashboard/WebhookLogsPage'

const CheckoutPage = lazy(() => import('./pages/CheckoutPage'))
const CustomerPortalPage = lazy(() => import('./pages/CustomerPortalPage'))

function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />
      <Route path="/reset-password" element={<ResetPassword />} />
      <Route path="/portal/:site" element={<Suspense fallback={<p className="p-6 text-body">Memuat portal pelanggan...</p>}><CustomerPortalPage /></Suspense>} />
      <Route path="/checkout/embed" element={<Suspense fallback={<p className="p-6 text-body">Memuat checkout...</p>}><CheckoutPage /></Suspense>} />
      <Route path="/checkout/:token" element={<Suspense fallback={<p className="p-6 text-body">Memuat checkout...</p>}><CheckoutPage /></Suspense>} />

      <Route element={<RequireGuest />}>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
      </Route>

      <Route element={<RequireAuth />}>
        <Route path="/dashboard" element={<DashboardLayout />}>
          <Route element={<RequireRole roles={['super_admin']} fallback="/dashboard" />}>
            <Route path="platform" element={<PlatformAdminPage />} />
          </Route>
          <Route index element={<DashboardIndex />} />
          <Route element={<RequireRole roles={['admin', 'sales']} fallback="/dashboard" />}>
            <Route path="pelanggan" element={<PelangganPage />} />
            <Route path="langganan" element={<LanggananPage />} />
            <Route path="katalog" element={<KatalogPage />} />
            <Route path="kupon" element={<KuponPage />} />
          </Route>
          <Route element={<RequireRole roles={['admin', 'sales', 'finance']} fallback="/dashboard" />}>
            <Route path="invoices" element={<InvoicePage />} />
            <Route path="email-pelanggan" element={<Navigate to="/dashboard" replace />} />
          </Route>
          <Route element={<RequireRole roles={['admin']} fallback="/dashboard" />}>
			<Route path="tagihan-platform" element={<PlatformBillingRedirect />} />
            <Route path="tagihan" element={<PlatformBillingRedirect />} />
            <Route path="pengaturan" element={<PengaturanPage />} />
            <Route path="pengaturan/payment-gateway" element={<PaymentGatewayPage />} />
            <Route path="webhook-logs" element={<WebhookLogsPage />} />
            <Route path="pengaturan/webhook-logs" element={<Navigate to="/dashboard/webhook-logs" replace />} />
            <Route path="integrasi" element={<IntegrationKeysPage />} />
            <Route path="pengaturan/invoice" element={<InvoiceCustomizationPage />} />
          </Route>
        </Route>
      </Route>
    </Routes>
  )
}

export default App
