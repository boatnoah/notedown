import { getBackendOrigin } from '../../../lib/config'

// The typed `redirect` search param is defined on this route for future
// email/password login support. Google OAuth redirects through the backend,
// so the return destination is controlled there rather than here.
export function LoginPage() {
  const signIn = () => {
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
