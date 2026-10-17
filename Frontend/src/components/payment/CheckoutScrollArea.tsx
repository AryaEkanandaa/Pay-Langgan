import { useEffect, useRef, useState, type ReactNode } from 'react'

export default function CheckoutScrollArea({ children, step }: { children: ReactNode; step: number }) {
  const container = useRef<HTMLDivElement>(null)
  const content = useRef<HTMLDivElement>(null)
  const [hasMore, setHasMore] = useState(false)
  useEffect(() => { container.current?.scrollTo({ top: 0 }) }, [step])
  useEffect(() => {
    const node = container.current
    const inner = content.current
    if (!node || !inner) return
    const update = () => setHasMore(node.scrollHeight - node.clientHeight - node.scrollTop > 2)
    const observer = new ResizeObserver(update)
    observer.observe(node)
    observer.observe(inner)
    node.addEventListener('scroll', update, { passive: true })
    update()
    return () => { observer.disconnect(); node.removeEventListener('scroll', update) }
  }, [])
  return <div className="relative min-h-0 flex-1">
    <div ref={container} tabIndex={0} aria-label="Isi checkout" className="checkout-scroll h-full overflow-y-auto focus-visible:outline focus-visible:outline-2 focus-visible:outline-inset focus-visible:outline-brand">
      <div ref={content} className="mx-auto w-full max-w-lg px-6 pb-1 pt-2">{children}</div>
    </div>
    {hasMore && <div aria-hidden="true" className="checkout-scroll-fade pointer-events-none absolute inset-x-0 bottom-0 h-5" />}
  </div>
}
