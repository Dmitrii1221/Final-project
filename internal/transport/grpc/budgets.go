package grpc

import (
	"context"
	"final-project/api/proto/budgetpb"
	"final-project/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *BudgetServer) CreateBudget(ctx context.Context, req *budgetpb.CreateBudgetRequest) (*budgetpb.Budget, error) {
	b, err := s.budgetRepo.Insert(ctx, domain.Budget{Name: req.Name})
	if err != nil {
		return nil, status.Error(codes.Internal, "create budget failed")
	}
	return &budgetpb.Budget{
		Id:   b.ID,
		Name: b.Name,
	}, nil
}

func (s *BudgetServer) ListSpendableBudgets(ctx context.Context, req *budgetpb.ListAvailableBudgetsRequest) (*budgetpb.ListAvailableBudgetsResponse, error) {
	spenderRole, err := s.roleRepo.GetByCode(ctx, "spender")
	if err != nil {
		return nil, status.Error(codes.Internal, "spender role not found")
	}

	roles, err := s.userBudgetRoleRepo.GetByUserID(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "list roles failed")
	}

	budgetIDs := make(map[int64]struct{})
	for _, r := range roles {
		if r.RoleID == spenderRole.ID {
			budgetIDs[r.BudgetID] = struct{}{}
		}
	}

	out := make([]*budgetpb.Budget, 0, len(budgetIDs))
	for id := range budgetIDs {
		b, err := s.budgetRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		out = append(out, &budgetpb.Budget{
			Id:   b.ID,
			Name: b.Name,
		})
	}

	return &budgetpb.ListAvailableBudgetsResponse{Budgets: out}, nil
}
