package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config for Postgres connection.
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	MaxConns     int32
	MinConns     int32
	MaxConnIdle  time.Duration
	MaxConnLife  time.Duration
}

// NewPostgresPool creates a connection pool to Postgres.
func NewPostgresPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolCfg.MinConns = cfg.MinConns
	}
	if cfg.MaxConnIdle > 0 {
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdle
	}
	if cfg.MaxConnLife > 0 {
		poolCfg.MaxConnLifetime = cfg.MaxConnLife
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
