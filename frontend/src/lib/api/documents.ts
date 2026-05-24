import { apiFetch, expectOk } from './client'
import type { DocumentRecord, Snapshot } from '../protocol'

export async function fetchDocuments(): Promise<DocumentRecord[]> {
  const res = await apiFetch('/documents')
  await expectOk(res, 'Failed to fetch documents')
  return res.json() as Promise<DocumentRecord[]>
}

export async function fetchDocument(id: string): Promise<Snapshot> {
  const res = await apiFetch(`/documents/${id}`)
  await expectOk(res, 'Failed to fetch document')
  return res.json() as Promise<Snapshot>
}

export async function createDocument(): Promise<DocumentRecord> {
  const res = await apiFetch('/documents', { method: 'POST' })
  await expectOk(res, 'Failed to create document')
  return res.json() as Promise<DocumentRecord>
}
