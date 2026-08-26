import { useState } from 'react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import ServicesTab from './catalog/ServicesTab'
import ProductsTab from './catalog/ProductsTab'
import PlansTab from './catalog/PlansTab'
import AddOnsTab from './catalog/AddOnsTab'

const tabs = [
  { key: 'layanan', label: 'Layanan' },
  { key: 'produk', label: 'Produk' },
  { key: 'plan', label: 'Plan' },
  { key: 'addon', label: 'Add-on' },
] as const

type TabKey = (typeof tabs)[number]['key']

export default function KatalogPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('layanan')

  return (
    <div>
      <DashboardPageHeader title="Katalog" description="Kelola layanan, produk, plan, dan add-on." />

      <div className="glass mb-6 inline-flex gap-1 rounded-full p-1">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            type="button"
            onClick={() => setActiveTab(tab.key)}
            className={`rounded-full px-4 py-1.5 text-[13px] font-semibold transition-colors ${
              activeTab === tab.key ? 'bg-brand text-white' : 'text-body hover:text-ink'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === 'layanan' && <ServicesTab />}
      {activeTab === 'produk' && <ProductsTab />}
      {activeTab === 'plan' && <PlansTab />}
      {activeTab === 'addon' && <AddOnsTab />}
    </div>
  )
}
