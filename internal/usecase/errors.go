package usecase

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidOperation = errors.New("invalid operation type")
	ErrInvalidAmount    = errors.New("invalid amount")
	ErrDocumentExists   = errors.New("document already exists")
	ErrInvalidDocument  = errors.New("invalid document")
	ErrNotImplemented   = errors.New("not implemented")
)
