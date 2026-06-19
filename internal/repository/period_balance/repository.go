package periodbalancerepo

import (
	"context"
	"final-project/internal/domain"
)

type Repository interface {
	Upsert(ctx context.Context, pb domain.PeriodBalance) (domain.PeriodBalance, error)
	GetByPeriodAndCurrency(ctx context.Context, periodID, currencyID int64) (domain.PeriodBalance, error)
	ListByPeriodID(ctx context.Context, periodID int64) ([]domain.PeriodBalance, error)
}
