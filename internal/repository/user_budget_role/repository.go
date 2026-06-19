package userbudgetrolerepo

import (
	"context"
	"final-project/internal/domain"
)

type Repository interface {
	Grant(ctx context.Context, ubr domain.UserBudgetRole) error
	Revoke(ctx context.Context, userID, budgetID, roleID int64) error
	GetByUserAndBudget(ctx context.Context, userID, budgetID int64) ([]domain.UserBudgetRole, error)
}
