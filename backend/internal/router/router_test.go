package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"learning-english/backend/internal/config"
	"learning-english/backend/internal/dto"
	"learning-english/backend/internal/utils"
)

func TestProtectedLearnerNamespaceRejectsMissingBearerToken(t *testing.T) {
	response := exerciseRouter(t, http.MethodGet, "/api/v1/dashboard/summary", "")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}

	if authenticate := response.Header().Get("WWW-Authenticate"); authenticate != "Bearer" {
		t.Fatalf("WWW-Authenticate = %q, want Bearer", authenticate)
	}

	body := decodeAPIResponse(t, response)
	if body.Error == nil || body.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("error = %#v, want UNAUTHORIZED", body.Error)
	}

	if requestID := body.Meta["request_id"]; requestID == nil || requestID == "" {
		t.Fatalf("meta.request_id = %#v, want populated request ID", requestID)
	}
}

func TestProtectedLearnerNamespaceAllowsAuthenticatedRequestThroughToRouteResolution(t *testing.T) {
	response := exerciseRouter(t, http.MethodGet, "/api/v1/dashboard/summary", issueAccessToken(t, "learner"))

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}

	body := decodeAPIResponse(t, response)
	if body.Error == nil || body.Error.Code != "NOT_FOUND" {
		t.Fatalf("error = %#v, want NOT_FOUND", body.Error)
	}
}

func TestProtectedAdminNamespaceRejectsLearnerRole(t *testing.T) {
	response := exerciseRouter(t, http.MethodGet, "/api/v1/admin/jobs/123", issueAccessToken(t, "learner"))

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}

	body := decodeAPIResponse(t, response)
	if body.Error == nil || body.Error.Code != "FORBIDDEN" {
		t.Fatalf("error = %#v, want FORBIDDEN", body.Error)
	}
}

func TestProtectedAdminNamespaceAllowsAdminRoleThroughToRouteResolution(t *testing.T) {
	response := exerciseRouter(t, http.MethodGet, "/api/v1/admin/jobs/123", issueAccessToken(t, "admin"))

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}

	body := decodeAPIResponse(t, response)
	if body.Error == nil || body.Error.Code != "NOT_FOUND" {
		t.Fatalf("error = %#v, want NOT_FOUND", body.Error)
	}
}

func TestPublicAuthNamespaceRemainsUnprotectedUntilHandlersLand(t *testing.T) {
	response := exerciseRouter(t, http.MethodPost, "/api/v1/auth/login", "")

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}

	body := decodeAPIResponse(t, response)
	if body.Error == nil || body.Error.Code != "NOT_FOUND" {
		t.Fatalf("error = %#v, want NOT_FOUND", body.Error)
	}
}

func exerciseRouter(t *testing.T, method, path, accessToken string) *httptest.ResponseRecorder {
	t.Helper()

	handler := New(Dependencies{
		Config: testConfig(),
		Logger: slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil)),
	})

	request := httptest.NewRequest(method, path, nil)
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func issueAccessToken(t *testing.T, role string) string {
	t.Helper()

	manager, err := utils.NewTokenManager(utils.TokenManagerConfig{
		Issuer:      "learning-english",
		Audience:    "learning-english-web",
		HS256Secret: "this-is-a-long-development-secret",
		TTL:         15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}

	token, _, err := manager.IssueAccessToken("user-123", role)
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	return token
}

func decodeAPIResponse(t *testing.T, response *httptest.ResponseRecorder) dto.APIResponse {
	t.Helper()

	var body dto.APIResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	return body
}

func testConfig() config.Config {
	return config.Config{
		App: config.AppConfig{
			Env: "test",
		},
		HTTP: config.HTTPConfig{
			RequestIDHeader: "X-Request-Id",
		},
		CORS: config.CORSConfig{
			AllowedOrigins: []string{"http://localhost:5173"},
		},
		Auth: config.AuthConfig{
			Issuer:          "learning-english",
			Audience:        "learning-english-web",
			HS256Secret:     "this-is-a-long-development-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 720 * time.Hour,
			RefreshCookie: config.CookieConfig{
				Name:     "le_refresh",
				Secure:   true,
				SameSite: "lax",
			},
		},
	}
}
