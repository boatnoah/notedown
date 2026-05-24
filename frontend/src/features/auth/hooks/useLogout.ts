import { useMutation, useQueryClient } from '@tanstack/react-query'

import { logout } from '../../../lib/api/auth'
import { clearAccessToken } from '../../../lib/auth'

export function useLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: logout,
    onSettled: () => {
      clearAccessToken()
      qc.clear()
    },
  })
}
