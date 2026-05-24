import { useNavigate, useSearch } from '@tanstack/react-router'
import { useEffect } from 'react'

import { useCreateDocument } from '../documents/hooks/useCreateDocument'
import { useDocument } from '../documents/hooks/useDocument'
import { Editor } from './components/Editor'

export function EditorPage() {
  const { room } = useSearch({ from: '/auth/editor' })
  const navigate = useNavigate()

  const { mutateAsync: createDoc, isPending: isCreating, error: createError } = useCreateDocument()
  const { data: snapshot, isPending: isFetchPending, error: fetchError } = useDocument(room)

  useEffect(() => {
    if (room) return
    let cancelled = false
    createDoc(undefined)
      .then((doc) => {
        if (!cancelled) {
          void navigate({ to: '/editor', search: { room: doc.id }, replace: true })
        }
      })
      .catch(() => {
        // createError captures this via useMutation state
      })
    return () => {
      cancelled = true
    }
  }, [room, createDoc, navigate])

  if (createError) {
    return (
      <p className="error">
        Failed to create document.{' '}
        {createError instanceof Error ? createError.message : 'Unknown error'}
      </p>
    )
  }

  if (isCreating || !room || isFetchPending) {
    return <p>Loading editor…</p>
  }

  if (fetchError || !snapshot) {
    return (
      <p className="error">
        Failed to load editor. {fetchError instanceof Error ? fetchError.message : 'Unknown error'}
      </p>
    )
  }

  return <Editor documentId={room} initialSnapshot={snapshot} />
}
