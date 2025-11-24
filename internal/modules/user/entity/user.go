package entity

import "time"

// User represents the core domain model for the user module.
type User struct {
	ID        int64
	Name      string
	Email     string
	PasswordHash string
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
