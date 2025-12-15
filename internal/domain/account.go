package domain

import "time"

// Account represents a customer account.
type Account struct {
	ID             int64
	DocumentNumber string
	CreatedAt      time.Time
}
