package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"final-project/internal/domain"

	"github.com/shopspring/decimal"
)

type SpendCommand struct {
	BudgetID       int64
	CurrencyCode   string
	Amount         decimal.Decimal
	SpentAt        time.Time
	UserID         int64
	IdempotencyKey string
	Source         string
}

type SpendingService struct {
	currencyRepo       interface{ GetByCode(ctx context.Context, code string) (domain.Currency, error) }
	periodRepo         interface{ ListByBudgetID(ctx context.Context, budgetID int64) ([]domain.BudgetPeriod, error) }
	periodLimitRepo    interface{ GetByPeriodAndCurrency(ctx context.Context, periodID, currencyID int64) (domain.PeriodLimit, error) }
	spendingRepo       interface {
		Insert(ctx context.Context, s domain.Spending) (domain.Spending, error)
		GetByIdempotencyKey(ctx context.Context, budgetID int64, key string) (domain.Spending, error)
	}
	periodBalanceRepo  interface{ Upsert(ctx context.Context, pb domain.PeriodBalance) (domain.PeriodBalance, error) }
	userBudgetRoleRepo interface{ GetByUserAndBudget(ctx context.Context, userID, budgetID int64) ([]domain.UserBudgetRole, error) }
	roleRepo           interface{ GetByCode(ctx context.Context, code string) (domain.Role, error) }
}

func NewSpendingService(
	currencyRepo interface{ GetByCode(ctx context.Context, code string) (domain.Currency, error) },
	periodRepo interface{ ListByBudgetID(ctx context.Context, budgetID int64) ([]domain.BudgetPeriod, error) },
	periodLimitRepo interface{ GetByPeriodAndCurrency(ctx context.Context, periodID, currencyID int64) (domain.PeriodLimit, error) },
	spendingRepo interface {
		Insert(ctx context.Context, s domain.Spending) (domain.Spending, error)
		GetByIdempotencyKey(ctx context.Context, budgetID int64, key string) (domain.Spending, error)
	},
	periodBalanceRepo interface{ Upsert(ctx context.Context, pb domain.PeriodBalance) (domain.PeriodBalance, error) },
	userBudgetRoleRepo interface{ GetByUserAndBudget(ctx context.Context, userID, budgetID int64) ([]domain.UserBudgetRole, error) },
	roleRepo interface{ GetByCode(ctx context.Context, code string) (domain.Role, error) },
) *SpendingService {
	return &SpendingService{
		currencyRepo:       currencyRepo,
		periodRepo:         periodRepo,
		periodLimitRepo:    periodLimitRepo,
		spendingRepo:       spendingRepo,
		periodBalanceRepo:  periodBalanceRepo,
		userBudgetRoleRepo: userBudgetRoleRepo,
		roleRepo:           roleRepo,
	}
}

func (s *SpendingService) Process(ctx context.Context, cmd SpendCommand) (domain.Spending, error) {
	currency, err := s.currencyRepo.GetByCode(ctx, cmd.CurrencyCode)
	if err != nil {
		return domain.Spending{}, fmt.Errorf("unknown currency %s: %w", cmd.CurrencyCode, err)
	}

	periods, err := s.periodRepo.ListByBudgetID(ctx, cmd.BudgetID)
	if err != nil {
		return domain.Spending{}, fmt.Errorf("list periods: %w", err)
	}

	var activePeriod *domain.BudgetPeriod
	for i, p := range periods {
		if (p.PeriodStart.Equal(cmd.SpentAt) || p.PeriodStart.Before(cmd.SpentAt)) && cmd.SpentAt.Before(p.PeriodEnd) {
			activePeriod = &periods[i]
			break
		}
	}
	if activePeriod == nil {
		return domain.Spending{}, fmt.Errorf("no active period for budget %d at %s", cmd.BudgetID, cmd.SpentAt)
	}

	_, err = s.periodLimitRepo.GetByPeriodAndCurrency(ctx, activePeriod.ID, currency.ID)
	if err != nil {
		return domain.Spending{}, fmt.Errorf("currency %s not allowed in period %d", cmd.CurrencyCode, activePeriod.ID)
	}

	role, err := s.roleRepo.GetByCode(ctx, "spender")
	if err != nil {
		return domain.Spending{}, fmt.Errorf("spender role not found: %w", err)
	}
	roles, err := s.userBudgetRoleRepo.GetByUserAndBudget(ctx, cmd.UserID, cmd.BudgetID)
	if err != nil {
		return domain.Spending{}, fmt.Errorf("check roles: %w", err)
	}
	hasRole := false
	for _, r := range roles {
		if r.RoleID == role.ID {
			hasRole = true
			break
		}
	}
	if !hasRole {
		return domain.Spending{}, fmt.Errorf("user %d is not spender on budget %d", cmd.UserID, cmd.BudgetID)
	}

	if cmd.IdempotencyKey != "" {
		_, err := s.spendingRepo.GetByIdempotencyKey(ctx, cmd.BudgetID, cmd.IdempotencyKey)
		if err == nil {
			slog.Warn("duplicate spending", "key", cmd.IdempotencyKey)
			return domain.Spending{}, nil
		}
	}

	spending, err := s.spendingRepo.Insert(ctx, domain.Spending{
		PeriodID:       activePeriod.ID,
		BudgetID:       cmd.BudgetID,
		CurrencyID:     currency.ID,
		Amount:         cmd.Amount,
		IdempotencyKey: cmd.IdempotencyKey,
		SpentAt:        cmd.SpentAt,
	})
	if err != nil {
		return domain.Spending{}, fmt.Errorf("insert spending: %w", err)
	}

	_, err = s.periodBalanceRepo.Upsert(ctx, domain.PeriodBalance{
		PeriodID:   activePeriod.ID,
		CurrencyID: currency.ID,
		Remaining:  cmd.Amount,
	})
	if err != nil {
		return domain.Spending{}, fmt.Errorf("update balance: %w", err)
	}

	slog.Info("spending processed",
		"id", spending.ID,
		"budget", cmd.BudgetID,
		"amount", cmd.Amount,
	)
	return spending, nil
}
