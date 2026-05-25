package domain

import "time"

// User represents a user in the system.
// This is the core business entity, independent of storage or transport.
type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
