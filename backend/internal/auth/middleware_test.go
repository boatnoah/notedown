package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/boatnoah/notedown/pkg/types"
)

const testSecret = "middleware-test-secret"

func mintToken(t *testing.T) string {
	t.Helper()
	user := &types.User{ID: "user-123", Name: "Ada", Username: "ada", Pfp: "blue"}
	token, err := issueAccessToken(user, testSecret)
	if err != nil {
		t.Fatalf("issueAccessToken: %v", err)
	}
	return token
}

func protectedHandler(t *testing.T, gotIdentity **Identity) http.Handler {
	t.Helper()
	return RequireAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := IdentityFromContext(r.Context())
		if !ok {
			t.Error("identity missing from context inside protected handler")
		}
		*gotIdentity = id
		w.WriteHeader(http.StatusOK)
	}))
}

func doAuthRequest(h http.Handler, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRequireAuthRejectsMissingHeader(t *testing.T) {
	var id *Identity
	rec := doAuthRequest(protectedHandler(t, &id), "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestRequireAuthRejectsMalformedToken(t *testing.T) {
	var id *Identity
	rec := doAuthRequest(protectedHandler(t, &id), "Bearer not-a-jwt")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestRequireAuthRejectsWrongSecret(t *testing.T) {
	user := &types.User{ID: "user-123", Name: "Ada", Username: "ada", Pfp: "blue"}
	token, err := issueAccessToken(user, "some-other-secret")
	if err != nil {
		t.Fatalf("issueAccessToken: %v", err)
	}
	var id *Identity
	rec := doAuthRequest(protectedHandler(t, &id), "Bearer "+token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestRequireAuthRejectsExpiredToken(t *testing.T) {
	claims := Claims{
		Name:     "Ada",
		Username: "ada",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	var id *Identity
	rec := doAuthRequest(protectedHandler(t, &id), "Bearer "+token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestRequireAuthRejectsUnsignedAlgorithm(t *testing.T) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	var id *Identity
	rec := doAuthRequest(protectedHandler(t, &id), "Bearer "+token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestRequireAuthAcceptsValidTokenAndExposesIdentity(t *testing.T) {
	var id *Identity
	rec := doAuthRequest(protectedHandler(t, &id), "Bearer "+mintToken(t))
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if id == nil {
		t.Fatal("identity not captured")
	}
	if id.UserID != "user-123" || id.Name != "Ada" || id.Username != "ada" || id.Pfp != "blue" {
		t.Fatalf("unexpected identity: %+v", id)
	}
}
