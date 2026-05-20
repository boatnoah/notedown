package documents

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/boatnoah/notedown/internal/crdt"
	"github.com/boatnoah/notedown/pkg/types"
)

// Service orchestrates document metadata, CRDT operations, and session state.
type Service struct {
	docs     DocumentRepository
	ops      OperationRepository
	sessions SessionRepository
	manager  *crdt.Manager

	loadMu sync.Mutex
	loaded map[string]struct{}
}

// Deps enumerates the collaborators required to construct the service.
type Deps struct {
	Documents  DocumentRepository
	Operations OperationRepository
	Sessions   SessionRepository
	Manager    *crdt.Manager
}

func NewService(deps Deps) *Service {
	return &Service{
		docs:     deps.Documents,
		ops:      deps.Operations,
		sessions: deps.Sessions,
		manager:  deps.Manager,
		loaded:   make(map[string]struct{}),
	}
}

// ensureLoaded guarantees the CRDT manager has the document initialized and all
// persisted operations replayed. On the first call for a given document after a
// server restart this fetches from the DB; subsequent calls are a map lookup.
func (s *Service) ensureLoaded(ctx context.Context, documentID string) error {
	s.loadMu.Lock()
	defer s.loadMu.Unlock()

	if _, ok := s.loaded[documentID]; ok {
		return nil
	}

	if _, err := s.docs.Get(ctx, documentID); err != nil {
		return err
	}

	s.manager.InitDocument(documentID)

	if s.ops != nil {
		ops, err := s.ops.List(ctx, documentID)
		if err != nil {
			return err
		}
		for _, op := range ops {
			if _, err := s.manager.Apply(documentID, op); err != nil {
				return err
			}
		}
	}

	s.loaded[documentID] = struct{}{}
	return nil
}

// CreateDocument registers a new document and initializes its CRDT state.
func (s *Service) CreateDocument(ctx context.Context, ownerID string) (*types.Document, error) {
	if ownerID == "" {
		return nil, errors.New("ownerID required")
	}

	doc := &types.Document{
		ID:        uuid.NewString(),
		OwnerID:   ownerID,
		Title:     "Untitled",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.docs.Save(ctx, doc); err != nil {
		return nil, err
	}

	s.manager.InitDocument(doc.ID)

	s.loadMu.Lock()
	s.loaded[doc.ID] = struct{}{}
	s.loadMu.Unlock()

	return doc, nil
}

// Snapshot returns the current canonical state for a document.
func (s *Service) Snapshot(ctx context.Context, documentID string) (crdt.Snapshot, error) {
	if err := s.ensureLoaded(ctx, documentID); err != nil {
		return crdt.Snapshot{}, err
	}
	return s.manager.Snapshot(documentID)
}

// ApplyOperation validates and merges a CRDT operation, returning the new
// document snapshot once the canonical state has been updated.
func (s *Service) ApplyOperation(ctx context.Context, documentID string, op crdt.Operation) (crdt.Snapshot, error) {
	if err := s.ensureLoaded(ctx, documentID); err != nil {
		return crdt.Snapshot{}, err
	}

	snapshot, err := s.manager.Apply(documentID, op)
	if err != nil {
		return crdt.Snapshot{}, err
	}

	if s.ops != nil {
		_ = s.ops.Append(ctx, documentID, op)
	}
	return snapshot, nil
}

// ListDocuments fetches all documents owned by the provided user.
func (s *Service) ListDocuments(ctx context.Context, ownerID string) ([]*types.Document, error) {
	return s.docs.ListByOwner(ctx, ownerID)
}
