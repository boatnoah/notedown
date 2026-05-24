export type Snapshot = {
  documentId: string
  version: number
  content: string
}

export type Operation = {
  kind: 'insert' | 'delete'
  offset: number
  length: number
  text: string
}

export type DocumentRecord = {
  id: string
}

export type Presence = {
  userId: string
  name: string
  color: string
  anchor: number
  head: number
}

export type SnapshotMessage = {
  type: 'snapshot'
  snapshot: Snapshot
}

export type PresenceSnapshotMessage = {
  type: 'presenceSnapshot'
  presences: Record<string, Presence>
}

export type PresenceUpdateMessage = {
  type: 'presenceUpdate'
  userId: string
  presence: Presence
}

export type ErrorMessage = {
  type: 'error'
  error: string
}

export type ClientMessage =
  | { type: 'operation'; operation: Operation }
  | { type: 'sync' }
  | { type: 'presence'; presence: { anchor: number; head: number } }

export type ServerMessage =
  | SnapshotMessage
  | PresenceSnapshotMessage
  | PresenceUpdateMessage
  | ErrorMessage

// Encodes a ClientMessage to a JSON string for sending over WebSocket.
// All client→server sends must go through this function.
export function encodeClientMessage(msg: ClientMessage): string {
  return JSON.stringify(msg)
}

// Decodes a raw WebSocket frame into a typed ServerMessage.
// Returns null if the payload is not a recognised server message kind.
export function parseServerMessage(data: string): ServerMessage | null {
  try {
    const parsed: unknown = JSON.parse(data)
    if (typeof parsed !== 'object' || parsed === null || !('type' in parsed)) {
      return null
    }
    const { type } = parsed as { type: unknown }
    if (typeof type !== 'string') {
      return null
    }
    switch (type) {
      case 'snapshot':
      case 'presenceSnapshot':
      case 'presenceUpdate':
      case 'error':
        return parsed as ServerMessage
      default:
        return null
    }
  } catch {
    return null
  }
}

// Use in the default branch of a switch over ServerMessage to get a
// compile error whenever a new message kind is added to the union.
export function assertNever(x: never): never {
  throw new Error(`Unhandled server message type: ${JSON.stringify(x)}`)
}
