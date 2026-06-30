package kafka

import (
	"context"
	"encoding/json"
	"final-project/internal/domain"
	budgetperiodrepo "final-project/internal/repository/budget_period"
	currencyrepo "final-project/internal/repository/currency"
	periodbalancerepo "final-project/internal/repository/period_balance"
	periodlimitrepo "final-project/internal/repository/period_limit"
	rolerepo "final-project/internal/repository/role"
	spendingrepo "final-project/internal/repository/spending"
	userbudgetrolerepo "final-project/internal/repository/user_budget_role"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/shopspring/decimal"
)

type SpendEvent struct {
	IdempotencyKey string `json:"idempotency_key"`
	BudgetID       int64  `json:"budget_id"`
	CurrencyCode   string `json:"currency_code"`
	Amount         string `json:"amount"`
	SpentAt        string `json:"spent_at"`
	UserID         int64  `json:"user_id"`
	Source         string `json:"source"`
}

type Handler struct {
	currencyRepo       currencyrepo.Repository
	periodRepo         budgetperiodrepo.Repository
	periodLimitRepo    periodlimitrepo.Repository
	spendingRepo       spendingrepo.Repository
	periodBalanceRepo  periodbalancerepo.Repository
	userBudgetRoleRepo userbudgetrolerepo.Repository
	roleRepo           rolerepo.Repository
	dlqWriter          *kafka.Writer
}

func NewHandler(
	currencyRepo currencyrepo.Repository,
	periodRepo budgetperiodrepo.Repository,
	periodLimitRepo periodlimitrepo.Repository,
	spendingRepo spendingrepo.Repository,
	periodBalanceRepo periodbalancerepo.Repository,
	userBudgetRoleRepo userbudgetrolerepo.Repository,
	roleRepo rolerepo.Repository,
) *Handler {
	return &Handler{
		currencyRepo:       currencyRepo,
		periodRepo:         periodRepo,
		periodLimitRepo:    periodLimitRepo,
		spendingRepo:       spendingRepo,
		periodBalanceRepo:  periodBalanceRepo,
		userBudgetRoleRepo: userBudgetRoleRepo,
		roleRepo:           roleRepo,
	}
}

func (h *Handler) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	var event SpendEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	if event.BudgetID == 0 || event.CurrencyCode == "" || event.Amount == "" || event.SpentAt == "" || event.UserID == 0 {
		return fmt.Errorf("missing required fields")
	}

	currency, err := h.currencyRepo.GetByCode(ctx, event.CurrencyCode)
	if err != nil {
		return fmt.Errorf("unknown currency %s: %w", event.CurrencyCode, err)
	}

	spentAt, err := time.Parse(time.RFC3339, event.SpentAt)
	if err != nil {
		return fmt.Errorf("invalid spent_at: %w", err)
	}

	periods, err := h.periodRepo.ListByBudgetID(ctx, event.BudgetID)
	if err != nil {
		return fmt.Errorf("list periods: %w", err)
	}

	var activePeriod *domain.BudgetPeriod
	for i, p := range periods {
		if (p.PeriodStart.Equal(spentAt) || p.PeriodStart.Before(spentAt)) && spentAt.Before(p.PeriodEnd) {
			activePeriod = &periods[i]
			break
		}
	}
	if activePeriod == nil {
		return fmt.Errorf("no active period for budget %d at %s", event.BudgetID, event.SpentAt)
	}

	periodLimit, err := h.periodLimitRepo.GetByPeriodAndCurrency(ctx, activePeriod.ID, currency.ID)
	if err != nil {
		return fmt.Errorf("currency %s not allowed in period %d", event.CurrencyCode, activePeriod.ID)
	}

	role, err := h.roleRepo.GetByCode(ctx, "spender")
	if err != nil {
		return fmt.Errorf("spender role not found: %w", err)
	}
	roles, err := h.userBudgetRoleRepo.GetByUserAndBudget(ctx, event.UserID, event.BudgetID)
	if err != nil {
		return fmt.Errorf("check roles: %w", err)
	}
	hasRole := false
	for _, r := range roles {
		if r.RoleID == role.ID {
			hasRole = true
			break
		}
	}
	if !hasRole {
		return fmt.Errorf("user %d is not spender on budget %d", event.UserID, event.BudgetID)
	}

	if event.IdempotencyKey != "" {
		_, err := h.spendingRepo.GetByIdempotencyKey(ctx, event.BudgetID, event.IdempotencyKey)
		if err == nil {
			slog.Warn("duplicate spending", "key", event.IdempotencyKey)
			return nil
		}
	}

	amount, err := decimal.NewFromString(event.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount %s: %w", event.Amount, err)
	}
	spending, err := h.spendingRepo.Insert(ctx, domain.Spending{
		PeriodID:       activePeriod.ID,
		BudgetID:       event.BudgetID,
		CurrencyID:     currency.ID,
		Amount:         amount,
		IdempotencyKey: event.IdempotencyKey,
		SpentAt:        spentAt,
	})
	if err != nil {
		return fmt.Errorf("insert spending: %w", err)
	}

	_, err = h.periodBalanceRepo.Upsert(ctx, domain.PeriodBalance{
		PeriodID:   activePeriod.ID,
		CurrencyID: currency.ID,
		Remaining:  periodLimit.LimitAmount.Sub(amount),
	})
	if err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	slog.Info("spending processed",
		"id", spending.ID,
		"budget", event.BudgetID,
		"amount", amount,
	)
	return nil
}

func (h *Handler) sendToDLQ(ctx context.Context, original kafka.Message, reason string) error {
	dlqMsg := struct {
		Original  json.RawMessage `json:"original"`
		Reason    string          `json:"error_reason"`
		Error     string          `json:"error_message"`
		FailedAt  string          `json:"failed_at"`
		Instance  string          `json:"consumer_instance"`
	}{
		Original: original.Value,
		Reason:   reason,
		Error:    reason,
		FailedAt: time.Now().UTC().Format(time.RFC3339),
		Instance: "budget-consumer-1",
	}

	body, err := json.Marshal(dlqMsg)
	if err != nil {
		return fmt.Errorf("marshal dlq: %w", err)
	}

	return h.dlqWriter.WriteMessages(ctx, kafka.Message{
		Key:   original.Key,
		Value: body,
	})
}
