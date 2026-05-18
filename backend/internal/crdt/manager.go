package crdt

import (
	"errors"
	"sync"
	"time"
)

// OperationKind represents mutation types supported by the server CRDT.
type OperationKind string

const (
	OperationInsert OperationKind = "insert"
	OperationDelete OperationKind = "delete"
)

// Operation captures a single text mutation emitted by a client.
type Operation struct {
	ID        string        `json:"id"`
	ClientID  string        `json:"clientId"`
	Kind      OperationKind `json:"kind"`
	Offset    int           `json:"offset"`
	Length    int           `json:"length"`
	Text      string        `json:"text"`
	Timestamp time.Time     `json:"timestamp"`
}

// Snapshot exposes the canonical document state shared with clients.
type Snapshot struct {
	DocumentID string `json:"documentId"`
	Version    int64  `json:"version"`
	Content    string `json:"content"`
}

// Manager keeps the authoritative CRDT state per document.
type Manager struct {
	mu   sync.RWMutex
	docs map[string]*docState
}

type docState struct {
	Version int64
	Content []rune
}

// NewManager constructs an empty CRDT manager.
func NewManager() *Manager {
	return &Manager{docs: make(map[string]*docState)}
}

// InitDocument ensures an entry exists for the provided document ID.
func (m *Manager) InitDocument(documentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.docs[documentID]; !exists {
		m.docs[documentID] = &docState{Version: 0, Content: []rune{}}
	}
}

// Snapshot retrieves the latest document state.
func (m *Manager) Snapshot(documentID string) (Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, ok := m.docs[documentID]
	if !ok {
		return Snapshot{}, errors.New("document not initialized")
	}

	return Snapshot{
		DocumentID: documentID,
		Version:    state.Version,
		Content:    string(state.Content),
	}, nil
}

// Apply merges an operation into the canonical document state.
func (m *Manager) Apply(documentID string, op Operation) (Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.docs[documentID]
	if !ok {
		return Snapshot{}, errors.New("document not initialized")
	}

	switch op.Kind {
	case OperationInsert:
		if op.Offset < 0 || op.Offset > len(state.Content) {
			return Snapshot{}, errors.New("insert offset out of bounds")
		}
		before := append([]rune{}, state.Content[:op.Offset]...)
		after := append([]rune{}, state.Content[op.Offset:]...)
		state.Content = append(before, append([]rune(op.Text), after...)...)
	case OperationDelete:
		if op.Offset < 0 || op.Offset+op.Length > len(state.Content) {
			return Snapshot{}, errors.New("delete range out of bounds")
		}
		state.Content = append(append([]rune{}, state.Content[:op.Offset]...), state.Content[op.Offset+op.Length:]...)
	default:
		return Snapshot{}, errors.New("unsupported operation kind")
	}

	state.Version++

	return Snapshot{
		DocumentID: documentID,
		Version:    state.Version,
		Content:    string(state.Content),
	}, nil
}
