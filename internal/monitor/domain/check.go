package domain

import "time"

type CheckResult struct {
	URLID          int64
	UserID         int64
	URL            string
	StatusCode     int
	ResponseTimeMs int
	IsUp           bool
	ErrorMessage   string
	CheckedAt      time.Time
}
