import { describe, expect, it } from 'vitest'

import type { Operation } from '../../../lib/protocol'
import { createOpDispatcher } from './opDispatcher'

function insert(text: string, offset: number): Operation {
  return { kind: 'insert', offset, length: text.length, text }
}

function harness(initialVersion = 0) {
  const sent: Operation[] = []
  let version = initialVersion
  const dispatcher = createOpDispatcher(
    () => version,
    (op) => sent.push(op)
  )
  return {
    sent,
    dispatcher,
    ack(newVersion: number) {
      version = newVersion
      dispatcher.onAck()
    },
  }
}

describe('createOpDispatcher', () => {
  it('holds ops until unblocked', () => {
    const h = harness()
    h.dispatcher.enqueue(insert('h', 0))
    expect(h.sent).toHaveLength(0)

    h.dispatcher.setBlocked(false)
    expect(h.sent).toHaveLength(1)
  })

  it('sends one op at a time, stamping the version at send time', () => {
    const h = harness()
    h.dispatcher.setBlocked(false)

    // Burst-typed "hel": three ops created before any ack arrives.
    h.dispatcher.enqueue(insert('h', 0))
    h.dispatcher.enqueue(insert('e', 1))
    h.dispatcher.enqueue(insert('l', 2))

    expect(h.sent).toHaveLength(1)
    expect(h.sent[0]).toMatchObject({ text: 'h', clientVersion: 0 })

    h.ack(1)
    expect(h.sent).toHaveLength(2)
    expect(h.sent[1]).toMatchObject({ text: 'e', offset: 1, clientVersion: 1 })

    h.ack(2)
    expect(h.sent).toHaveLength(3)
    expect(h.sent[2]).toMatchObject({ text: 'l', offset: 2, clientVersion: 2 })
  })

  it('pauses while blocked and resumes the queue afterwards', () => {
    const h = harness()
    h.dispatcher.setBlocked(false)
    h.dispatcher.enqueue(insert('h', 0))
    h.ack(1)

    h.dispatcher.setBlocked(true)
    h.dispatcher.enqueue(insert('e', 1))
    expect(h.sent).toHaveLength(1)

    h.dispatcher.setBlocked(false)
    expect(h.sent).toHaveLength(2)
    expect(h.sent[1]).toMatchObject({ text: 'e', clientVersion: 1 })
  })

  it('continues past an op the server refused', () => {
    const h = harness()
    h.dispatcher.setBlocked(false)
    h.dispatcher.enqueue(insert('h', 0))
    h.dispatcher.enqueue(insert('e', 1))

    // Error frame: no version advance, but the in-flight op is settled.
    h.ack(0)
    expect(h.sent).toHaveLength(2)
    expect(h.sent[1]).toMatchObject({ text: 'e', clientVersion: 0 })
  })
})
