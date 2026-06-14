package userrepo

import (
	"context"

	"final-project/internal/domain"
)

type Repository interface {
	Insert(ctx context.Context, c domain.User) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
}
