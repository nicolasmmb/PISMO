package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
	"github.com/nicolasmmb/pismo-challenge/internal/usecase"
)

type AccountHandler struct {
	createUC *usecase.CreateAccount
	getUC    *usecase.GetAccount
}

func NewAccountHandler(createUC *usecase.CreateAccount, getUC *usecase.GetAccount) *AccountHandler {
	return &AccountHandler{
		createUC: createUC,
		getUC:    getUC,
	}
}

type CreateAccountRequest struct {
	DocumentNumber string `json:"document_number"`
}

type AccountResponse struct {
	ID             int64  `json:"account_id"`
	DocumentNumber string `json:"document_number"`
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	output, err := h.createUC.Execute(r.Context(), req.DocumentNumber)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidDocument) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, usecase.ErrDocumentExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(AccountResponse{
		ID:             output.ID,
		DocumentNumber: output.DocumentNumber,
	})
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("accountID")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid account id", http.StatusBadRequest)
		return
	}

	output, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(AccountResponse{
		ID:             output.ID,
		DocumentNumber: output.DocumentNumber,
	})
}
