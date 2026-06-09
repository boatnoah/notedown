package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/boatnoah/notedown/internal/documents"
	"github.com/boatnoah/notedown/internal/storage/postgres/pgstore"
	"github.com/boatnoah/notedown/pkg/types"
)

type DocumentRepository struct {
	q *pgstore.Queries
}

func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{q: pgstore.New(db)}
}

var _ documents.DocumentRepository = (*DocumentRepository)(nil)

func (r *DocumentRepository) Save(ctx context.Context, doc *types.Document) error {
	return r.q.UpsertDocument(ctx, pgstore.UpsertDocumentParams{
		ID:        doc.ID,
		OwnerID:   doc.OwnerID,
		Title:     doc.Title,
		ShareMode: string(doc.ShareMode),
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	})
}

func (r *DocumentRepository) Get(ctx context.Context, id string) (*types.Document, error) {
	row, err := r.q.GetDocument(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, documents.ErrNotFound
		}
		return nil, err
	}
	return &types.Document{
		ID:        row.ID,
		OwnerID:   row.OwnerID,
		Title:     row.Title,
		ShareMode: types.ShareMode(row.ShareMode),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id string) error {
	affected, err := r.q.DeleteDocument(ctx, id)
	if err != nil {
		return err
	}
	if affected == 0 {
		return documents.ErrNotFound
	}
	return nil
}

func (r *DocumentRepository) ListByOwner(ctx context.Context, ownerID string) ([]*types.Document, error) {
	rows, err := r.q.ListDocumentsByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	out := make([]*types.Document, len(rows))
	for i, row := range rows {
		out[i] = &types.Document{
			ID:        row.ID,
			OwnerID:   row.OwnerID,
			Title:     row.Title,
			ShareMode: types.ShareMode(row.ShareMode),
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}
	return out, nil
}
