import type { Operation } from '../../../lib/protocol'

export type OpDispatcher = {
  /** Queue an operation for delivery. */
  enqueue: (op: Operation) => void
  /** Call when the server acknowledged the in-flight op (snapshot or error frame). */
  onAck: () => void
  /** Pause/resume delivery (socket closed, awaiting initial sync). Resuming pumps the queue. */
  setBlocked: (blocked: boolean) => void
}

// Serializes operation delivery: one op in flight at a time, stamped with the
// latest server version the client has applied. Sequential local edits are
// offset-relative to the previous edit's result, so each op must only be sent
// after the server acknowledged the one before it — otherwise the server
// would transform an op against its own predecessor and push it out of
// bounds.
export function createOpDispatcher(
  getVersion: () => number,
  sendNow: (op: Operation) => void
): OpDispatcher {
  const queue: Operation[] = []
  let inflight = false
  let blocked = true

  const pump = () => {
    if (inflight || blocked) {
      return
    }
    const next = queue.shift()
    if (!next) {
      return
    }
    inflight = true
    sendNow({ ...next, clientVersion: getVersion() })
  }

  return {
    enqueue(op) {
      queue.push(op)
      pump()
    },
    onAck() {
      inflight = false
      pump()
    },
    setBlocked(value) {
      blocked = value
      if (!blocked) {
        pump()
      }
    },
  }
}
