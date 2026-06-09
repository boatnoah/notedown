import { useQuery } from '@tanstack/react-query'

import { fetchDocumentMeta } from '../../../lib/api/documents'
import type { DocumentMeta } from '../../../lib/protocol'

export function useDocumentMeta(id: string | undefined) {
  return useQuery<DocumentMeta>({
    queryKey: ['document-meta', id],
    queryFn: () => {
      if (!id) throw new Error('useDocumentMeta called without an id')
      return fetchDocumentMeta(id)
    },
    enabled: Boolean(id),
  })
}
