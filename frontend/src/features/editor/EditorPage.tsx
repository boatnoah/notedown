import { useNavigate, useSearch } from '@tanstack/react-router'
import { useEffect, useState } from 'react'

import { createDocument, fetchSnapshot } from '../../lib/api'
import type { Snapshot } from '../../lib/protocol'
import { Editor } from './components/Editor'

export function EditorPage() {
  const { room } = useSearch({ from: '/auth/editor' })
  const navigate = useNavigate()

  const [snapshot, setSnapshot] = useState<Snapshot | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false

    async function load() {
      setLoading(true)
      setError(null)
      setSnapshot(null)
      try {
        if (!room) {
          const doc = await createDocument()
          if (!cancelled) {
            // Navigate updates the URL (and room), triggering this effect again
            // to fetch the snapshot. Return here to avoid a duplicate fetch.
            await navigate({ to: '/editor', search: { room: doc.id }, replace: true })
          }
          return
        }

        const snap = await fetchSnapshot(room)
        if (!cancelled) {
          setSnapshot(snap)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : String(err))
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    load()

    return () => {
      cancelled = true
    }
  }, [room, navigate])

  if (loading) {
    return <p>Loading editor…</p>
  }

  if (error || !room || !snapshot) {
    return <p className="error">Failed to load editor. {error ?? 'Unknown error'}</p>
  }

  return <Editor documentId={room} initialSnapshot={snapshot} />
}
