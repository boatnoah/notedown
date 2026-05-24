import { apiFetch, expectOk } from './client'

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
  const res = await apiFetch('/auth/logout', { method: 'POST' })
  await expectOk(res, 'Logout failed')
}
