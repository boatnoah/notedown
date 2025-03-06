package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

type Server struct {
	mu          sync.Mutex
	connections map[*websocket.Conn]bool
	document    *Document
}

type Document struct {
	documentId uuid.UUID
	content    []Operation
}

type Operation struct {
	ClientID string  `json:"clientID"`
	Value    string  `json:"value"`
	CharID   string  `json:"charID"`
	Action   string  `json:"action"`
	Position float32 `json:"position"`
}

func (d *Document) addOperation(operation Operation) {
	d.content = append(d.content, operation)
}

func (d *Document) deleteOperation(operation Operation) int {
	var index int
	left := 0
	right := len(d.content) - 1

	for left <= right {
		mid := (left + right) / 2
		fmt.Println("left: ", left)
		fmt.Println("right: ", right)
		fmt.Println("middle: ", mid)
		fmt.Println(d.content[mid].CharID, operation.CharID)
		if d.content[mid].CharID == operation.CharID {
			fmt.Println(d.content[mid].CharID, operation.CharID)
			index = mid
			break
		} else if d.content[mid].Position < operation.Position {
			left = mid + 1
		} else if d.content[mid].Position > operation.Position {
			right = mid - 1
		}
	}
	fmt.Println("left: ", left)
	fmt.Println("right: ", right)

	d.content = append(d.content[:index], d.content[index+1:]...)
	return index
}

func NewServer() *Server {
	return &Server{
		connections: make(map[*websocket.Conn]bool),
		document: &Document{
			documentId: uuid.New(),
			content:    []Operation{},
		},
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

		var operation Operation
		err = json.Unmarshal(buffer[:n], &operation)
		if err != nil {
			log.Fatal(err)
		}

		if operation.Action == "INSERT" {
			s.document.addOperation(operation)
			sort.Slice(s.document.content, func(i, j int) bool {
				return s.document.content[i].Position < s.document.content[j].Position
			})
		}

		if operation.Action == "DELETE" {
			i := s.document.deleteOperation(operation)
			fmt.Println(i)
		}

		jsonBytes, err := json.Marshal(s.document.content)
		if err != nil {
			log.Fatal(err)
		}

		s.broadcast(jsonBytes)

		fmt.Println(string(buffer[:n]))

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
