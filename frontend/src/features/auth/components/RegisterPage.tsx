import { Link, useNavigate } from '@tanstack/react-router'
import { type FormEvent, useState } from 'react'

import { useRegister } from '../hooks/useRegister'

const PFP_PRESETS = ['blue', 'green', 'red', 'yellow', 'purple', 'orange'] as const
type PfpPreset = (typeof PFP_PRESETS)[number]

const PFP_COLORS: Record<PfpPreset, string> = {
  blue: '#3b82f6',
  green: '#22c55e',
  red: '#ef4444',
  yellow: '#eab308',
  purple: '#a855f7',
  orange: '#f97316',
}

interface FieldErrors {
  name?: string
  email?: string
  username?: string
  password?: string
  pfp?: string
}

function validate(fields: {
  name: string
  email: string
  username: string
  password: string
  pfp: string
}): FieldErrors {
  const errors: FieldErrors = {}
  if (!fields.name.trim()) errors.name = 'Name is required'
  if (!fields.email) {
    errors.email = 'Email is required'
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(fields.email)) {
    errors.email = 'Invalid email address'
  }
  if (!fields.username.trim()) {
    errors.username = 'Username is required'
  } else if (!/^[a-zA-Z0-9_]+$/.test(fields.username)) {
    errors.username = 'Username may only contain letters, numbers, and underscores'
  }
  if (!fields.password) {
    errors.password = 'Password is required'
  } else if (fields.password.length < 8) {
    errors.password = 'Password must be at least 8 characters'
  }
  if (!fields.pfp) errors.pfp = 'Please choose an avatar'
  return errors
}

export function RegisterPage() {
  const navigate = useNavigate()

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [pfp, setPfp] = useState<PfpPreset | ''>('')
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})

  const { mutateAsync: register, isPending, error } = useRegister()

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const errors = validate({ name, email, username, password, pfp })
    setFieldErrors(errors)
    if (Object.keys(errors).length > 0) return

    try {
      await register({ name, email, username, password, pfp: pfp as PfpPreset })
      await navigate({ to: '/login', search: { redirect: undefined, registered: true } })
    } catch {
      // error captured by mutation state
    }
  }

  return (
    <div style={styles.page}>
      <div style={styles.card}>
        <h1 style={styles.heading}>Create an account</h1>

        <form onSubmit={handleSubmit} noValidate style={styles.form}>
          <div style={styles.field}>
            <label htmlFor="name" style={styles.label}>
              Name
            </label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={isPending}
              autoComplete="name"
              style={styles.input}
            />
            {fieldErrors.name && <p style={styles.fieldError}>{fieldErrors.name}</p>}
          </div>

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
            <label htmlFor="username" style={styles.label}>
              Username
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={isPending}
              autoComplete="username"
              style={styles.input}
            />
            {fieldErrors.username && <p style={styles.fieldError}>{fieldErrors.username}</p>}
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
              autoComplete="new-password"
              style={styles.input}
            />
            {fieldErrors.password && <p style={styles.fieldError}>{fieldErrors.password}</p>}
          </div>

          <div style={styles.field}>
            <span style={styles.label}>Avatar</span>
            <div style={styles.pfpGrid}>
              {PFP_PRESETS.map((preset) => (
                <button
                  key={preset}
                  type="button"
                  aria-label={preset}
                  aria-pressed={pfp === preset}
                  onClick={() => setPfp(preset)}
                  disabled={isPending}
                  style={{
                    ...styles.pfpButton,
                    background: PFP_COLORS[preset],
                    boxShadow:
                      pfp === preset ? `0 0 0 3px #fff, 0 0 0 5px ${PFP_COLORS[preset]}` : 'none',
                  }}
                />
              ))}
            </div>
            {fieldErrors.pfp && <p style={styles.fieldError}>{fieldErrors.pfp}</p>}
          </div>

          {error && (
            <p style={styles.formError}>
              {error instanceof Error ? error.message : 'Registration failed'}
            </p>
          )}

          <button type="submit" disabled={isPending} style={styles.submitButton}>
            {isPending ? 'Creating account…' : 'Create account'}
          </button>
        </form>

        <p style={styles.footer}>
          Already have an account?{' '}
          <Link to="/login" search={{ redirect: undefined, registered: undefined }}>
            Sign in
          </Link>
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
  pfpGrid: {
    display: 'flex',
    gap: '0.75rem',
    flexWrap: 'wrap' as const,
    paddingTop: '0.25rem',
  },
  pfpButton: {
    width: '40px',
    height: '40px',
    borderRadius: '50%',
    border: 'none',
    cursor: 'pointer',
    transition: 'box-shadow 0.1s',
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
  },
  footer: {
    marginTop: '1.5rem',
    textAlign: 'center' as const,
    fontSize: '0.875rem',
  },
} as const
