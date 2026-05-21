package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/boatnoah/notedown/internal/auth"
	"github.com/boatnoah/notedown/internal/storage/postgres/pgstore"
)

type AuthSessionRepository struct {
	q *pgstore.Queries
}

func NewAuthSessionRepository(db *sql.DB) *AuthSessionRepository {
	return &AuthSessionRepository{q: pgstore.New(db)}
}

var _ auth.SessionRepository = (*AuthSessionRepository)(nil)

func (r *AuthSessionRepository) Create(ctx context.Context, s *auth.AuthSession) error {
	row, err := r.q.CreateAuthSession(ctx, pgstore.CreateAuthSessionParams{
		UserID:           s.UserID,
		RefreshTokenHash: s.RefreshTokenHash,
		ExpiresAt:        s.ExpiresAt,
	})
	if err != nil {
		return err
	}
	s.ID = row.ID
	return nil
}

func (r *AuthSessionRepository) GetByTokenHash(ctx context.Context, hash string) (*auth.AuthSession, error) {
	row, err := r.q.GetAuthSessionByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrSessionNotFound
		}
		return nil, err
	}
	return &auth.AuthSession{
		ID:               row.ID,
		UserID:           row.UserID,
		RefreshTokenHash: row.RefreshTokenHash,
		ExpiresAt:        row.ExpiresAt,
	}, nil
}

func (r *AuthSessionRepository) Delete(ctx context.Context, id string) error {
	return r.q.DeleteAuthSession(ctx, id)
}

func (r *AuthSessionRepository) DeleteByTokenHash(ctx context.Context, hash string) error {
	return r.q.DeleteAuthSessionByTokenHash(ctx, hash)
}
