package domain

import "time"

// Transaction represents a financial event tied to an account.
type Transaction struct {
	ID              int64
	AccountID       int64
	OperationTypeID int
	AmountCents     int64
	EventDate       time.Time
	CreatedAt       time.Time
}
