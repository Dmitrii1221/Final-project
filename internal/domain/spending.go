package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Spending struct {
	ID             int64
	PeriodID       int64
	BudgetID       int64
	CurrencyID     int64
	Amount         decimal.Decimal
	IdempotencyKey string
	SpentAt        time.Time
	CreatedAt      time.Time
}
