import { useMutation, useQueryClient } from '@tanstack/react-query'

import { deleteDocument } from '../../../lib/api/documents'

export function useDeleteDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: deleteDocument,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['documents'] })
    },
  })
}
