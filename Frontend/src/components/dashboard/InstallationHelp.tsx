import { useState, type FormEvent } from 'react'
import CodeBlock from '../ui/CodeBlock'
import GlassPanel from '../ui/GlassPanel'
import Input from '../ui/Input'
import SubmitButton from '../ui/SubmitButton'

interface InstallationHelpProps {
  headerCode: string
  checkoutCode: string
}

export default function InstallationHelp({ headerCode, checkoutCode }: InstallationHelpProps) {
  const [url, setUrl] = useState('')
  const [notice, setNotice] = useState<string | null>(null)
  function verify(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    // TODO: Connect to an installation verification API when one is available.
    // A valid URL alone is not evidence that the script has been installed.
    setNotice('Verifikasi otomatis belum tersedia. Buka halaman penjualan Anda dan coba tombol checkout untuk memeriksa pemasangan.')
  }
  const example = `<!doctype html>
<html lang="id">
<head>
  <!-- Skrip header dipasang di sini -->
${headerCode.replace(/\n\s*/g, ' ') || '  <!-- Salin skrip header dari Langkah 1 -->'}
</head>
<body>
  <!-- Tombol dipasang di halaman penjualan -->
${checkoutCode.replace(/\n\s*/g, ' ') || '  <!-- Salin tombol dari Langkah 2 -->'}
</body>
</html>`
  return <>
    <GlassPanel>
      <div className="flex flex-wrap items-center justify-between gap-3"><h2 className="text-base font-semibold text-ink">Status instalasi</h2><span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-700 dark:bg-slate-800 dark:text-slate-200">Belum terdeteksi</span></div>
      <p className="mt-3 text-sm leading-6 text-body">Status belum diperiksa. Verifikasi otomatis belum tersedia.</p>
      <form onSubmit={verify} className="mt-4 space-y-4">
        <Input id="installation-url" label="URL halaman penjualan" type="url" required placeholder="https://website-anda.com/langganan" value={url} onChange={event => { setUrl(event.target.value); setNotice(null) }} />
        <SubmitButton type="submit" disabled={!url.trim()} className="text-sm">Verifikasi</SubmitButton>
      </form>
      {notice && <p role="status" aria-live="polite" className="mt-4 text-sm leading-6 text-body">{notice}</p>}
    </GlassPanel>
    <GlassPanel>
      <h2 className="text-base font-semibold text-ink">Contoh penempatan</h2>
      <p className="mb-4 mt-3 text-sm leading-6 text-body">Letakkan skrip di <code>&lt;head&gt;</code> dan tombol di <code>&lt;body&gt;</code> halaman penjualan.</p>
      <CodeBlock code={example} label="contoh penempatan" />
    </GlassPanel>
  </>
}
