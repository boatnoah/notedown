import { getBackendOrigin } from './config'
import type { DocumentRecord, Snapshot } from './protocol'

async function safeText(response: Response): Promise<string | null> {
  try {
    return await response.text()
  } catch {
    return null
  }
}

export async function createDocument(): Promise<DocumentRecord> {
  const response = await fetch(`${getBackendOrigin()}/documents`, {
    method: 'POST',
  })

  if (!response.ok) {
    const body = await safeText(response)
    throw new Error(`Failed to create document (${response.status}): ${body ?? ''}`)
  }

  return response.json()
}

export async function fetchSnapshot(documentId: string): Promise<Snapshot> {
  const response = await fetch(`${getBackendOrigin()}/documents/${documentId}`, {
    method: 'GET',
  })

  if (!response.ok) {
    const body = await safeText(response)
    throw new Error(`Failed to fetch document (${response.status}): ${body ?? ''}`)
  }

  return response.json()
}
