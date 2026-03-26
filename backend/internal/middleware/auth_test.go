package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"learning-english/backend/internal/dto"
	"learning-english/backend/internal/utils"

	jwt "github.com/golang-jwt/jwt/v5"
)

type stubAccessTokenParser struct {
	claims utils.AccessTokenClaims
	err    error
}

func (s stubAccessTokenParser) ParseAccessToken(_ string) (utils.AccessTokenClaims, error) {
	return s.claims, s.err
}

func TestBearerAuthRejectsExpiredToken(t *testing.T) {
	handler := BearerAuth(stubAccessTokenParser{err: utils.ErrTokenExpired})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.Header.Set("Authorization", "Bearer expired-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if authenticate := response.Header().Get("WWW-Authenticate"); authenticate != "Bearer" {
		t.Fatalf("WWW-Authenticate = %q, want Bearer", authenticate)
	}

	body := decodeMiddlewareResponse(t, response)
	if body.Error == nil || body.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("error = %#v, want UNAUTHORIZED", body.Error)
	}
	if body.Error.Message != "access token is expired" {
		t.Fatalf("error.message = %q, want access token is expired", body.Error.Message)
	}
}

func TestBearerAuthStoresAuthenticatedUserInContext(t *testing.T) {
	var capturedUser AuthenticatedUser
	handler := BearerAuth(stubAccessTokenParser{
		claims: utils.AccessTokenClaims{
			Role:    "learner",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user-123",
			},
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := AuthenticatedUserFromContext(r.Context())
		if !ok {
			t.Fatal("authenticated user missing from context")
		}

		capturedUser = user
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if capturedUser.UserID != "user-123" || capturedUser.Role != "learner" {
		t.Fatalf("captured user = %#v, want user-123/learner", capturedUser)
	}
}

func TestRequireRolesRejectsUnauthenticatedRequest(t *testing.T) {
	handler := RequireRoles("admin")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/jobs/123", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestRequireRolesRejectsWrongRole(t *testing.T) {
	handler := RequireRoles("admin")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/jobs/123", nil)
	request = request.WithContext(ContextWithAuthenticatedUser(request.Context(), AuthenticatedUser{
		UserID: "user-123",
		Role:   "learner",
	}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}

	body := decodeMiddlewareResponse(t, response)
	if body.Error == nil || body.Error.Code != "FORBIDDEN" {
		t.Fatalf("error = %#v, want FORBIDDEN", body.Error)
	}
}

func TestBearerTokenFromHeaderRejectsMalformedValues(t *testing.T) {
	_, err := bearerTokenFromHeader("Basic abc123")
	if err == nil {
		t.Fatal("bearerTokenFromHeader() error = nil, want error")
	}
	if err.Error() == "" {
		t.Fatalf("bearerTokenFromHeader() error = %v, want non-empty error", err)
	}
}

func decodeMiddlewareResponse(t *testing.T, response *httptest.ResponseRecorder) dto.APIResponse {
	t.Helper()

	var body dto.APIResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	return body
}
