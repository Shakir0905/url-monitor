package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shakir/url-monitor/internal/analytics/domain"
)

type StatsRepository struct {
	pool *pgxpool.Pool
}

func NewStatsRepository(pool *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{pool: pool}
}

func (r *StatsRepository) GetURLStats(ctx context.Context, urlID int64) (*domain.URLStats, error) {
	const query = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE is_up) AS successful,
			COUNT(*) FILTER (WHERE NOT is_up) AS failed,
			COALESCE(AVG(response_time_ms)::float, 0) AS avg_response
		FROM checks
		WHERE url_id = $1
	`
	stats := &domain.URLStats{URLID: urlID}
	err := r.pool.QueryRow(ctx, query, urlID).Scan(
		&stats.TotalChecks,
		&stats.SuccessfulChecks,
		&stats.FailedChecks,
		&stats.AvgResponseTimeMs,
	)
	if err != nil {
		return nil, fmt.Errorf("get url stats: %w", err)
	}

	if stats.TotalChecks > 0 {
		stats.UptimePercentage = float64(stats.SuccessfulChecks) / float64(stats.TotalChecks) * 100
	}

	const lastQuery = `
		SELECT COALESCE(status_code, 0), is_up
		FROM checks
		WHERE url_id = $1
		ORDER BY checked_at DESC
		LIMIT 1
	`
	_ = r.pool.QueryRow(ctx, lastQuery, urlID).Scan(&stats.LastStatusCode, &stats.IsCurrentlyUp)

	return stats, nil
}

func (r *StatsRepository) GetUserDashboard(ctx context.Context, userID int64) (*domain.UserDashboard, error) {
	const query = `
		SELECT
			COUNT(*) AS total_urls,
			COUNT(*) FILTER (WHERE is_active) AS active_urls
		FROM urls
		WHERE user_id = $1
	`
	dash := &domain.UserDashboard{UserID: userID}
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&dash.TotalURLs, &dash.ActiveURLs); err != nil {
		return nil, fmt.Errorf("get user dashboard: %w", err)
	}

	const upDownQuery = `
		SELECT
			COUNT(*) FILTER (WHERE c.is_up) AS up,
			COUNT(*) FILTER (WHERE NOT c.is_up) AS down
		FROM urls u
		JOIN LATERAL (
			SELECT is_up FROM checks
			WHERE url_id = u.id
			ORDER BY checked_at DESC
			LIMIT 1
		) c ON true
		WHERE u.user_id = $1
	`
	_ = r.pool.QueryRow(ctx, upDownQuery, userID).Scan(&dash.URLsCurrentlyUp, &dash.URLsCurrentlyDown)

	return dash, nil
}
