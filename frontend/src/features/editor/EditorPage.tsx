import { useNavigate, useSearch } from '@tanstack/react-router'
import { useEffect } from 'react'

import { HttpError } from '../../lib/api/client'
import { useCurrentUser } from '../auth/hooks/useCurrentUser'
import { useCreateDocument } from '../documents/hooks/useCreateDocument'
import { useDocument } from '../documents/hooks/useDocument'
import { Editor } from './components/Editor'
import { useDocumentMeta } from './hooks/useDocumentMeta'

export function EditorPage() {
  const { room } = useSearch({ from: '/auth/editor' })
  const navigate = useNavigate()
  const currentUser = useCurrentUser()

  const { mutateAsync: createDoc, isPending: isCreating, error: createError } = useCreateDocument()
  const { data: snapshot, isPending: isFetchPending, error: fetchError } = useDocument(room)
  const { data: meta, isPending: isMetaPending, error: metaError } = useDocumentMeta(room)

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

  if (isCreating || !room || isFetchPending || isMetaPending) {
    return <p>Loading editor…</p>
  }

  const error = fetchError ?? metaError
  if (error || !snapshot || !meta) {
    if (error instanceof HttpError && error.status === 403) {
      return (
        <p className="error">
          You don&apos;t have access to this document. Ask the owner to share it with you.
        </p>
      )
    }
    return (
      <p className="error">
        Failed to load editor. {error instanceof Error ? error.message : 'Unknown error'}
      </p>
    )
  }

  return (
    <Editor
      documentId={room}
      initialSnapshot={snapshot}
      meta={meta}
      currentUserId={currentUser?.id}
    />
  )
}
