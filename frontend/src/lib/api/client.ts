import { getAccessToken, setAccessToken, clearAccessToken } from '../auth'
import { getBackendOrigin } from '../config'

export class AuthError extends Error {
  constructor() {
    super('Authentication failed')
    this.name = 'AuthError'
  }
}

// Exported so callers can throw consistent errors without duplicating the
// "read body on failure" pattern.
export async function expectOk(res: Response, label: string): Promise<Response> {
  if (!res.ok) {
    const body = await res.text().catch(() => '')
    throw new Error(`${label} (${res.status})${body ? `: ${body}` : ''}`)
  }
  return res
}

// Lives here (not lib/auth.ts) so all network I/O is under src/lib/api/.
// Uses raw fetch directly to avoid recursion through apiFetch.
export async function refreshAuth(): Promise<boolean> {
  try {
    const res = await fetch(`${getBackendOrigin()}/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    })
    if (!res.ok) {
      clearAccessToken()
      return false
    }
    const data = (await res.json()) as { accessToken: string }
    setAccessToken(data.accessToken)
    return true
  } catch {
    return false
  }
}

async function doFetch(path: string, init: RequestInit, token: string | null): Promise<Response> {
  const headers = new Headers(init.headers)
  if (token) headers.set('Authorization', `Bearer ${token}`)
  if (init.body !== undefined && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  return fetch(`${getBackendOrigin()}${path}`, { ...init, headers })
}

export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const response = await doFetch(path, init, getAccessToken())

  if (response.status !== 401) {
    return response
  }

  const refreshed = await refreshAuth()
  if (!refreshed) {
    throw new AuthError()
  }

  const retry = await doFetch(path, init, getAccessToken())
  if (retry.status === 401) {
    throw new AuthError()
  }
  return retry
}
