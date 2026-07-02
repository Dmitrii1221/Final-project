package grpc

import (
	"context"

	"final-project/api/proto/budgetpb"
	"final-project/internal/domain"
)

type (
	budgetRepo interface {
		Insert(ctx context.Context, budget domain.Budget) (domain.Budget, error)
		GetByID(ctx context.Context, id int64) (domain.Budget, error)
	}

	periodRepo interface {
		Create(ctx context.Context, p domain.BudgetPeriod) (domain.BudgetPeriod, error)
		ListByBudgetID(ctx context.Context, budgetID int64) ([]domain.BudgetPeriod, error)
	}

	periodLimitRepo interface {
		Create(ctx context.Context, pl domain.PeriodLimit) (domain.PeriodLimit, error)
	}

	roleRepo interface {
		GetByCode(ctx context.Context, code string) (domain.Role, error)
	}

	userBudgetRoleRepo interface {
		Grant(ctx context.Context, ubr domain.UserBudgetRole) error
		GetByUserID(ctx context.Context, userID int64) ([]domain.UserBudgetRole, error)
	}
)

type BudgetServer struct {
	budgetpb.UnimplementedBudgetServiceServer
	budgetRepo          budgetRepo
	periodRepo          periodRepo
	periodLimitRepo     periodLimitRepo
	roleRepo            roleRepo
	userBudgetRoleRepo  userBudgetRoleRepo
}

func NewBudgetServer(br budgetRepo, pr periodRepo, plr periodLimitRepo, rr roleRepo, ubrr userBudgetRoleRepo) *BudgetServer {
	return &BudgetServer{
		budgetRepo:         br,
		periodRepo:         pr,
		periodLimitRepo:    plr,
		roleRepo:           rr,
		userBudgetRoleRepo: ubrr,
	}
}
