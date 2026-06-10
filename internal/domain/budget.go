package domain

import "time"

type Budget struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
