import { useEffect, useRef, useState } from 'react'

export default function NavBadge({ count }: { count: number }) {
  const previous = useRef(count)
  const [revision, setRevision] = useState(0)
  useEffect(() => {
    if (count > previous.current) setRevision(value => value + 1)
    previous.current = count
  }, [count])
  if (count < 1) return null
  return <span key={revision} aria-hidden="true" className="nav-new-badge">{count > 99 ? '99+' : count}</span>
}
