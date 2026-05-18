package memory

import (
	"context"
	"sync"

	"github.com/boatnoah/notedown/internal/documents"
	"github.com/boatnoah/notedown/pkg/types"
)

// SessionRepository stores session metadata for connected users.
type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*types.Session
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{sessions: make(map[string]*types.Session)}
}

var _ documents.SessionRepository = (*SessionRepository)(nil)

func (r *SessionRepository) Create(ctx context.Context, session *types.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := *session
	r.sessions[session.ID] = &clone
	return nil
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, id)
	return nil
}

func (r *SessionRepository) ListByDocument(ctx context.Context, documentID string) ([]*types.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*types.Session
	for _, session := range r.sessions {
		if session.DocumentID == documentID {
			clone := *session
			result = append(result, &clone)
		}
	}
	return result, nil
}
