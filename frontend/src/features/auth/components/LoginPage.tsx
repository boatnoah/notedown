import { Link, useSearch } from '@tanstack/react-router'
import { type FormEvent, useState } from 'react'

import { useLogin } from '../hooks/useLogin'

interface FieldErrors {
  email?: string
  password?: string
}

function validate(email: string, password: string): FieldErrors {
  const errors: FieldErrors = {}
  if (!email) {
    errors.email = 'Email is required'
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    errors.email = 'Invalid email address'
  }
  if (!password) errors.password = 'Password is required'
  return errors
}

export function LoginPage() {
  const { redirect, registered } = useSearch({ from: '/login' })

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})

  const { mutateAsync: login, isPending, error } = useLogin()

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const errors = validate(email, password)
    setFieldErrors(errors)
    if (Object.keys(errors).length > 0) return

    try {
      await login({ email, password })
      // redirect is pre-validated as a safe relative path by validateSearch
      window.location.replace(redirect ?? '/documents')
    } catch {
      // error captured by mutation state
    }
  }

  return (
    <div style={styles.page}>
      <div style={styles.card}>
        <h1 style={styles.heading}>Sign in</h1>

        {registered && <p style={styles.successBanner}>Account created! Please sign in.</p>}

        <form onSubmit={handleSubmit} noValidate style={styles.form}>
          <div style={styles.field}>
            <label htmlFor="email" style={styles.label}>
              Email
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isPending}
              autoComplete="email"
              style={styles.input}
            />
            {fieldErrors.email && <p style={styles.fieldError}>{fieldErrors.email}</p>}
          </div>

          <div style={styles.field}>
            <label htmlFor="password" style={styles.label}>
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isPending}
              autoComplete="current-password"
              style={styles.input}
            />
            {fieldErrors.password && <p style={styles.fieldError}>{fieldErrors.password}</p>}
          </div>

          {error && (
            <p style={styles.formError}>
              {error instanceof Error ? error.message : 'Login failed'}
            </p>
          )}

          <button type="submit" disabled={isPending} style={styles.submitButton}>
            {isPending ? 'Signing in…' : 'Sign in'}
          </button>
        </form>

        <p style={styles.footer}>
          {"Don't have an account? "}
          <Link to="/register">Create one</Link>
        </p>
      </div>
    </div>
  )
}

const styles = {
  page: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: '100vh',
    padding: '1rem',
    boxSizing: 'border-box' as const,
  },
  card: {
    background: '#fff',
    borderRadius: '8px',
    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
    padding: '2rem',
    width: '100%',
    maxWidth: '400px',
  },
  heading: {
    margin: '0 0 1.5rem',
    fontSize: '1.5rem',
  },
  successBanner: {
    background: '#dcfce7',
    color: '#166534',
    borderRadius: '4px',
    padding: '0.75rem 1rem',
    margin: '0 0 1rem',
    fontSize: '0.875rem',
  },
  form: {
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '1rem',
  },
  field: {
    display: 'flex',
    flexDirection: 'column' as const,
    gap: '0.25rem',
  },
  label: {
    fontSize: '0.875rem',
    fontWeight: 500,
  },
  input: {
    padding: '0.5rem 0.75rem',
    border: '1px solid #d1d5db',
    borderRadius: '4px',
    fontSize: '1rem',
    width: '100%',
    boxSizing: 'border-box' as const,
  },
  fieldError: {
    margin: 0,
    fontSize: '0.8rem',
    color: '#dc2626',
  },
  formError: {
    margin: 0,
    padding: '0.75rem 1rem',
    background: '#fef2f2',
    color: '#dc2626',
    borderRadius: '4px',
    fontSize: '0.875rem',
  },
  submitButton: {
    padding: '0.625rem',
    background: '#2563eb',
    color: '#fff',
    border: 'none',
    borderRadius: '4px',
    fontSize: '1rem',
    cursor: 'pointer',
    opacity: 1,
  },
  footer: {
    marginTop: '1.5rem',
    textAlign: 'center' as const,
    fontSize: '0.875rem',
  },
} as const
