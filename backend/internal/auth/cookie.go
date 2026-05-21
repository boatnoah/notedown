package auth

import "net/http"

func refreshCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Path:     "/auth/refresh",
		HttpOnly: true,
		Secure:   false, // set true behind TLS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	}
}
