import { useEffect, useRef } from 'react'

export default function useLandingReveal<T extends HTMLElement>() {
  const ref = useRef<T>(null)

  useEffect(() => {
    const element = ref.current
    const scrollRoot = document.querySelector('.landing-scroll')
    if (!element) return

    if (!scrollRoot || !('IntersectionObserver' in window)) {
      element.classList.add('is-visible')
      return
    }

    const observer = new IntersectionObserver(
      ([entry]) => element.classList.toggle('is-visible', entry.isIntersecting),
      { root: scrollRoot, threshold: 0.18 },
    )

    observer.observe(element)
    return () => observer.disconnect()
  }, [])

  return ref
}
