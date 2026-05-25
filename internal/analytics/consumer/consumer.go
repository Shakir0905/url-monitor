package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

type Consumer struct {
	reader *kafka.Reader
}

func New(cfg Config) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	})
	return &Consumer{reader: r}
}

func (c *Consumer) Run(ctx context.Context) error {
	slog.Info("kafka consumer started", "topic", c.reader.Config().Topic)
	defer c.reader.Close()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Error("kafka read", "err", err)
			continue
		}
		var payload map[string]any
		_ = json.Unmarshal(m.Value, &payload)
		slog.Info("event received",
			"topic", m.Topic,
			"partition", m.Partition,
			"offset", m.Offset,
			"url_id", payload["url_id"],
			"status_code", payload["status_code"],
		)
	}
}
