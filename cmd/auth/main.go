package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/shakir/url-monitor/internal/auth/config"
	"github.com/shakir/url-monitor/internal/auth/repository"
	"github.com/shakir/url-monitor/internal/auth/server"
	"github.com/shakir/url-monitor/internal/auth/service"
	"github.com/shakir/url-monitor/internal/pkg/db"
	authpb "github.com/shakir/url-monitor/proto/auth"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. Logger setup.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// 2. Load config.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	slog.Info("config loaded",
		"grpc_port", cfg.GRPCPort,
		"metrics_port", cfg.MetricsPort,
		"db_host", cfg.DBHost,
	)

	// 3. Context with cancellation on SIGTERM/SIGINT.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// 4. Database pool.
	pool, err := db.NewPostgresPool(ctx, db.Config{
		Host:        cfg.DBHost,
		Port:        cfg.DBPort,
		User:        cfg.DBUser,
		Password:    cfg.DBPassword,
		Database:    cfg.DBName,
		MaxConns:    cfg.DBMaxConns,
		MinConns:    cfg.DBMinConns,
		MaxConnIdle: 5 * time.Minute,
		MaxConnLife: 30 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer pool.Close()
	slog.Info("connected to postgres")

	// 5. Wire layers: repository -> service -> grpc server.
	userRepo := repository.NewUserRepository(pool)

	authSvc := service.NewAuthService(userRepo, service.Config{
		JWTSecret:      cfg.JWTSecret,
		JWTTTL:         cfg.JWTTTL,
		BcryptCost:     cfg.BcryptCost,
		MinPasswordLen: cfg.MinPasswordLen,
	})

	grpcSrv := server.NewGRPCServer(authSvc)

	// 6. gRPC server.
	gs := grpc.NewServer()
	authpb.RegisterAuthServiceServer(gs, grpcSrv)
	reflection.Register(gs) // enables grpcurl introspection — useful for debugging

	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	// 7. Metrics server (separate port for Prometheus).
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	metricsSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler: metricsMux,
	}

	// 8. Run both servers.
	errCh := make(chan error, 2)

	go func() {
		slog.Info("grpc server listening", "addr", grpcLis.Addr().String())
		if err := gs.Serve(grpcLis); err != nil {
			errCh <- fmt.Errorf("grpc serve: %w", err)
		}
	}()

	go func() {
		slog.Info("metrics server listening", "addr", metricsSrv.Addr)
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("metrics serve: %w", err)
		}
	}()

	// 9. Wait for signal or error.
	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-errCh:
		slog.Error("server failed", "err", err)
	}

	// 10. Graceful shutdown.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("shutting down grpc server")
	gracefulStop := make(chan struct{})
	go func() {
		gs.GracefulStop()
		close(gracefulStop)
	}()
	select {
	case <-gracefulStop:
		slog.Info("grpc stopped gracefully")
	case <-shutdownCtx.Done():
		slog.Warn("grpc graceful stop timed out, forcing")
		gs.Stop()
	}

	slog.Info("shutting down metrics server")
	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		slog.Warn("metrics shutdown error", "err", err)
	}

	slog.Info("auth service stopped")
	return nil
}
