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

	"github.com/shakir/url-monitor/internal/analytics/config"
	"github.com/shakir/url-monitor/internal/analytics/consumer"
	"github.com/shakir/url-monitor/internal/analytics/repository"
	"github.com/shakir/url-monitor/internal/analytics/server"
	"github.com/shakir/url-monitor/internal/analytics/service"
	"github.com/shakir/url-monitor/internal/pkg/db"
	analyticspb "github.com/shakir/url-monitor/proto/analytics"
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
		return err
	}
	slog.Info("config loaded", "grpc_port", cfg.GRPCPort, "metrics_port", cfg.MetricsPort)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pool, err := db.NewPostgresPool(ctx, db.Config{
		Host: cfg.DBHost, Port: cfg.DBPort, User: cfg.DBUser, Password: cfg.DBPassword,
		Database: cfg.DBName, MaxConns: 5, MinConns: 1,
	})
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer pool.Close()

	repo := repository.NewStatsRepository(pool)
	svc := service.NewAnalyticsService(repo)
	grpcSrv := server.NewGRPCServer(svc)

	gs := grpc.NewServer()
	analyticspb.RegisterAnalyticsServiceServer(gs, grpcSrv)
	reflection.Register(gs)

	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return err
	}

	cons := consumer.New(consumer.Config{
		Brokers: cfg.KafkaBrokerList(),
		Topic:   cfg.KafkaTopic,
		GroupID: cfg.KafkaGroup,
	})

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	metricsSrv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.MetricsPort), Handler: mux}

	errCh := make(chan error, 3)
	go func() {
		slog.Info("grpc listening", "addr", grpcLis.Addr().String())
		if err := gs.Serve(grpcLis); err != nil {
			errCh <- err
		}
	}()
	go func() {
		slog.Info("metrics listening", "addr", metricsSrv.Addr)
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	go func() {
		if err := cons.Run(ctx); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		slog.Error("component failed", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	gs.GracefulStop()
	_ = metricsSrv.Shutdown(shutdownCtx)
	slog.Info("analytics stopped")
	return nil
}
