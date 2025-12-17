package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
	"github.com/nicolasmmb/pismo-challenge/internal/usecase"
)

// =============================================================================
// MOCKS
// =============================================================================

// mockAccountRepo is a mock for AccountRepository.
type mockAccountRepo struct {
	createFn          func(ctx context.Context, acc domain.Account) (int64, error)
	findByIDFn        func(ctx context.Context, id int64) (domain.Account, error)
	findByIDForUpdate func(ctx context.Context, id int64) (domain.Account, error)
}

func (m *mockAccountRepo) Create(ctx context.Context, acc domain.Account) (int64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, acc)
	}
	return 1, nil
}

func (m *mockAccountRepo) FindByID(ctx context.Context, id int64) (domain.Account, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return domain.Account{ID: id, DocumentNumber: "12345678900"}, nil
}

func (m *mockAccountRepo) FindByIDForUpdate(ctx context.Context, id int64) (domain.Account, error) {
	if m.findByIDForUpdate != nil {
		return m.findByIDForUpdate(ctx, id)
	}
	return domain.Account{ID: id, DocumentNumber: "12345678900"}, nil
}

// mockOperationTypeRepo is a mock for OperationTypeRepository.
type mockOperationTypeRepo struct {
	findByIDFn func(ctx context.Context, id int) (domain.OperationType, error)
}

func (m *mockOperationTypeRepo) FindByID(ctx context.Context, id int) (domain.OperationType, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	// Default: debit operations (1,2,3) have Sign=-1, credit (4) has Sign=+1
	sign := -1
	if id == domain.OperationTypeCreditVoucher {
		sign = 1
	}
	return domain.OperationType{ID: id, Sign: sign}, nil
}

func (m *mockOperationTypeRepo) SeedDefaults(ctx context.Context) error {
	return nil
}

// mockTransactionRepo is a mock for TransactionRepository.
type mockTransactionRepo struct {
	createFn func(ctx context.Context, tx domain.Transaction) (int64, error)
}

func (m *mockTransactionRepo) Create(ctx context.Context, tx domain.Transaction) (int64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, tx)
	}
	return 1, nil
}

// mockTransactionManager is a mock for TransactionManager.
type mockTransactionManager struct {
	runFn func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTransactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.runFn != nil {
		return m.runFn(ctx, fn)
	}
	// Default: just execute the function directly
	return fn(ctx)
}

// =============================================================================
// CreateAccount Tests
// =============================================================================

func TestCreateAccount_Execute(t *testing.T) {
	tests := []struct {
		name           string
		documentNumber string
		setupRepo      func(*mockAccountRepo)
		wantErr        error
		wantID         int64
	}{
		{
			name:           "success - creates account with valid document",
			documentNumber: "12345678900",
			setupRepo: func(m *mockAccountRepo) {
				m.createFn = func(ctx context.Context, acc domain.Account) (int64, error) {
					return 42, nil
				}
			},
			wantErr: nil,
			wantID:  42,
		},
		{
			name:           "error - empty document number",
			documentNumber: "",
			setupRepo:      nil,
			wantErr:        usecase.ErrInvalidDocument,
			wantID:         0,
		},
		{
			name:           "error - repository fails",
			documentNumber: "12345678900",
			setupRepo: func(m *mockAccountRepo) {
				m.createFn = func(ctx context.Context, acc domain.Account) (int64, error) {
					return 0, errors.New("db connection failed")
				}
			},
			wantErr: errors.New("db connection failed"),
			wantID:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountRepo{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			uc := usecase.CreateAccount{
				Accounts: repo,
			}

			acc, err := uc.Execute(context.Background(), tt.documentNumber)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if acc.ID != tt.wantID {
				t.Errorf("expected ID %d, got %d", tt.wantID, acc.ID)
			}

			if acc.DocumentNumber != tt.documentNumber {
				t.Errorf("expected DocumentNumber %s, got %s", tt.documentNumber, acc.DocumentNumber)
			}

			// CreatedAt should be recent (within last second)
			if time.Since(acc.CreatedAt) > time.Second {
				t.Errorf("CreatedAt should be recent, got %v", acc.CreatedAt)
			}
		})
	}
}

