import { apiFetch } from './client'
import type { DocumentRecord, Snapshot } from '../protocol'

async function expectOk(res: Response, label: string): Promise<Response> {
  if (!res.ok) {
    const body = await res.text().catch(() => '')
    throw new Error(`${label} (${res.status})${body ? `: ${body}` : ''}`)
  }
  return res
}

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
