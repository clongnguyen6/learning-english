package config

import (
	"net/http"
	"testing"
)

func TestLoadRejectsInvalidSameSite(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("REFRESH_COOKIE_SAME_SITE", "weird")

	if _, err := Load(); err == nil {
		t.Fatalf("Load() error = nil, want invalid same-site failure")
	}
}

func TestLoadRejectsInsecureSameSiteNone(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("REFRESH_COOKIE_SAME_SITE", "none")
	t.Setenv("REFRESH_COOKIE_SECURE", "false")

	if _, err := Load(); err == nil {
		t.Fatalf("Load() error = nil, want secure-cookie validation failure")
	}
}

func TestLoadNormalizesCookiePolicy(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("REFRESH_COOKIE_SAME_SITE", "Strict")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Auth.RefreshCookie.SameSite != "strict" {
		t.Fatalf("Load() same_site = %q, want strict", cfg.Auth.RefreshCookie.SameSite)
	}

	if cfg.Auth.RefreshCookie.SameSiteMode() != http.SameSiteStrictMode {
		t.Fatalf("Load() same_site mode = %v, want %v", cfg.Auth.RefreshCookie.SameSiteMode(), http.SameSiteStrictMode)
	}
}

func TestLoadRejectsRefreshTTLNotLongerThanAccessTTL(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("ACCESS_TOKEN_TTL", "15m")
	t.Setenv("REFRESH_TOKEN_TTL", "15m")

	if _, err := Load(); err == nil {
		t.Fatalf("Load() error = nil, want ttl ordering failure")
	}
}

func setRequiredEnv(t *testing.T) {
	t.Helper()

	t.Setenv("APP_ENV", "development")
	t.Setenv("APP_PUBLIC_URL", "http://localhost:5173")
	t.Setenv("API_BASE_URL", "http://localhost:8080")
	t.Setenv("API_ADDR", ":8080")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("REQUEST_ID_HEADER", "X-Request-Id")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	t.Setenv("DATABASE_URL", "postgres://learning_english:learning_english_dev@localhost:5432/learning_english?sslmode=disable")
	t.Setenv("JWT_ISSUER", "learning-english")
	t.Setenv("JWT_AUDIENCE", "learning-english-web")
	t.Setenv("JWT_HS256_SECRET", "this-is-a-long-development-secret")
	t.Setenv("ACCESS_TOKEN_TTL", "15m")
	t.Setenv("REFRESH_TOKEN_TTL", "720h")
	t.Setenv("REFRESH_COOKIE_NAME", "le_refresh")
	t.Setenv("REFRESH_COOKIE_DOMAIN", "")
	t.Setenv("REFRESH_COOKIE_SECURE", "true")
	t.Setenv("REFRESH_COOKIE_SAME_SITE", "lax")
	t.Setenv("STORAGE_BACKEND", "local")
	t.Setenv("STORAGE_BUCKET", "learning-english-local")
	t.Setenv("STORAGE_REGION", "us-east-1")
	t.Setenv("STORAGE_ENDPOINT", "")
	t.Setenv("STORAGE_PUBLIC_BASE_URL", "")
}
