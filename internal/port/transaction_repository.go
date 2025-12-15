package port

import (
	"context"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
)

// TransactionRepository defines persistence operations for transactions.
type TransactionRepository interface {
	Create(ctx context.Context, tx domain.Transaction) (int64, error)
}
