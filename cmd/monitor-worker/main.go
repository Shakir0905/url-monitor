package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/shakir/url-monitor/internal/monitor/config"
	"github.com/shakir/url-monitor/internal/monitor/repository"
	"github.com/shakir/url-monitor/internal/monitor/worker"
	"github.com/shakir/url-monitor/internal/pkg/db"
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
	slog.Info("config loaded",
		"poll_interval", cfg.PollInterval,
		"url_service", cfg.URLServiceAddr,
		"kafka", cfg.KafkaBrokers,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	pool, err := db.NewPostgresPool(ctx, db.Config{
		Host: cfg.DBHost, Port: cfg.DBPort,
		User: cfg.DBUser, Password: cfg.DBPassword,
		Database: cfg.DBName,
		MaxConns: 5, MinConns: 1,
	})
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer pool.Close()
	slog.Info("connected to postgres")

	repo := repository.NewCheckRepository(pool)

	w, err := worker.New(worker.Config{
		URLServiceAddr:     cfg.URLServiceAddr,
		KafkaBrokers:       cfg.KafkaBrokerList(),
		PollInterval:       cfg.PollInterval,
		HTTPTimeout:        cfg.HTTPTimeout,
		MaxConcurrency:     cfg.MaxConcurrency,
		CheckedTopic:       cfg.CheckedTopic,
		StatusChangedTopic: cfg.StatusChangedTopic,
	}, repo)
	if err != nil {
		return fmt.Errorf("create worker: %w", err)
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
		slog.Info("metrics server listening", "addr", metricsSrv.Addr)
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("metrics: %w", err)
		}
	}()
	go func() {
		if err := w.Run(ctx); err != nil {
			errCh <- fmt.Errorf("worker: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-errCh:
		slog.Error("component failed", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = metricsSrv.Shutdown(shutdownCtx)

	slog.Info("monitor worker stopped")
	return nil
}
