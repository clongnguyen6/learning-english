package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"learning-english/backend/internal/config"
	"learning-english/backend/internal/database"
	"learning-english/backend/internal/router"
)

const serviceName = "learning-english-api"

var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("failed to load config: %v", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.HTTP.LogLevel)
	db := database.New(cfg.Database.URL)

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           router.New(router.Dependencies{Config: cfg, Logger: logger, Database: db, ServiceName: serviceName, Version: version}),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting api server", "service", serviceName, "addr", cfg.HTTP.Addr, "env", cfg.App.Env, "version", version)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("api server exited with error", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutdown requested", "signal", ctx.Err())

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
	}
}

func newLogger(level string) *slog.Logger {
	var slogLevel slog.Level

	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel}))
}
