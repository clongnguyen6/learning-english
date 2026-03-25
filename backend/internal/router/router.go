package router

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"learning-english/backend/internal/config"
	"learning-english/backend/internal/handlers"
	"learning-english/backend/internal/middleware"
	"learning-english/backend/internal/policies"
	"learning-english/backend/internal/utils"
)

type readinessPinger interface {
	Ping(context.Context) error
}

type Dependencies struct {
	Config      config.Config
	Logger      *slog.Logger
	Database    readinessPinger
	AuthHandler *handlers.AuthHandler
	ServiceName string
	Version     string
}

func New(deps Dependencies) http.Handler {
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	if deps.ServiceName == "" {
		deps.ServiceName = "learning-english-api"
	}

	healthHandler := handlers.NewHealthHandler(deps.ServiceName, deps.Version, deps.Database)
	authNoStore := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			next.ServeHTTP(w, r)
		})
	}
	tokenManager, err := utils.NewTokenManager(utils.TokenManagerConfig{
		Issuer:      deps.Config.Auth.Issuer,
		Audience:    deps.Config.Auth.Audience,
		HS256Secret: deps.Config.Auth.HS256Secret,
		TTL:         deps.Config.Auth.AccessTokenTTL,
	})
	if err != nil {
		deps.Logger.Error("failed to initialize token manager", "error", err)

		return middleware.Chain(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlers.WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "auth runtime is not configured", nil, nil)
			}),
			middleware.Recovery(deps.Logger),
			middleware.RequestID(deps.Config.HTTP.RequestIDHeader),
			middleware.Logging(deps.Logger, deps.Config.HTTP.RequestIDHeader),
			middleware.CORS(deps.Config.CORS.AllowedOrigins),
		)
	}

	learnerProtectedNotFound := middleware.Chain(
		http.HandlerFunc(handlers.NotFound),
		middleware.BearerAuth(tokenManager),
		middleware.RequireAuthenticatedUser(),
	)
	authProtectedRoutes := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if deps.AuthHandler == nil {
				handlers.NotFound(w, r)
				return
			}
			user, ok := middleware.AuthenticatedUserFromContext(r.Context())
			if !ok || strings.TrimSpace(user.UserID) == "" {
				handlers.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication is required", nil, nil)
				return
			}

			switch r.URL.Path {
			case "/api/v1/auth/me":
				if r.Method != http.MethodGet {
					handlers.MethodNotAllowed(w, r, http.MethodGet)
					return
				}
				deps.AuthHandler.MeForUser(w, r, user.UserID)
			case "/api/v1/auth/logout":
				if r.Method != http.MethodPost {
					handlers.MethodNotAllowed(w, r, http.MethodPost)
					return
				}
				deps.AuthHandler.LogoutForUser(w, r, user.UserID)
			case "/api/v1/auth/logout-all":
				if r.Method != http.MethodPost {
					handlers.MethodNotAllowed(w, r, http.MethodPost)
					return
				}
				deps.AuthHandler.LogoutAllForUser(w, r, user.UserID)
			default:
				handlers.NotFound(w, r)
			}
		}),
		authNoStore,
		middleware.BearerAuth(tokenManager),
		middleware.RequireAuthenticatedUser(),
	)
	authPublicRoutes := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if deps.AuthHandler == nil {
			handlers.NotFound(w, r)
			return
		}

		switch r.URL.Path {
		case "/api/v1/auth/login":
			if r.Method != http.MethodPost {
				handlers.MethodNotAllowed(w, r, http.MethodPost)
				return
			}
			deps.AuthHandler.Login(w, r)
		case "/api/v1/auth/refresh":
			if r.Method != http.MethodPost {
				handlers.MethodNotAllowed(w, r, http.MethodPost)
				return
			}
			deps.AuthHandler.Refresh(w, r)
		default:
			handlers.NotFound(w, r)
		}
	})
	adminProtectedNotFound := middleware.Chain(
		http.HandlerFunc(handlers.NotFound),
		middleware.BearerAuth(tokenManager),
		middleware.RequireRoles(policies.RoleAdmin),
	)

	mux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1":
			if r.Method != http.MethodGet {
				handlers.MethodNotAllowed(w, r, http.MethodGet)
				return
			}
			handlers.Root(deps.Config).ServeHTTP(w, r)
		case "/api/v1/healthz":
			if r.Method != http.MethodGet {
				handlers.MethodNotAllowed(w, r, http.MethodGet)
				return
			}
			healthHandler.Healthz(w, r)
		case "/api/v1/readyz":
			if r.Method != http.MethodGet {
				handlers.MethodNotAllowed(w, r, http.MethodGet)
				return
			}
			healthHandler.Readyz(w, r)
		default:
			if pathMatchesPrefix(r.URL.Path, "/api/v1/auth") {
				if isProtectedAuthPath(r.URL.Path) {
					authProtectedRoutes.ServeHTTP(w, r)
					return
				}
				authPublicRoutes.ServeHTTP(w, r)
				return
			}
			if pathMatchesPrefix(r.URL.Path, "/api/v1/admin") {
				adminProtectedNotFound.ServeHTTP(w, r)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/api/v1/") && !pathMatchesPrefix(r.URL.Path, "/api/v1/auth") {
				learnerProtectedNotFound.ServeHTTP(w, r)
				return
			}
			handlers.NotFound(w, r)
		}
	})

	return middleware.Chain(
		mux,
		middleware.Recovery(deps.Logger),
		middleware.RequestID(deps.Config.HTTP.RequestIDHeader),
		middleware.Logging(deps.Logger, deps.Config.HTTP.RequestIDHeader),
		middleware.CORS(deps.Config.CORS.AllowedOrigins),
	)
}

func pathMatchesPrefix(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func isProtectedAuthPath(path string) bool {
	switch path {
	case "/api/v1/auth/me", "/api/v1/auth/logout", "/api/v1/auth/logout-all":
		return true
	default:
		return false
	}
}
