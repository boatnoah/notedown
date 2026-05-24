package ot

import (
	"errors"
	"sync"
	"time"
)

// OperationKind represents mutation types supported by the server OT.
type OperationKind string

const (
	OperationInsert OperationKind = "insert"
	OperationDelete OperationKind = "delete"
)

// Operation captures a single text mutation emitted by a client.
type Operation struct {
	ID            string        `json:"id"`
	ClientID      string        `json:"clientId"`
	ClientVersion int64         `json:"clientVersion"`
	Kind          OperationKind `json:"kind"`
	Offset        int           `json:"offset"`
	Length        int           `json:"length"`
	Text          string        `json:"text"`
	Timestamp     time.Time     `json:"timestamp"`
}

// Snapshot exposes the canonical document state shared with clients.
type Snapshot struct {
	DocumentID string `json:"documentId"`
	Version    int64  `json:"version"`
	Content    string `json:"content"`
}

// Manager keeps the authoritative OT state per document.
type Manager struct {
	mu   sync.RWMutex
	docs map[string]*docState
}

type docState struct {
	Version int64
	Content []rune
	history []Operation // canonical operations in application order
}

// NewManager constructs an empty OT manager.
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

// Apply merges an operation into the canonical document state. If the
// operation's ClientVersion is behind the current document version, it is
// transformed against every operation applied since that version before being
// applied, ensuring convergent state across concurrent editors.
func (m *Manager) Apply(documentID string, op Operation) (Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.docs[documentID]
	if !ok {
		return Snapshot{}, errors.New("document not initialized")
	}

	if op.ClientVersion < 0 || op.ClientVersion > state.Version {
		return Snapshot{}, errors.New("invalid client version")
	}

	for _, concurrent := range state.history[op.ClientVersion:] {
		op = transform(op, concurrent)
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

	state.history = append(state.history, op)
	state.Version++

	return Snapshot{
		DocumentID: documentID,
		Version:    state.Version,
		Content:    string(state.Content),
	}, nil
}

// transform adjusts op so it produces the correct result when applied after
// against has already been applied. The server always has priority for the
// already-applied operation; at equal positions an insert-vs-insert tie is
// broken by shifting the incoming op to the right.
func transform(op, against Operation) Operation {
	result := op
	switch against.Kind {
	case OperationInsert:
		aLen := len([]rune(against.Text))
		switch op.Kind {
		case OperationInsert:
			// against sits at or before op's insertion point — shift right.
			// Equal-position tie-break: against wins (was applied first).
			if against.Offset <= op.Offset {
				result.Offset += aLen
			}
		case OperationDelete:
			if against.Offset <= op.Offset {
				result.Offset += aLen
			} else if against.Offset < op.Offset+op.Length {
				// against inserted inside the range op wants to delete — expand.
				result.Length += aLen
			}
		}
	case OperationDelete:
		aEnd := against.Offset + against.Length
		switch op.Kind {
		case OperationInsert:
			if aEnd <= op.Offset {
				// against's delete is entirely before op's insert point.
				result.Offset -= against.Length
			} else if against.Offset < op.Offset {
				// op's insert point fell inside the deleted region — clamp.
				result.Offset = against.Offset
			}
		case OperationDelete:
			oEnd := op.Offset + op.Length
			if aEnd <= op.Offset {
				// against is entirely before op.
				result.Offset -= against.Length
			} else if against.Offset < oEnd {
				// Ranges overlap: remove double-counted characters.
				leftShift := 0
				if against.Offset < op.Offset {
					leftShift = min(aEnd, op.Offset) - against.Offset
				}
				overlapStart := max(against.Offset, op.Offset)
				overlapEnd := min(aEnd, oEnd)
				overlap := overlapEnd - overlapStart
				result.Offset -= leftShift
				result.Length -= overlap
				if result.Length < 0 {
					result.Length = 0
				}
			}
		}
	}
	return result
}
