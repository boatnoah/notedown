package main

import (
	"fmt"
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
	connections map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		connections: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	fmt.Println("hello I am a new connection and I coming from: ", ws.RemoteAddr())

	s.mu.Lock()
	s.connections[ws] = true
	s.mu.Unlock()

	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
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

		s.broadcast(buffer[:n])

		fmt.Println(string(buffer[:n]))

	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.connections {
		go func(ws *websocket.Conn) {
			_, err := ws.Write(b)
			if err != nil {
			}
		}(ws)
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

func getGoogleAuthCallbackFunc(w http.ResponseWriter, r *http.Request) {
	for _, c := range r.Cookies() {
		fmt.Printf("→ incoming cookie: %s=%s; Domain=%s; Path=%s; Secure=%v\n",
			c.Name, c.Value, c.Domain, c.Path, c.Secure)
	}
	// We have to check if we are getting {provider} from the path
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

	// Log the authenticated user
	fmt.Println(user)

	// Redirect the user after authentication
	http.Redirect(w, r, "http://localhost:5173/editor", http.StatusFound)
}

func main() {
	NewAuth()
	router := http.NewServeMux()
	router.HandleFunc("/auth/google", handleProviderLogin)
	router.HandleFunc("/auth/google/callback", getGoogleAuthCallbackFunc)

	server := NewServer()
	router.Handle("/ws", websocket.Handler(server.handleWS))

	fmt.Println("listening on :3000")
	log.Fatal(http.ListenAndServe(":3000", router))
}
