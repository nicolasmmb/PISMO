package port

import (
	"context"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
)

// AccountRepository defines persistence operations for accounts.
type AccountRepository interface {
	Create(ctx context.Context, account domain.Account) (int64, error)
	FindByID(ctx context.Context, id int64) (domain.Account, error)
}
