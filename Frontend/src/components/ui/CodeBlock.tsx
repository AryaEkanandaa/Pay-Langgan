import CopyButton from './CopyButton'

interface CodeBlockProps {
  code: string
  language?: string
  label?: string
}

export default function CodeBlock({ code, language = 'HTML', label = 'Kode' }: CodeBlockProps) {
  return <div className="min-w-0 overflow-hidden rounded-xl border border-slate-700 bg-slate-900 text-slate-100">
    <div className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-700 px-4 py-2">
      <span className="text-[12px] font-medium text-slate-300">{language.toUpperCase()}</span>
      <CopyButton text={code} label={`Salin ${label}`} />
    </div>
    <pre aria-label={label} tabIndex={0} className="max-w-full overflow-x-auto whitespace-pre p-4 font-mono text-[12px] leading-6 focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-brand"><code>{code}</code></pre>
  </div>
}
