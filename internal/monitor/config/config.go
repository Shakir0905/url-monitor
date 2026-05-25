package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MetricsPort int    `envconfig:"METRICS_PORT" default:"9194"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	URLServiceAddr string        `envconfig:"URL_SERVICE_ADDR" default:"localhost:50052"`
	KafkaBrokers   string        `envconfig:"KAFKA_BROKERS" default:"localhost:9092"`
	PollInterval   time.Duration `envconfig:"POLL_INTERVAL" default:"30s"`
	HTTPTimeout    time.Duration `envconfig:"HTTP_TIMEOUT" default:"10s"`
	MaxConcurrency int           `envconfig:"MAX_CONCURRENCY" default:"10"`

	CheckedTopic       string `envconfig:"CHECKED_TOPIC" default:"url.checked"`
	StatusChangedTopic string `envconfig:"STATUS_CHANGED_TOPIC" default:"url.status_changed"`

	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     int    `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER" default:"app"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"app_password"`
	DBName     string `envconfig:"DB_NAME" default:"url_monitor"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) KafkaBrokerList() []string {
	parts := strings.Split(c.KafkaBrokers, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
