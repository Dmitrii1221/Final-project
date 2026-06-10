package budgetrepo

import (
	"context"

	"final-project/internal/domain"
)

type Repository interface {
	Insert(ctx context.Context, name string) (domain.Budget, error)
	list(ctx context.Context) ([]domain.Budget, error)
}
