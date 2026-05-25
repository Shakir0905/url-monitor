package config

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCPort    int    `envconfig:"GRPC_PORT" default:"50053"`
	MetricsPort int    `envconfig:"METRICS_PORT" default:"9193"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     int    `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER" default:"app"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"app_password"`
	DBName     string `envconfig:"DB_NAME" default:"url_monitor"`

	KafkaBrokers string `envconfig:"KAFKA_BROKERS" default:"localhost:9092"`
	KafkaTopic   string `envconfig:"KAFKA_TOPIC" default:"url.checked"`
	KafkaGroup   string `envconfig:"KAFKA_GROUP" default:"analytics-group"`
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
