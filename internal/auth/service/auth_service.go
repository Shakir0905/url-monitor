package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/shakir/url-monitor/internal/auth/domain"
)

// UserRepository defines what AuthService needs from the persistence layer.
// Using an interface here allows mocking for tests and decouples from pgx.
type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

// Config for AuthService.
type Config struct {
	JWTSecret      string
	JWTTTL         time.Duration
	BcryptCost     int // 10-12 is fine, 14+ for paranoid
	MinPasswordLen int
}

// AuthService implements the authentication business logic.
type AuthService struct {
	repo UserRepository
	cfg  Config
}

// NewAuthService constructs a new AuthService.
func NewAuthService(repo UserRepository, cfg Config) *AuthService {
	if cfg.BcryptCost == 0 {
		cfg.BcryptCost = 10
	}
	if cfg.MinPasswordLen == 0 {
		cfg.MinPasswordLen = 8
	}
	if cfg.JWTTTL == 0 {
		cfg.JWTTTL = 24 * time.Hour
	}
	return &AuthService{repo: repo, cfg: cfg}
}

// Register creates a new user with hashed password and returns a JWT token.
func (s *AuthService) Register(ctx context.Context, email, password string) (token string, userID int64, err error) {
	email = normalizeEmail(email)

	if err := validateEmail(email); err != nil {
		return "", 0, err
	}
	if err := s.validatePassword(password); err != nil {
		return "", 0, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.BcryptCost)
	if err != nil {
		return "", 0, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.Create(ctx, email, string(hash))
	if err != nil {
		return "", 0, err
	}

	token, err = s.generateJWT(user.ID)
	if err != nil {
		return "", 0, fmt.Errorf("generate jwt: %w", err)
	}

	return token, user.ID, nil
}

// Login verifies credentials and returns a JWT token.
func (s *AuthService) Login(ctx context.Context, email, password string) (token string, userID int64, err error) {
	email = normalizeEmail(email)

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Don't reveal whether the email exists.
			return "", 0, domain.ErrInvalidPassword
		}
		return "", 0, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", 0, domain.ErrInvalidPassword
	}

	token, err = s.generateJWT(user.ID)
	if err != nil {
		return "", 0, fmt.Errorf("generate jwt: %w", err)
	}

	return token, user.ID, nil
}

// ValidateToken parses and verifies the JWT, returning the user ID if valid.
func (s *AuthService) ValidateToken(ctx context.Context, tokenStr string) (userID int64, err error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, domain.ErrTokenExpired
		}
		return 0, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, domain.ErrInvalidToken
	}

	sub, ok := claims["sub"].(float64) // JSON numbers are float64 in Go
	if !ok {
		return 0, domain.ErrInvalidToken
	}

	return int64(sub), nil
}

// generateJWT creates a signed JWT for the given user ID.
func (s *AuthService) generateJWT(userID int64) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(s.cfg.JWTTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// validatePassword applies basic password policy.
func (s *AuthService) validatePassword(password string) error {
	if len(password) < s.cfg.MinPasswordLen {
		return domain.ErrWeakPassword
	}
	return nil
}

// normalizeEmail lowercases and trims the email.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// validateEmail is a minimal sanity check, not RFC-compliant.
func validateEmail(email string) error {
	if !strings.Contains(email, "@") || len(email) < 5 {
		return domain.ErrInvalidEmail
	}
	return nil
}
