package domain

import "github.com/shopspring/decimal"

type PeriodLimit struct {
	ID          int64
	PeriodID    int64
	CurrencyID  int64
	LimitAmount decimal.Decimal
}
