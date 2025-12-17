package usecase

import (
	"context"
	"time"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
	"github.com/nicolasmmb/pismo-challenge/internal/port"
)

type CreateAccount struct {
	Accounts port.AccountRepository
}

func (uc CreateAccount) Execute(ctx context.Context, documentNumber string) (domain.Account, error) {
	if documentNumber == "" {
		return domain.Account{}, ErrInvalidDocument
	}

	acc := domain.Account{DocumentNumber: documentNumber, CreatedAt: time.Now()}
	id, err := uc.Accounts.Create(ctx, acc)
	if err != nil {
		return domain.Account{}, err
	}

	acc.ID = id
	return acc, nil
}
