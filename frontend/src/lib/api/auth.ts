import { apiFetch } from './client'

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterCredentials {
  name: string
  email: string
  username: string
  password: string
  pfp: string
}

export interface LoginResponse {
  accessToken: string
}

export interface RegisteredUser {
  id: string
  name: string
  email: string
  username: string
  pfp: string
  createdAt: string
}

async function expectOk(res: Response, label: string): Promise<Response> {
  if (!res.ok) {
    const body = await res.text().catch(() => '')
    throw new Error(`${label} (${res.status})${body ? `: ${body}` : ''}`)
  }
  return res
}

export async function login(credentials: LoginCredentials): Promise<LoginResponse> {
  const res = await apiFetch('/auth/login', {
    method: 'POST',
    body: JSON.stringify(credentials),
  })
  await expectOk(res, 'Login failed')
  return res.json() as Promise<LoginResponse>
}

export async function register(credentials: RegisterCredentials): Promise<RegisteredUser> {
  const res = await apiFetch('/auth/register', {
    method: 'POST',
    body: JSON.stringify(credentials),
  })
  await expectOk(res, 'Registration failed')
  return res.json() as Promise<RegisteredUser>
}

export async function logout(): Promise<void> {
  await apiFetch('/auth/logout', { method: 'POST' })
}
