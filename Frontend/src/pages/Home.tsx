import Navbar from '../components/layout/Navbar'
import Footer from '../components/layout/Footer'
import Hero from '../components/sections/Hero'
import Features from '../components/sections/Features'
import HowItWorks from '../components/sections/HowItWorks'
import CTA from '../components/sections/CTA'
import useLandingReveal from '../hooks/useLandingReveal'

export default function Home() {
  const finalSectionRef = useLandingReveal<HTMLElement>()

  return (
    <div className="landing-scroll min-h-screen">
      <Navbar />
      <main>
        <Hero />
        <Features />
        <HowItWorks />
        <section ref={finalSectionRef} id="tentang" className="landing-section flex flex-col">
          <div className="flex flex-1 items-center">
            <CTA />
          </div>
          <Footer />
        </section>
      </main>
    </div>
  )
}
