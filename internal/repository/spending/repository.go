package spendingrepo

import (
	"context"
	"final-project/internal/domain"
)

type Repository interface {
	Insert(ctx context.Context, s domain.Spending) (domain.Spending, error)
	GetByID(ctx context.Context, id int64) (domain.Spending, error)
	GetByIdempotencyKey(ctx context.Context, budgetID int64, key string) (domain.Spending, error)
}
