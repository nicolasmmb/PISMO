package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
)

const operationTypeSelectSQL = `SELECT id, description, sign FROM operation_types WHERE id = $1`

type OperationTypeRepository struct {
	tm *TransactionManagerDB
}

func NewOperationTypeRepository(db *sql.DB) *OperationTypeRepository {
	return &OperationTypeRepository{
		tm: NewTransactionManager(db),
	}
}

func (r *OperationTypeRepository) FindByID(ctx context.Context, id int) (domain.OperationType, error) {
	var ot domain.OperationType
	err := r.tm.GetExecutor(ctx).QueryRowContext(ctx, operationTypeSelectSQL, id).Scan(&ot.ID, &ot.Description, &ot.Sign)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.OperationType{}, domain.ErrOperationTypeNotFound
		}
		return domain.OperationType{}, fmt.Errorf("failed to find operation type: %w", err)
	}
	return ot, nil
}

func (r *OperationTypeRepository) SeedDefaults(ctx context.Context) error {
	return nil
}
