package budgetrepo

import (
	"context"

	"final-project/internal/domain"
)

type Repository interface {
	Insert(ctx context.Context, budget domain.Budget) (domain.Budget, error)
	GetByID(ctx context.Context, id int64) (domain.Budget, error)
	List(ctx context.Context) ([]domain.Budget, error)
}
