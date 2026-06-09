package realtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"

	"github.com/boatnoah/notedown/internal/documents"
	"github.com/boatnoah/notedown/internal/ot"
	"github.com/boatnoah/notedown/internal/storage/memory"
	"github.com/boatnoah/notedown/pkg/types"
)

const testSecret = "hub-test-secret"

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
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func newTestHub(t *testing.T) (*documents.Service, *httptest.Server) {
	t.Helper()
	svc := documents.NewService(documents.Deps{
		Documents:  memory.NewDocumentRepository(),
		Operations: memory.NewOperationRepository(),
		Sessions:   memory.NewSessionRepository(),
		Manager:    ot.NewManager(),
	})
	hub := NewHub(svc, testSecret)
	srv := httptest.NewServer(http.HandlerFunc(hub.HandleWebsocket))
	t.Cleanup(srv.Close)
	return svc, srv
}

func wsURL(srv *httptest.Server, documentID string) string {
	return "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?documentId=" + documentID
}

// dial attempts a WebSocket handshake, optionally with a bearer token.
func dial(t *testing.T, srv *httptest.Server, documentID, token string) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	dialer := websocket.Dialer{}
	if token != "" {
		dialer.Subprotocols = []string{"bearer", token}
	}
	return dialer.Dial(wsURL(srv, documentID), nil)
}

func mustDial(t *testing.T, srv *httptest.Server, documentID, token string) *websocket.Conn {
	t.Helper()
	conn, _, err := dial(t, srv, documentID, token)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// readMessageOfType reads frames until one with the given type arrives.
func readMessageOfType(t *testing.T, conn *websocket.Conn, msgType string) map[string]json.RawMessage {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	_ = conn.SetReadDeadline(deadline)
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("waiting for %q message: %v", msgType, err)
		}
		var frame map[string]json.RawMessage
		if err := json.Unmarshal(data, &frame); err != nil {
			t.Fatalf("invalid frame: %v", err)
		}
		var kind string
		_ = json.Unmarshal(frame["type"], &kind)
		if kind == msgType {
			return frame
		}
	}
}

func sendOperation(t *testing.T, conn *websocket.Conn, text string) {
	t.Helper()
	op := map[string]any{
		"type": "operation",
		"operation": map[string]any{
			"kind":   "insert",
			"offset": 0,
			"length": len(text),
			"text":   text,
		},
	}
	if err := conn.WriteJSON(op); err != nil {
		t.Fatalf("write operation: %v", err)
	}
}

func TestWebsocketRejectsUnauthenticatedWith401(t *testing.T) {
	svc, srv := newTestHub(t)
	doc, err := svc.CreateDocument(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("create document: %v", err)
	}

	_, resp, dialErr := dial(t, srv, doc.ID, "")
	if dialErr == nil {
		t.Fatal("expected handshake to fail without a token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got status %v, want 401", resp)
	}
}

func TestWebsocketRejectsInvalidTokenWith401(t *testing.T) {
	svc, srv := newTestHub(t)
	doc, err := svc.CreateDocument(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("create document: %v", err)
	}

	_, resp, dialErr := dial(t, srv, doc.ID, "garbage.token.value")
	if dialErr == nil {
		t.Fatal("expected handshake to fail with an invalid token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got status %v, want 401", resp)
	}
}

func TestWebsocketRejectsNonOwnerOnPrivateDocWith403(t *testing.T) {
	svc, srv := newTestHub(t)
	doc, err := svc.CreateDocument(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("create document: %v", err)
	}

	_, resp, dialErr := dial(t, srv, doc.ID, mintToken(t, "intruder"))
	if dialErr == nil {
		t.Fatal("expected handshake to fail for non-owner on private doc")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Fatalf("got status %v, want 403", resp)
	}
}

func TestWebsocketReturns404ForUnknownDocument(t *testing.T) {
	_, srv := newTestHub(t)

	_, resp, dialErr := dial(t, srv, "no-such-doc", mintToken(t, "someone"))
	if dialErr == nil {
		t.Fatal("expected handshake to fail for unknown document")
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		t.Fatalf("got status %v, want 404", resp)
	}
}

