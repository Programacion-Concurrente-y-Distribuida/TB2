import { useCallback, useEffect, useRef, useState } from 'react'
import { connectWebSocket } from '../api/client.js'

/**
 * Manages a persistent WebSocket connection.
 * Re-connects automatically with exponential back-off when the connection drops.
 *
 * Returns { connected, lastMessage, resetLastMessage }
 * `lastMessage` is the latest parsed JSON object received.
 */
export function useWebSocket() {
  const [connected, setConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState(null)
  const wsRef = useRef(null)
  const retryRef = useRef(null)
  const retryDelay = useRef(1000)

  const connect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.onclose = null
      wsRef.current.close()
    }

    wsRef.current = connectWebSocket({
      onOpen: () => {
        setConnected(true)
        retryDelay.current = 1000
      },
      onClose: () => {
        setConnected(false)
        // Exponential back-off, max 30s
        const delay = Math.min(retryDelay.current, 30_000)
        retryDelay.current = delay * 2
        retryRef.current = setTimeout(connect, delay)
      },
      onError: () => {
        wsRef.current?.close()
      },
      onMessage: (data) => setLastMessage(data),
    })
  }, [])

  useEffect(() => {
    connect()
    return () => {
      clearTimeout(retryRef.current)
      if (wsRef.current) {
        wsRef.current.onclose = null
        wsRef.current.close()
      }
    }
  }, [connect])

  const resetLastMessage = useCallback(() => setLastMessage(null), [])

  return { connected, lastMessage, resetLastMessage }
}
