package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
)

const (
	accountInsertSQL      = `INSERT INTO accounts (document_number, created_at) VALUES ($1, $2) RETURNING id`
	accountSelectSQL      = `SELECT id, document_number, created_at FROM accounts WHERE id = $1`
	accountSelectForUpSQL = `SELECT id, document_number, created_at FROM accounts WHERE id = $1 FOR UPDATE`
)

type AccountRepository struct {
	tm *TransactionManagerDB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{
		tm: NewTransactionManager(db),
	}
}

func (r *AccountRepository) Create(ctx context.Context, account domain.Account) (int64, error) {
	createdAt := account.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	var id int64
	err := r.tm.GetExecutor(ctx).QueryRowContext(ctx, accountInsertSQL, account.DocumentNumber, createdAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create account: %w", err)
	}
	return id, nil
}

func (r *AccountRepository) FindByID(ctx context.Context, id int64) (domain.Account, error) {
	return r.findOne(ctx, accountSelectSQL, id)
}

func (r *AccountRepository) FindByIDForUpdate(ctx context.Context, id int64) (domain.Account, error) {
	return r.findOne(ctx, accountSelectForUpSQL, id)
}

func (r *AccountRepository) findOne(ctx context.Context, query string, id int64) (domain.Account, error) {
	var acc domain.Account
	err := r.tm.GetExecutor(ctx).QueryRowContext(ctx, query, id).Scan(&acc.ID, &acc.DocumentNumber, &acc.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Account{}, domain.ErrAccountNotFound
		}
		return domain.Account{}, fmt.Errorf("failed to find account: %w", err)
	}
	return acc, nil
}
