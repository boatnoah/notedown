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
// Create is the source of truth for uniqueness — it must return ErrDuplicateEmail
// or ErrDuplicateUsername when a conflict is detected (under a lock or DB constraint).
type Repository interface {
	Create(ctx context.Context, user *types.User, passwordHash string) error
}
