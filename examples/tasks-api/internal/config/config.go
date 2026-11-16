// Package config loads settings from environment variables (12-factor style).
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env             string // "development" or "production"
	Addr            string // listen address, e.g. ":8080"
	DatabaseURL     string // postgres://user:pass@host:5432/db?sslmode=disable
	AutoMigrate     bool   // apply migrations on startup
	JWTSecret       string // >= 32 bytes, keep it secret
	JWTIssuer       string
	JWTTTL          time.Duration // access token lifetime
	CORSOrigins     []string
	LogLevel        slog.Level
	LogFormat       string // "json" or "text"
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
	AuthRateLimit   int // login/register requests per minute per IP
}

// Load reads the environment, applies defaults and validates the result.
// It reports every problem at once instead of failing on the first one.
func Load() (Config, error) {
	var errs []error
	env := func(key, def string) string {
		if v, ok := os.LookupEnv(key); ok && v != "" {
			return v
		}
		return def
	}
	duration := func(key, def string) time.Duration {
		d, err := time.ParseDuration(env(key, def))
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", key, err))
		}
		return d
	}
	integer := func(key, def string) int {
		n, err := strconv.Atoi(env(key, def))
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", key, err))
		}
		return n
	}
	boolean := func(key, def string) bool {
		b, err := strconv.ParseBool(env(key, def))
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", key, err))
		}
		return b
	}

	cfg := Config{
		Env:             env("APP_ENV", "development"),
		Addr:            env("HTTP_ADDR", ":8080"),
		DatabaseURL:     env("DATABASE_URL", ""),
		AutoMigrate:     boolean("DB_AUTO_MIGRATE", "true"),
		JWTSecret:       env("JWT_SECRET", ""),
		JWTIssuer:       env("JWT_ISSUER", "tasks-api"),
		JWTTTL:          duration("JWT_TTL", "15m"),
		LogFormat:       env("LOG_FORMAT", "json"),
		RequestTimeout:  duration("REQUEST_TIMEOUT", "10s"),
		ShutdownTimeout: duration("SHUTDOWN_TIMEOUT", "15s"),
		AuthRateLimit:   integer("AUTH_RATE_LIMIT", "10"),
	}
	for o := range strings.SplitSeq(env("CORS_ALLOWED_ORIGINS", ""), ",") {
		if o = strings.TrimSpace(o); o != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, o)
		}
	}
	if err := cfg.LogLevel.UnmarshalText([]byte(env("LOG_LEVEL", "info"))); err != nil {
		errs = append(errs, fmt.Errorf("LOG_LEVEL: %w", err))
	}

	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if len(cfg.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET is required and must be at least 32 bytes"))
	}
	if cfg.LogFormat != "json" && cfg.LogFormat != "text" {
		errs = append(errs, errors.New(`LOG_FORMAT must be "json" or "text"`))
	}
	if cfg.AuthRateLimit < 1 {
		errs = append(errs, errors.New("AUTH_RATE_LIMIT must be >= 1"))
	}

	return cfg, errors.Join(errs...)
}
