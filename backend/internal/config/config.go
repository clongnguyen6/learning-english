package config

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the shared backend runtime contract for the first application shell.
type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	CORS     CORSConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Storage  StorageConfig
}

type AppConfig struct {
	Env        string
	PublicURL  string
	APIBaseURL string
}

type HTTPConfig struct {
	Addr            string
	LogLevel        string
	RequestIDHeader string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type DatabaseConfig struct {
	URL string
}

type AuthConfig struct {
	Issuer          string
	Audience        string
	HS256Secret     string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	RefreshCookie   CookieConfig
}

type CookieConfig struct {
	Name     string
	Domain   string
	Secure   bool
	SameSite string
}

type StorageConfig struct {
	Backend       string
	Bucket        string
	Region        string
	Endpoint      string
	PublicBaseURL string
}

// Load reads the environment contract defined in the root .env.example.
func Load() (Config, error) {
	accessTokenTTL, err := requiredDuration("ACCESS_TOKEN_TTL")
	if err != nil {
		return Config{}, err
	}

	refreshTokenTTL, err := requiredDuration("REFRESH_TOKEN_TTL")
	if err != nil {
		return Config{}, err
	}

	refreshCookieSecure, err := requiredBool("REFRESH_COOKIE_SECURE")
	if err != nil {
		return Config{}, err
	}

	allowedOrigins, err := requiredCSV("CORS_ALLOWED_ORIGINS")
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		App: AppConfig{
			Env:        requiredString("APP_ENV"),
			PublicURL:  requiredString("APP_PUBLIC_URL"),
			APIBaseURL: requiredString("API_BASE_URL"),
		},
		HTTP: HTTPConfig{
			Addr:            requiredString("API_ADDR"),
			LogLevel:        requiredString("LOG_LEVEL"),
			RequestIDHeader: requiredString("REQUEST_ID_HEADER"),
		},
		CORS: CORSConfig{
			AllowedOrigins: allowedOrigins,
		},
		Database: DatabaseConfig{
			URL: requiredString("DATABASE_URL"),
		},
		Auth: AuthConfig{
			Issuer:          requiredString("JWT_ISSUER"),
			Audience:        requiredString("JWT_AUDIENCE"),
			HS256Secret:     requiredString("JWT_HS256_SECRET"),
			AccessTokenTTL:  accessTokenTTL,
			RefreshTokenTTL: refreshTokenTTL,
			RefreshCookie: CookieConfig{
				Name:     requiredString("REFRESH_COOKIE_NAME"),
				Domain:   optionalString("REFRESH_COOKIE_DOMAIN"),
				Secure:   refreshCookieSecure,
				SameSite: requiredString("REFRESH_COOKIE_SAME_SITE"),
			},
		},
		Storage: StorageConfig{
			Backend:       requiredString("STORAGE_BACKEND"),
			Bucket:        requiredString("STORAGE_BUCKET"),
			Region:        requiredString("STORAGE_REGION"),
			Endpoint:      optionalString("STORAGE_ENDPOINT"),
			PublicBaseURL: optionalString("STORAGE_PUBLIC_BASE_URL"),
		},
	}

	for key, value := range map[string]string{
		"APP_ENV":                  cfg.App.Env,
		"APP_PUBLIC_URL":           cfg.App.PublicURL,
		"API_BASE_URL":             cfg.App.APIBaseURL,
		"API_ADDR":                 cfg.HTTP.Addr,
		"LOG_LEVEL":                cfg.HTTP.LogLevel,
		"REQUEST_ID_HEADER":        cfg.HTTP.RequestIDHeader,
		"DATABASE_URL":             cfg.Database.URL,
		"JWT_ISSUER":               cfg.Auth.Issuer,
		"JWT_AUDIENCE":             cfg.Auth.Audience,
		"JWT_HS256_SECRET":         cfg.Auth.HS256Secret,
		"REFRESH_COOKIE_NAME":      cfg.Auth.RefreshCookie.Name,
		"REFRESH_COOKIE_SAME_SITE": cfg.Auth.RefreshCookie.SameSite,
		"STORAGE_BACKEND":          cfg.Storage.Backend,
		"STORAGE_BUCKET":           cfg.Storage.Bucket,
		"STORAGE_REGION":           cfg.Storage.Region,
	} {
		if strings.TrimSpace(value) == "" {
			return Config{}, fmt.Errorf("%s is required", key)
		}
	}

	if cfg.Auth.AccessTokenTTL <= 0 {
		return Config{}, fmt.Errorf("ACCESS_TOKEN_TTL must be greater than zero")
	}

	if cfg.Auth.RefreshTokenTTL <= 0 {
		return Config{}, fmt.Errorf("REFRESH_TOKEN_TTL must be greater than zero")
	}

	if cfg.Auth.RefreshTokenTTL <= cfg.Auth.AccessTokenTTL {
		return Config{}, fmt.Errorf("REFRESH_TOKEN_TTL must be greater than ACCESS_TOKEN_TTL")
	}

	sameSite, err := normalizeSameSite(cfg.Auth.RefreshCookie.SameSite)
	if err != nil {
		return Config{}, err
	}

	cfg.Auth.RefreshCookie.SameSite = sameSite
	if cfg.Auth.RefreshCookie.SameSite == "none" && !cfg.Auth.RefreshCookie.Secure {
		return Config{}, fmt.Errorf("REFRESH_COOKIE_SECURE must be true when REFRESH_COOKIE_SAME_SITE=none")
	}

	return cfg, nil
}

func (c CookieConfig) SameSiteMode() http.SameSite {
	switch c.SameSite {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func requiredString(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func optionalString(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func requiredDuration(key string) (time.Duration, error) {
	value := requiredString(key)
	if value == "" {
		return 0, fmt.Errorf("%s is required", key)
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}

	return duration, nil
}

func requiredBool(key string) (bool, error) {
	value := requiredString(key)
	if value == "" {
		return false, fmt.Errorf("%s is required", key)
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a valid boolean: %w", key, err)
	}

	return parsed, nil
}
func requiredCSV(key string) ([]string, error) {
	value := requiredString(key)
	if value == "" {
		return nil, fmt.Errorf("%s is required", key)
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("%s must contain at least one value", key)
	}

	return values, nil
}

func normalizeSameSite(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))

	switch normalized {
	case "lax", "strict", "none":
		return normalized, nil
	default:
		return "", fmt.Errorf("REFRESH_COOKIE_SAME_SITE must be one of lax, strict, or none")
	}
}
