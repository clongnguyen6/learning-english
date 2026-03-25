package router

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"learning-english/backend/internal/config"
	"learning-english/backend/internal/dto"
	"learning-english/backend/internal/handlers"
	"learning-english/backend/internal/models"
	"learning-english/backend/internal/services"
	"learning-english/backend/internal/utils"

	jwt "github.com/golang-jwt/jwt/v5"
)

func TestProtectedLearnerNamespaceRejectsMissingBearerToken(t *testing.T) {
	response := exerciseRouter(t, routerRequest{
		method: http.MethodGet,
		path:   "/api/v1/dashboard/summary",
	})

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
	response := exerciseRouter(t, routerRequest{
		method:      http.MethodGet,
		path:        "/api/v1/dashboard/summary",
		accessToken: issueAccessToken(t, "learner"),
	})

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}

	body := decodeAPIResponse(t, response)
	if body.Error == nil || body.Error.Code != "NOT_FOUND" {
		t.Fatalf("error = %#v, want NOT_FOUND", body.Error)
	}
}

func TestProtectedAdminNamespaceRejectsLearnerRole(t *testing.T) {
	response := exerciseRouter(t, routerRequest{
		method:      http.MethodGet,
		path:        "/api/v1/admin/jobs/123",
		accessToken: issueAccessToken(t, "learner"),
	})

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}

	body := decodeAPIResponse(t, response)
	if body.Error == nil || body.Error.Code != "FORBIDDEN" {
		t.Fatalf("error = %#v, want FORBIDDEN", body.Error)
	}
}

func TestProtectedAdminNamespaceAllowsAdminRoleThroughToRouteResolution(t *testing.T) {
	response := exerciseRouter(t, routerRequest{
		method:      http.MethodGet,
		path:        "/api/v1/admin/jobs/123",
		accessToken: issueAccessToken(t, "admin"),
	})

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}

	body := decodeAPIResponse(t, response)
	if body.Error == nil || body.Error.Code != "NOT_FOUND" {
		t.Fatalf("error = %#v, want NOT_FOUND", body.Error)
	}
}

func TestPublicAuthLoginRouteReturnsTokenAndSetsRefreshCookie(t *testing.T) {
	now := time.Date(2026, 3, 26, 4, 0, 0, 0, time.UTC)
	service := &fakeAuthService{
		loginResult: services.LoginResult{
			User:        models.UserSummary{ID: "user-123", Username: "long", DisplayName: "Long", Role: "learner"},
			AccessToken: "access-token",
			AccessTokenClaims: utils.AccessTokenClaims{
				Role: "learner",
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   "user-123",
					ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
				},
			},
			RefreshToken:     "refresh-token",
			RefreshExpiresAt: now.Add(24 * time.Hour),
		},
	}

	response := exerciseRouter(t, routerRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth/login",
		body:        `{"username":"long","password":"secret123"}`,
		authService: service,
	})

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	if cacheControl := response.Header().Get("Cache-Control"); cacheControl != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", cacheControl)
	}

	body := decodeAPIResponse(t, response)
	data := responseDataMap(t, body)
	if data["access_token"] != "access-token" {
		t.Fatalf("data.access_token = %#v, want access-token", data["access_token"])
	}
	if data["token_type"] != "Bearer" {
		t.Fatalf("data.token_type = %#v, want Bearer", data["token_type"])
	}
	if _, ok := data["refresh_token"]; ok {
		t.Fatalf("data.refresh_token unexpectedly present: %#v", data["refresh_token"])
	}

	cookie := findCookie(t, response, testConfig().Auth.RefreshCookie.Name)
	if cookie.Value != "refresh-token" {
		t.Fatalf("refresh cookie value = %q, want refresh-token", cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Fatalf("refresh cookie HttpOnly = false, want true")
	}

	if service.loginInput.Username != "long" || service.loginInput.Password != "secret123" {
		t.Fatalf("login input = %#v, want username/password propagated", service.loginInput)
	}
	if service.loginInput.IPAddress != "192.0.2.1" {
		t.Fatalf("login input ip = %q, want 192.0.2.1", service.loginInput.IPAddress)
	}
}

func TestPublicAuthRefreshMissingCookieReturnsSessionRevoked(t *testing.T) {
	service := &fakeAuthService{refreshErr: services.ErrSessionRevoked}
	response := exerciseRouter(t, routerRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth/refresh",
		authService: service,
	})

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}

	body := decodeAPIResponse(t, response)
	if body.Error == nil || body.Error.Code != "SESSION_REVOKED" {
		t.Fatalf("error = %#v, want SESSION_REVOKED", body.Error)
	}
}

func TestProtectedAuthMeRequiresBearerToken(t *testing.T) {
	response := exerciseRouter(t, routerRequest{
		method:      http.MethodGet,
		path:        "/api/v1/auth/me",
		authService: &fakeAuthService{},
	})

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}

	body := decodeAPIResponse(t, response)
	if body.Error == nil || body.Error.Code != "UNAUTHORIZED" {
		t.Fatalf("error = %#v, want UNAUTHORIZED", body.Error)
	}
}

func TestProtectedAuthMeReturnsCurrentUser(t *testing.T) {
	service := &fakeAuthService{
		meResult: models.UserSummary{ID: "user-123", Username: "long", DisplayName: "Long", Role: "learner"},
	}
	response := exerciseRouter(t, routerRequest{
		method:      http.MethodGet,
		path:        "/api/v1/auth/me",
		accessToken: issueAccessToken(t, "learner"),
		authService: service,
	})

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeAPIResponse(t, response)
	data := responseDataMap(t, body)
	user, ok := data["user"].(map[string]any)
	if !ok {
		t.Fatalf("data.user = %#v, want object", data["user"])
	}
	if user["username"] != "long" {
		t.Fatalf("data.user.username = %#v, want long", user["username"])
	}
	if service.meUserID != "user-123" {
		t.Fatalf("me user id = %q, want user-123", service.meUserID)
	}
}

