package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/boatnoah/notedown/internal/auth"
	"github.com/boatnoah/notedown/internal/storage/postgres/pgstore"
)

type AuthSessionRepository struct {
	db *sql.DB
	q  *pgstore.Queries
}

func NewAuthSessionRepository(db *sql.DB) *AuthSessionRepository {
	return &AuthSessionRepository{db: db, q: pgstore.New(db)}
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

func (r *AuthSessionRepository) RotateSession(ctx context.Context, oldHash string, newSession *auth.AuthSession) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE refresh_token_hash = $1", oldHash)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return auth.ErrSessionNotFound
	}

	row := tx.QueryRowContext(ctx,
		"INSERT INTO sessions (user_id, refresh_token_hash, expires_at) VALUES ($1, $2, $3) RETURNING id, created_at",
		newSession.UserID, newSession.RefreshTokenHash, newSession.ExpiresAt,
	)
	var createdAt time.Time
	if err := row.Scan(&newSession.ID, &createdAt); err != nil {
		return err
	}

	return tx.Commit()
}
