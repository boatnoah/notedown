import { useCallback, useEffect, useRef, type RefObject } from 'react'

import { getAccessToken } from '../../../lib/auth'
import { getWebSocketUrl } from '../../../lib/config'
import type { Operation, ServerMessage } from '../../../lib/protocol'
import { encodeClientMessage, parseServerMessage } from '../../../lib/protocol'
import { createOpDispatcher, type OpDispatcher } from '../lib/opDispatcher'

// Browsers cannot attach an Authorization header to WebSocket handshakes, so
// the access token rides along in the Sec-WebSocket-Protocol header as
// ["bearer", "<token>"]. The server validates it before upgrading and echoes
// "bearer" back. Must match backend/internal/realtime/hub.go.
function openAuthenticatedSocket(url: string): WebSocket {
  const token = getAccessToken()
  return token ? new WebSocket(url, ['bearer', token]) : new WebSocket(url)
}

type UseCollaborationSessionOptions = {
  documentId: string
  initialVersion: number
  socketRef: RefObject<WebSocket | null>
  isApplyingRemoteRef: RefObject<boolean>
  onSnapshot: (content: string, version: number) => void
  onServerMessage: (msg: ServerMessage) => void
  onConnectionLost: () => void
}

export function useCollaborationSession({
  documentId,
  initialVersion,
  socketRef,
  isApplyingRemoteRef,
  onSnapshot,
  onServerMessage,
  onConnectionLost,
}: UseCollaborationSessionOptions) {
  const latestVersionRef = useRef(initialVersion)
  const awaitingSyncRef = useRef(true)
  const forceNextSnapshotRef = useRef(false)
  const dispatcherRef = useRef<OpDispatcher | null>(null)

  const sendOperation = useCallback(
    (op: Operation) => {
      if (isApplyingRemoteRef.current) {
        return
      }
      dispatcherRef.current?.enqueue(op)
    },
    [isApplyingRemoteRef]
  )

  const handleServerMessage = useCallback(
    (msg: ServerMessage, socket: WebSocket) => {
      if (msg.type === 'snapshot') {
        if (forceNextSnapshotRef.current || msg.snapshot.version > latestVersionRef.current) {
          forceNextSnapshotRef.current = false
          latestVersionRef.current = msg.snapshot.version
          onSnapshot(msg.snapshot.content, msg.snapshot.version)
        }

        if (awaitingSyncRef.current) {
          awaitingSyncRef.current = false
          dispatcherRef.current?.setBlocked(false)
        }
        // Every snapshot settles the in-flight op (if any) and lets the next
        // queued op go out stamped with the version we just recorded.
        dispatcherRef.current?.onAck()
        return
      }

      if (msg.type === 'error') {
        console.error('Server error:', msg.error)
        // The local editor already applied a change the server refused
        // (e.g. read-only access), so the documents have diverged. Resync
        // and force-apply the result — its version won't have advanced, so
        // the version guard alone would ignore it.
        forceNextSnapshotRef.current = true
        socket.send(encodeClientMessage({ type: 'sync' }))
        dispatcherRef.current?.onAck()
        return
      }

      onServerMessage(msg)
    },
    [onSnapshot, onServerMessage]
  )

  useEffect(() => {
    latestVersionRef.current = initialVersion
  }, [initialVersion])

  useEffect(() => {
    const socket = openAuthenticatedSocket(getWebSocketUrl(documentId))
    socketRef.current = socket
    dispatcherRef.current = createOpDispatcher(
      () => latestVersionRef.current,
      (op) => {
        if (socket.readyState === WebSocket.OPEN) {
          socket.send(encodeClientMessage({ type: 'operation', operation: op }))
        }
      }
    )

    socket.addEventListener('open', () => {
      awaitingSyncRef.current = true
      socket.send(encodeClientMessage({ type: 'sync' }))
    })

    socket.addEventListener('message', (event) => {
      const msg = parseServerMessage(event.data as string)
      if (!msg) {
        console.error('Invalid WebSocket message')
        return
      }
      handleServerMessage(msg, socket)
    })

    socket.addEventListener('close', () => {
      dispatcherRef.current?.setBlocked(true)
      onConnectionLost()
    })
    socket.addEventListener('error', (err) => {
      console.error('WebSocket error:', err)
    })

    return () => {
      socket.close()
      socketRef.current = null
      dispatcherRef.current = null
    }
  }, [documentId, handleServerMessage, onConnectionLost, socketRef])

  return { sendOperation }
}
