import { Routes, Route } from 'react-router-dom'
import Home from './pages/Home'
import Login from './pages/Login'
import Register from './pages/Register'
import RequireAuth from './routes/RequireAuth'
import RequireGuest from './routes/RequireGuest'
import DashboardLayout from './components/layout/DashboardLayout'
import DashboardHome from './pages/dashboard/DashboardHome'
import PelangganPage from './pages/dashboard/PelangganPage'
import LanggananPage from './pages/dashboard/LanggananPage'
import KatalogPage from './pages/dashboard/KatalogPage'
import KuponPage from './pages/dashboard/KuponPage'
import PengaturanPage from './pages/dashboard/PengaturanPage'

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
          <Route index element={<DashboardHome />} />
          <Route path="pelanggan" element={<PelangganPage />} />
          <Route path="langganan" element={<LanggananPage />} />
          <Route path="katalog" element={<KatalogPage />} />
          <Route path="kupon" element={<KuponPage />} />
          <Route path="pengaturan" element={<PengaturanPage />} />
        </Route>
      </Route>
    </Routes>
  )
}

export default App
