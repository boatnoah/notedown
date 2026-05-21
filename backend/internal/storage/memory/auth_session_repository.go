package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/boatnoah/notedown/internal/auth"
)

type AuthSessionRepository struct {
	mu     sync.RWMutex
	byID   map[string]*auth.AuthSession
	byHash map[string]*auth.AuthSession
}

func NewAuthSessionRepository() *AuthSessionRepository {
	return &AuthSessionRepository{
		byID:   make(map[string]*auth.AuthSession),
		byHash: make(map[string]*auth.AuthSession),
	}
}

var _ auth.SessionRepository = (*AuthSessionRepository)(nil)

func (r *AuthSessionRepository) Create(_ context.Context, s *auth.AuthSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s.ID = uuid.NewString()
	clone := *s
	r.byID[s.ID] = &clone
	r.byHash[s.RefreshTokenHash] = &clone
	return nil
}

func (r *AuthSessionRepository) GetByTokenHash(_ context.Context, hash string) (*auth.AuthSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.byHash[hash]
	if !ok {
		return nil, auth.ErrSessionNotFound
	}
	if time.Now().After(s.ExpiresAt) {
		return nil, auth.ErrSessionExpired
	}
	clone := *s
	return &clone, nil
}

func (r *AuthSessionRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.byID[id]
	if ok {
		delete(r.byHash, s.RefreshTokenHash)
		delete(r.byID, id)
	}
	return nil
}

func (r *AuthSessionRepository) DeleteByTokenHash(_ context.Context, hash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.byHash[hash]
	if ok {
		delete(r.byID, s.ID)
		delete(r.byHash, hash)
	}
	return nil
}

func (r *AuthSessionRepository) RotateSession(_ context.Context, oldHash string, newSession *auth.AuthSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.byHash[oldHash]
	if !ok {
		return auth.ErrSessionNotFound
	}
	delete(r.byID, old.ID)
	delete(r.byHash, oldHash)
	newSession.ID = uuid.NewString()
	clone := *newSession
	r.byID[newSession.ID] = &clone
	r.byHash[newSession.RefreshTokenHash] = &clone
	return nil
}
