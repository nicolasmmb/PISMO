package usecase

import (
	"context"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
	"github.com/nicolasmmb/pismo-challenge/internal/port"
)

// GetAccount handles account retrieval.
type GetAccount struct {
	Accounts port.AccountRepository
}

// Execute returns an account by ID.
func (uc GetAccount) Execute(ctx context.Context, id int64) (domain.Account, error) {
	return uc.Accounts.FindByID(ctx, id)
}
