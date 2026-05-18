import { useCallback, useRef, useState, type RefObject } from 'react'

import type { Presence, ServerMessage } from '../../../lib/protocol'

export function usePresence(socketRef: RefObject<WebSocket | null>) {
  const [remotePresence, setRemotePresence] = useState<Map<string, Presence>>(() => new Map())
  const presenceTimerRef = useRef<number | null>(null)

  const sendPresence = useCallback(
    (anchor: number, head: number) => {
      if (presenceTimerRef.current !== null) {
        window.clearTimeout(presenceTimerRef.current)
      }
      presenceTimerRef.current = window.setTimeout(() => {
        const socket = socketRef.current
        if (!socket || socket.readyState !== WebSocket.OPEN) {
          return
        }
        socket.send(
          JSON.stringify({
            type: 'presence',
            presence: { anchor, head },
          })
        )
        presenceTimerRef.current = null
      }, 100)
    },
    [socketRef]
  )

  const handlePresenceMessage = useCallback((msg: ServerMessage) => {
    if (msg.type === 'presenceSnapshot') {
      setRemotePresence(() => {
        const next = new Map<string, Presence>()
        Object.entries(msg.presences).forEach(([id, presence]) => {
          next.set(id, presence)
        })
        return next
      })
      return
    }

    if (msg.type === 'presenceUpdate') {
      setRemotePresence((prev) => {
        const next = new Map(prev)
        if (!msg.presence || (!msg.presence.color && !msg.presence.name)) {
          next.delete(msg.userId)
        } else {
          next.set(msg.userId, msg.presence)
        }
        return next
      })
    }
  }, [])

  return { remotePresence, sendPresence, handlePresenceMessage }
}
