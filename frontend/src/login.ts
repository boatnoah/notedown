// src/auth.ts
export function initAuth() {
  const root = document.getElementById("app")!;
  root.innerHTML = `
    <h1>Please sign in</h1>
    <button id="btn-google">Sign in with Google</button>
  `;

  document.getElementById("btn-google")!.addEventListener("click", () => {
    window.location.href = "https://backend-ph82.onrender.com/auth/google";
  });
}
