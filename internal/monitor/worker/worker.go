package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"

	"github.com/shakir/url-monitor/internal/monitor/domain"
	urlpb "github.com/shakir/url-monitor/proto/url"
)

type CheckRepository interface {
	SaveCheck(ctx context.Context, c *domain.CheckResult) error
	UpdateLastChecked(ctx context.Context, urlID int64, checkedAt any) error
	GetLastStatus(ctx context.Context, urlID int64) (isUp bool, exists bool, err error)
}

type Config struct {
	URLServiceAddr   string
	KafkaBrokers     []string
	PollInterval     time.Duration
	HTTPTimeout      time.Duration
	MaxConcurrency   int
	CheckedTopic     string
	StatusChangedTopic string
}

type Worker struct {
	cfg         Config
	repo        CheckRepository
	urlClient   urlpb.URLServiceClient
	httpClient  *http.Client
	checkedW    *kafka.Writer
	statusChW   *kafka.Writer
}

func New(cfg Config, repo CheckRepository) (*Worker, error) {
	conn, err := grpc.Dial(cfg.URLServiceAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("dial url service: %w", err)
	}
	urlClient := urlpb.NewURLServiceClient(conn)

	checkedW := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBrokers...),
		Topic:    cfg.CheckedTopic,
		Balancer: &kafka.Hash{},
	}
	statusW := &kafka.Writer{
		Addr:     kafka.TCP(cfg.KafkaBrokers...),
		Topic:    cfg.StatusChangedTopic,
		Balancer: &kafka.Hash{},
	}

	httpClient := &http.Client{
		Timeout: cfg.HTTPTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	return &Worker{
		cfg:        cfg,
		repo:       repo,
		urlClient:  urlClient,
		httpClient: httpClient,
		checkedW:   checkedW,
		statusChW:  statusW,
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	slog.Info("worker started", "poll_interval", w.cfg.PollInterval, "max_concurrency", w.cfg.MaxConcurrency)

	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	w.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker stopping")
			w.checkedW.Close()
			w.statusChW.Close()
			return nil
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	listCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := w.urlClient.ListActiveURLs(listCtx, &urlpb.ListActiveURLsRequest{Limit: 1000})
	if err != nil {
		slog.Error("list active urls failed", "err", err)
		return
	}
	urls := resp.GetUrls()
	if len(urls) == 0 {
		slog.Debug("no active urls")
		return
	}
	slog.Info("checking urls", "count", len(urls))

	sem := make(chan struct{}, w.cfg.MaxConcurrency)
	var wg sync.WaitGroup
	for _, u := range urls {
		wg.Add(1)
		sem <- struct{}{}
		go func(u *urlpb.URL) {
			defer wg.Done()
			defer func() { <-sem }()
			w.checkOne(ctx, u)
		}(u)
	}
	wg.Wait()
}

func (w *Worker) checkOne(ctx context.Context, u *urlpb.URL) {
	start := time.Now()
	statusCode, errMsg := w.ping(ctx, u.GetUrl())
	elapsed := time.Since(start)

	isUp := statusCode >= 200 && statusCode < 400 && errMsg == ""

	result := &domain.CheckResult{
		URLID:          u.GetId(),
		UserID:         u.GetUserId(),
		URL:            u.GetUrl(),
		StatusCode:     statusCode,
		ResponseTimeMs: int(elapsed.Milliseconds()),
		IsUp:           isUp,
		ErrorMessage:   errMsg,
		CheckedAt:      time.Now().UTC(),
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := w.repo.SaveCheck(dbCtx, result); err != nil {
		slog.Error("save check failed", "url_id", u.GetId(), "err", err)
	}
	if err := w.repo.UpdateLastChecked(dbCtx, u.GetId(), result.CheckedAt); err != nil {
		slog.Error("update last_checked failed", "url_id", u.GetId(), "err", err)
	}

	w.publishChecked(ctx, result)

	prevUp, exists, _ := w.repo.GetLastStatus(dbCtx, u.GetId())
	if exists && prevUp != isUp {
		w.publishStatusChange(ctx, result, prevUp)
	}

	slog.Info("checked",
		"url_id", u.GetId(),
		"url", u.GetUrl(),
		"status", statusCode,
		"is_up", isUp,
		"latency_ms", result.ResponseTimeMs,
	)
}

func (w *Worker) ping(ctx context.Context, url string) (int, string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("User-Agent", "url-monitor/1.0")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	return resp.StatusCode, ""
}

func (w *Worker) publishChecked(ctx context.Context, r *domain.CheckResult) {
	payload, _ := json.Marshal(map[string]any{
		"url_id":           r.URLID,
		"user_id":          r.UserID,
		"url":              r.URL,
		"status_code":      r.StatusCode,
		"response_time_ms": r.ResponseTimeMs,
		"is_up":            r.IsUp,
		"checked_at":       r.CheckedAt.Format(time.RFC3339),
	})
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", r.URLID)),
		Value: payload,
	}
	pubCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := w.checkedW.WriteMessages(pubCtx, msg); err != nil {
		slog.Error("publish url.checked failed", "err", err)
	}
}

func (w *Worker) publishStatusChange(ctx context.Context, r *domain.CheckResult, prevUp bool) {
	prev := "up"
	if !prevUp {
		prev = "down"
	}
	curr := "up"
	if !r.IsUp {
		curr = "down"
	}
	payload, _ := json.Marshal(map[string]any{
		"url_id":          r.URLID,
		"user_id":         r.UserID,
		"url":             r.URL,
		"previous_status": prev,
		"current_status":  curr,
		"error_message":   r.ErrorMessage,
		"changed_at":      r.CheckedAt.Format(time.RFC3339),
	})
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", r.URLID)),
		Value: payload,
	}
	pubCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := w.statusChW.WriteMessages(pubCtx, msg); err != nil {
		slog.Error("publish url.status_changed failed", "err", err)
	}
}
