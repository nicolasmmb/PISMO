package domain

import "time"

type Transaction struct {
	ID              int64
	AccountID       int64
	OperationTypeID int
	AmountCents     int64
	EventDate       time.Time
	CreatedAt       time.Time
}
