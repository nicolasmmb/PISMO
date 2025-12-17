package usecase

import (
	"context"
	"time"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
	"github.com/nicolasmmb/pismo-challenge/internal/port"
)

type CreateTransaction struct {
	Accounts           port.AccountRepository
	OperationTypes     port.OperationTypeRepository
	Transactions       port.TransactionRepository
	TransactionManager port.TransactionManager
}

func (uc CreateTransaction) Execute(ctx context.Context, accountID int64, operationTypeID int, amountCents int64) (domain.Transaction, error) {
	if amountCents <= 0 {
		return domain.Transaction{}, ErrInvalidAmount
	}

	var tx domain.Transaction

	err := uc.TransactionManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if _, err := uc.Accounts.FindByIDForUpdate(txCtx, accountID); err != nil {
			return err
		}

		op, err := uc.OperationTypes.FindByID(txCtx, operationTypeID)
		if err != nil {
			return err
		}

		if op.Sign != -1 && op.Sign != 1 {
			return ErrInvalidOperation
		}

		normalized := amountCents
		if op.Sign < 0 {
			normalized = -amountCents
		}

		now := time.Now()
		tx = domain.Transaction{
			AccountID:       accountID,
			OperationTypeID: operationTypeID,
			AmountCents:     normalized,
			EventDate:       now,
			CreatedAt:       now,
		}

		id, err := uc.Transactions.Create(txCtx, tx)
		if err != nil {
			return err
		}

		tx.ID = id
		return nil
	})

	if err != nil {
		return domain.Transaction{}, err
	}

	return tx, nil
}
