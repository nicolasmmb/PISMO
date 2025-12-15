package port

import (
	"context"

	"github.com/nicolasmmb/pismo-challenge/internal/domain"
)

// OperationTypeRepository defines access to operation types.
type OperationTypeRepository interface {
	FindByID(ctx context.Context, id int) (domain.OperationType, error)
	SeedDefaults(ctx context.Context) error
}
