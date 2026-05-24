import { useMutation, useQueryClient } from '@tanstack/react-query'

import { createDocument } from '../../../lib/api/documents'

export function useCreateDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: createDocument,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['documents'] })
    },
  })
}
