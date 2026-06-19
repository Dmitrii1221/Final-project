package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type PeriodBalance struct {
	PeriodID   int64
	CurrencyID int64
	Remaining  decimal.Decimal
	UpdatedAt  time.Time
}
