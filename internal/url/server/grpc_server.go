package server

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/shakir/url-monitor/internal/url/domain"
	urlpb "github.com/shakir/url-monitor/proto/url"
)

type URLService interface {
	Create(ctx context.Context, userID int64, urlStr string, interval int32) (*domain.URL, error)
	Get(ctx context.Context, id, userID int64) (*domain.URL, error)
	List(ctx context.Context, userID int64, limit, offset int32) ([]*domain.URL, int32, error)
	Update(ctx context.Context, id, userID int64, urlStr string, interval int32, isActive bool) (*domain.URL, error)
	Delete(ctx context.Context, id, userID int64) error
	ListActive(ctx context.Context, limit int32) ([]*domain.URL, error)
}

type GRPCServer struct {
	urlpb.UnimplementedURLServiceServer
	svc URLService
}

func NewGRPCServer(svc URLService) *GRPCServer {
	return &GRPCServer{svc: svc}
}

func (s *GRPCServer) CreateURL(ctx context.Context, req *urlpb.CreateURLRequest) (*urlpb.URL, error) {
	u, err := s.svc.Create(ctx, req.GetUserId(), req.GetUrl(), req.GetCheckIntervalSeconds())
	if err != nil {
		return nil, mapError(err)
	}
	return toProto(u), nil
}

func (s *GRPCServer) GetURL(ctx context.Context, req *urlpb.GetURLRequest) (*urlpb.URL, error) {
	u, err := s.svc.Get(ctx, req.GetId(), req.GetUserId())
	if err != nil {
		return nil, mapError(err)
	}
	return toProto(u), nil
}

func (s *GRPCServer) ListURLs(ctx context.Context, req *urlpb.ListURLsRequest) (*urlpb.ListURLsResponse, error) {
	urls, total, err := s.svc.List(ctx, req.GetUserId(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]*urlpb.URL, 0, len(urls))
	for _, u := range urls {
		out = append(out, toProto(u))
	}
	return &urlpb.ListURLsResponse{Urls: out, Total: total}, nil
}

func (s *GRPCServer) UpdateURL(ctx context.Context, req *urlpb.UpdateURLRequest) (*urlpb.URL, error) {
	u, err := s.svc.Update(ctx, req.GetId(), req.GetUserId(), req.GetUrl(), req.GetCheckIntervalSeconds(), req.GetIsActive())
	if err != nil {
		return nil, mapError(err)
	}
	return toProto(u), nil
}

func (s *GRPCServer) DeleteURL(ctx context.Context, req *urlpb.DeleteURLRequest) (*urlpb.DeleteURLResponse, error) {
	if err := s.svc.Delete(ctx, req.GetId(), req.GetUserId()); err != nil {
		return nil, mapError(err)
	}
	return &urlpb.DeleteURLResponse{Success: true}, nil
}

func (s *GRPCServer) ListActiveURLs(ctx context.Context, req *urlpb.ListActiveURLsRequest) (*urlpb.ListURLsResponse, error) {
	urls, err := s.svc.ListActive(ctx, req.GetLimit())
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]*urlpb.URL, 0, len(urls))
	for _, u := range urls {
		out = append(out, toProto(u))
	}
	return &urlpb.ListURLsResponse{Urls: out, Total: int32(len(out))}, nil
}

func toProto(u *domain.URL) *urlpb.URL {
	out := &urlpb.URL{
		Id:                   u.ID,
		UserId:               u.UserID,
		Url:                  u.URL,
		CheckIntervalSeconds: u.CheckIntervalSeconds,
		IsActive:             u.IsActive,
		CreatedAt:            timestamppb.New(u.CreatedAt),
		UpdatedAt:            timestamppb.New(u.UpdatedAt),
	}
	if u.LastCheckedAt != nil {
		out.LastCheckedAt = timestamppb.New(*u.LastCheckedAt)
	}
	return out
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrURLNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidURL), errors.Is(err, domain.ErrInvalidInterval):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
