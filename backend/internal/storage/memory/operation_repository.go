package memory

import (
	"context"
	"sync"

	"github.com/boatnoah/notedown/internal/crdt"
	"github.com/boatnoah/notedown/internal/documents"
)

// OperationRepository persists operations in-memory for replay/testing.
type OperationRepository struct {
	mu  sync.RWMutex
	ops map[string][]crdt.Operation
}

func NewOperationRepository() *OperationRepository {
	return &OperationRepository{ops: make(map[string][]crdt.Operation)}
}

var _ documents.OperationRepository = (*OperationRepository)(nil)

func (r *OperationRepository) Append(ctx context.Context, documentID string, op crdt.Operation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ops[documentID] = append(r.ops[documentID], op)
	return nil
}

func (r *OperationRepository) List(ctx context.Context, documentID string) ([]crdt.Operation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ops := append([]crdt.Operation{}, r.ops[documentID]...)
	return ops, nil
}
