package domain

import "time"

type BudgetPeriod struct {
	ID          int64
	BudgetID    int64
	PeriodStart time.Time
	PeriodEnd   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
