package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCPort    int    `envconfig:"GRPC_PORT" default:"50052"`
	MetricsPort int    `envconfig:"METRICS_PORT" default:"9092"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     int    `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER" default:"app"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"app_password"`
	DBName     string `envconfig:"DB_NAME" default:"url_monitor"`
	DBMaxConns int32  `envconfig:"DB_MAX_CONNS" default:"10"`
	DBMinConns int32  `envconfig:"DB_MIN_CONNS" default:"2"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return &cfg, nil
}
