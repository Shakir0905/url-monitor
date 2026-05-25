package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration for the Auth Service.
type Config struct {
	// Server
	GRPCPort    int    `envconfig:"GRPC_PORT" default:"50051"`
	MetricsPort int    `envconfig:"METRICS_PORT" default:"9091"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	// Database
	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     int    `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER" default:"app"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"app_password"`
	DBName     string `envconfig:"DB_NAME" default:"url_monitor"`
	DBMaxConns int32  `envconfig:"DB_MAX_CONNS" default:"10"`
	DBMinConns int32  `envconfig:"DB_MIN_CONNS" default:"2"`

	// JWT
	JWTSecret string        `envconfig:"JWT_SECRET" required:"true"`
	JWTTTL    time.Duration `envconfig:"JWT_TTL" default:"24h"`

	// Bcrypt
	BcryptCost     int `envconfig:"BCRYPT_COST" default:"10"`
	MinPasswordLen int `envconfig:"MIN_PASSWORD_LEN" default:"8"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return &cfg, nil
}
