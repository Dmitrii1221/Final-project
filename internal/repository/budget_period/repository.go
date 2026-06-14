package budgetperiodrepo

import (
	"context"
	"final-project/internal/domain"
)

// repository.go
type Repository interface {
	Create(ctx context.Context, p domain.BudgetPeriod) (domain.BudgetPeriod, error)
	GetByID(ctx context.Context, id int64) (domain.BudgetPeriod, error)
	ListByBudgetID(ctx context.Context, budgetID int64) ([]domain.BudgetPeriod, error)
}
