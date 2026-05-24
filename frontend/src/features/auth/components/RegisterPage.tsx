import { Link } from '@tanstack/react-router'

/** Placeholder for the registration page (see issue #14). */
export function RegisterPage() {
  return (
    <div>
      <h1>Create an account</h1>
      <p>Registration coming soon.</p>
      <Link to="/login" search={{ redirect: undefined }}>
        Back to sign in
      </Link>
    </div>
  )
}
