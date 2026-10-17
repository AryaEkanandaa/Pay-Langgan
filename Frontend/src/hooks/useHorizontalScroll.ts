import { useCallback, useEffect, useRef, useState } from 'react'

export default function useHorizontalScroll(itemCount: number) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const [position, setPosition] = useState({ left: false, right: false, progress: 0 })
  const update = useCallback(() => {
    const node = scrollRef.current
    if (!node) return
    const maximum = Math.max(0, node.scrollWidth - node.clientWidth)
    setPosition({ left: node.scrollLeft > 1, right: node.scrollLeft < maximum - 1, progress: maximum > 0 ? node.scrollLeft / maximum : 0 })
  }, [])
  useEffect(() => {
    const node = scrollRef.current
    if (!node) return
    node.addEventListener('scroll', update, { passive: true })
    const observer = new ResizeObserver(update)
    observer.observe(node)
    update()
    return () => { node.removeEventListener('scroll', update); observer.disconnect() }
  }, [itemCount, update])
  const scrollByPage = useCallback((direction: -1 | 1) => {
    const node = scrollRef.current
    if (!node) return
    node.scrollBy({ left: direction * node.clientWidth, behavior: window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth' })
  }, [])
  return { scrollRef, ...position, scrollByPage }
}
