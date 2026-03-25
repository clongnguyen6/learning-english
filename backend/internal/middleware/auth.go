package middleware

import (
	"errors"
	"net/http"
	"strings"

	"learning-english/backend/internal/handlers"
	"learning-english/backend/internal/policies"
	"learning-english/backend/internal/utils"
)

type AccessTokenParser interface {
	ParseAccessToken(tokenString string) (utils.AccessTokenClaims, error)
}

func BearerAuth(parser AccessTokenParser) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if parser == nil {
				handlers.WriteError(
					w,
					http.StatusServiceUnavailable,
					"SERVICE_UNAVAILABLE",
					"auth runtime is not configured",
					nil,
					authResponseMeta(r),
				)
				return
			}

			token, err := bearerTokenFromHeader(r.Header.Get("Authorization"))
			if err != nil {
				writeUnauthorized(w, r, "authorization header must contain a bearer token", nil)
				return
			}

			claims, err := parser.ParseAccessToken(token)
			if err != nil {
				writeUnauthorized(w, r, authErrorMessage(err), nil)
				return
			}

			user := AuthenticatedUser{
				UserID: strings.TrimSpace(claims.Subject),
				Role:   policies.NormalizeRole(claims.Role),
			}
			if user.UserID == "" || !policies.IsSupportedRole(user.Role) {
				writeUnauthorized(w, r, "access token is invalid", nil)
				return
			}

			next.ServeHTTP(w, r.WithContext(ContextWithAuthenticatedUser(r.Context(), user)))
		})
	}
}

func RequireAuthenticatedUser() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := AuthenticatedUserFromContext(r.Context()); !ok {
				writeUnauthorized(w, r, "authentication is required", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireRoles(allowedRoles ...string) Middleware {
	normalizedRoles := normalizeRoles(allowedRoles)
	if len(normalizedRoles) == 0 {
		return RequireAuthenticatedUser()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := AuthenticatedUserFromContext(r.Context())
			if !ok {
				writeUnauthorized(w, r, "authentication is required", nil)
				return
			}

			if !policies.HasAnyRole(user.Role, normalizedRoles...) {
				handlers.WriteError(
					w,
					http.StatusForbidden,
					"FORBIDDEN",
					"insufficient permissions for this route",
					map[string]any{
						"required_roles": normalizedRoles,
						"current_role":   user.Role,
					},
					authResponseMeta(r),
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bearerTokenFromHeader(headerValue string) (string, error) {
	fields := strings.Fields(strings.TrimSpace(headerValue))
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return "", errors.New("authorization header must use bearer scheme")
	}

	token := strings.TrimSpace(fields[1])
	if token == "" {
		return "", errors.New("authorization bearer token is blank")
	}

	return token, nil
}

func authErrorMessage(err error) string {
	switch {
	case errors.Is(err, utils.ErrTokenExpired):
		return "access token is expired"
	case errors.Is(err, utils.ErrTokenAudienceInvalid), errors.Is(err, utils.ErrTokenIssuerInvalid):
		return "access token is invalid"
	case errors.Is(err, utils.ErrTokenRoleInvalid), errors.Is(err, utils.ErrTokenRoleBlank), errors.Is(err, utils.ErrTokenSubjectBlank):
		return "access token is invalid"
	default:
		return "access token is invalid"
	}
}

func authResponseMeta(r *http.Request) map[string]any {
	meta := map[string]any{}
	if requestID := utils.RequestIDFromContext(r.Context()); requestID != "" {
		meta["request_id"] = requestID
	}

	return meta
}

func normalizeRoles(roles []string) []string {
	normalized := make([]string, 0, len(roles))
	seen := map[string]struct{}{}

	for _, role := range roles {
		candidate := policies.NormalizeRole(role)
		if !policies.IsSupportedRole(candidate) {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}

		seen[candidate] = struct{}{}
		normalized = append(normalized, candidate)
	}

	return normalized
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request, message string, details any) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	handlers.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", message, details, authResponseMeta(r))
}
