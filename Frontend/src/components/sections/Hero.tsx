import { Link } from 'react-router-dom'
import Badge from '../ui/Badge'
import Button from '../ui/Button'
import { buttonBaseClass, buttonVariantClass } from '../ui/buttonStyles'
import useLandingReveal from '../../hooks/useLandingReveal'
import HeroDashboard from './HeroDashboard'

export default function Hero() {
  const sectionRef = useLandingReveal<HTMLElement>()

  return (
    <section ref={sectionRef} className="landing-section relative flex items-center overflow-hidden">
      <div className="mx-auto grid w-full max-w-6xl grid-cols-1 gap-16 px-5 pb-20 pt-16 md:grid-cols-2 md:items-center md:pb-32 md:pt-24">
        <div className="landing-reveal min-w-0">
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

        <div className="landing-reveal min-w-0" style={{ transitionDelay: '120ms' }}>
          <HeroDashboard />
        </div>
      </div>
    </section>
  )
}
