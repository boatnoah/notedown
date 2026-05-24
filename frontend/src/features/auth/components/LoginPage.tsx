import { useSearch } from '@tanstack/react-router'

import { getBackendOrigin } from '../../../lib/config'

export function LoginPage() {
  const { redirect } = useSearch({ from: '/login' })

  const signIn = () => {
    // Store the intended destination so the app can redirect after OAuth completes.
    if (redirect) {
      sessionStorage.setItem('postLoginRedirect', redirect)
    }
    window.location.href = `${getBackendOrigin()}/auth/google`
  }

  return (
    <div>
      <h1>Please sign in</h1>
      <button type="button" onClick={signIn}>
        Sign in with Google
      </button>
    </div>
  )
}
