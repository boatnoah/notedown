package realtime

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/boatnoah/notedown/internal/auth"
	"github.com/boatnoah/notedown/internal/documents"
)

// bearerSubprotocol is the Sec-WebSocket-Protocol entry that marks the
// companion entry as a bearer access token. Browsers cannot attach an
// Authorization header to WebSocket handshakes, so the client sends
// ["bearer", "<access token>"] as subprotocols and the server echoes
// "bearer" back. Using the header (rather than a query parameter) keeps the
// token out of request logs.
const bearerSubprotocol = "bearer"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub coordinates WebSocket connections per document.
type Hub struct {
	service   *documents.Service
	jwtSecret string
	mu        sync.RWMutex
	rooms     map[string]map[*Client]struct{}
	presence  *PresenceStore
}

func NewHub(service *documents.Service, jwtSecret string) *Hub {
	return &Hub{
		service:   service,
		jwtSecret: jwtSecret,
		rooms:     make(map[string]map[*Client]struct{}),
		presence:  NewPresenceStore(60 * time.Second),
	}
}

// bearerTokenFromSubprotocols extracts the access token smuggled through the
// Sec-WebSocket-Protocol header alongside the "bearer" marker.
func bearerTokenFromSubprotocols(r *http.Request) string {
	for _, proto := range websocket.Subprotocols(r) {
		if proto != bearerSubprotocol {
			return proto
		}
	}
	return ""
}

