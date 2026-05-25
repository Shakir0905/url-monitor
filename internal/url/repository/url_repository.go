package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shakir/url-monitor/internal/url/domain"
)

type URLRepository struct {
	pool *pgxpool.Pool
}

func NewURLRepository(pool *pgxpool.Pool) *URLRepository {
	return &URLRepository{pool: pool}
}

func (r *URLRepository) Create(ctx context.Context, userID int64, url string, interval int32) (*domain.URL, error) {
	const query = `
		INSERT INTO urls (user_id, url, check_interval_seconds)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, url, check_interval_seconds, is_active, created_at, updated_at, last_checked_at
	`
	var u domain.URL
	err := r.pool.QueryRow(ctx, query, userID, url, interval).Scan(
		&u.ID, &u.UserID, &u.URL, &u.CheckIntervalSeconds,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastCheckedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert url: %w", err)
	}
	return &u, nil
}

func (r *URLRepository) GetByID(ctx context.Context, id int64) (*domain.URL, error) {
	const query = `
		SELECT id, user_id, url, check_interval_seconds, is_active, created_at, updated_at, last_checked_at
		FROM urls WHERE id = $1
	`
	var u domain.URL
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.UserID, &u.URL, &u.CheckIntervalSeconds,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastCheckedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrURLNotFound
		}
		return nil, fmt.Errorf("select url by id: %w", err)
	}
	return &u, nil
}

func (r *URLRepository) ListByUser(ctx context.Context, userID int64, limit, offset int32) ([]*domain.URL, int32, error) {
	const countQuery = `SELECT COUNT(*) FROM urls WHERE user_id = $1`
	var total int32
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count urls: %w", err)
	}

	if limit <= 0 {
		limit = 100
	}

	const listQuery = `
		SELECT id, user_id, url, check_interval_seconds, is_active, created_at, updated_at, last_checked_at
		FROM urls
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, listQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query urls: %w", err)
	}
	defer rows.Close()

	var urls []*domain.URL
	for rows.Next() {
		var u domain.URL
		if err := rows.Scan(
			&u.ID, &u.UserID, &u.URL, &u.CheckIntervalSeconds,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastCheckedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan url: %w", err)
		}
		urls = append(urls, &u)
	}
	return urls, total, nil
}

func (r *URLRepository) Update(ctx context.Context, id, userID int64, url string, interval int32, isActive bool) (*domain.URL, error) {
	const query = `
		UPDATE urls
		SET url = $1, check_interval_seconds = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4 AND user_id = $5
		RETURNING id, user_id, url, check_interval_seconds, is_active, created_at, updated_at, last_checked_at
	`
	var u domain.URL
	err := r.pool.QueryRow(ctx, query, url, interval, isActive, id, userID).Scan(
		&u.ID, &u.UserID, &u.URL, &u.CheckIntervalSeconds,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastCheckedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrURLNotFound
		}
		return nil, fmt.Errorf("update url: %w", err)
	}
	return &u, nil
}

func (r *URLRepository) Delete(ctx context.Context, id, userID int64) error {
	const query = `DELETE FROM urls WHERE id = $1 AND user_id = $2`
	tag, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("delete url: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrURLNotFound
	}
	return nil
}

func (r *URLRepository) ListActive(ctx context.Context, limit int32) ([]*domain.URL, error) {
	if limit <= 0 {
		limit = 1000
	}
	const query = `
		SELECT id, user_id, url, check_interval_seconds, is_active, created_at, updated_at, last_checked_at
		FROM urls
		WHERE is_active = TRUE
		ORDER BY COALESCE(last_checked_at, '1970-01-01'::timestamptz) ASC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query active urls: %w", err)
	}
	defer rows.Close()

	var urls []*domain.URL
	for rows.Next() {
		var u domain.URL
		if err := rows.Scan(
			&u.ID, &u.UserID, &u.URL, &u.CheckIntervalSeconds,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt, &u.LastCheckedAt,
		); err != nil {
			return nil, fmt.Errorf("scan url: %w", err)
		}
		urls = append(urls, &u)
	}
	return urls, nil
}
