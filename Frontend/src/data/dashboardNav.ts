import { LayoutDashboard, Users, Repeat, LayoutGrid, Ticket, Settings, type LucideIcon } from 'lucide-react'

export interface DashboardNavItem {
  label: string
  to: string
  icon: LucideIcon
}

export const dashboardNavItems: DashboardNavItem[] = [
  { label: 'Ringkasan', to: '/dashboard', icon: LayoutDashboard },
  { label: 'Pelanggan', to: '/dashboard/pelanggan', icon: Users },
  { label: 'Langganan', to: '/dashboard/langganan', icon: Repeat },
  { label: 'Katalog', to: '/dashboard/katalog', icon: LayoutGrid },
  { label: 'Kupon', to: '/dashboard/kupon', icon: Ticket },
  { label: 'Pengaturan Bisnis', to: '/dashboard/pengaturan', icon: Settings },
]