func TestWebsocketOwnerCanConnectAndEdit(t *testing.T) {
	svc, srv := newTestHub(t)
	doc, err := svc.CreateDocument(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("create document: %v", err)
	}

	conn := mustDial(t, srv, doc.ID, mintToken(t, "owner-1"))
	readMessageOfType(t, conn, "snapshot") // initial snapshot on register

	sendOperation(t, conn, "hello")
	frame := readMessageOfType(t, conn, "snapshot")
	var snap ot.Snapshot
	if err := json.Unmarshal(frame["snapshot"], &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snap.Content != "hello" {
		t.Fatalf("content = %q, want %q", snap.Content, "hello")
	}
}

func TestWebsocketReadModeAllowsConnectButRejectsOperations(t *testing.T) {
	svc, srv := newTestHub(t)
	doc, err := svc.CreateDocument(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("create document: %v", err)
	}
	if _, err := svc.SetShareMode(context.Background(), doc.ID, "owner-1", types.ShareModeRead); err != nil {
		t.Fatalf("set share mode: %v", err)
	}

	conn := mustDial(t, srv, doc.ID, mintToken(t, "viewer"))
	readMessageOfType(t, conn, "snapshot")

	sendOperation(t, conn, "sneaky edit")
	frame := readMessageOfType(t, conn, "error")
	var errText string
	_ = json.Unmarshal(frame["error"], &errText)
	if !strings.Contains(errText, "read-only") {
		t.Fatalf("error = %q, want read-only rejection", errText)
	}

	// Document must remain unchanged.
	snap, err := svc.Snapshot(context.Background(), doc.ID)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if snap.Content != "" {
		t.Fatalf("content = %q, want empty (operation must not apply)", snap.Content)
	}
}

func TestWebsocketEditModeAllowsNonOwnerToEdit(t *testing.T) {
	svc, srv := newTestHub(t)
	doc, err := svc.CreateDocument(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("create document: %v", err)
	}
	if _, err := svc.SetShareMode(context.Background(), doc.ID, "owner-1", types.ShareModeEdit); err != nil {
		t.Fatalf("set share mode: %v", err)
	}

	conn := mustDial(t, srv, doc.ID, mintToken(t, "collaborator"))
	readMessageOfType(t, conn, "snapshot")

	sendOperation(t, conn, "shared edit")
	frame := readMessageOfType(t, conn, "snapshot")
	var snap ot.Snapshot
	if err := json.Unmarshal(frame["snapshot"], &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snap.Content != "shared edit" {
		t.Fatalf("content = %q, want %q", snap.Content, "shared edit")
	}
}

func TestWebsocketEchoesBearerSubprotocol(t *testing.T) {
	svc, srv := newTestHub(t)
	doc, err := svc.CreateDocument(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("create document: %v", err)
	}

	conn, resp, dialErr := dial(t, srv, doc.ID, mintToken(t, "owner-1"))
	if dialErr != nil {
		t.Fatalf("dial: %v", dialErr)
	}
	defer func() { _ = conn.Close() }()
	if got := resp.Header.Get("Sec-WebSocket-Protocol"); got != "bearer" {
		t.Fatalf("negotiated subprotocol = %q, want %q", got, "bearer")
	}
}

// readSnapshotContent reads frames until a snapshot arrives and returns its
// raw snapshot payload.
func readSnapshotContent(t *testing.T, conn *websocket.Conn) string {
	t.Helper()
	frame := readMessageOfType(t, conn, "snapshot")
	return string(frame["snapshot"])
}

func TestOperationsRelayBetweenClients(t *testing.T) {
	svc, srv := newTestHub(t)
	doc, err := svc.CreateDocument(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("create doc: %v", err)
	}
	if _, err := svc.SetShareMode(context.Background(), doc.ID, "owner-1", "edit"); err != nil {
		t.Fatalf("set share mode: %v", err)
	}

	ownerConn := mustDial(t, srv, doc.ID, mintToken(t, "owner-1"))
	guestConn := mustDial(t, srv, doc.ID, mintToken(t, "guest-1"))
	readSnapshotContent(t, ownerConn)
	readSnapshotContent(t, guestConn)

	sendOperation(t, ownerConn, "afa")
	if got := readSnapshotContent(t, guestConn); !strings.Contains(got, "afa") {
		t.Fatalf("guest did not receive owner's edit: %s", got)
	}
	readSnapshotContent(t, ownerConn) // drain owner's own broadcast

	sendOperation(t, guestConn, "asdf")
	if got := readSnapshotContent(t, ownerConn); !strings.Contains(got, "asdf") {
		t.Fatalf("owner did not receive guest's edit: %s", got)
	}
}

// Share-mode changes must take effect on connections that are already open:
// a viewer who connected under "read" gains editing as soon as the owner
// widens the mode to "edit". Regression test for live clients silently
// diverging because canEdit was frozen at connect time.
func TestShareModeWideningAppliesToLiveConnections(t *testing.T) {
	svc, srv := newTestHub(t)
	doc, err := svc.CreateDocument(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("create doc: %v", err)
	}
	if _, err := svc.SetShareMode(context.Background(), doc.ID, "owner-1", "read"); err != nil {
		t.Fatalf("set share mode read: %v", err)
	}

	ownerConn := mustDial(t, srv, doc.ID, mintToken(t, "owner-1"))
	guestConn := mustDial(t, srv, doc.ID, mintToken(t, "guest-1"))
	readSnapshotContent(t, ownerConn)
	readSnapshotContent(t, guestConn)

	if _, err := svc.SetShareMode(context.Background(), doc.ID, "owner-1", "edit"); err != nil {
		t.Fatalf("set share mode edit: %v", err)
	}

	sendOperation(t, guestConn, "asdf")
	if got := readSnapshotContent(t, ownerConn); !strings.Contains(got, "asdf") {
		t.Fatalf("owner did not receive guest's edit after share mode widened: %s", got)
	}
}

// The inverse: tightening edit -> read revokes editing on live connections.
func TestShareModeTighteningRevokesLiveConnections(t *testing.T) {
	svc, srv := newTestHub(t)
	doc, err := svc.CreateDocument(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("create doc: %v", err)
	}
	if _, err := svc.SetShareMode(context.Background(), doc.ID, "owner-1", "edit"); err != nil {
		t.Fatalf("set share mode edit: %v", err)
	}

	guestConn := mustDial(t, srv, doc.ID, mintToken(t, "guest-1"))
	readSnapshotContent(t, guestConn)

	if _, err := svc.SetShareMode(context.Background(), doc.ID, "owner-1", "read"); err != nil {
		t.Fatalf("set share mode read: %v", err)
	}

	sendOperation(t, guestConn, "asdf")
	frame := readMessageOfType(t, guestConn, "error")
	var errMsg string
	_ = json.Unmarshal(frame["error"], &errMsg)
	if !strings.Contains(errMsg, "read-only") {
		t.Fatalf("expected read-only rejection, got: %s", errMsg)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	snapshot, err := svc.Snapshot(ctx, doc.ID)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if strings.Contains(snapshot.Content, "asdf") {
		t.Fatalf("document mutated by revoked editor: %q", snapshot.Content)
	}
}
