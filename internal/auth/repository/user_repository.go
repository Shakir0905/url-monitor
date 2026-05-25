package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shakir/url-monitor/internal/auth/domain"
)

// UserRepository handles user persistence in Postgres.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository constructs a new UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserts a new user and returns the created entity with ID and timestamps populated.
func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at, updated_at
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		// Check for unique constraint violation on email.
		if isUniqueViolation(err) {
			return nil, domain.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &user, nil
}

// GetByEmail returns a user by email or ErrUserNotFound.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("select user by email: %w", err)
	}

	return &user, nil
}

// GetByID returns a user by ID or ErrUserNotFound.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	const query = `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("select user by id: %w", err)
	}

	return &user, nil
}

// isUniqueViolation checks if the error is a Postgres unique constraint violation (code 23505).
func isUniqueViolation(err error) bool {
	type pgError interface {
		SQLState() string
	}
	var pgErr pgError
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}
