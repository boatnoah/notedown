package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when an access token is missing, malformed,
// expired, or signed with the wrong key.
var ErrInvalidToken = errors.New("invalid or expired access token")

// Identity is the authenticated user extracted from a verified access token.
type Identity struct {
	UserID   string
	Name     string
	Username string
	Pfp      string
}

type identityCtxKey struct{}

// WithIdentity returns a child context carrying the authenticated identity.
func WithIdentity(ctx context.Context, id *Identity) context.Context {
	return context.WithValue(ctx, identityCtxKey{}, id)
}

// IdentityFromContext extracts the authenticated identity stored by RequireAuth.
func IdentityFromContext(ctx context.Context) (*Identity, bool) {
	id, ok := ctx.Value(identityCtxKey{}).(*Identity)
	return id, ok && id != nil
}

// ParseAccessToken verifies an access token's signature and expiry and
// returns the identity it asserts. It is shared by the HTTP middleware and
// the WebSocket handshake.
func ParseAccessToken(token, secret string) (*Identity, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !parsed.Valid || claims.Subject == "" {
		return nil, ErrInvalidToken
	}
	return &Identity{
		UserID:   claims.Subject,
		Name:     claims.Name,
		Username: claims.Username,
		Pfp:      claims.Pfp,
	}, nil
}

// RequireAuth returns middleware that rejects requests lacking a valid
// Bearer access token with 401 and otherwise stores the caller's Identity in
// the request context for downstream handlers.
func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const prefix = "Bearer "
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, prefix) {
				http.Error(w, "missing access token", http.StatusUnauthorized)
				return
			}
			identity, err := ParseAccessToken(strings.TrimSpace(header[len(prefix):]), secret)
			if err != nil {
				http.Error(w, "invalid or expired access token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithIdentity(r.Context(), identity)))
		})
	}
}
