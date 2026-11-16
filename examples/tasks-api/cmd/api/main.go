// Command api runs the tasks REST API.
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

	"github.com/example/tasks-api/internal/auth"
	"github.com/example/tasks-api/internal/config"
	mw "github.com/example/tasks-api/internal/middleware"
	"github.com/example/tasks-api/internal/postgres"
	"github.com/example/tasks-api/internal/server"
	"github.com/example/tasks-api/internal/task"
	"github.com/example/tasks-api/internal/user"
)

// version is set at build time: go build -ldflags "-X main.version=1.2.3"
var version = "dev"

func main() {
	// main only reports errors; all real work happens in run, which is
	// easy to test and lets deferred cleanups run before the process exits.
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	// ctx is cancelled on Ctrl+C (SIGINT) or SIGTERM (docker stop, Kubernetes).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := newLogger(cfg)
	slog.SetDefault(logger)

	// ---- Infrastructure ----
	pool, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if cfg.AutoMigrate {
		if err := postgres.Migrate(ctx, pool); err != nil {
			return err
		}
		logger.Info("database migrated")
	}

	tokens, err := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, "tasks-api", cfg.JWTTTL)
	if err != nil {
		return err
	}

	// ---- Features (dependency injection by hand) ----
	userSvc, err := user.NewService(postgres.NewUserRepo(pool), tokens)
	if err != nil {
		return err
	}
	taskSvc := task.NewService(postgres.NewTaskRepo(pool))

	authLimiter := mw.NewRateLimiter(cfg.AuthRateLimit, time.Minute)
	go authLimiter.Cleanup(ctx, time.Minute)

	handler := server.NewHandler(server.Deps{
		Logger:         logger,
		DB:             pool,
		Tokens:         tokens,
		Users:          user.NewHandler(userSvc),
		Tasks:          task.NewHandler(taskSvc),
		AuthLimiter:    authLimiter,
		CORSOrigins:    cfg.CORSOrigins,
		RequestTimeout: cfg.RequestTimeout,
	})
	srv := server.New(cfg.Addr, handler, logger)

	// ---- Run until a signal arrives, then shut down gracefully ----
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", cfg.Addr, "env", cfg.Env, "version", version)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		// The server failed to start (e.g. port already in use).
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	// Stop accepting new connections and wait for in-flight requests.
	// Use a fresh context: ctx is already cancelled.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	if err := <-serverErr; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("server stopped cleanly")
	return nil
}

func newLogger(cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	var h slog.Handler = slog.NewJSONHandler(os.Stdout, opts)
	if cfg.LogFormat == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(h)
}
