import { describe, expect, it } from 'vitest'

import type { DocumentMeta, ShareMode } from '../../../lib/protocol'
import { canEditDocument, isDocumentOwner } from './access'

function doc(shareMode: ShareMode): DocumentMeta {
  return {
    id: 'doc-1',
    ownerId: 'owner-1',
    title: 'Untitled',
    shareMode,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  }
}

describe('isDocumentOwner', () => {
  it('is true for the owner', () => {
    expect(isDocumentOwner(doc('private'), 'owner-1')).toBe(true)
  })

  it('is false for other users and missing user', () => {
    expect(isDocumentOwner(doc('edit'), 'someone-else')).toBe(false)
    expect(isDocumentOwner(doc('edit'), undefined)).toBe(false)
  })
})

describe('canEditDocument', () => {
  it('always allows the owner', () => {
    for (const mode of ['private', 'read', 'edit'] as const) {
      expect(canEditDocument(doc(mode), 'owner-1')).toBe(true)
    }
  })

  it('allows non-owners only in edit mode', () => {
    expect(canEditDocument(doc('edit'), 'viewer')).toBe(true)
    expect(canEditDocument(doc('read'), 'viewer')).toBe(false)
    expect(canEditDocument(doc('private'), 'viewer')).toBe(false)
  })

  it('denies edit when the user is unknown', () => {
    expect(canEditDocument(doc('edit'), undefined)).toBe(false)
  })
})
