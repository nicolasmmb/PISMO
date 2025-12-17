package usecase

import (
	"context"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
	"github.com/nicolasmmb/pismo-challenge/internal/port"
)

type GetAccount struct {
	Accounts port.AccountRepository
}

func (uc GetAccount) Execute(ctx context.Context, id int64) (domain.Account, error) {
	return uc.Accounts.FindByID(ctx, id)
}
