import { Link, useNavigate } from '@tanstack/react-router'

import type { DocumentRecord } from '../../../lib/protocol'
import { useCreateDocument } from '../hooks/useCreateDocument'
import { useDeleteDocument } from '../hooks/useDeleteDocument'
import { useDocuments } from '../hooks/useDocuments'

function formatTimestamp(iso: string): string {
  const date = new Date(iso)
  return Number.isNaN(date.getTime()) ? iso : date.toLocaleString()
}

function DocumentCard({ doc }: { doc: DocumentRecord }) {
  const { mutate: deleteDoc, isPending: isDeleting, error: deleteError } = useDeleteDocument()

  function handleDelete() {
    if (window.confirm('Delete this document? This cannot be undone.')) {
      deleteDoc(doc.id)
    }
  }

  const shareMode = doc.shareMode ?? 'private'

  return (
    <li style={styles.card}>
      <div style={styles.cardHeader}>
        <h2 style={styles.cardTitle}>{doc.title || 'Untitled'}</h2>
        <span style={styles.badge}>{shareMode}</span>
      </div>
      <p style={styles.cardMeta}>
        Updated <time dateTime={doc.updatedAt}>{formatTimestamp(doc.updatedAt)}</time>
      </p>
      <div style={styles.cardActions}>
        <Link to="/editor" search={{ room: doc.id }}>
          Open
        </Link>
        <button
          type="button"
          onClick={handleDelete}
          disabled={isDeleting}
          style={styles.deleteButton}
        >
          {isDeleting ? 'Deleting…' : 'Delete'}
        </button>
      </div>
      {deleteError && (
        <p role="alert" style={styles.error}>
          {deleteError instanceof Error ? deleteError.message : 'Failed to delete document'}
        </p>
      )}
    </li>
  )
}

export function DocumentsPage() {
  const navigate = useNavigate()
  const { data: documents, isPending, error } = useDocuments()
  const { mutateAsync: createDoc, isPending: isCreating, error: createError } = useCreateDocument()

  async function handleCreate() {
    try {
      const doc = await createDoc()
      void navigate({ to: '/editor', search: { room: doc.id } })
    } catch {
      // error captured by mutation state
    }
  }

  return (
    <main style={styles.page}>
      <header style={styles.header}>
        <h1 style={styles.heading}>Your documents</h1>
        <button type="button" onClick={handleCreate} disabled={isCreating} style={styles.newButton}>
          {isCreating ? 'Creating…' : 'New document'}
        </button>
      </header>

      {createError && (
        <p role="alert" style={styles.error}>
          {createError instanceof Error ? createError.message : 'Failed to create document'}
        </p>
      )}

      {isPending && <p>Loading documents…</p>}

      {error && (
        <p role="alert" style={styles.error}>
          {error instanceof Error ? error.message : 'Failed to load documents'}
        </p>
      )}

      {documents && documents.length === 0 && (
        <section style={styles.empty}>
          <p>You don’t have any documents yet.</p>
          <p>Create your first one with the “New document” button.</p>
        </section>
      )}

      {documents && documents.length > 0 && (
        <ul style={styles.list}>
          {documents.map((doc) => (
            <DocumentCard key={doc.id} doc={doc} />
          ))}
        </ul>
      )}
    </main>
  )
}

const styles = {
  page: {
    maxWidth: '720px',
    margin: '0 auto',
    padding: '2rem 1rem',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '1.5rem',
  },
  heading: {
    margin: 0,
    fontSize: '1.5rem',
  },
  newButton: {
    padding: '0.5rem 1rem',
    background: '#2563eb',
    color: '#fff',
    border: 'none',
    borderRadius: '4px',
    fontSize: '0.875rem',
    cursor: 'pointer',
  },
  error: {
    padding: '0.75rem 1rem',
    background: '#fef2f2',
    color: '#dc2626',
    borderRadius: '4px',
    fontSize: '0.875rem',
  },
  empty: {
    padding: '2rem',
    textAlign: 'center' as const,
    color: '#6b7280',
    border: '1px dashed #d1d5db',
    borderRadius: '8px',
  },
  list: {
    listStyle: 'none',
    margin: 0,
    padding: 0,
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '0.75rem',
  },
  card: {
    background: '#fff',
    border: '1px solid #e5e7eb',
    borderRadius: '8px',
    padding: '1rem 1.25rem',
  },
  cardHeader: {
    display: 'flex',
    alignItems: 'center',
    gap: '0.5rem',
  },
  cardTitle: {
    margin: 0,
    fontSize: '1.125rem',
  },
  badge: {
    fontSize: '0.75rem',
    textTransform: 'uppercase' as const,
    background: '#f3f4f6',
    color: '#374151',
    borderRadius: '9999px',
    padding: '0.125rem 0.5rem',
  },
  cardMeta: {
    margin: '0.25rem 0 0.75rem',
    fontSize: '0.875rem',
    color: '#6b7280',
  },
  cardActions: {
    display: 'flex',
    alignItems: 'center',
    gap: '0.75rem',
  },
  deleteButton: {
    padding: '0.25rem 0.75rem',
    background: 'transparent',
    color: '#dc2626',
    border: '1px solid #dc2626',
    borderRadius: '4px',
    fontSize: '0.8rem',
    cursor: 'pointer',
  },
} as const
