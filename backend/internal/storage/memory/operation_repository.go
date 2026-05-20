package memory

import (
	"context"
	"sync"

	"github.com/boatnoah/notedown/internal/documents"
	"github.com/boatnoah/notedown/internal/ot"
)

// OperationRepository persists operations in-memory for replay/testing.
type OperationRepository struct {
	mu  sync.RWMutex
	ops map[string][]ot.Operation
}

func NewOperationRepository() *OperationRepository {
	return &OperationRepository{ops: make(map[string][]ot.Operation)}
}

var _ documents.OperationRepository = (*OperationRepository)(nil)

func (r *OperationRepository) Append(ctx context.Context, documentID string, op ot.Operation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ops[documentID] = append(r.ops[documentID], op)
	return nil
}

func (r *OperationRepository) List(ctx context.Context, documentID string) ([]ot.Operation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ops := append([]ot.Operation{}, r.ops[documentID]...)
	return ops, nil
}
