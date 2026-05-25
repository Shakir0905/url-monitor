package service

import (
	"context"

	"github.com/shakir/url-monitor/internal/analytics/domain"
)

type StatsRepository interface {
	GetURLStats(ctx context.Context, urlID int64) (*domain.URLStats, error)
	GetUserDashboard(ctx context.Context, userID int64) (*domain.UserDashboard, error)
}

type AnalyticsService struct {
	repo StatsRepository
}

func NewAnalyticsService(repo StatsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) GetURLStats(ctx context.Context, urlID int64) (*domain.URLStats, error) {
	return s.repo.GetURLStats(ctx, urlID)
}

func (s *AnalyticsService) GetUserDashboard(ctx context.Context, userID int64) (*domain.UserDashboard, error) {
	return s.repo.GetUserDashboard(ctx, userID)
}
