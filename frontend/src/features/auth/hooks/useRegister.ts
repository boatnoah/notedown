import { useMutation } from '@tanstack/react-query'

import { register } from '../../../lib/api/auth'

export function useRegister() {
  return useMutation({ mutationFn: register })
}
