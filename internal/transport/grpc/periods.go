package grpc

import (
	"context"

	"final-project/api/proto/budgetpb"
	"final-project/internal/domain"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *BudgetServer) CreatePeriod(ctx context.Context, req *budgetpb.CreatePeriodRequest) (*budgetpb.BudgetPeriod, error) {
	p, err := s.periodRepo.Create(ctx, domain.BudgetPeriod{
		BudgetID:    req.BudgetId,
		PeriodStart: req.PeriodStart.AsTime(),
		PeriodEnd:   req.PeriodEnd.AsTime(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "create period failed")
	}
	return &budgetpb.BudgetPeriod{
		Id:          p.ID,
		BudgetId:    p.BudgetID,
		PeriodStart: timestamppb.New(p.PeriodStart),
		PeriodEnd:   timestamppb.New(p.PeriodEnd),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}, nil
}

func (s *BudgetServer) ListPeriods(ctx context.Context, req *budgetpb.ListPeriodsRequest) (*budgetpb.ListPeriodsResponse, error) {
	periods, err := s.periodRepo.ListByBudgetID(ctx, req.BudgetId)
	if err != nil {
		return nil, status.Error(codes.Internal, "list periods failed")
	}
	out := make([]*budgetpb.BudgetPeriod, 0, len(periods))
	for _, p := range periods {
		out = append(out, &budgetpb.BudgetPeriod{
			Id:          p.ID,
			BudgetId:    p.BudgetID,
			PeriodStart: timestamppb.New(p.PeriodStart),
			PeriodEnd:   timestamppb.New(p.PeriodEnd),
			CreatedAt:   timestamppb.New(p.CreatedAt),
			UpdatedAt:   timestamppb.New(p.UpdatedAt),
		})
	}
	return &budgetpb.ListPeriodsResponse{Periods: out}, nil
}

// TODO: реализовать SetPeriodLimit
// 1. Вызвать s.periodLimitRepo.Create с req.PeriodId, req.CurrencyId, req.LimitAmount
// 2. Amount приходит как string — преобразовать через decimal.NewFromString(req.LimitAmount)
// 3. Вернуть *budgetpb.PeriodLimit
func (s *BudgetServer) SetPeriodLimit(ctx context.Context, req *budgetpb.SetPeriodLimitRequest) (*budgetpb.PeriodLimit, error) {

	amount, err := decimal.NewFromString(req.LimitAmount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid limit_amount")
	}

	periodLimit, err := s.periodLimitRepo.Create(ctx, domain.PeriodLimit{
		PeriodID:    req.PeriodId,
		CurrencyID:  req.CurrencyId,
		LimitAmount: amount,
	})
	if err != nil {
		return nil, status.Error(codes.Unimplemented, "create limit failed")
	}
	return &budgetpb.PeriodLimit{
		Id:          periodLimit.ID,
		PeriodId:    periodLimit.PeriodID,
		CurrencyId:  periodLimit.CurrencyID,
		LimitAmount: amount.String(),
	}, nil
}
