import { useQuery } from '@tanstack/react-query'

import { fetchDocuments } from '../../../lib/api/documents'
import type { DocumentRecord } from '../../../lib/protocol'

export function useDocuments() {
  return useQuery<DocumentRecord[]>({
    queryKey: ['documents'],
    queryFn: fetchDocuments,
  })
}
