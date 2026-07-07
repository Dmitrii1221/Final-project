package grpc

import (
	"context"

	"final-project/api/proto/budgetpb"
	"final-project/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: реализовать GrantRole
// 1. Получить роль по role_code через s.roleRepo.GetByCode
// 2. Вызвать s.userBudgetRoleRepo.Grant
// 3. Вернуть GrantRoleResponse{Success: true}
func (s *BudgetServer) GrantRole(ctx context.Context, req *budgetpb.GrantRoleRequest) (*budgetpb.GrantRoleResponse, error) {
	role, err := s.roleRepo.GetByCode(ctx, req.RoleCode)
	if err != nil {
		return nil, status.Error(codes.NotFound, "role not found")
	}

	err = s.userBudgetRoleRepo.Grant(ctx, domain.UserBudgetRole{
		UserID:   req.UserId,
		BudgetID: req.BudgetId,
		RoleID:   role.ID,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "grant role failed")
	}

	return &budgetpb.GrantRoleResponse{Success: true}, nil
}
