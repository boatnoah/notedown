import { useNavigate, useSearch } from '@tanstack/react-router'
import { useEffect, useState } from 'react'

import { createDocument, fetchSnapshot } from '../../lib/api'
import type { Snapshot } from '../../lib/protocol'
import { Editor } from './components/Editor'

export function EditorPage() {
  const { room } = useSearch({ from: '/auth/editor' })
  const navigate = useNavigate()

  const [documentId, setDocumentId] = useState<string | undefined>(room)
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        let id = documentId

        if (!id) {
          const doc = await createDocument()
          id = doc.id
          if (!cancelled) {
            setDocumentId(id)
            await navigate({ to: '/editor', search: { room: id }, replace: true })
          }
        }

        if (!id) {
          throw new Error('Unable to determine document ID')
        }

        const snap = await fetchSnapshot(id)
        if (!cancelled) {
          setSnapshot(snap)
          setError(null)
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
  }, [documentId, navigate])

  if (loading) {
    return <p>Loading editor…</p>
  }

  if (error || !documentId || !snapshot) {
    return <p className="error">Failed to load editor. {error ?? 'Unknown error'}</p>
  }

  return <Editor documentId={documentId} initialSnapshot={snapshot} />
}
