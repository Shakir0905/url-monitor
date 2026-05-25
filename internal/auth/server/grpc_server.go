package server

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/shakir/url-monitor/internal/auth/domain"
	authpb "github.com/shakir/url-monitor/proto/auth"
)

// AuthService is the interface the gRPC server depends on.
// Matches the methods of service.AuthService.
type AuthService interface {
	Register(ctx context.Context, email, password string) (token string, userID int64, err error)
	Login(ctx context.Context, email, password string) (token string, userID int64, err error)
	ValidateToken(ctx context.Context, token string) (userID int64, err error)
}

// GRPCServer implements the AuthService gRPC server.
type GRPCServer struct {
	authpb.UnimplementedAuthServiceServer
	svc AuthService
}

// NewGRPCServer constructs a new GRPCServer.
func NewGRPCServer(svc AuthService) *GRPCServer {
	return &GRPCServer{svc: svc}
}

// Register handles the Register RPC.
func (s *GRPCServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	token, userID, err := s.svc.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapError(err)
	}
	return &authpb.RegisterResponse{
		UserId: userID,
		Token:  token,
	}, nil
}

// Login handles the Login RPC.
func (s *GRPCServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	token, userID, err := s.svc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapError(err)
	}
	return &authpb.LoginResponse{
		Token:  token,
		UserId: userID,
	}, nil
}

// ValidateToken handles the ValidateToken RPC.
func (s *GRPCServer) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	userID, err := s.svc.ValidateToken(ctx, req.GetToken())
	if err != nil {
		// For ValidateToken we return valid=false instead of a gRPC error.
		// This is a design choice — token validation is a common, expected failure.
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}
	return &authpb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
		Role:   "user",
	}, nil
}

// mapError converts a domain error to an appropriate gRPC status error.
func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidPassword), errors.Is(err, domain.ErrInvalidToken), errors.Is(err, domain.ErrTokenExpired):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrWeakPassword):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
