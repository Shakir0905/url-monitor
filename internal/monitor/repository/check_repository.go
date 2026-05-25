package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shakir/url-monitor/internal/monitor/domain"
)

type CheckRepository struct {
	pool *pgxpool.Pool
}

func NewCheckRepository(pool *pgxpool.Pool) *CheckRepository {
	return &CheckRepository{pool: pool}
}

func (r *CheckRepository) SaveCheck(ctx context.Context, c *domain.CheckResult) error {
	const query = `
		INSERT INTO checks (url_id, status_code, response_time_ms, is_up, error_message, checked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	var statusCode any = nil
	if c.StatusCode > 0 {
		statusCode = c.StatusCode
	}
	var errMsg any = nil
	if c.ErrorMessage != "" {
		errMsg = c.ErrorMessage
	}

	_, err := r.pool.Exec(ctx, query, c.URLID, statusCode, c.ResponseTimeMs, c.IsUp, errMsg, c.CheckedAt)
	if err != nil {
		return fmt.Errorf("insert check: %w", err)
	}
	return nil
}

func (r *CheckRepository) UpdateLastChecked(ctx context.Context, urlID int64, checkedAt any) error {
	const query = `UPDATE urls SET last_checked_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, checkedAt, urlID)
	if err != nil {
		return fmt.Errorf("update last_checked_at: %w", err)
	}
	return nil
}

// GetLastStatus returns the most recent is_up value for the URL, or true if no checks yet.
// Used to detect status changes (up -> down or down -> up).
func (r *CheckRepository) GetLastStatus(ctx context.Context, urlID int64) (bool, bool, error) {
	const query = `
		SELECT is_up FROM checks
		WHERE url_id = $1
		ORDER BY checked_at DESC
		LIMIT 1 OFFSET 1
	`
	var isUp bool
	err := r.pool.QueryRow(ctx, query, urlID).Scan(&isUp)
	if err != nil {
		// No previous check exists.
		return false, false, nil
	}
	return isUp, true, nil
}
