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
//
// Create contract: implementations must (1) populate user.ID and user.CreatedAt
// before returning, (2) enforce email+username uniqueness, returning
// ErrDuplicateEmail or ErrDuplicateUsername on conflict.
type Repository interface {
	Create(ctx context.Context, user *types.User, passwordHash string) error
}