// =============================================================================
// GetAccount Tests
// =============================================================================

func TestGetAccount_Execute(t *testing.T) {
	tests := []struct {
		name      string
		accountID int64
		setupRepo func(*mockAccountRepo)
		wantErr   error
		wantAcc   domain.Account
	}{
		{
			name:      "success - finds existing account",
			accountID: 1,
			setupRepo: func(m *mockAccountRepo) {
				m.findByIDFn = func(ctx context.Context, id int64) (domain.Account, error) {
					return domain.Account{
						ID:             1,
						DocumentNumber: "12345678900",
					}, nil
				}
			},
			wantErr: nil,
			wantAcc: domain.Account{ID: 1, DocumentNumber: "12345678900"},
		},
		{
			name:      "error - account not found",
			accountID: 999,
			setupRepo: func(m *mockAccountRepo) {
				m.findByIDFn = func(ctx context.Context, id int64) (domain.Account, error) {
					return domain.Account{}, usecase.ErrNotFound
				}
			},
			wantErr: usecase.ErrNotFound,
			wantAcc: domain.Account{},
		},
		{
			name:      "error - repository fails",
			accountID: 1,
			setupRepo: func(m *mockAccountRepo) {
				m.findByIDFn = func(ctx context.Context, id int64) (domain.Account, error) {
					return domain.Account{}, errors.New("db error")
				}
			},
			wantErr: errors.New("db error"),
			wantAcc: domain.Account{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountRepo{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			uc := usecase.GetAccount{Accounts: repo}

			acc, err := uc.Execute(context.Background(), tt.accountID)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if acc.ID != tt.wantAcc.ID || acc.DocumentNumber != tt.wantAcc.DocumentNumber {
				t.Errorf("expected account %+v, got %+v", tt.wantAcc, acc)
			}
		})
	}
}

// =============================================================================
// CreateTransaction Tests
// =============================================================================

func TestCreateTransaction_Execute(t *testing.T) {
	tests := []struct {
		name            string
		accountID       int64
		operationTypeID int
		amountCents     int64
		setupMocks      func(*mockAccountRepo, *mockOperationTypeRepo, *mockTransactionRepo, *mockTransactionManager)
		wantErr         error
		wantAmount      int64 // Expected normalized amount
	}{
		{
			name:            "success - debit transaction (purchase)",
			accountID:       1,
			operationTypeID: domain.OperationTypeNormalPurchase,
			amountCents:     10000, // R$100.00
			setupMocks:      nil,   // Use defaults
			wantErr:         nil,
			wantAmount:      -10000, // Should be negative for debit
		},
		{
			name:            "success - debit transaction (withdrawal)",
			accountID:       1,
			operationTypeID: domain.OperationTypeWithdrawal,
			amountCents:     5000,
			setupMocks:      nil,
			wantErr:         nil,
			wantAmount:      -5000,
		},
		{
			name:            "success - credit transaction (voucher)",
			accountID:       1,
			operationTypeID: domain.OperationTypeCreditVoucher,
			amountCents:     15000,
			setupMocks:      nil,
			wantErr:         nil,
			wantAmount:      15000, // Should stay positive for credit
		},
		{
			name:            "error - invalid amount (zero)",
			accountID:       1,
			operationTypeID: domain.OperationTypeNormalPurchase,
			amountCents:     0,
			setupMocks:      nil,
			wantErr:         usecase.ErrInvalidAmount,
			wantAmount:      0,
		},
		{
			name:            "error - invalid amount (negative)",
			accountID:       1,
			operationTypeID: domain.OperationTypeNormalPurchase,
			amountCents:     -100,
			setupMocks:      nil,
			wantErr:         usecase.ErrInvalidAmount,
			wantAmount:      0,
		},
		{
			name:            "error - account not found",
			accountID:       999,
			operationTypeID: domain.OperationTypeNormalPurchase,
			amountCents:     10000,
			setupMocks: func(accRepo *mockAccountRepo, opRepo *mockOperationTypeRepo, txRepo *mockTransactionRepo, txMgr *mockTransactionManager) {
				accRepo.findByIDForUpdate = func(ctx context.Context, id int64) (domain.Account, error) {
					return domain.Account{}, usecase.ErrNotFound
				}
			},
			wantErr:    usecase.ErrNotFound,
			wantAmount: 0,
		},
		{
			name:            "error - operation type not found",
			accountID:       1,
			operationTypeID: 999,
			amountCents:     10000,
			setupMocks: func(accRepo *mockAccountRepo, opRepo *mockOperationTypeRepo, txRepo *mockTransactionRepo, txMgr *mockTransactionManager) {
				opRepo.findByIDFn = func(ctx context.Context, id int) (domain.OperationType, error) {
					return domain.OperationType{}, usecase.ErrNotFound
				}
			},
			wantErr:    usecase.ErrNotFound,
			wantAmount: 0,
		},
		{
			name:            "error - invalid operation type sign",
			accountID:       1,
			operationTypeID: 1,
			amountCents:     10000,
			setupMocks: func(accRepo *mockAccountRepo, opRepo *mockOperationTypeRepo, txRepo *mockTransactionRepo, txMgr *mockTransactionManager) {
				opRepo.findByIDFn = func(ctx context.Context, id int) (domain.OperationType, error) {
					return domain.OperationType{ID: 1, Sign: 0}, nil // Invalid sign
				}
			},
			wantErr:    usecase.ErrInvalidOperation,
			wantAmount: 0,
		},
		{
			name:            "error - transaction create fails",
			accountID:       1,
			operationTypeID: domain.OperationTypeNormalPurchase,
			amountCents:     10000,
			setupMocks: func(accRepo *mockAccountRepo, opRepo *mockOperationTypeRepo, txRepo *mockTransactionRepo, txMgr *mockTransactionManager) {
				txRepo.createFn = func(ctx context.Context, tx domain.Transaction) (int64, error) {
					return 0, errors.New("db insert failed")
				}
			},
			wantErr:    errors.New("db insert failed"),
			wantAmount: 0,
		},
		{
			name:            "error - transaction manager fails",
			accountID:       1,
			operationTypeID: domain.OperationTypeNormalPurchase,
			amountCents:     10000,
			setupMocks: func(accRepo *mockAccountRepo, opRepo *mockOperationTypeRepo, txRepo *mockTransactionRepo, txMgr *mockTransactionManager) {
				txMgr.runFn = func(ctx context.Context, fn func(ctx context.Context) error) error {
					return errors.New("transaction failed")
				}
			},
			wantErr:    errors.New("transaction failed"),
			wantAmount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accRepo := &mockAccountRepo{}
			opRepo := &mockOperationTypeRepo{}
			txRepo := &mockTransactionRepo{}
			txMgr := &mockTransactionManager{}

			if tt.setupMocks != nil {
				tt.setupMocks(accRepo, opRepo, txRepo, txMgr)
			}

			uc := usecase.CreateTransaction{
				Accounts:           accRepo,
				OperationTypes:     opRepo,
				Transactions:       txRepo,
				TransactionManager: txMgr,
			}

			tx, err := uc.Execute(context.Background(), tt.accountID, tt.operationTypeID, tt.amountCents)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tx.AmountCents != tt.wantAmount {
				t.Errorf("expected AmountCents %d, got %d", tt.wantAmount, tx.AmountCents)
			}

			if tx.AccountID != tt.accountID {
				t.Errorf("expected AccountID %d, got %d", tt.accountID, tx.AccountID)
			}

			if tx.OperationTypeID != tt.operationTypeID {
				t.Errorf("expected OperationTypeID %d, got %d", tt.operationTypeID, tx.OperationTypeID)
			}

			// EventDate should be recent (within last second)
			if time.Since(tx.EventDate) > time.Second {
				t.Errorf("EventDate should be recent, got %v", tx.EventDate)
			}
		})
	}
}
