import { QueryCache, QueryClient } from '@tanstack/react-query'

import { AuthError } from './api/client'
import { clearAccessToken } from './auth'

export const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (error) => {
      if (error instanceof AuthError) {
        clearAccessToken()
        window.location.href = '/login'
      }
    },
  }),
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
