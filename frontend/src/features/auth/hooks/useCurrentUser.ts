import { getAccessToken } from '../../../lib/auth'

export interface User {
  id: string
  name: string
  username: string
  avatarUrl: string
}

function parseJwt(token: string): Record<string, unknown> | null {
  try {
    const [, payload] = token.split('.')
    const padded = payload + '='.repeat((4 - (payload.length % 4)) % 4)
    return JSON.parse(atob(padded.replace(/-/g, '+').replace(/_/g, '/'))) as Record<string, unknown>
  } catch {
    return null
  }
}

export function useCurrentUser(): User | null {
  const token = getAccessToken()
  if (!token) return null

  const claims = parseJwt(token)
  if (!claims) return null

  const { sub, name, username, pfp } = claims
  if (
    typeof sub !== 'string' ||
    typeof name !== 'string' ||
    typeof username !== 'string' ||
    typeof pfp !== 'string'
  ) {
    return null
  }

  return { id: sub, name, username, avatarUrl: pfp }
}
