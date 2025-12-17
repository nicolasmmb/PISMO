package port

import (
	"context"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
)

type AccountRepository interface {
	Create(ctx context.Context, account domain.Account) (int64, error)
	FindByID(ctx context.Context, id int64) (domain.Account, error)
	FindByIDForUpdate(ctx context.Context, id int64) (domain.Account, error)
}
