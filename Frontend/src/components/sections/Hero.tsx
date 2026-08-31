import { Link } from 'react-router-dom'
import Badge from '../ui/Badge'
import Button from '../ui/Button'
import GlassPanel from '../ui/GlassPanel'
import { buttonBaseClass, buttonVariantClass } from '../ui/buttonStyles'
import { statusRows, invoiceExample } from '../../data/subscriptionPreview'
import useLandingReveal from '../../hooks/useLandingReveal'

export default function Hero() {
  const sectionRef = useLandingReveal<HTMLElement>()

  return (
    <section ref={sectionRef} className="landing-section relative flex items-center overflow-hidden">
      <div className="mx-auto grid max-w-6xl gap-16 px-5 pb-20 pt-16 md:grid-cols-2 md:items-center md:pb-32 md:pt-24">
        <div className="landing-reveal">
          <Badge>Manajemen Langganan SaaS</Badge>

          <h1 className="mt-5 font-display text-[40px] font-bold leading-[1.08] tracking-[-0.03em] text-ink md:text-[48px]">
            Kelola pelanggan, paket langganan, dan penagihan dari satu tempat.
          </h1>

          <p className="mt-6 max-w-md text-[18px] leading-relaxed text-body">
            PayLanggan menyatukan katalog produk, data pelanggan, dan siklus
            hidup langganan dari trial, aktif, jeda, hingga batal dalam
            satu platform multi-tenant untuk tim yang menjual layanan
            berbasis langganan.
          </p>

          <div className="mt-8 flex flex-wrap items-center gap-4">
            <Link to="/register" className={`${buttonBaseClass} ${buttonVariantClass.primary}`}>
              Daftar Sekarang
            </Link>
            <Button href="#fitur" variant="secondary">
              Lihat Fitur
            </Button>
          </div>
        </div>

        <div className="landing-reveal relative" style={{ transitionDelay: '120ms' }}>
          <GlassPanel className="mx-auto max-w-sm space-y-5">
            <div className="flex items-center justify-between">
              <p className="text-[13px] font-semibold text-muted">Status Langganan</p>
              <Badge>Pratinjau</Badge>
            </div>

            <div className="space-y-3">
              {statusRows.map((row) => (
                <div key={row.label} className="space-y-1.5">
                  <div className="flex items-center justify-between text-[13px]">
                    <span className="text-body">{row.label}</span>
                    <span className="font-semibold text-ink">{row.value}%</span>
                  </div>
                  <div className="h-2 w-full rounded-full bg-white/50">
                    <div
                      className={`h-2 rounded-full ${row.className}`}
                      style={{ width: `${row.value}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>

            <div className="rounded-xl border border-white/50 bg-white/30 p-4">
              <p className="text-[11px] font-medium uppercase tracking-wide text-muted">
                Contoh Invoice
              </p>
              <div className="mt-3 flex items-center justify-between text-[13px]">
                <span className="text-body">{invoiceExample.plan}</span>
                <span className="font-semibold text-ink">{invoiceExample.price}</span>
              </div>
              <div className="mt-2 flex items-center justify-between text-[13px]">
                <span className="text-body">{invoiceExample.coupon}</span>
                <span className="font-semibold text-success">{invoiceExample.discount}</span>
              </div>
            </div>
          </GlassPanel>
        </div>
      </div>
    </section>
  )
}
