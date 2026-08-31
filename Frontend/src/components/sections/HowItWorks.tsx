import Badge from '../ui/Badge'
import { steps } from '../../data/howItWorks'
import useLandingReveal from '../../hooks/useLandingReveal'

export default function HowItWorks() {
  const sectionRef = useLandingReveal<HTMLElement>()

  return (
    <section ref={sectionRef} id="cara-kerja" className="landing-section flex items-center py-20 md:py-28">
      <div className="mx-auto max-w-6xl px-5">
        <div className="landing-reveal mx-auto max-w-xl text-center">
          <Badge>Cara Kerja</Badge>
          <h2 className="mt-4 font-display text-[32px] font-bold leading-tight tracking-[-0.02em] text-ink md:text-[36px]">
            Tiga langkah untuk mulai mengelola langganan.
          </h2>
          <p className="mt-4 text-[16px] leading-relaxed text-body">
            Dari katalog produk sampai pelanggan berlangganan, alurnya
            sederhana.
          </p>
        </div>

        <div className="mt-14 grid gap-8 md:grid-cols-3">
          {steps.map((step, index) => (
            <div
              key={step.number}
              className="landing-reveal relative"
              style={{ transitionDelay: `${100 + index * 90}ms` }}
            >
              <span className="font-display text-[40px] font-bold text-brand/20">
                {step.number}
              </span>
              <h3 className="mt-2 font-display text-[18px] font-semibold text-ink">
                {step.title}
              </h3>
              <p className="mt-2 text-[14px] leading-relaxed text-body">
                {step.description}
              </p>
              {index < steps.length - 1 && (
                <span className="absolute right-[-16px] top-3 hidden text-muted md:block">
                  →
                </span>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
