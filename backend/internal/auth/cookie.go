package auth

import "net/http"

func refreshCookie(value string, maxAge int, secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	}
}

// isSecure reports whether the request arrived over TLS (directly or via proxy).
func isSecure(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}
