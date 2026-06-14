package currencyrepo

import (
	"context"

	"final-project/internal/domain"
)

type Repository interface {
	Insert(ctx context.Context, c domain.Currency) (domain.Currency, error)
	GetByCode(ctx context.Context, code string) (domain.Currency, error)
	List(ctx context.Context) ([]domain.Currency, error)
}
