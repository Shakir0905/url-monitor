package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shakir/url-monitor/internal/auth/domain"
)

type mockRepo struct {
	users     map[string]*domain.User
	createErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{users: make(map[string]*domain.User)}
}

func (m *mockRepo) Create(ctx context.Context, email, hash string) (*domain.User, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if _, ok := m.users[email]; ok {
		return nil, domain.ErrUserAlreadyExists
	}
	u := &domain.User{ID: int64(len(m.users) + 1), Email: email, PasswordHash: hash, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	m.users[email] = u
	return u, nil
}

func (m *mockRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func TestRegister_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo, Config{
		JWTSecret:      "test-secret-at-least-32-chars-long-here",
		JWTTTL:         time.Hour,
		BcryptCost:     4,
		MinPasswordLen: 8,
	})

	token, id, err := svc.Register(context.Background(), "Test@Example.com", "password123")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}
}

func TestRegister_NormalizesEmail(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo, Config{
		JWTSecret:      "test-secret-at-least-32-chars-long-here",
		BcryptCost:     4,
		MinPasswordLen: 8,
	})

	_, _, _ = svc.Register(context.Background(), "  USER@Example.com  ", "password123")

	if _, ok := repo.users["user@example.com"]; !ok {
		t.Error("expected email to be normalized to lowercase and trimmed")
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo, Config{
		JWTSecret:      "test-secret-at-least-32-chars-long-here",
		BcryptCost:     4,
		MinPasswordLen: 8,
	})

	_, _, err := svc.Register(context.Background(), "user@example.com", "short")
	if !errors.Is(err, domain.ErrWeakPassword) {
		t.Errorf("expected ErrWeakPassword, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo, Config{
		JWTSecret:      "test-secret-at-least-32-chars-long-here",
		BcryptCost:     4,
		MinPasswordLen: 8,
	})

	_, _, err := svc.Register(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	_, _, err = svc.Login(context.Background(), "user@example.com", "wrong_password")
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo, Config{
		JWTSecret:      "test-secret-at-least-32-chars-long-here",
		BcryptCost:     4,
		MinPasswordLen: 8,
	})

	_, _, err := svc.Login(context.Background(), "nobody@example.com", "password123")
	// Должна вернуться та же ошибка что и при неверном пароле — защита от user enumeration
	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword (not ErrUserNotFound), got %v", err)
	}
}

func TestValidateToken_Roundtrip(t *testing.T) {
	repo := newMockRepo()
	svc := NewAuthService(repo, Config{
		JWTSecret:      "test-secret-at-least-32-chars-long-here",
		JWTTTL:         time.Hour,
		BcryptCost:     4,
		MinPasswordLen: 8,
	})

	token, id, err := svc.Register(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}

	gotID, err := svc.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if gotID != id {
		t.Errorf("expected id %d, got %d", id, gotID)
	}
}
