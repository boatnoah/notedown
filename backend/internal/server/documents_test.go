package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/boatnoah/notedown/internal/documents"
	"github.com/boatnoah/notedown/internal/ot"
	"github.com/boatnoah/notedown/internal/realtime"
	"github.com/boatnoah/notedown/internal/storage/memory"
	"github.com/boatnoah/notedown/pkg/types"
)

const testJWTSecret = "router-test-secret"

func mintToken(t *testing.T, userID string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub":      userID,
		"name":     "User " + userID,
		"username": userID,
		"pfp":      "blue",
		"exp":      time.Now().Add(time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	svc := documents.NewService(documents.Deps{
		Documents:  memory.NewDocumentRepository(),
		Operations: memory.NewOperationRepository(),
		Sessions:   memory.NewSessionRepository(),
		Manager:    ot.NewManager(),
	})
	return NewRouter(Dependencies{
		DocumentService: svc,
		RealtimeHub:     realtime.NewHub(svc, testJWTSecret),
		FrontendURL:     "http://localhost:5173",
		JWTSecret:       testJWTSecret,
	})
}

func request(t *testing.T, h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func createDoc(t *testing.T, h http.Handler, ownerToken string) types.Document {
	t.Helper()
	rec := request(t, h, http.MethodPost, "/documents", ownerToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("create document: got %d: %s", rec.Code, rec.Body.String())
	}
	var doc types.Document
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("decode document: %v", err)
	}
	return doc
}

func TestDocumentsEndpointsRequireAuth(t *testing.T) {
	h := newTestRouter(t)
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/documents"},
		{http.MethodGet, "/documents/some-id"},
		{http.MethodGet, "/documents/some-id/meta"},
		{http.MethodPost, "/documents/some-id/share"},
	} {
		rec := request(t, h, tc.method, tc.path, "", "")
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: got %d, want 401", tc.method, tc.path, rec.Code)
		}
	}
}

func TestCreateDocumentUsesAuthenticatedOwnerAndDefaultsPrivate(t *testing.T) {
	h := newTestRouter(t)
	doc := createDoc(t, h, mintToken(t, "owner-1"))
	if doc.OwnerID != "owner-1" {
		t.Fatalf("ownerId = %q, want owner-1", doc.OwnerID)
	}
	if doc.ShareMode != types.ShareModePrivate {
		t.Fatalf("shareMode = %q, want private", doc.ShareMode)
	}
}

func TestGetDocumentForbiddenForNonOwnerWhenPrivate(t *testing.T) {
	h := newTestRouter(t)
	doc := createDoc(t, h, mintToken(t, "owner-1"))

	for _, path := range []string{"/documents/" + doc.ID, "/documents/" + doc.ID + "/meta"} {
		rec := request(t, h, http.MethodGet, path, mintToken(t, "intruder"), "")
		if rec.Code != http.StatusForbidden {
			t.Errorf("GET %s: got %d, want 403", path, rec.Code)
		}
	}
}

func TestShareModeUpdateByOwnerGrantsAccess(t *testing.T) {
	h := newTestRouter(t)
	ownerToken := mintToken(t, "owner-1")
	otherToken := mintToken(t, "friend")
	doc := createDoc(t, h, ownerToken)

	rec := request(t, h, http.MethodPost, "/documents/"+doc.ID+"/share", ownerToken, `{"shareMode":"read"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("share update: got %d: %s", rec.Code, rec.Body.String())
	}
	var updated types.Document
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode updated document: %v", err)
	}
	if updated.ShareMode != types.ShareModeRead {
		t.Fatalf("shareMode = %q, want read", updated.ShareMode)
	}

	// The link is now readable by any authenticated user.
	rec = request(t, h, http.MethodGet, "/documents/"+doc.ID, otherToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET shared document: got %d, want 200", rec.Code)
	}
}

func TestShareModeUpdateForbiddenForNonOwner(t *testing.T) {
	h := newTestRouter(t)
	doc := createDoc(t, h, mintToken(t, "owner-1"))

	rec := request(t, h, http.MethodPost, "/documents/"+doc.ID+"/share", mintToken(t, "intruder"), `{"shareMode":"edit"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("got %d, want 403", rec.Code)
	}
}

func TestShareModeUpdateRejectsInvalidMode(t *testing.T) {
	h := newTestRouter(t)
	ownerToken := mintToken(t, "owner-1")
	doc := createDoc(t, h, ownerToken)

	rec := request(t, h, http.MethodPost, "/documents/"+doc.ID+"/share", ownerToken, `{"shareMode":"public"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", rec.Code)
	}
}

func TestShareModeUpdateUnknownDocumentIs404(t *testing.T) {
	h := newTestRouter(t)
	rec := request(t, h, http.MethodPost, "/documents/no-such-doc/share", mintToken(t, "owner-1"), `{"shareMode":"read"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", rec.Code)
	}
}
