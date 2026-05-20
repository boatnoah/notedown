package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/boatnoah/notedown/internal/documents"
	"github.com/boatnoah/notedown/internal/ot"
	"github.com/boatnoah/notedown/internal/storage/postgres/pgstore"
)

type OperationRepository struct {
	q *pgstore.Queries
}

func NewOperationRepository(db *sql.DB) *OperationRepository {
	return &OperationRepository{q: pgstore.New(db)}
}

var _ documents.OperationRepository = (*OperationRepository)(nil)

func (r *OperationRepository) Append(ctx context.Context, documentID string, op ot.Operation) error {
	id := op.ID
	if id == "" {
		id = uuid.NewString()
	}
	return r.q.AppendOperation(ctx, pgstore.AppendOperationParams{
		ID:         id,
		DocumentID: documentID,
		Kind:       string(op.Kind),
		CharOffset: int32(op.Offset),
		Length:     int32(op.Length),
		Text:       op.Text,
		CreatedAt:  op.Timestamp,
	})
}

func (r *OperationRepository) List(ctx context.Context, documentID string) ([]ot.Operation, error) {
	rows, err := r.q.ListOperations(ctx, documentID)
	if err != nil {
		return nil, err
	}
	ops := make([]ot.Operation, len(rows))
	for i, row := range rows {
		ops[i] = ot.Operation{
			ID:        row.ID,
			Kind:      ot.OperationKind(row.Kind),
			Offset:    int(row.CharOffset),
			Length:    int(row.Length),
			Text:      row.Text,
			Timestamp: row.CreatedAt,
		}
	}
	return ops, nil
}
