package port

import (
	"context"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx domain.Transaction) (int64, error)
}
