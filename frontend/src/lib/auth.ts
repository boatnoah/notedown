const ACCESS_TOKEN_KEY = 'accessToken'

export function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

export function setAccessToken(token: string): void {
  localStorage.setItem(ACCESS_TOKEN_KEY, token)
}

export function clearAccessToken(): void {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
}

export function isAuthenticated(): boolean {
  return getAccessToken() !== null
}

// Exchanges the HttpOnly refresh_token cookie (set by the backend after login
// or OAuth) for a new access token. Returns true on success and persists the
// token so subsequent isAuthenticated() calls return true without a network hit.
export async function refreshAuth(backendOrigin: string): Promise<boolean> {
  try {
    const res = await fetch(`${backendOrigin}/auth/refresh`, {
      method: 'POST',
      credentials: 'include',
    })
    if (!res.ok) return false
    const data = (await res.json()) as { accessToken: string }
    setAccessToken(data.accessToken)
    return true
  } catch {
    return false
  }
}
