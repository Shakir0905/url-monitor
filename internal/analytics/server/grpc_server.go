package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/shakir/url-monitor/internal/analytics/domain"
	analyticspb "github.com/shakir/url-monitor/proto/analytics"
)

type AnalyticsService interface {
	GetURLStats(ctx context.Context, urlID int64) (*domain.URLStats, error)
	GetUserDashboard(ctx context.Context, userID int64) (*domain.UserDashboard, error)
}

type GRPCServer struct {
	analyticspb.UnimplementedAnalyticsServiceServer
	svc AnalyticsService
}

func NewGRPCServer(svc AnalyticsService) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) GetURLStats(ctx context.Context, req *analyticspb.GetURLStatsRequest) (*analyticspb.URLStats, error) {
	stats, err := s.svc.GetURLStats(ctx, req.GetUrlId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &analyticspb.URLStats{
		UrlId:             stats.URLID,
		TotalChecks:       stats.TotalChecks,
		SuccessfulChecks:  stats.SuccessfulChecks,
		FailedChecks:      stats.FailedChecks,
		UptimePercentage:  stats.UptimePercentage,
		AvgResponseTimeMs: stats.AvgResponseTimeMs,
		LastStatusCode:    stats.LastStatusCode,
		IsCurrentlyUp:     stats.IsCurrentlyUp,
	}, nil
}

func (s *GRPCServer) GetUserDashboard(ctx context.Context, req *analyticspb.GetUserDashboardRequest) (*analyticspb.UserDashboard, error) {
	dash, err := s.svc.GetUserDashboard(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &analyticspb.UserDashboard{
		UserId:            dash.UserID,
		TotalUrls:         dash.TotalURLs,
		ActiveUrls:        dash.ActiveURLs,
		UrlsCurrentlyUp:   dash.URLsCurrentlyUp,
		UrlsCurrentlyDown: dash.URLsCurrentlyDown,
	}, nil
}
