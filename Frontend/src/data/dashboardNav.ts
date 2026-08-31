import { FileText, LayoutDashboard, Users, Repeat, LayoutGrid, Ticket, Settings, type LucideIcon } from 'lucide-react'
import type { UserRole } from '../lib/api'

export interface DashboardNavItem {
  label: string
  to: string
  icon: LucideIcon
  allowedRoles?: UserRole[]
  children?: Array<{ label: string; to: string }>
}

export const dashboardNavItems: DashboardNavItem[] = [
  { label: 'Ringkasan', to: '/dashboard', icon: LayoutDashboard, allowedRoles: ['admin', 'sales', 'finance'] },
  { label: 'Pelanggan', to: '/dashboard/pelanggan', icon: Users, allowedRoles: ['admin', 'sales'] },
  { label: 'Langganan', to: '/dashboard/langganan', icon: Repeat, allowedRoles: ['admin', 'sales'] },
  { label: 'Invoice', to: '/dashboard/invoices', icon: FileText, allowedRoles: ['admin', 'sales', 'finance'] },
  {
    label: 'Katalog',
    to: '/dashboard/katalog',
    icon: LayoutGrid,
    allowedRoles: ['admin', 'sales'],
    children: [
      { label: 'Layanan', to: '/dashboard/katalog#layanan' },
      { label: 'Produk', to: '/dashboard/katalog#produk' },
      { label: 'Plan', to: '/dashboard/katalog#plan' },
      { label: 'Add-on', to: '/dashboard/katalog#addon' },
    ],
  },
  { label: 'Kupon', to: '/dashboard/kupon', icon: Ticket, allowedRoles: ['admin', 'sales'] },
  { label: 'Pengaturan Bisnis', to: '/dashboard/pengaturan', icon: Settings, allowedRoles: ['admin'] },
]
