import Badge from '../ui/Badge'
import GlassPanel from '../ui/GlassPanel'
import { features } from '../../data/features'
import useLandingReveal from '../../hooks/useLandingReveal'

export default function Features() {
  const sectionRef = useLandingReveal<HTMLElement>()

  return (
    <section ref={sectionRef} id="fitur" className="landing-section flex items-center py-20 md:py-28">
      <div className="mx-auto max-w-6xl px-5">
        <div className="landing-reveal max-w-xl">
          <Badge>Fitur</Badge>
          <h2 className="mt-4 font-display text-[32px] font-bold leading-tight tracking-[-0.02em] text-ink md:text-[36px]">
            Semua yang dibutuhkan untuk mengelola langganan.
          </h2>
          <p className="mt-4 text-[16px] leading-relaxed text-body">
            Dari katalog produk hingga siklus hidup langganan, setiap bagian
            saling terhubung dalam satu platform multi-tenant.
          </p>
        </div>

        <div className="mt-12 grid gap-5 md:grid-cols-12">
          {features.map((feature, index) => (
            <GlassPanel
              key={feature.title}
              className={`landing-reveal ${feature.span}`}
              style={{ transitionDelay: `${80 + index * 65}ms` }}
            >
              <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-lg bg-brand/[0.12]">
                <span className="h-2.5 w-2.5 rounded-sm bg-brand" />
              </div>
              <h3 className="font-display text-[18px] font-semibold text-ink">
                {feature.title}
              </h3>
              <p className="mt-2 text-[14px] leading-relaxed text-body">
                {feature.description}
              </p>
            </GlassPanel>
          ))}
        </div>
      </div>
    </section>
  )
}
