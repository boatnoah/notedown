const envBackend =
  (import.meta.env.VITE_API_URL as string | undefined) ||
  (import.meta.env.VITE_WS_URL as string | undefined) ||
  "";

function defaultBackendOrigin(): string {
  if (typeof window === "undefined") {
    return "http://localhost:3000";
  }

  const { protocol, hostname } = window.location;
  const port = protocol === "https:" ? "443" : "3000";
  return `${protocol}//${hostname}:${port}`;
}

function normalizeUrl(url: string): string {
  if (!url) {
    return url;
  }
  return url.endsWith("/") ? url.slice(0, -1) : url;
}

export function getBackendOrigin(): string {
  if (envBackend) {
    return normalizeUrl(envBackend);
  }

  return normalizeUrl(defaultBackendOrigin());
}

export function getWebSocketUrl(documentId: string): string {
  const backend = new URL(getBackendOrigin());
  const protocol = backend.protocol === "https:" ? "wss" : "ws";
  const query = new URLSearchParams({ documentId }).toString();
  return `${protocol}://${backend.host}/ws?${query}`;
}
