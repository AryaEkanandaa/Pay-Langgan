import { useRef, useState } from 'react'
import Modal from '../ui/Modal'
import SubmitButton from '../ui/SubmitButton'
import Input from '../ui/Input'
import { getDashboardSummary } from '../../lib/api'
import { exportBusinessReportPdf } from '../../lib/reportPdf'
import { addDays, addMonths, businessToday, periodDays, periodError, periodLabel, type DatePeriod } from '../../lib/dashboardPeriod'

const presets = [{ id: 'month', label: 'Bulan ini' }, { id: '1b', label: '1 bulan terakhir' }, { id: '3b', label: '3 bulan terakhir' }, { id: '6b', label: '6 bulan terakhir' }, { id: '12b', label: '12 bulan terakhir' }, { id: 'year', label: 'Tahun ini' }, { id: 'custom', label: 'Rentang kustom' }] as const
export default function SummaryExportDialog({ onClose, initialPeriod, initialPreset, timezone, token, businessID, salesOnly }: { onClose: () => void; initialPeriod: DatePeriod; initialPreset: string; timezone: string; token: string; businessID: string; salesOnly: boolean }) {
  const [preset, setPreset] = useState(initialPreset), [period, setPeriod] = useState(initialPeriod), [busy, setBusy] = useState(false), [error, setError] = useState<string | null>(null)
  const submitting = useRef(false), today = businessToday(timezone), validation = periodError(period, today)
  function choose(id: string) {
    setPreset(id); setError(null)
    if (id === 'custom') return
    setPeriod({ to: today, from: id === 'month' ? `${today.slice(0, 7)}-01` : id === 'year' ? `${today.slice(0, 4)}-01-01` : id === '1b' ? addDays(today, -29) : addDays(addMonths(today, -Number.parseInt(id)), 1) })
  }
  async function download() {
    if (validation || submitting.current) return
    submitting.current = true; setBusy(true); setError(null)
    try {
      const data = await getDashboardSummary(token, { ...period, granularity: periodDays(period) > 90 ? 'month' : 'day' })
      if (!data.range_stats) throw new Error('Statistik rentang belum tersedia.')
      await exportBusinessReportPdf(data, businessID, salesOnly, { ...period, timezone })
    } catch { setError('Gagal membuat ekspor. Coba lagi.') }
    finally { submitting.current = false; setBusy(false) }
  }
  return <Modal open onClose={() => { if (!busy) onClose() }} title="Ekspor ringkasan" trapFocus>
    <fieldset disabled={busy}><legend className="mb-3 text-sm font-semibold">Pilihan periode</legend><div className="grid grid-cols-2 gap-2">{presets.map(option => <label key={option.id} className={`flex cursor-pointer items-center gap-2 rounded-xl border p-3 text-xs font-medium ${preset === option.id ? 'border-brand bg-brand/10' : 'border-border'}`}><input type="radio" name="export-period" value={option.id} checked={preset === option.id} onChange={() => choose(option.id)} className="accent-brand" />{option.label}</label>)}</div>
      {preset === 'custom' && <div className="mt-4 grid grid-cols-2 gap-3"><Input id="report-from" label="Dari" type="date" max={today} value={period.from} onChange={event => setPeriod({ ...period, from: event.target.value })} /><Input id="report-to" label="Sampai" type="date" max={today} value={period.to} onChange={event => setPeriod({ ...period, to: event.target.value })} /></div>}
    </fieldset>
    <div aria-live="polite" className="mt-4 text-sm">{validation ? <p className="text-red-600">{validation}</p> : <p className="text-body">Periode: {periodLabel(period)} ({periodDays(period)} hari)<br />Zona waktu: {timezone}</p>}{error && <p role="alert" className="mt-3 text-red-600">{error}</p>}</div>
    <SubmitButton type="button" onClick={() => void download()} disabled={busy || Boolean(validation)} className="mt-6 w-full">{busy ? 'Menyiapkan...' : 'Unduh PDF'}</SubmitButton>
  </Modal>
}
