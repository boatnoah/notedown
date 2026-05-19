package users

import (
	"context"
	"errors"

	"github.com/boatnoah/notedown/pkg/types"
)

var (
	ErrDuplicateEmail    = errors.New("email already registered")
	ErrDuplicateUsername = errors.New("username already taken")
)

// Repository defines persistence operations for user accounts.
type Repository interface {
	Create(ctx context.Context, user *types.User, passwordHash string) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}
