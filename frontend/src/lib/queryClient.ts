import { MutationCache, QueryCache, QueryClient } from '@tanstack/react-query'

import { AuthError, HttpError } from './api/client'
import { clearAccessToken } from './auth'

function handleAuthError(error: unknown): void {
  if (error instanceof AuthError) {
    clearAccessToken()
    // Preserve the current location so login can return the user here —
    // important for share links opened with an expired session.
    const destination = window.location.pathname + window.location.search
    window.location.href = `/login?redirect=${encodeURIComponent(destination)}`
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
        // Client errors (403 no access, 404 missing, …) won't heal on retry.
        if (error instanceof HttpError && error.status < 500) return false
        return failureCount < 2
      },
    },
  },
})
