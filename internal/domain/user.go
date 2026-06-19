package domain

import "time"

type User struct {
	ID           int64
	ExternalID   string
	Username     string
	PasswordHash string
	Email        string
	CreatedAt    time.Time
}
