package usecase

import (
	"context"
	"math"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
	"github.com/nicolasmmb/pismo-challenge/internal/port"
)

// CreateTransaction handles transaction creation.
type CreateTransaction struct {
	Accounts       port.AccountRepository
	OperationTypes port.OperationTypeRepository
	Transactions   port.TransactionRepository
	Clock          port.Clock
}

// Execute validates and creates a transaction.
func (uc CreateTransaction) Execute(ctx context.Context, accountID int64, operationTypeID int, amountCents int64) (domain.Transaction, error) {
	if amountCents <= 0 {
		return domain.Transaction{}, ErrInvalidAmount
	}

	_, err := uc.Accounts.FindByID(ctx, accountID)
	if err != nil {
		return domain.Transaction{}, err
	}

	op, err := uc.OperationTypes.FindByID(ctx, operationTypeID)
	if err != nil {
		return domain.Transaction{}, err
	}
	if op.Sign != -1 && op.Sign != 1 {
		return domain.Transaction{}, ErrInvalidOperation
	}

	normalized := amountCents
	if op.Sign < 0 {
		normalized = -int64(math.Abs(float64(amountCents)))
	} else {
		normalized = int64(math.Abs(float64(amountCents)))
	}

	tx := domain.Transaction{
		AccountID:       accountID,
		OperationTypeID: operationTypeID,
		AmountCents:     normalized,
		EventDate:       uc.Clock.Now(),
		CreatedAt:       uc.Clock.Now(),
	}

	id, err := uc.Transactions.Create(ctx, tx)
	if err != nil {
		return domain.Transaction{}, err
	}

	tx.ID = id
	return tx, nil
}
