import { useEffect, useState, type ChangeEvent, type FormEvent } from 'react'
import { CheckCircle2, Copy, CreditCard, ShieldCheck } from 'lucide-react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import GlassPanel from '../../components/ui/GlassPanel'
import Input from '../../components/ui/Input'
import SubmitButton from '../../components/ui/SubmitButton'
import { useAuth } from '../../context/AuthContext'
import {
  ApiError,
  getBusinessPaymentGateway,
  saveBusinessPaymentGateway,
  type BusinessPaymentGatewayDTO,
  type BusinessPaymentGatewayInput,
} from '../../lib/api'

type Provider = BusinessPaymentGatewayInput['provider']

export default function PaymentGatewayPage() {
  const { auth } = useAuth()
  const token = auth?.token
  const [provider, setProvider] = useState<Provider>('xendit')
  const [environment, setEnvironment] = useState<'sandbox' | 'production'>('sandbox')
  const [values, setValues] = useState({ xenditSecret: '', xenditCallback: '', midtransMID: '', midtransClient: '', midtransServer: '' })
  const [saved, setSaved] = useState<BusinessPaymentGatewayDTO | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

  useEffect(() => {
    if (!token) return
    getBusinessPaymentGateway(token)
      .then((gateway) => {
        setSaved(gateway)
        if (gateway) {
          setProvider(gateway.provider)
          setEnvironment(gateway.environment)
        }
      })
      .catch((err) => setError(err instanceof ApiError ? err.message : 'Gagal memuat konfigurasi gateway.'))
      .finally(() => setLoading(false))
  }, [token])

  const set = (key: keyof typeof values) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (!token) return
    setError(null)
    setNotice(null)
    setSaving(true)
    const payload: BusinessPaymentGatewayInput = { provider, environment }
    if (provider === 'xendit') {
      payload.xendit_secret_key = values.xenditSecret
      payload.xendit_callback_token = values.xenditCallback
    } else {
      payload.midtrans_merchant_id = values.midtransMID
      payload.midtrans_client_key = values.midtransClient
      payload.midtrans_server_key = values.midtransServer
    }
    try {
      setSaved(await saveBusinessPaymentGateway(token, payload))
      setValues({ xenditSecret: '', xenditCallback: '', midtransMID: '', midtransClient: '', midtransServer: '' })
      setNotice('Kredensial tersimpan dengan aman. Nilai rahasia tidak akan ditampilkan kembali.')
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal menyimpan konfigurasi gateway.')
    } finally {
      setSaving(false)
    }
  }

  async function copyWebhook() {
    if (!saved) return
    await navigator.clipboard.writeText(saved.webhook_url)
    setNotice('URL webhook disalin.')
  }

  if (loading) return <div className="p-8 text-body">Memuat pengaturan pembayaran…</div>

  return (
    <div className="space-y-6 p-5 md:p-8">
      <DashboardPageHeader title="Payment Gateway" description="Hubungkan satu akun payment gateway untuk menerima pembayaran pelanggan bisnis Anda." />

      <GlassPanel className="max-w-4xl p-6">
        <div className="mb-6 flex items-start gap-3">
          <span className="rounded-xl bg-brand/10 p-3 text-brand"><CreditCard size={20} /></span>
          <div>
            <h2 className="font-semibold text-ink">Pilih satu payment gateway</h2>
            <p className="mt-1 text-sm text-muted">Kredensial milik bisnis Anda dipakai untuk transaksi pelanggan. Data tagihan PayLanggan tetap terpisah.</p>
          </div>
        </div>

        <div className="mb-6 grid gap-3 sm:grid-cols-2">
          {(['xendit', 'midtrans'] as const).map((item) => (
            <button key={item} type="button" onClick={() => setProvider(item)}
              className={`rounded-xl border p-4 text-left transition ${provider === item ? 'border-brand bg-brand/5 ring-2 ring-brand/20' : 'border-line bg-white/60 hover:border-brand/40'}`}>
              <span className="font-semibold capitalize text-ink">{item === 'xendit' ? 'Xendit' : 'Midtrans'}</span>
              <span className="mt-1 block text-xs text-muted">{item === 'xendit' ? 'Pembayaran tertanam dengan Xendit Components' : 'Pembayaran melalui Midtrans Snap'}</span>
            </button>
          ))}
        </div>

        <form className="space-y-4" onSubmit={submit}>
          <label className="block text-sm font-medium text-ink" htmlFor="gateway-environment">Mode akun</label>
          <select id="gateway-environment" className="w-full rounded-xl border border-line bg-white px-4 py-3 text-sm text-ink focus:border-brand focus:outline-none"
            value={environment} onChange={(event) => setEnvironment(event.target.value as 'sandbox' | 'production')}>
            <option value="sandbox">Sandbox (pengujian)</option>
            <option value="production">Production (transaksi nyata)</option>
          </select>

          {provider === 'xendit' ? <>
            <Input id="xendit-secret" label="Secret API Key" type="password" autoComplete="new-password" value={values.xenditSecret} onChange={set('xenditSecret')} required />
            <Input id="xendit-callback" label="Webhook Verification Token" type="password" autoComplete="new-password" value={values.xenditCallback} onChange={set('xenditCallback')} required />
            <p className="text-xs text-muted">Masukkan Secret API Key dan Callback Verification Token dari mode akun yang dipilih. Daftarkan URL webhook di bawah pada dashboard Xendit.</p>
          </> : <>
            <Input id="midtrans-mid" label="Merchant ID" value={values.midtransMID} onChange={set('midtransMID')} required />
            <Input id="midtrans-client" label="Client Key" value={values.midtransClient} onChange={set('midtransClient')} required />
            <Input id="midtrans-server" label="Server Key" type="password" autoComplete="new-password" value={values.midtransServer} onChange={set('midtransServer')} required />
            <p className="text-xs text-muted">Ambil ketiga nilai dari Access Keys pada mode akun yang dipilih. Server Key hanya dikirim ke backend.</p>
          </>}

          {saved && <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-sm text-emerald-900">
            <div className="flex items-center gap-2 font-medium"><CheckCircle2 size={17} /> {saved.provider === 'xendit' ? 'Xendit' : 'Midtrans'} tersambung · {saved.environment === 'sandbox' ? 'Sandbox' : 'Production'}</div>
            {saved.account_label && <p className="mt-1">Merchant ID: {saved.account_label}</p>}
            <div className="mt-3 flex flex-wrap items-center gap-2">
              <code className="max-w-full break-all rounded bg-white px-2 py-1 text-xs">{saved.webhook_url}</code>
              <button type="button" onClick={copyWebhook} className="inline-flex items-center gap-1 rounded-lg border border-emerald-300 px-2 py-1 text-xs"><Copy size={13} /> Salin URL webhook</button>
            </div>
            <p className="mt-2 text-xs">{saved.provider === 'midtrans' ? 'Atur URL ini sebagai Payment Notification URL di akun Midtrans Anda.' : 'Daftarkan URL ini sebagai webhook pembayaran di akun Xendit Anda.'}</p>
          </div>}

          <div className="flex gap-2 rounded-xl bg-slate-50 p-4 text-xs text-muted">
            <ShieldCheck size={17} className="shrink-0 text-emerald-600" />
            Secret disimpan terenkripsi di server dan tidak pernah ditampilkan kembali. Simpan kunci enkripsi backend dengan aman.
          </div>
          {error && <p role="alert" className="text-sm text-red-600">{error}</p>}
          {notice && <p role="status" className="text-sm text-emerald-700">{notice}</p>}
          <SubmitButton type="submit" disabled={saving}>{saving ? 'Menyimpan…' : 'Simpan konfigurasi'}</SubmitButton>
        </form>
      </GlassPanel>
    </div>
  )
}
