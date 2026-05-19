package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/boatnoah/notedown/pkg/types"
)

type UserRepository struct {
	mu     sync.RWMutex
	users  map[string]*types.User // keyed by id
	hashes map[string]string      // id -> password_hash
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:  make(map[string]*types.User),
		hashes: make(map[string]string),
	}
}

func (r *UserRepository) Create(_ context.Context, user *types.User, passwordHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.ID = uuid.NewString()
	user.CreatedAt = time.Now().UTC()

	copy := *user
	r.users[user.ID] = &copy
	r.hashes[user.ID] = passwordHash
	return nil
}

func (r *UserRepository) ExistsByEmail(_ context.Context, email string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (r *UserRepository) ExistsByUsername(_ context.Context, username string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Username == username {
			return true, nil
		}
	}
	return false, nil
}
