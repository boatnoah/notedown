package realtime

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"

	"github.com/boatnoah/notedown/internal/crdt"
	"github.com/boatnoah/notedown/internal/documents"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub coordinates WebSocket connections per document.
type Hub struct {
	service  *documents.Service
	mu       sync.RWMutex
	rooms    map[string]map[*Client]struct{}
	presence *PresenceStore
}

func NewHub(service *documents.Service) *Hub {
	return &Hub{
		service:  service,
		rooms:    make(map[string]map[*Client]struct{}),
		presence: NewPresenceStore(60 * time.Second),
	}
}

// HandleWebsocket upgrades HTTP connections and registers clients.
func (h *Hub) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	documentID := r.URL.Query().Get("documentId")
	if documentID == "" {
		documentID = r.URL.Query().Get("room")
	}
	if documentID == "" {
		http.Error(w, "missing documentId", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:        h,
		conn:       conn,
		send:       make(chan []byte, 32),
		documentID: documentID,
		userID:     middleware.GetReqID(r.Context()),
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

	// initialize presence entry for this user
	pres := Presence{
		UserID: c.userID,
		Name:   c.userID,
		Color:  assignColor(c.userID),
		Anchor: 0,
		Head:   0,
	}
	h.presence.Update(c.documentID, pres)

	ctx, cancel := rWithTimeout()
	defer cancel()
	snapshot, err := h.service.Snapshot(ctx, c.documentID)
	if err == nil {
		payload, _ := json.Marshal(outboundMessage{Type: "snapshot", Snapshot: snapshot})
		c.send <- payload
	}

	presencePayload, _ := json.Marshal(presenceSnapshot{Type: "presenceSnapshot", Presences: h.presence.Snapshot(c.documentID)})
	c.send <- presencePayload
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
	update := presenceUpdate{Type: "presenceUpdate", UserID: c.userID, Presence: Presence{}}
	payload, _ := json.Marshal(update)
	go h.broadcast(c.documentID, payload)

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
}

type inboundMessage struct {
	Type      string          `json:"type"`
	Operation *crdt.Operation `json:"operation,omitempty"`
	Presence  *CursorPayload  `json:"presence,omitempty"`
}

type outboundMessage struct {
	Type     string        `json:"type"`
	Snapshot crdt.Snapshot `json:"snapshot"`
}

type errorMessage struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

type CursorPayload struct {
	Anchor int `json:"anchor"`
	Head   int `json:"head"`
}

type presenceSnapshot struct {
	Type      string              `json:"type"`
	Presences map[string]Presence `json:"presences"`
}

type presenceUpdate struct {
	Type     string   `json:"type"`
	UserID   string   `json:"userId"`
	Presence Presence `json:"presence"`
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

		var msg inboundMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("invalid ws message: %v", err)
			continue
		}

		switch msg.Type {
		case "operation":
			if msg.Operation == nil {
				continue
			}
			msg.Operation.Timestamp = time.Now().UTC()
			ctx, cancel := rWithTimeout()
			snapshot, err := c.hub.service.ApplyOperation(ctx, c.documentID, *msg.Operation)
			cancel()
			if err != nil {
				log.Printf("apply op failed: %v", err)
				payload, _ := json.Marshal(errorMessage{Type: "error", Error: err.Error()})
				select {
				case c.send <- payload:
				default:
				}
				continue
			}
			payload, _ := json.Marshal(outboundMessage{Type: "snapshot", Snapshot: snapshot})
			c.hub.broadcast(c.documentID, payload)
		case "sync":
			ctx, cancel := rWithTimeout()
			snapshot, err := c.hub.service.Snapshot(ctx, c.documentID)
			cancel()
			if err != nil {
				log.Printf("snapshot failed: %v", err)
				payload, _ := json.Marshal(errorMessage{Type: "error", Error: err.Error()})
				select {
				case c.send <- payload:
				default:
				}
				continue
			}
			payload, _ := json.Marshal(outboundMessage{Type: "snapshot", Snapshot: snapshot})
			select {
			case c.send <- payload:
			default:
			}
		case "presence":
			if msg.Presence == nil {
				continue
			}
			pres := Presence{
				UserID: c.userID,
				Name:   c.userID,
				Color:  assignColor(c.userID),
				Anchor: msg.Presence.Anchor,
				Head:   msg.Presence.Head,
			}
			c.hub.presence.Update(c.documentID, pres)
			payload, _ := json.Marshal(presenceUpdate{Type: "presenceUpdate", UserID: c.userID, Presence: pres})
			c.hub.broadcast(c.documentID, payload)
		default:
			log.Printf("unsupported message type: %s", msg.Type)
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
