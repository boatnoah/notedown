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

func (r *UserRepository) GetByID(ctx context.Context, id string) (*types.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, users.ErrNotFound
		}
		return nil, err
	}
	return &types.User{
		ID:        row.ID,
		Name:      row.Name,
		Email:     row.Email,
		Username:  row.Username,
		Pfp:       types.PfpPreset(row.Pfp),
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*types.User, string, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", users.ErrNotFound
		}
		return nil, "", err
	}
	user := &types.User{
		ID:        row.ID,
		Name:      row.Name,
		Email:     row.Email,
		Username:  row.Username,
		Pfp:       types.PfpPreset(row.Pfp),
		CreatedAt: row.CreatedAt,
	}
	return user, row.PasswordHash, nil
}
