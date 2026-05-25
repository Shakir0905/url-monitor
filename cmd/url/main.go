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

	"github.com/shakir/url-monitor/internal/pkg/db"
	"github.com/shakir/url-monitor/internal/url/config"
	"github.com/shakir/url-monitor/internal/url/repository"
	"github.com/shakir/url-monitor/internal/url/server"
	"github.com/shakir/url-monitor/internal/url/service"
	urlpb "github.com/shakir/url-monitor/proto/url"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	slog.Info("config loaded", "grpc_port", cfg.GRPCPort, "metrics_port", cfg.MetricsPort, "db_host", cfg.DBHost)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pool, err := db.NewPostgresPool(ctx, db.Config{
		Host: cfg.DBHost, Port: cfg.DBPort,
		User: cfg.DBUser, Password: cfg.DBPassword,
		Database: cfg.DBName,
		MaxConns: cfg.DBMaxConns, MinConns: cfg.DBMinConns,
		MaxConnIdle: 5 * time.Minute, MaxConnLife: 30 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer pool.Close()
	slog.Info("connected to postgres")

	repo := repository.NewURLRepository(pool)
	svc := service.NewURLService(repo)
	grpcSrv := server.NewGRPCServer(svc)

	gs := grpc.NewServer()
	urlpb.RegisterURLServiceServer(gs, grpcSrv)
	reflection.Register(gs)

	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	metricsSrv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.MetricsPort), Handler: metricsMux}

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

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-errCh:
		slog.Error("server failed", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

	_ = metricsSrv.Shutdown(shutdownCtx)
	slog.Info("url service stopped")
	return nil
}
