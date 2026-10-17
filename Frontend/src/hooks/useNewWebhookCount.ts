import { createContext, useCallback, useEffect, useRef, useState } from 'react'
import { getNewWebhookCount, markWebhookLogsSeen, type WebhookUnreadDTO } from '../lib/api'

export function useNewWebhookCount(token: string | undefined, enabled: boolean) {
  const [state, setState] = useState<WebhookUnreadDTO | null>(null)
  const generation = useRef(0)
  const pending = useRef<AbortController | null>(null)

  const refresh = useCallback(async () => {
    if (!token || !enabled || document.hidden) return
    const revision = ++generation.current
    pending.current?.abort()
    const controller = new AbortController()
    pending.current = controller
    try {
      const next = await getNewWebhookCount(token, controller.signal)
      if (!controller.signal.aborted && revision === generation.current) setState(next)
    } catch { /* Keep the last known badge on temporary network failure. */ }
  }, [token, enabled])

  const markSeen = useCallback(async (through: string, signal: AbortSignal) => {
    if (!token || !enabled) return
    const revision = ++generation.current
    pending.current?.abort()
    const next = await markWebhookLogsSeen(token, through, signal)
    if (!signal.aborted && revision === generation.current) setState(next)
  }, [token, enabled])

  useEffect(() => {
    setState(null)
    void refresh()
    const timer = window.setInterval(() => void refresh(), 30_000)
    function visibility() {
      if (document.hidden) { generation.current++; pending.current?.abort() }
      else void refresh()
    }
    document.addEventListener('visibilitychange', visibility)
    function stop() {
      generation.current++
      pending.current?.abort()
      window.clearInterval(timer)
      document.removeEventListener('visibilitychange', visibility)
    }
    return stop
  }, [refresh])

  return { count: state?.count ?? 0, refresh, markSeen }
}

export const WebhookNotificationsContext = createContext<ReturnType<typeof useNewWebhookCount> | null>(null)
