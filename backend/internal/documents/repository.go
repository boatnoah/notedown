package documents

import (
	"context"

	"github.com/boatnoah/notedown/internal/ot"
	"github.com/boatnoah/notedown/pkg/types"
)

// DocumentRepository defines persistence operations for document metadata.
type DocumentRepository interface {
	Save(ctx context.Context, doc *types.Document) error
	Get(ctx context.Context, id string) (*types.Document, error)
	ListByOwner(ctx context.Context, ownerID string) ([]*types.Document, error)
}

// OperationRepository stores the ordered list of CRDT operations for each document.
type OperationRepository interface {
	Append(ctx context.Context, documentID string, op ot.Operation) error
	List(ctx context.Context, documentID string) ([]ot.Operation, error)
}

// SessionRepository tracks live collaborative sessions. Included for future
// expansion when we enforce max concurrent editors, etc.
type SessionRepository interface {
	Create(ctx context.Context, session *types.Session) error
	Delete(ctx context.Context, id string) error
	ListByDocument(ctx context.Context, documentID string) ([]*types.Session, error)
}
