import type { DocumentMeta } from '../../../lib/protocol'

// Mirrors the server-side access rules (backend/pkg/types/document.go).
// The server is the source of truth; these helpers only drive UI state.

export function isDocumentOwner(doc: DocumentMeta, userId: string | undefined): boolean {
  return Boolean(userId) && doc.ownerId === userId
}

export function canEditDocument(doc: DocumentMeta, userId: string | undefined): boolean {
  if (!userId) return false
  if (isDocumentOwner(doc, userId)) return true
  return doc.shareMode === 'edit'
}
