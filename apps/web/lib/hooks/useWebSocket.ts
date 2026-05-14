'use client'

import { useCallback, useEffect, useRef, useState } from 'react'
import { API_BASE_URL } from '@/lib/env'

export interface WebSocketMessage {
  type: 'run_update' | 'step_update' | 'message_new' | 'approval_update' | 'tool_call_update'
  data: any
}

interface UseWebSocketOptions {
  runId: string | null
  onMessage?: (msg: WebSocketMessage) => void
  enabled?: boolean
}

/**
 * Derive WebSocket base URL from the HTTP API base URL.
 * e.g. http://localhost:8080 -> ws://localhost:8080
 */
function getWsUrl(runId: string): string {
  const wsBase = API_BASE_URL.replace(/^http/, 'ws')
  return `${wsBase}/ws/agent-runs/${runId}`
}

const MAX_RECONNECT_DELAY = 30000 // 30 seconds

/**
 * A custom React hook that manages a WebSocket connection with:
 * - Auto-reconnect with exponential backoff (1s, 2s, 4s, 8s, max 30s)
 * - Clean up on unmount
 * - Manual reconnect via `reconnect()`
 * - `isConnected` state for UI indicators
 */
export function useWebSocket({ runId, onMessage, enabled = true }: UseWebSocketOptions) {
  const [isConnected, setIsConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const retryCountRef = useRef(0)
  const onMessageRef = useRef(onMessage)
  const enabledRef = useRef(enabled)
  const runIdRef = useRef(runId)

  // Keep refs in sync with latest props
  onMessageRef.current = onMessage
  enabledRef.current = enabled
  runIdRef.current = runId

  // Store connect in a ref so setTimeout callbacks always call the latest version
  const connectRef = useRef<() => void>(() => {})

  const clearReconnectTimer = useCallback(() => {
    if (reconnectTimerRef.current !== null) {
      clearTimeout(reconnectTimerRef.current)
      reconnectTimerRef.current = null
    }
  }, [])

  const scheduleReconnect = useCallback(() => {
    if (!enabledRef.current) return

    const delay = Math.min(
      1000 * Math.pow(2, retryCountRef.current),
      MAX_RECONNECT_DELAY
    )
    retryCountRef.current++
    clearReconnectTimer()
    reconnectTimerRef.current = setTimeout(() => {
      connectRef.current()
    }, delay)
  }, [clearReconnectTimer])

  const connect = useCallback(() => {
    const currentRunId = runIdRef.current
    if (!currentRunId || !enabledRef.current) {
      setIsConnected(false)
      return
    }

    // Clean up existing WebSocket without triggering reconnect
    if (wsRef.current) {
      const oldWs = wsRef.current
      oldWs.onopen = null
      oldWs.onmessage = null
      oldWs.onclose = null
      oldWs.onerror = null
      oldWs.close()
      wsRef.current = null
    }

    try {
      const url = getWsUrl(currentRunId)
      const ws = new WebSocket(url)
      wsRef.current = ws

      ws.onopen = () => {
        setIsConnected(true)
        retryCountRef.current = 0
      }

      ws.onmessage = (event: MessageEvent) => {
        try {
          const msg = JSON.parse(event.data) as WebSocketMessage
          onMessageRef.current?.(msg)
        } catch {
          // Invalid JSON from server, ignore gracefully
        }
      }

      ws.onclose = () => {
        setIsConnected(false)
        if (wsRef.current === ws) {
          wsRef.current = null
        }
        scheduleReconnect()
      }

      ws.onerror = () => {
        // onclose will fire after onerror, which handles reconnect scheduling
        ws.close()
      }
    } catch {
      // WebSocket constructor can throw (e.g., bad URL)
      scheduleReconnect()
    }
  }, [scheduleReconnect])

  // Keep connectRef up to date
  connectRef.current = connect

  useEffect(() => {
    connect()

    return () => {
      clearReconnectTimer()
      if (wsRef.current) {
        const ws = wsRef.current
        ws.onopen = null
        ws.onmessage = null
        ws.onclose = null
        ws.onerror = null
        ws.close()
        wsRef.current = null
      }
      setIsConnected(false)
    }
  }, [connect, clearReconnectTimer, runId, enabled])

  const reconnect = useCallback(() => {
    retryCountRef.current = 0
    clearReconnectTimer()
    connect()
  }, [clearReconnectTimer, connect])

  return { isConnected, reconnect }
}
