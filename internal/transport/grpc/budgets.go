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
