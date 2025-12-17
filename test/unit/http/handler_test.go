package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	adapterhttp "github.com/nicolasmmb/pismo-challenge/internal/adapter/http"
	"github.com/nicolasmmb/pismo-challenge/internal/domain"
	"github.com/nicolasmmb/pismo-challenge/internal/usecase"
)

// FakeAccountRepo implements port.AccountRepository
type FakeAccountRepo struct {
	accounts map[int64]domain.Account
	nextID   int64
}

func NewFakeAccountRepo() *FakeAccountRepo {
	return &FakeAccountRepo{
		accounts: make(map[int64]domain.Account),
		nextID:   1,
	}
}

func (r *FakeAccountRepo) Create(ctx context.Context, account domain.Account) (int64, error) {
	id := r.nextID
	r.nextID++
	account.ID = id
	r.accounts[id] = account
	return id, nil
}

func (r *FakeAccountRepo) FindByID(ctx context.Context, id int64) (domain.Account, error) {
	acc, ok := r.accounts[id]
	if !ok {
		return domain.Account{}, domain.ErrAccountNotFound
	}
	return acc, nil
}

func (r *FakeAccountRepo) FindByIDForUpdate(ctx context.Context, id int64) (domain.Account, error) {
	return r.FindByID(ctx, id)
}

// AccountResponse mirrors the handler response for testing
type AccountResponse struct {
	ID             int64  `json:"account_id"`
	DocumentNumber string `json:"document_number"`
}

func TestCreateAccount(t *testing.T) {
	repo := NewFakeAccountRepo()
	createUC := &usecase.CreateAccount{Accounts: repo}
	getUC := &usecase.GetAccount{Accounts: repo}
	handler := adapterhttp.NewAccountHandler(createUC, getUC)

	t.Run("success", func(t *testing.T) {
		reqBody := `{"document_number": "12345678900"}`
		req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		handler.CreateAccount(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("expected status 201, got %d", resp.StatusCode)
		}

		var body AccountResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.ID == 0 {
			t.Error("expected non-zero account ID")
		}
		if body.DocumentNumber != "12345678900" {
			t.Errorf("expected document number 12345678900, got %s", body.DocumentNumber)
		}
	})

	t.Run("invalid document", func(t *testing.T) {
		reqBody := `{"document_number": ""}`
		req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBufferString(reqBody))
		w := httptest.NewRecorder()

		handler.CreateAccount(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})
}

func TestGetAccount(t *testing.T) {
	repo := NewFakeAccountRepo()
	repo.accounts[1] = domain.Account{ID: 1, DocumentNumber: "123", CreatedAt: time.Now()}

	createUC := &usecase.CreateAccount{Accounts: repo}
	getUC := &usecase.GetAccount{Accounts: repo}
	handler := adapterhttp.NewAccountHandler(createUC, getUC)

	t.Run("found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/accounts/1", nil)
		req.SetPathValue("accountID", "1") // Only works in Go 1.22+

		w := httptest.NewRecorder()

		handler.GetAccount(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/accounts/999", nil)
		req.SetPathValue("accountID", "999")

		w := httptest.NewRecorder()

		handler.GetAccount(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})
}
