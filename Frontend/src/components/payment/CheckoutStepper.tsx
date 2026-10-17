import { Check } from 'lucide-react'

export default function CheckoutStepper({ currentStep, skipCard = false }: { currentStep: number; skipCard?: boolean }) {
  return (
    <div className="relative mx-2">
    <div aria-hidden="true" className="checkout-step-connectors">
      {[1, 2, 3].map(number => <span key={number} className="checkout-step-line"><span className={currentStep > number ? 'is-complete' : ''} /></span>)}
    </div>
    <ol aria-label="Langkah checkout" className="checkout-stepper relative grid grid-cols-4 py-3">
      {['Paket', 'Pelanggan', 'Kartu', 'Selesai'].map((label, index) => {
        const number = index + 1
        const skipped = skipCard && number === 3
        const done = currentStep > number && !skipped
        const active = currentStep === number
        return (
          <li key={label} aria-current={active ? 'step' : undefined} className="relative text-center">
            <span className={`checkout-step-circle relative z-10 mx-auto flex size-8 items-center justify-center rounded-full text-xs font-bold ${done || active ? 'bg-brand text-white' : 'border border-slate-400 bg-white text-slate-600 dark:bg-slate-900 dark:text-slate-300'} ${active ? 'checkout-step-active' : ''}`}>
              {done ? <Check size={16} aria-hidden="true" /> : number}
            </span>
            <span className={`mt-2 block text-xs ${active ? 'font-bold text-brand dark:text-blue-300' : done ? 'font-semibold text-ink' : 'font-medium text-slate-600 dark:text-slate-300'}`}>{skipped ? 'Tanpa kartu' : label}</span>
            <span className="sr-only">{skipped ? 'Dilewati, tidak diperlukan pembayaran kartu sekarang' : done ? 'Selesai' : active ? 'Sedang berlangsung' : 'Belum selesai'}</span>
          </li>
        )
      })}
    </ol>
    </div>
  )
}