// HandleWebsocket authenticates and authorizes the caller, then upgrades the
// connection and registers the client. Access control must happen before the
// upgrade so the handshake can fail with a meaningful HTTP status:
// 401 for missing/invalid tokens, 403 for callers without document access.
func (h *Hub) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	documentID := r.URL.Query().Get("documentId")
	if documentID == "" {
		documentID = r.URL.Query().Get("room")
	}
	if documentID == "" {
		http.Error(w, "missing documentId", http.StatusBadRequest)
		return
	}

	token := bearerTokenFromSubprotocols(r)
	if token == "" {
		http.Error(w, "missing access token", http.StatusUnauthorized)
		return
	}
	identity, err := auth.ParseAccessToken(token, h.jwtSecret)
	if err != nil {
		http.Error(w, "invalid or expired access token", http.StatusUnauthorized)
		return
	}

	doc, err := h.service.GetDocument(r.Context(), documentID)
	if err != nil {
		if errors.Is(err, documents.ErrNotFound) {
			http.Error(w, "document not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
	if !doc.CanRead(identity.UserID) {
		http.Error(w, "you do not have access to this document", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, http.Header{
		"Sec-WebSocket-Protocol": {bearerSubprotocol},
	})
	if err != nil {
		log.Printf("ws upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:        h,
		conn:       conn,
		send:       make(chan []byte, 32),
		documentID: documentID,
		userID:     identity.UserID,
		name:       identity.Name,
		username:   identity.Username,
		pfp:        identity.Pfp,
	}

	h.register(client)

	go client.writeLoop()
	go client.readLoop()
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[c.documentID]; !ok {
		h.rooms[c.documentID] = make(map[*Client]struct{})
	}
	h.rooms[c.documentID][c] = struct{}{}

	pres := Presence{
		UserID:   c.userID,
		Name:     c.name,
		Username: c.username,
		Pfp:      c.pfp,
		Color:    assignColor(c.userID),
		Anchor:   0,
		Head:     0,
	}
	h.presence.Update(c.documentID, pres)

	ctx, cancel := rWithTimeout()
	defer cancel()
	snapshot, err := h.service.Snapshot(ctx, c.documentID)
	if err == nil {
		if payload, merr := MarshalServer(SnapshotMsg{Snapshot: snapshot}); merr == nil {
			c.send <- payload
		} else {
			log.Printf("marshal snapshot: %v", merr)
		}
	}

	if payload, merr := MarshalServer(PresenceSnapshotMsg{Presences: h.presence.Snapshot(c.documentID)}); merr == nil {
		c.send <- payload
	} else {
		log.Printf("marshal presence snapshot: %v", merr)
	}
}

func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[c.documentID]; ok {
		delete(room, c)
		if len(room) == 0 {
			delete(h.rooms, c.documentID)
		}
	}

	h.presence.Remove(c.documentID, c.userID)
	if payload, err := MarshalServer(PresenceUpdateMsg{UserID: c.userID, Presence: Presence{}}); err == nil {
		go h.broadcast(c.documentID, payload)
	} else {
		log.Printf("marshal presence update: %v", err)
	}

	close(c.send)
}

func (h *Hub) broadcast(documentID string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, ok := h.rooms[documentID]
	if !ok {
		return
	}
	for client := range room {
		select {
		case client.send <- payload:
		default:
			log.Printf("dropping client send buffer for doc=%s", documentID)
		}
	}
}

type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	documentID string
	userID     string
	name       string
	username   string
	pfp        string
}

// canEdit checks the document's current share mode so that owners changing
// sharing mid-session take effect on live connections — both widening
// (read → edit lets connected viewers start editing) and tightening
// (edit → read revokes editing without requiring a reconnect).
func (c *Client) canEdit() bool {
	ctx, cancel := rWithTimeout()
	defer cancel()
	doc, err := c.hub.service.GetDocument(ctx, c.documentID)
	if err != nil {
		return false
	}
	return doc.CanEdit(c.userID)
}

func (c *Client) readLoop() {
	defer func() {
		c.hub.unregister(c)
		_ = c.conn.Close()
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		msg, err := UnmarshalClient(data)
		if err != nil {
			log.Printf("invalid ws message: %v", err)
			continue
		}

		switch m := msg.(type) {
		case OperationMsg:
			if !c.canEdit() {
				if payload, merr := MarshalServer(ErrorMsg{Error: "read-only access: operations are not permitted"}); merr == nil {
					select {
					case c.send <- payload:
					default:
					}
				} else {
					log.Printf("marshal error msg: %v", merr)
				}
				continue
			}
			m.Operation.Timestamp = time.Now().UTC()
			ctx, cancel := rWithTimeout()
			snapshot, err := c.hub.service.ApplyOperation(ctx, c.documentID, m.Operation)
			cancel()
			if err != nil {
				log.Printf("apply op failed: %v", err)
				if payload, merr := MarshalServer(ErrorMsg{Error: err.Error()}); merr == nil {
					select {
					case c.send <- payload:
					default:
					}
				} else {
					log.Printf("marshal error msg: %v", merr)
				}
				continue
			}
			if payload, err := MarshalServer(SnapshotMsg{Snapshot: snapshot}); err == nil {
				c.hub.broadcast(c.documentID, payload)
			} else {
				log.Printf("marshal snapshot: %v", err)
			}
		case SyncMsg:
			ctx, cancel := rWithTimeout()
			snapshot, err := c.hub.service.Snapshot(ctx, c.documentID)
			cancel()
			if err != nil {
				log.Printf("snapshot failed: %v", err)
				if payload, merr := MarshalServer(ErrorMsg{Error: err.Error()}); merr == nil {
					select {
					case c.send <- payload:
					default:
					}
				} else {
					log.Printf("marshal error msg: %v", merr)
				}
				continue
			}
			if payload, err := MarshalServer(SnapshotMsg{Snapshot: snapshot}); err == nil {
				select {
				case c.send <- payload:
				default:
				}
			} else {
				log.Printf("marshal snapshot: %v", err)
			}
		case PresenceMsg:
			pres := Presence{
				UserID:   c.userID,
				Name:     c.name,
				Username: c.username,
				Pfp:      c.pfp,
				Color:    assignColor(c.userID),
				Anchor:   m.Presence.Anchor,
				Head:     m.Presence.Head,
			}
			c.hub.presence.Update(c.documentID, pres)
			if payload, err := MarshalServer(PresenceUpdateMsg{UserID: c.userID, Presence: pres}); err == nil {
				c.hub.broadcast(c.documentID, payload)
			} else {
				log.Printf("marshal presence update: %v", err)
			}
		}
	}
}

func (c *Client) writeLoop() {
	defer func() { _ = c.conn.Close() }()
	for payload := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			return
		}
	}
}

func rWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
