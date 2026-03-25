package utils

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHashPasswordAndVerify(t *testing.T) {
	passwordHash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if passwordHash == "correct horse battery staple" {
		t.Fatalf("HashPassword() returned the raw password")
	}

	if err := VerifyPassword(passwordHash, "correct horse battery staple"); err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}
}

func TestHashPasswordRejectsBlankInput(t *testing.T) {
	if _, err := HashPassword("   "); !errors.Is(err, ErrPasswordBlank) {
		t.Fatalf("HashPassword() error = %v, want %v", err, ErrPasswordBlank)
	}
}

func TestTokenManagerRoundTrip(t *testing.T) {
	manager, err := NewTokenManager(TokenManagerConfig{
		Issuer:      "learning-english",
		Audience:    "learning-english-web",
		HS256Secret: "this-is-a-long-development-secret",
		TTL:         15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}

	fixedNow := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	manager.now = func() time.Time { return fixedNow }

	token, claims, err := manager.IssueAccessToken("user-123", "admin")
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	if claims.Subject != "user-123" {
		t.Fatalf("IssueAccessToken() subject = %q, want user-123", claims.Subject)
	}

	parsed, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}

	if parsed.Subject != "user-123" {
		t.Fatalf("ParseAccessToken() subject = %q, want user-123", parsed.Subject)
	}

	if parsed.Role != "admin" {
		t.Fatalf("ParseAccessToken() role = %q, want admin", parsed.Role)
	}
}

func TestTokenManagerRejectsExpiredToken(t *testing.T) {
	manager, err := NewTokenManager(TokenManagerConfig{
		Issuer:      "learning-english",
		Audience:    "learning-english-web",
		HS256Secret: "this-is-a-long-development-secret",
		TTL:         time.Minute,
	})
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}

	issuedAt := time.Now().UTC().Add(-3 * time.Minute).Truncate(time.Second)
	manager.now = func() time.Time { return issuedAt }

	token, _, err := manager.IssueAccessToken("user-123", "learner")
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	manager.now = func() time.Time { return issuedAt.Add(2 * time.Minute) }

	if _, err := manager.ParseAccessToken(token); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("ParseAccessToken() error = %v, want %v", err, ErrTokenExpired)
	}
}

func TestTokenManagerRejectsUnexpectedAudience(t *testing.T) {
	manager, err := NewTokenManager(TokenManagerConfig{
		Issuer:      "learning-english",
		Audience:    "learning-english-web",
		HS256Secret: "this-is-a-long-development-secret",
		TTL:         15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}

	manager.now = func() time.Time { return time.Now().UTC().Add(-time.Minute).Truncate(time.Second) }
	token, _, err := manager.IssueAccessToken("user-123", "learner")
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	otherManager, err := NewTokenManager(TokenManagerConfig{
		Issuer:      "learning-english",
		Audience:    "other-audience",
		HS256Secret: "this-is-a-long-development-secret",
		TTL:         15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}

	otherManager.now = manager.now
	if _, err := otherManager.ParseAccessToken(token); !errors.Is(err, ErrTokenAudienceInvalid) {
		t.Fatalf("ParseAccessToken() error = %v, want %v", err, ErrTokenAudienceInvalid)
	}
}

func TestTokenManagerRejectsUnknownRole(t *testing.T) {
	manager, err := NewTokenManager(TokenManagerConfig{
		Issuer:      "learning-english",
		Audience:    "learning-english-web",
		HS256Secret: "this-is-a-long-development-secret",
		TTL:         15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}

	if _, _, err := manager.IssueAccessToken("user-123", "owner"); !errors.Is(err, ErrTokenRoleInvalid) {
		t.Fatalf("IssueAccessToken() error = %v, want %v", err, ErrTokenRoleInvalid)
	}
}

func TestBuildRefreshCookie(t *testing.T) {
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	cookie, err := BuildRefreshCookie(RefreshCookieConfig{
		Name:     "le_refresh",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}, "refresh-token-value", expiresAt)
	if err != nil {
		t.Fatalf("BuildRefreshCookie() error = %v", err)
	}

	if cookie.Name != "le_refresh" {
		t.Fatalf("BuildRefreshCookie() name = %q", cookie.Name)
	}

	if !cookie.HttpOnly {
		t.Fatalf("BuildRefreshCookie() HttpOnly = false, want true")
	}

	if !cookie.Secure {
		t.Fatalf("BuildRefreshCookie() Secure = false, want true")
	}

	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("BuildRefreshCookie() SameSite = %v, want %v", cookie.SameSite, http.SameSiteStrictMode)
	}

	if cookie.MaxAge < 1 {
		t.Fatalf("BuildRefreshCookie() MaxAge = %d, want >= 1", cookie.MaxAge)
	}
}

func TestExpireRefreshCookie(t *testing.T) {
	cookie, err := ExpireRefreshCookie(RefreshCookieConfig{Name: "le_refresh"})
	if err != nil {
		t.Fatalf("ExpireRefreshCookie() error = %v", err)
	}

	if cookie.MaxAge != -1 {
		t.Fatalf("ExpireRefreshCookie() MaxAge = %d, want -1", cookie.MaxAge)
	}

	if cookie.Value != "" {
		t.Fatalf("ExpireRefreshCookie() value = %q, want empty", cookie.Value)
	}
}

func TestVerifyPasswordRejectsWrongPassword(t *testing.T) {
	passwordHash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	err = VerifyPassword(passwordHash, "not-the-same")
	if err == nil {
		t.Fatalf("VerifyPassword() error = nil, want mismatch")
	}

	if !strings.Contains(err.Error(), "verify password") {
		t.Fatalf("VerifyPassword() error = %v, want verify password context", err)
	}
}
