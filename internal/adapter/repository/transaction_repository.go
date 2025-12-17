package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
)

const transactionInsertSQL = `INSERT INTO transactions (account_id, operation_type_id, amount_cents, event_date, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`

type TransactionRepository struct {
	tm *TransactionManagerDB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{
		tm: NewTransactionManager(db),
	}
}

func (r *TransactionRepository) Create(ctx context.Context, tx domain.Transaction) (int64, error) {
	createdAt := tx.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	var id int64
	err := r.tm.GetExecutor(ctx).QueryRowContext(ctx, transactionInsertSQL,
		tx.AccountID, tx.OperationTypeID, tx.AmountCents, tx.EventDate, createdAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	return id, nil
}
