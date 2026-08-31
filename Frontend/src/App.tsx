import { Routes, Route } from 'react-router-dom'
import Home from './pages/Home'
import Login from './pages/Login'
import Register from './pages/Register'
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

function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />

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
          </Route>
          <Route element={<RequireRole roles={['admin']} fallback="/dashboard" />}>
            <Route path="pengaturan" element={<PengaturanPage />} />
            <Route path="pengaturan/invoice" element={<InvoiceCustomizationPage />} />
          </Route>
        </Route>
      </Route>
    </Routes>
  )
}

export default App
