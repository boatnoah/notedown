import { getBackendOrigin } from "../../../lib/config";

export function LoginPage() {
  const signIn = () => {
    window.location.href = `${getBackendOrigin()}/auth/google`;
  };

  return (
    <div>
      <h1>Please sign in</h1>
      <button type="button" onClick={signIn}>
        Sign in with Google
      </button>
    </div>
  );
}
