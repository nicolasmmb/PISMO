package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
	"github.com/nicolasmmb/pismo-challenge/internal/usecase"
)

type TransactionHandler struct {
	createUC *usecase.CreateTransaction
}

func NewTransactionHandler(createUC *usecase.CreateTransaction) *TransactionHandler {
	return &TransactionHandler{
		createUC: createUC,
	}
}

type CreateTransactionRequest struct {
	AccountID       int64   `json:"account_id"`
	OperationTypeID int     `json:"operation_type_id"`
	Amount          float64 `json:"amount"`
}

type TransactionResponse struct {
	ID int64 `json:"transaction_id"`
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	amountCents := int64(req.Amount * 100)

	output, err := h.createUC.Execute(r.Context(), req.AccountID, req.OperationTypeID, amountCents)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, domain.ErrAccountNotFound), errors.Is(err, domain.ErrOperationTypeNotFound):
			status = http.StatusNotFound
		case errors.Is(err, usecase.ErrInvalidAmount), errors.Is(err, usecase.ErrInvalidOperation), errors.Is(err, domain.ErrInsufficientFunds):
			status = http.StatusBadRequest
		}

		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TransactionResponse{ID: output.ID})
}
