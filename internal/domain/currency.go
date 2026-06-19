package domain

import "time"

type Currency struct {
	ID          int64
	Code        string
	DisplayName string
	Precision   int
	CreatedAt   time.Time
}
