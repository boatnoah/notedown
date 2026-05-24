import { useQuery } from '@tanstack/react-query'

import { fetchDocument } from '../../../lib/api/documents'
import type { Snapshot } from '../../../lib/protocol'

export function useDocument(id: string | undefined) {
  return useQuery<Snapshot>({
    queryKey: ['document', id],
    queryFn: () => fetchDocument(id!),
    enabled: id !== undefined,
  })
}
