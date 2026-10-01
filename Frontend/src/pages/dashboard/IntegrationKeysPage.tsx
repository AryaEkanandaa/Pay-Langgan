import { useEffect, useState, type FormEvent } from 'react'
import { Check, Copy, KeyRound, Plus, RotateCw, ShieldCheck, Trash2 } from 'lucide-react'
import DashboardPageHeader from '../../components/dashboard/DashboardPageHeader'
import GlassPanel from '../../components/ui/GlassPanel'
import Input from '../../components/ui/Input'
import Modal from '../../components/ui/Modal'
import SubmitButton from '../../components/ui/SubmitButton'
import { useAuth } from '../../context/AuthContext'
import EmbedSitesPanel from './EmbedSitesPanel'
import {
  ApiError,
  createIntegrationKey,
  listIntegrationKeys,
  revokeIntegrationKey,
  type CreatedIntegrationKey,
  type IntegrationKey,
} from '../../lib/api'

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime())
    ? 'Tanggal tidak tersedia'
    : date.toLocaleString('id-ID', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

export default function IntegrationKeysPage() {
  const { auth } = useAuth()
  const token = auth?.token
  const [keys, setKeys] = useState<IntegrationKey[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [creating, setCreating] = useState(false)
  const [revokingId, setRevokingId] = useState<number | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [keyName, setKeyName] = useState('')
  const [newKey, setNewKey] = useState<CreatedIntegrationKey | null>(null)
  const [copied, setCopied] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const [mode, setMode] = useState<'dropin' | 'api'>('dropin')

  useEffect(() => {
    if (!token) return
    let active = true
    listIntegrationKeys(token)
      .then((result) => { if (active) setKeys(result) })
      .catch((err) => { if (active) setError(err instanceof ApiError ? err.message : 'Gagal memuat kunci integrasi.') })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [token])

  async function refresh() {
    if (!token) return
    setRefreshing(true)
    setError(null)
    try {
      setKeys(await listIntegrationKeys(token))
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal memperbarui daftar kunci.')
    } finally {
      setRefreshing(false)
    }
  }

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!token || creating) return
    const name = keyName.trim()
    if (!name) {
      setError('Nama kunci wajib diisi.')
      return
    }
    setCreating(true)
    setError(null)
    setNotice(null)
    try {
      const created = await createIntegrationKey(token, name)
      const keyInfo: IntegrationKey = {
        id: created.id,
        name: created.name,
        created_at: created.created_at,
        revoked_at: created.revoked_at,
      }
      setKeys((current) => [keyInfo, ...current])
      setNewKey(created)
      setCopied(false)
      setKeyName('')
      setCreateOpen(false)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal membuat kunci integrasi.')
    } finally {
      setCreating(false)
    }
  }

  async function copySecret() {
    if (!newKey) return
    try {
      await navigator.clipboard.writeText(newKey.secret_key)
      setCopied(true)
      setError(null)
    } catch {
      setError('Gagal menyalin otomatis. Pilih dan salin kunci secara manual.')
    }
  }

  async function revoke(key: IntegrationKey) {
    if (!token || revokingId !== null) return
    if (!window.confirm(`Cabut kunci "${key.name}"? Website yang memakai kunci ini tidak bisa lagi membuat atau membaca pesanan.`)) return
    setRevokingId(key.id)
    setError(null)
    setNotice(null)
    try {
      await revokeIntegrationKey(token, key.id)
      if (newKey?.id === key.id) setNewKey(null)
      setKeys((current) => current.map((item) => item.id === key.id
        ? { ...item, revoked_at: new Date().toISOString() }
        : item))
      setNotice(`Kunci "${key.name}" berhasil dicabut.`)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Gagal mencabut kunci integrasi.')
    } finally {
      setRevokingId(null)
    }
  }

  const activeKeys = keys.filter((key) => !key.revoked_at)

  return (
    <div className="max-w-6xl">
      <DashboardPageHeader
        title="Integrasi Website"
        description="Pilih skrip siap pakai atau integrasi API melalui server website Anda."
      />

      <div className="mb-6 overflow-hidden rounded-2xl border border-white/80 bg-white/75 shadow-sm">
        <div className="border-b border-slate-200 px-5 pt-5 sm:px-7">
          <h2 className="font-display text-[19px] font-bold text-ink">Hubungkan PayLanggan ke website</h2>
          <p className="mt-1 text-[13px] text-body">Pilih cara integrasi sesuai kebutuhan website Anda.</p>
          <div role="tablist" aria-label="Cara integrasi" className="mt-5 flex gap-6">
            <button type="button" role="tab" id="dropin-tab" aria-controls="integration-panel" aria-selected={mode === 'dropin'} onClick={() => setMode('dropin')}
              className={`border-b-[3px] pb-3 text-[13px] font-semibold transition-colors ${mode === 'dropin' ? 'border-brand text-ink' : 'border-transparent text-muted hover:text-ink'}`}>
              Via skrip drop-in
            </button>
            <button type="button" role="tab" id="api-tab" aria-controls="integration-panel" aria-selected={mode === 'api'} onClick={() => setMode('api')}
              className={`border-b-[3px] pb-3 text-[13px] font-semibold transition-colors ${mode === 'api' ? 'border-brand text-ink' : 'border-transparent text-muted hover:text-ink'}`}>
              Via API
            </button>
          </div>
        </div>
        <div id="integration-panel" role="tabpanel" aria-labelledby={mode === 'dropin' ? 'dropin-tab' : 'api-tab'} className="p-5 sm:p-7">
          {mode === 'dropin' ? (
            <EmbedSitesPanel />
          ) : (
            <div className="grid gap-4 md:grid-cols-[minmax(0,1fr)_minmax(0,2fr)]">
              <div>
                <h3 className="font-display text-[16px] font-bold text-ink">Integrasi melalui server</h3>
                <p className="mt-2 text-[13px] leading-relaxed text-body">Buat pesanan dan periksa status pembayaran dari server website Anda. Simpan secret key hanya di server.</p>
              </div>
              <div className="rounded-xl border border-slate-200 bg-slate-50 p-4 text-[13px] text-body">
                <p className="font-semibold text-ink">Langkah singkat</p>
                <ol className="mt-2 list-inside list-decimal space-y-1">
                  <li>Buat kunci integrasi di bawah.</li>
                  <li>Simpan kunci pada server website bisnis.</li>
                  <li>Gunakan API PayLanggan untuk membuat pesanan dan membaca statusnya.</li>
                </ol>
                <p className="mt-3 text-[12px] text-muted">Gunakan endpoint API integrasi yang didokumentasikan bersama proyek.</p>
              </div>
            </div>
          )}
        </div>
      </div>

      {mode === 'api' && <>

      <div className="mb-6 grid gap-4 sm:grid-cols-2">
        <GlassPanel className="flex items-start gap-4">
          <span className="rounded-xl bg-brand/10 p-3 text-brand"><KeyRound size={22} /></span>
          <div>
            <p className="text-[13px] text-muted">Kunci aktif</p>
            <p className="mt-1 font-display text-[28px] font-bold text-ink">{activeKeys.length}</p>
            <p className="text-[12px] text-body">Kunci berlaku sampai Anda mencabutnya.</p>
          </div>
        </GlassPanel>
        <GlassPanel className="flex items-start gap-4">
          <span className="rounded-xl bg-emerald-100 p-3 text-emerald-700"><ShieldCheck size={22} /></span>
          <div>
            <p className="text-[14px] font-semibold text-ink">Simpan kunci di server bisnis</p>
            <p className="mt-1 text-[13px] leading-relaxed text-body">Kunci ini digunakan oleh server website untuk membuat dan memeriksa pesanan. Jangan masukkan ke kode browser.</p>
          </div>
        </GlassPanel>
      </div>

      {error && <p role="alert" className="mb-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-[13px] text-red-700">{error}</p>}
      {notice && <p role="status" className="mb-4 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-[13px] text-emerald-800">{notice}</p>}

      {newKey && (
        <section aria-labelledby="new-key-title" className="mb-6 rounded-2xl border border-amber-300 bg-amber-50 p-5 shadow-sm sm:p-6">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <h2 id="new-key-title" className="font-display text-[17px] font-bold text-ink">Kunci baru: {newKey.name}</h2>
              <p className="mt-1 text-[13px] leading-relaxed text-amber-900">Salin dan simpan sekarang. Kunci rahasia ini tidak akan ditampilkan lagi setelah panel ditutup atau halaman dimuat ulang.</p>
            </div>
            <span className="rounded-full bg-amber-100 px-3 py-1 text-[11px] font-semibold text-amber-900">Ditampilkan sekali</span>
          </div>
          <label htmlFor="integration-secret" className="mt-5 block text-[12px] font-semibold text-ink">Secret key</label>
          <div className="mt-2 flex flex-col gap-2 sm:flex-row">
            <input id="integration-secret" readOnly value={newKey.secret_key} onFocus={(event) => event.currentTarget.select()}
              className="min-w-0 flex-1 rounded-xl border border-amber-300 bg-white px-4 py-2.5 font-mono text-[13px] text-ink" />
            <SubmitButton type="button" onClick={copySecret} className="justify-center gap-2">
              {copied ? <Check size={16} /> : <Copy size={16} />}
              {copied ? 'Tersalin' : 'Salin kunci'}
            </SubmitButton>
          </div>
          <div className="mt-4 flex flex-wrap items-center justify-between gap-3">
            <p className="text-[12px] text-amber-900">Simpan sebagai <code className="font-semibold">PAYLANGGAN_SECRET_KEY</code> di server website bisnis.</p>
            <button type="button" onClick={() => { setNewKey(null); setCopied(false) }}
              className="rounded-lg border border-amber-500 px-3 py-2 text-[12px] font-semibold text-amber-950 hover:bg-amber-100">
              Saya sudah menyimpan kunci
            </button>
          </div>
        </section>
      )}

      <GlassPanel>
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h2 className="font-display text-[18px] font-bold text-ink">Kunci integrasi</h2>
            <p className="mt-1 text-[13px] text-body">Gunakan nama yang menunjukkan website atau sistem pemakainya.</p>
          </div>
          <div className="flex gap-2">
            <button type="button" onClick={refresh} disabled={refreshing || loading}
              className="flex h-10 items-center gap-2 rounded-xl border border-white/70 bg-white/50 px-3 text-[13px] font-semibold text-body hover:bg-white disabled:opacity-50">
              <RotateCw size={16} className={refreshing ? 'animate-spin' : ''} /> Perbarui
            </button>
            <SubmitButton type="button" disabled={Boolean(newKey)} onClick={() => { setError(null); setCreateOpen(true) }}
              className="gap-2 px-4 py-2 text-[13px]"><Plus size={16} /> Buat kunci</SubmitButton>
          </div>
        </div>

        {loading ? (
          <p role="status" className="mt-8 text-[13px] text-muted">Memuat kunci integrasi...</p>
        ) : keys.length === 0 ? (
          <div className="mt-7 rounded-xl border border-dashed border-white/70 bg-white/40 px-5 py-10 text-center">
            <KeyRound className="mx-auto text-muted" size={28} />
            <p className="mt-3 font-semibold text-ink">Belum ada kunci integrasi</p>
            <p className="mt-1 text-[13px] text-body">Buat kunci pertama untuk menghubungkan server website bisnis.</p>
          </div>
        ) : (
          <div className="mt-6 space-y-3">
            {keys.map((key) => (
              <div key={key.id} className="flex flex-col gap-4 rounded-xl border border-white/60 bg-white/40 p-4 sm:flex-row sm:items-center sm:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <p className="font-semibold text-ink">{key.name}</p>
                    <span className={`rounded-full px-2.5 py-0.5 text-[11px] font-semibold ${key.revoked_at ? 'bg-slate-200 text-slate-600' : 'bg-emerald-100 text-emerald-800'}`}>
                      {key.revoked_at ? 'Dicabut' : 'Aktif'}
                    </span>
                  </div>
                  <p className="mt-1 text-[12px] text-muted">Dibuat {formatDate(key.created_at)}{key.revoked_at ? ` · Dicabut ${formatDate(key.revoked_at)}` : ''}</p>
                </div>
                {!key.revoked_at && (
                  <button type="button" disabled={revokingId !== null} onClick={() => revoke(key)}
                    className="inline-flex items-center justify-center gap-2 rounded-lg border border-red-200 px-3 py-2 text-[12px] font-semibold text-red-700 hover:bg-red-50 disabled:opacity-50">
                    <Trash2 size={14} /> {revokingId === key.id ? 'Mencabut...' : 'Cabut kunci'}
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </GlassPanel>

      </>}

      <Modal open={createOpen} onClose={() => { if (!creating) setCreateOpen(false) }} title="Buat kunci integrasi">
        <form onSubmit={create} className="space-y-5">
          <p className="text-[13px] leading-relaxed text-body">Beri nama sesuai website atau aplikasi yang akan menggunakan kunci ini. Secret key akan ditampilkan satu kali setelah dibuat.</p>
          <Input id="integration-name" label="Nama kunci" value={keyName} onChange={(event) => setKeyName(event.target.value)}
            placeholder="Website Kelana Workspace" maxLength={100} required autoFocus />
          {error && <p role="alert" className="text-[13px] text-red-700">{error}</p>}
          <div className="flex justify-end gap-2">
            <button type="button" disabled={creating} onClick={() => setCreateOpen(false)}
              className="rounded-xl border border-white/70 px-4 py-2 text-[13px] font-semibold text-body hover:bg-white/50">Batal</button>
            <SubmitButton type="submit" disabled={creating} className="px-4 py-2 text-[13px]">
              {creating ? 'Membuat...' : 'Buat kunci'}
            </SubmitButton>
          </div>
        </form>
      </Modal>
    </div>
  )
}
