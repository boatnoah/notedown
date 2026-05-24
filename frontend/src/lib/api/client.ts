import { getAccessToken, refreshAuth } from '../auth'
import { getBackendOrigin } from '../config'

export class AuthError extends Error {
  constructor() {
    super('Authentication failed')
    this.name = 'AuthError'
  }
}

async function doFetch(path: string, init: RequestInit, token: string | null): Promise<Response> {
  const headers: Record<string, string> = {
    ...(init.headers as Record<string, string>),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(init.body !== undefined ? { 'Content-Type': 'application/json' } : {}),
  }
  return fetch(`${getBackendOrigin()}${path}`, { ...init, headers })
}

export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const response = await doFetch(path, init, getAccessToken())

  if (response.status !== 401) {
    return response
  }

  const refreshed = await refreshAuth(getBackendOrigin())
  if (!refreshed) {
    throw new AuthError()
  }

  return doFetch(path, init, getAccessToken())
}
