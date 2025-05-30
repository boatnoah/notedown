package main

import (
	"fmt"
	"github.com/google/uuid" // ← add this
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/markbates/goth/gothic"
	"golang.org/x/net/websocket"
)

type Server struct {
	mu          sync.Mutex
	connections map[string]map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		connections: make(map[string]map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	roomID := ws.Request().URL.Query().Get("room")

	s.mu.Lock()
	if _, ok := s.connections[roomID]; !ok {
		s.connections[roomID] = make(map[*websocket.Conn]bool)
	}
	s.connections[roomID][ws] = true
	s.mu.Unlock()

	s.readLoop(ws, roomID)
}

func (s *Server) readLoop(ws *websocket.Conn, roomID string) {
	buffer := make([]byte, 1024)

	for {
		n, err := ws.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected: ", ws.RemoteAddr())
			} else {
				fmt.Println("read error", err)
			}
			return
		}

		s.broadcast(buffer[:n], roomID)

		fmt.Println(string(buffer[:n]))

	}
}

func (s *Server) broadcast(b []byte, roomID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if room, exists := s.connections[roomID]; exists {
		for ws := range room {
			go func(ws *websocket.Conn) {
				_, err := ws.Write(b)
				if err != nil {
					fmt.Printf("Error broadcasting to client: %v\n", err)
				}
			}(ws)
		}
	}
}

func handleProviderLogin(w http.ResponseWriter, r *http.Request) {
	// Extract the provider (e.g., "google")
	pathSegments := strings.Split(r.URL.Path, "/")
	if len(pathSegments) < 3 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	provider := pathSegments[2] // Extracting 'google' from '/auth/google'

	// Add provider as a query parameter
	q := r.URL.Query()
	q.Add("provider", provider)
	r.URL.RawQuery = q.Encode()

	// Begin the authentication process
	gothic.BeginAuthHandler(w, r)
}

func (s *Server) getGoogleAuthCallbackFunc(w http.ResponseWriter, r *http.Request) {

	value := r.URL.Path
	fmt.Println("value", value)

	// Extract the provider from the URL path
	pathSegments := strings.Split(r.URL.Path, "/")
	if len(pathSegments) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	provider := pathSegments[2] // Extracting 'google' from '/auth/google/callback'

	// Log the extracted provider
	fmt.Println("provider", provider)

	// Add provider as a query parameter (this is what `gothic.CompleteUserAuth` expects)
	q := r.URL.Query()
	q.Add("provider", provider)
	r.URL.RawQuery = q.Encode()

	// Complete the authentication process
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	// 1) generate a random room ID
	roomID := uuid.New().String()

	// 2) init the room in memory
	s.mu.Lock()
	s.connections[roomID] = make(map[*websocket.Conn]bool)
	s.mu.Unlock()

	log.Printf("🎉 user %s logged in → new room %s created\n", user.Email, roomID)

	// 3) redirect them into the editor (front-end reads `room` query param and dials /ws)
	http.Redirect(w, r,
		fmt.Sprintf("http://localhost:5173/editor?room=%s", roomID),
		http.StatusFound,
	)

	// Log the authenticated user
	fmt.Println(user)

	// Redirect the user after authentication
	http.Redirect(w, r, "http://localhost:5173/editor", http.StatusFound)
}

func main() {

	server := NewServer()

	NewAuth()
	router := http.NewServeMux()
	router.Handle("/ws", websocket.Handler(server.handleWS))
	router.HandleFunc("/auth/google", handleProviderLogin)
	router.HandleFunc("/auth/google/callback", server.getGoogleAuthCallbackFunc)

	fmt.Println("listening on :3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}
