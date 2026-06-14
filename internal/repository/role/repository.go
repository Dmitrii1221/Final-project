package rolerepo

import (
	"context"

	"final-project/internal/domain"
)

type Repository interface {
	Insert(ctx context.Context, c domain.Role) (domain.Role, error)
	GetByCode(ctx context.Context, code string) (domain.Role, error)
	List(ctx context.Context) ([]domain.Role, error)
}
