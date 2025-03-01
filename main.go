package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"

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
				break
			}
			fmt.Println("read error", err)
		}

		message := buffer[:n]
		fmt.Println(string(message))

		s.broadcast(message)

	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.connections {
		go func(ws *websocket.Conn) {
			_, err := ws.Write(b)
			if err != nil {
				fmt.Println("write error", err)
			}
		}(ws)
	}
}

func main() {
	server := NewServer()
	http.Handle("/ws", websocket.Handler(server.handleWS))
	fs := http.FileServer(http.Dir("client"))
	http.Handle("/", fs)
	fmt.Println("Started server")
	http.ListenAndServe(":3000", nil)
}
