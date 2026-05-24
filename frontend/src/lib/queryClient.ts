import { MutationCache, QueryCache, QueryClient } from '@tanstack/react-query'

import { AuthError } from './api/client'
import { clearAccessToken } from './auth'

function handleAuthError(error: unknown): void {
  if (error instanceof AuthError) {
    clearAccessToken()
    window.location.href = '/login'
  }
}

export const queryClient = new QueryClient({
  queryCache: new QueryCache({ onError: handleAuthError }),
  mutationCache: new MutationCache({ onError: handleAuthError }),
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: (failureCount, error) => {
        if (error instanceof AuthError) return false
        return failureCount < 2
      },
    },
  },
})
