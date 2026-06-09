import { useMutation, useQueryClient } from '@tanstack/react-query'

import { updateShareMode } from '../../../lib/api/documents'
import type { ShareMode } from '../../../lib/protocol'

export function useUpdateShareMode(documentId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (shareMode: ShareMode) => updateShareMode(documentId, shareMode),
    onSuccess: (meta) => {
      qc.setQueryData(['document-meta', documentId], meta)
    },
  })
}
