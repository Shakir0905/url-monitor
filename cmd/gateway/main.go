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

	"github.com/shakir/url-monitor/internal/gateway/clients"
	"github.com/shakir/url-monitor/internal/gateway/config"
	"github.com/shakir/url-monitor/internal/gateway/handlers"
	"github.com/shakir/url-monitor/internal/gateway/middleware"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
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
	slog.Info("config loaded", "http_port", cfg.HTTPPort)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cls, cleanup, err := clients.New(cfg.AuthServiceAddr, cfg.URLServiceAddr, cfg.AnalyticsServiceAddr)
	if err != nil {
		return fmt.Errorf("clients: %w", err)
	}
	defer cleanup()
	slog.Info("grpc clients connected")

	authH := handlers.NewAuthHandler(cls.Auth)
	urlH := handlers.NewURLHandler(cls.URL)
	analH := handlers.NewAnalyticsHandler(cls.Analytics)

	mux := http.NewServeMux()

	// Public routes.
	mux.HandleFunc("POST /api/auth/register", authH.Register)
	mux.HandleFunc("POST /api/auth/login", authH.Login)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Protected routes.
	authMW := middleware.AuthMiddleware(cls.Auth)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/urls", urlH.List)
	protected.HandleFunc("POST /api/urls", urlH.Create)
	protected.HandleFunc("GET /api/urls/{id}", urlH.Get)
	protected.HandleFunc("PUT /api/urls/{id}", urlH.Update)
	protected.HandleFunc("DELETE /api/urls/{id}", urlH.Delete)
	protected.HandleFunc("GET /api/dashboard", analH.Dashboard)
	protected.HandleFunc("GET /api/urls/{id}/stats", analH.URLStats)

	mux.Handle("/api/urls", authMW(protected))
	mux.Handle("/api/urls/", authMW(protected))
	mux.Handle("/api/dashboard", authMW(protected))

	handler := middleware.CORS(mux)

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: handler,
	}

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.MetricsPort), Handler: metricsMux}

	errCh := make(chan error, 2)
	go func() {
		slog.Info("http server listening", "addr", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	go func() {
		slog.Info("metrics listening", "addr", metricsSrv.Addr)
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		slog.Error("server failed", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	_ = metricsSrv.Shutdown(shutdownCtx)
	slog.Info("gateway stopped")
	return nil
}
