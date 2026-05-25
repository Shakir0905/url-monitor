package service

import (
	"context"
	"net/url"
	"strings"

	"github.com/shakir/url-monitor/internal/url/domain"
)

type URLRepository interface {
	Create(ctx context.Context, userID int64, urlStr string, interval int32) (*domain.URL, error)
	GetByID(ctx context.Context, id int64) (*domain.URL, error)
	ListByUser(ctx context.Context, userID int64, limit, offset int32) ([]*domain.URL, int32, error)
	Update(ctx context.Context, id, userID int64, urlStr string, interval int32, isActive bool) (*domain.URL, error)
	Delete(ctx context.Context, id, userID int64) error
	ListActive(ctx context.Context, limit int32) ([]*domain.URL, error)
}

type URLService struct {
	repo URLRepository
}

func NewURLService(repo URLRepository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) Create(ctx context.Context, userID int64, urlStr string, interval int32) (*domain.URL, error) {
	urlStr = strings.TrimSpace(urlStr)
	if err := validateURL(urlStr); err != nil {
		return nil, err
	}
	if interval < 10 || interval > 86400 {
		return nil, domain.ErrInvalidInterval
	}
	return s.repo.Create(ctx, userID, urlStr, interval)
}

func (s *URLService) Get(ctx context.Context, id, userID int64) (*domain.URL, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u.UserID != userID {
		return nil, domain.ErrPermissionDenied
	}
	return u, nil
}

func (s *URLService) List(ctx context.Context, userID int64, limit, offset int32) ([]*domain.URL, int32, error) {
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

func (s *URLService) Update(ctx context.Context, id, userID int64, urlStr string, interval int32, isActive bool) (*domain.URL, error) {
	urlStr = strings.TrimSpace(urlStr)
	if err := validateURL(urlStr); err != nil {
		return nil, err
	}
	if interval < 10 || interval > 86400 {
		return nil, domain.ErrInvalidInterval
	}
	return s.repo.Update(ctx, id, userID, urlStr, interval, isActive)
}

func (s *URLService) Delete(ctx context.Context, id, userID int64) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *URLService) ListActive(ctx context.Context, limit int32) ([]*domain.URL, error) {
	return s.repo.ListActive(ctx, limit)
}

func validateURL(urlStr string) error {
	if urlStr == "" {
		return domain.ErrInvalidURL
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return domain.ErrInvalidURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return domain.ErrInvalidURL
	}
	if u.Host == "" {
		return domain.ErrInvalidURL
	}
	return nil
}
