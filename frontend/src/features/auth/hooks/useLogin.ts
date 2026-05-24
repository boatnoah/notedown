import { useMutation, useQueryClient } from '@tanstack/react-query'

import { login } from '../../../lib/api/auth'
import { setAccessToken } from '../../../lib/auth'

export function useLogin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: login,
    onSuccess: ({ accessToken }) => {
      setAccessToken(accessToken)
      void qc.invalidateQueries()
    },
  })
}
