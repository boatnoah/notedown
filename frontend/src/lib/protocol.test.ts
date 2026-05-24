import { describe, expect, it } from 'vitest'

import {
  encodeClientMessage,
  parseServerMessage,
  type ClientMessage,
  type ServerMessage,
} from './protocol'

// ---------- encodeClientMessage round-trip ----------

describe('encodeClientMessage', () => {
  it('encodes operation message', () => {
    const msg: ClientMessage = {
      type: 'operation',
      operation: { kind: 'insert', offset: 3, length: 0, text: 'hi' },
    }
    const encoded = encodeClientMessage(msg)
    const parsed = JSON.parse(encoded) as unknown
    expect((parsed as { type: string }).type).toBe('operation')
    expect((parsed as { operation: { text: string } }).operation.text).toBe('hi')
  })

  it('encodes sync message', () => {
    const encoded = encodeClientMessage({ type: 'sync' })
    const parsed = JSON.parse(encoded) as unknown
    expect((parsed as { type: string }).type).toBe('sync')
  })

  it('encodes presence message', () => {
    const msg: ClientMessage = {
      type: 'presence',
      presence: { anchor: 4, head: 9 },
    }
    const encoded = encodeClientMessage(msg)
    const parsed = JSON.parse(encoded) as { presence: { anchor: number; head: number } }
    expect(parsed.presence.anchor).toBe(4)
    expect(parsed.presence.head).toBe(9)
  })
})

// ---------- parseServerMessage round-trip ----------

describe('parseServerMessage', () => {
  it('parses snapshot message', () => {
    const raw = JSON.stringify({
      type: 'snapshot',
      snapshot: { documentId: 'doc1', version: 2, content: 'hello' },
    })
    const msg = parseServerMessage(raw) as ServerMessage
    expect(msg).not.toBeNull()
    expect(msg.type).toBe('snapshot')
    if (msg.type === 'snapshot') {
      expect(msg.snapshot.content).toBe('hello')
      expect(msg.snapshot.version).toBe(2)
    }
  })

  it('parses presenceSnapshot message', () => {
    const raw = JSON.stringify({
      type: 'presenceSnapshot',
      presences: {
        u1: { userId: 'u1', name: 'Alice', color: '#f00', anchor: 0, head: 1 },
      },
    })
    const msg = parseServerMessage(raw) as ServerMessage
    expect(msg).not.toBeNull()
    expect(msg.type).toBe('presenceSnapshot')
    if (msg.type === 'presenceSnapshot') {
      expect(msg.presences['u1'].name).toBe('Alice')
    }
  })

  it('parses presenceUpdate message', () => {
    const raw = JSON.stringify({
      type: 'presenceUpdate',
      userId: 'u2',
      presence: { userId: 'u2', name: 'Bob', color: '#0f0', anchor: 5, head: 8 },
    })
    const msg = parseServerMessage(raw) as ServerMessage
    expect(msg).not.toBeNull()
    expect(msg.type).toBe('presenceUpdate')
    if (msg.type === 'presenceUpdate') {
      expect(msg.userId).toBe('u2')
      expect(msg.presence.anchor).toBe(5)
    }
  })

  it('parses error message', () => {
    const raw = JSON.stringify({ type: 'error', error: 'something went wrong' })
    const msg = parseServerMessage(raw) as ServerMessage
    expect(msg).not.toBeNull()
    expect(msg.type).toBe('error')
    if (msg.type === 'error') {
      expect(msg.error).toBe('something went wrong')
    }
  })

  it('returns null for unknown type', () => {
    expect(parseServerMessage(JSON.stringify({ type: 'bogus' }))).toBeNull()
  })

  it('returns null for malformed JSON', () => {
    expect(parseServerMessage('not json')).toBeNull()
  })

  it('returns null for missing type field', () => {
    expect(parseServerMessage(JSON.stringify({ data: 'oops' }))).toBeNull()
  })
})
