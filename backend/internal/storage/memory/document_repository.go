package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/boatnoah/notedown/internal/documents"
	"github.com/boatnoah/notedown/pkg/types"
)

// DocumentRepository is an in-memory map-backed implementation suitable for
// local development and tests.
type DocumentRepository struct {
	mu   sync.RWMutex
	data map[string]*types.Document
}

func NewDocumentRepository() *DocumentRepository {
	return &DocumentRepository{data: make(map[string]*types.Document)}
}

var _ documents.DocumentRepository = (*DocumentRepository)(nil)

func (r *DocumentRepository) Save(ctx context.Context, doc *types.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	clone := *doc
	r.data[doc.ID] = &clone
	return nil
}

func (r *DocumentRepository) Get(ctx context.Context, id string) (*types.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	doc, ok := r.data[id]
	if !ok {
		return nil, errors.New("document not found")
	}
	clone := *doc
	return &clone, nil
}

func (r *DocumentRepository) ListByOwner(ctx context.Context, ownerID string) ([]*types.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*types.Document
	for _, doc := range r.data {
		if doc.OwnerID == ownerID {
			clone := *doc
			result = append(result, &clone)
		}
	}
	return result, nil
}
