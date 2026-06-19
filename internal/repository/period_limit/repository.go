package periodlimitrepo

import (
	"context"
	"final-project/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, pl domain.PeriodLimit) (domain.PeriodLimit, error)
	GetByPeriodAndCurrency(ctx context.Context, periodID, currencyID int64) (domain.PeriodLimit, error)
	ListByPeriodID(ctx context.Context, periodID int64) ([]domain.PeriodLimit, error)
}
