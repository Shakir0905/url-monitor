package domain

import "time"

// URL represents a monitored URL in the system.
type URL struct {
	ID                   int64
	UserID               int64
	URL                  string
	CheckIntervalSeconds int32
	IsActive             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
	LastCheckedAt        *time.Time // pointer because it can be NULL
}
