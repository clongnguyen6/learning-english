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
	"learning-english/backend/internal/handlers"
	"learning-english/backend/internal/repositories"
	"learning-english/backend/internal/router"
	"learning-english/backend/internal/services"
	"learning-english/backend/internal/utils"
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
	if err := db.Ping(context.Background()); err != nil {
		logger.Warn("database ping failed during startup; readiness will stay degraded until the dependency is reachable", "error", err)
	}
	defer db.Close()

	tokenManager, err := utils.NewTokenManager(utils.TokenManagerConfig{
		Issuer:      cfg.Auth.Issuer,
		Audience:    cfg.Auth.Audience,
		HS256Secret: cfg.Auth.HS256Secret,
		TTL:         cfg.Auth.AccessTokenTTL,
	})
	if err != nil {
		log.Printf("failed to initialize token manager: %v", err)
		os.Exit(1)
	}

	authService, err := services.NewAuthService(services.AuthServiceConfig{
		Repositories:    repositories.NewSQLAuthRepositoryProvider(db),
		Tokens:          tokenManager,
		RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
	})
	if err != nil {
		log.Printf("failed to initialize auth service: %v", err)
		os.Exit(1)
	}

	authHandler, err := handlers.NewAuthHandler(handlers.AuthHandlerConfig{
		Service: authService,
		RefreshCookie: utils.RefreshCookieConfig{
			Name:     cfg.Auth.RefreshCookie.Name,
			Domain:   cfg.Auth.RefreshCookie.Domain,
			Secure:   cfg.Auth.RefreshCookie.Secure,
			SameSite: cfg.Auth.RefreshCookie.SameSiteMode(),
		},
	})
	if err != nil {
		log.Printf("failed to initialize auth handler: %v", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           router.New(router.Dependencies{Config: cfg, Logger: logger, Database: db, AuthHandler: authHandler, ServiceName: serviceName, Version: version}),
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
