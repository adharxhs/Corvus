package config

import (
	"errors"
	"time"
)

// Config holds the immutable application configuration. It is populated once at
// startup via Load and must not be mutated afterwards.
type Config struct {
	HTTPPort      string
	DatabasePath  string
	JWTSecret     string
	JWTExpiration time.Duration
	LogLevel      string
	Environment   string
}

const (
	defaultHTTPPort      = "8080"
	defaultDatabasePath  = "corvus.db"
	defaultLogLevel      = "info"
	defaultEnvironment   = "development"
	defaultJWTExpiration = 24 * time.Hour
)

// Load reads configuration from the environment, applying defaults for any
// unset values, then validates the result. The returned Config is immutable.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPPort:      getEnv("HTTP_PORT", defaultHTTPPort),
		DatabasePath:  getEnv("DATABASE_PATH", defaultDatabasePath),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTExpiration: getEnvDuration("JWT_EXPIRATION", defaultJWTExpiration),
		LogLevel:      getEnv("LOG_LEVEL", defaultLogLevel),
		Environment:   getEnv("ENVIRONMENT", defaultEnvironment),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.HTTPPort == "" {
		return errors.New("config: HTTP_PORT must not be empty")
	}
	if c.DatabasePath == "" {
		return errors.New("config: DATABASE_PATH must not be empty")
	}
	if c.JWTSecret == "" {
		return errors.New("config: JWT_SECRET must be set")
	}
	if c.JWTExpiration <= 0 {
		return errors.New("config: JWT_EXPIRATION must be positive")
	}
	return nil
}
