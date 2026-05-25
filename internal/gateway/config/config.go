package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTPPort    int `envconfig:"HTTP_PORT" default:"8000"`
	MetricsPort int `envconfig:"METRICS_PORT" default:"9195"`

	AuthServiceAddr      string `envconfig:"AUTH_SERVICE_ADDR" default:"localhost:50051"`
	URLServiceAddr       string `envconfig:"URL_SERVICE_ADDR" default:"localhost:50052"`
	AnalyticsServiceAddr string `envconfig:"ANALYTICS_SERVICE_ADDR" default:"localhost:50053"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return &cfg, nil
}
