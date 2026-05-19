package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/boatnoah/notedown/internal/users"
	"github.com/boatnoah/notedown/pkg/types"
)

type UserRepository struct {
	mu     sync.Mutex
	byID   map[string]*types.User
	hashes map[string]string // id -> password_hash
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		byID:   make(map[string]*types.User),
		hashes: make(map[string]string),
	}
}

// Create enforces email and username uniqueness under the lock before inserting.
func (r *UserRepository) Create(_ context.Context, user *types.User, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.byID {
		if u.Email == user.Email {
			return users.ErrDuplicateEmail
		}
		if u.Username == user.Username {
			return users.ErrDuplicateUsername
		}
	}

	user.ID = uuid.NewString()
	user.CreatedAt = time.Now().UTC()

	copy := *user
	r.byID[user.ID] = &copy
	r.hashes[user.ID] = passwordHash
	return nil
}
