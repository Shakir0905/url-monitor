package domain

type URLStats struct {
	URLID             int64
	TotalChecks       int64
	SuccessfulChecks  int64
	FailedChecks      int64
	UptimePercentage  float64
	AvgResponseTimeMs float64
	LastStatusCode    int32
	IsCurrentlyUp     bool
}

type UserDashboard struct {
	UserID            int64
	TotalURLs         int32
	ActiveURLs        int32
	URLsCurrentlyUp   int32
	URLsCurrentlyDown int32
	TopURLs           []*URLStats
}
