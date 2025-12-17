package domain

import "errors"

var (
	ErrAccountNotFound       = errors.New("account not found")
	ErrOperationTypeNotFound = errors.New("operation type not found")
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrTransactionNotFound   = errors.New("transaction not found")
	ErrInvalidDocumentNumber = errors.New("invalid document number")
)
