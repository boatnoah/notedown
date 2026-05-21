package auth

import (
	"context"
	"errors"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")
var ErrSessionExpired = errors.New("session expired")

// AuthSession represents a persisted refresh token session.
type AuthSession struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
}

// SessionRepository persists JWT refresh token sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *AuthSession) error
	GetByTokenHash(ctx context.Context, hash string) (*AuthSession, error)
	Delete(ctx context.Context, id string) error
	DeleteByTokenHash(ctx context.Context, hash string) error
}
