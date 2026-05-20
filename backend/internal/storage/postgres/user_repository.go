package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/boatnoah/notedown/internal/storage/postgres/pgstore"
	"github.com/boatnoah/notedown/internal/users"
	"github.com/boatnoah/notedown/pkg/types"
)

type UserRepository struct {
	q *pgstore.Queries
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{q: pgstore.New(db)}
}

var _ users.Repository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, user *types.User, passwordHash string) error {
	row, err := r.q.CreateUser(ctx, pgstore.CreateUserParams{
		Name:         user.Name,
		Email:        user.Email,
		Username:     user.Username,
		PasswordHash: passwordHash,
		Pfp:          string(user.Pfp),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return users.ErrDuplicateEmail
			case "users_username_key":
				return users.ErrDuplicateUsername
			}
		}
		return err
	}
	user.ID = row.ID
	user.CreatedAt = row.CreatedAt
	return nil
}
