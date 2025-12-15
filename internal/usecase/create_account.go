package usecase

import (
	"context"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
	"github.com/nicolasmmb/pismo-challenge/internal/port"
)

// CreateAccount handles creation of accounts.
type CreateAccount struct {
	Accounts port.AccountRepository
	Clock    port.Clock
}

// Execute validates and creates an account.
func (uc CreateAccount) Execute(ctx context.Context, documentNumber string) (domain.Account, error) {
	if documentNumber == "" {
		return domain.Account{}, ErrInvalidDocument
	}

	acc := domain.Account{DocumentNumber: documentNumber, CreatedAt: uc.Clock.Now()}
	id, err := uc.Accounts.Create(ctx, acc)
	if err != nil {
		return domain.Account{}, err
	}
	acc.ID = id
	return acc, nil
}