func TestProtectedAuthLogoutClearsRefreshCookie(t *testing.T) {
	service := &fakeAuthService{
		logoutResult: services.LogoutResult{Revoked: true},
	}
	response := exerciseRouter(t, routerRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth/logout",
		accessToken: issueAccessToken(t, "learner"),
		cookies: []*http.Cookie{
			{Name: testConfig().Auth.RefreshCookie.Name, Value: "refresh-token"},
		},
		authService: service,
	})

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeAPIResponse(t, response)
	data := responseDataMap(t, body)
	if data["revoked"] != true {
		t.Fatalf("data.revoked = %#v, want true", data["revoked"])
	}

	cookie := findCookie(t, response, testConfig().Auth.RefreshCookie.Name)
	if cookie.Value != "" || cookie.MaxAge != -1 {
		t.Fatalf("logout cookie = %#v, want expired cookie", cookie)
	}

	if service.logoutInput.UserID != "user-123" || service.logoutInput.RefreshToken != "refresh-token" {
		t.Fatalf("logout input = %#v, want authenticated user and refresh token", service.logoutInput)
	}
}

func TestProtectedAuthLogoutAllClearsRefreshCookie(t *testing.T) {
	service := &fakeAuthService{
		logoutAllResult: services.LogoutAllResult{RevokedSessions: 3},
	}
	response := exerciseRouter(t, routerRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth/logout-all",
		accessToken: issueAccessToken(t, "learner"),
		cookies: []*http.Cookie{
			{Name: testConfig().Auth.RefreshCookie.Name, Value: "refresh-token"},
		},
		authService: service,
	})

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := decodeAPIResponse(t, response)
	data := responseDataMap(t, body)
	if data["revoked_sessions"] != float64(3) {
		t.Fatalf("data.revoked_sessions = %#v, want 3", data["revoked_sessions"])
	}

	cookie := findCookie(t, response, testConfig().Auth.RefreshCookie.Name)
	if cookie.Value != "" || cookie.MaxAge != -1 {
		t.Fatalf("logout-all cookie = %#v, want expired cookie", cookie)
	}
}

type routerRequest struct {
	method      string
	path        string
	accessToken string
	body        string
	cookies     []*http.Cookie
	authService handlers.AuthHTTPService
}

type fakeAuthService struct {
	loginInput      services.LoginInput
	loginResult     services.LoginResult
	loginErr        error
	refreshInput    services.RefreshInput
	refreshResult   services.RefreshResult
	refreshErr      error
	meUserID        string
	meResult        models.UserSummary
	meErr           error
	logoutInput     services.LogoutInput
	logoutResult    services.LogoutResult
	logoutErr       error
	logoutAllInput  services.LogoutAllInput
	logoutAllResult services.LogoutAllResult
	logoutAllErr    error
}

func (f *fakeAuthService) Login(_ context.Context, input services.LoginInput) (services.LoginResult, error) {
	f.loginInput = input
	return f.loginResult, f.loginErr
}

func (f *fakeAuthService) Refresh(_ context.Context, input services.RefreshInput) (services.RefreshResult, error) {
	f.refreshInput = input
	return f.refreshResult, f.refreshErr
}

func (f *fakeAuthService) Me(_ context.Context, userID string) (models.UserSummary, error) {
	f.meUserID = userID
	return f.meResult, f.meErr
}

func (f *fakeAuthService) Logout(_ context.Context, input services.LogoutInput) (services.LogoutResult, error) {
	f.logoutInput = input
	return f.logoutResult, f.logoutErr
}

func (f *fakeAuthService) LogoutAll(_ context.Context, input services.LogoutAllInput) (services.LogoutAllResult, error) {
	f.logoutAllInput = input
	return f.logoutAllResult, f.logoutAllErr
}

func exerciseRouter(t *testing.T, input routerRequest) *httptest.ResponseRecorder {
	t.Helper()

	deps := Dependencies{
		Config: testConfig(),
		Logger: slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil)),
	}
	if input.authService != nil {
		authHandler, err := handlers.NewAuthHandler(handlers.AuthHandlerConfig{
			Service: input.authService,
			RefreshCookie: utils.RefreshCookieConfig{
				Name:     testConfig().Auth.RefreshCookie.Name,
				Domain:   testConfig().Auth.RefreshCookie.Domain,
				Secure:   testConfig().Auth.RefreshCookie.Secure,
				SameSite: testConfig().Auth.RefreshCookie.SameSiteMode(),
			},
		})
		if err != nil {
			t.Fatalf("NewAuthHandler() error = %v", err)
		}
		deps.AuthHandler = authHandler
	}

	handler := New(deps)

	var body *bytes.Buffer
	if input.body != "" {
		body = bytes.NewBufferString(input.body)
	} else {
		body = bytes.NewBuffer(nil)
	}
	request := httptest.NewRequest(input.method, input.path, body)
	if input.body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if input.accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+input.accessToken)
	}
	for _, cookie := range input.cookies {
		request.AddCookie(cookie)
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

func responseDataMap(t *testing.T, body dto.APIResponse) map[string]any {
	t.Helper()

	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("body.Data = %#v, want object", body.Data)
	}

	return data
}

func findCookie(t *testing.T, response *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}

	t.Fatalf("cookie %q not found in response", name)
	return nil
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
