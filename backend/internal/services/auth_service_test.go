package services

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"learning-english/backend/internal/models"
	"learning-english/backend/internal/repositories"
	"learning-english/backend/internal/utils"
)

func TestAuthServiceLoginSuccess(t *testing.T) {
	passwordHash, err := utils.HashPassword("LearnerDemo123!")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	fixedNow := time.Date(2026, time.March, 26, 10, 0, 0, 0, time.UTC)
	userRepo := &fakeUserRepository{
		usersByUsername: map[string]models.User{
			"learner_alex": {
				ID:           "user-1",
				Username:     "learner_alex",
				PasswordHash: passwordHash,
				DisplayName:  "Alex Nguyen",
				Role:         "learner",
				Status:       models.UserStatusActive,
			},
		},
	}
	sessionRepo := &fakeSessionRepository{}
	eventRepo := &fakeAuthEventRepository{}
	provider := fakeAuthRepositoryProvider{
		readOnlySet: repositories.AuthRepositorySet{Users: userRepo, Sessions: sessionRepo, AuthEvents: eventRepo},
		txSet:       repositories.AuthRepositorySet{Users: userRepo, Sessions: sessionRepo, AuthEvents: eventRepo},
	}

	service, err := NewAuthService(AuthServiceConfig{
		Repositories:    provider,
		Tokens:          fakeAccessTokenIssuer{token: "issued-access-token"},
		RefreshTokenTTL: 24 * time.Hour,
		Now:             func() time.Time { return fixedNow },
		Random:          bytes.NewReader(bytes.Repeat([]byte{0xAB}, 64)),
	})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	result, err := service.Login(context.Background(), LoginInput{
		Username:  "learner_alex",
		Password:  "LearnerDemo123!",
		UserAgent: "Mozilla/5.0",
		IPAddress: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if result.User.ID != "user-1" || result.User.Username != "learner_alex" || result.User.Role != "learner" {
		t.Fatalf("Login() user = %#v", result.User)
	}
	if result.AccessToken != "issued-access-token" {
		t.Fatalf("Login() access token = %q, want issued-access-token", result.AccessToken)
	}
	if result.RefreshToken == "" {
		t.Fatalf("Login() refresh token is blank")
	}
	if got, want := result.RefreshExpiresAt, fixedNow.Add(24*time.Hour); !got.Equal(want) {
		t.Fatalf("Login() refresh expiry = %v, want %v", got, want)
	}
	if len(sessionRepo.createdSessions) != 1 {
		t.Fatalf("created sessions = %d, want 1", len(sessionRepo.createdSessions))
	}
	if sessionRepo.createdSessions[0].RefreshTokenHash == result.RefreshToken {
		t.Fatalf("refresh token hash unexpectedly equals raw refresh token")
	}
	if len(eventRepo.createdEvents) != 1 || eventRepo.createdEvents[0].EventType != models.AuthEventLoginSuccess {
		t.Fatalf("login success event = %#v", eventRepo.createdEvents)
	}
	if !userRepo.lastLoginAt.Equal(fixedNow) {
		t.Fatalf("UpdateLastLoginAt() time = %v, want %v", userRepo.lastLoginAt, fixedNow)
	}
}

func TestAuthServiceLoginInvalidPasswordRecordsFailure(t *testing.T) {
	passwordHash, err := utils.HashPassword("LearnerDemo123!")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	userRepo := &fakeUserRepository{
		usersByUsername: map[string]models.User{
			"learner_alex": {
				ID:           "user-1",
				Username:     "learner_alex",
				PasswordHash: passwordHash,
				DisplayName:  "Alex Nguyen",
				Role:         "learner",
				Status:       models.UserStatusActive,
			},
		},
	}
	sessionRepo := &fakeSessionRepository{}
	eventRepo := &fakeAuthEventRepository{}

	service, err := NewAuthService(AuthServiceConfig{
		Repositories: fakeAuthRepositoryProvider{
			readOnlySet: repositories.AuthRepositorySet{Users: userRepo, Sessions: sessionRepo, AuthEvents: eventRepo},
			txSet:       repositories.AuthRepositorySet{Users: userRepo, Sessions: sessionRepo, AuthEvents: eventRepo},
		},
		Tokens:          fakeAccessTokenIssuer{token: "issued-access-token"},
		RefreshTokenTTL: 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	_, err = service.Login(context.Background(), LoginInput{
		Username: "learner_alex",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
	if len(sessionRepo.createdSessions) != 0 {
		t.Fatalf("created sessions = %d, want 0", len(sessionRepo.createdSessions))
	}
	if len(eventRepo.createdEvents) != 1 || eventRepo.createdEvents[0].EventType != models.AuthEventLoginFailed {
		t.Fatalf("login failed event = %#v", eventRepo.createdEvents)
	}
}

func TestAuthServiceMeReturnsActiveUserSummary(t *testing.T) {
	userRepo := &fakeUserRepository{
		usersByID: map[string]models.User{
			"user-1": {
				ID:          "user-1",
				Username:    "learner_alex",
				DisplayName: "Alex Nguyen",
				Role:        "learner",
				Status:      models.UserStatusActive,
			},
		},
	}

	service, err := NewAuthService(AuthServiceConfig{
		Repositories: fakeAuthRepositoryProvider{
			readOnlySet: repositories.AuthRepositorySet{Users: userRepo, Sessions: &fakeSessionRepository{}, AuthEvents: &fakeAuthEventRepository{}},
			txSet:       repositories.AuthRepositorySet{Users: userRepo, Sessions: &fakeSessionRepository{}, AuthEvents: &fakeAuthEventRepository{}},
		},
		Tokens:          fakeAccessTokenIssuer{token: "issued-access-token"},
		RefreshTokenTTL: 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	user, err := service.Me(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("Me() error = %v", err)
	}
	if user.ID != "user-1" || user.Username != "learner_alex" || user.Role != "learner" {
		t.Fatalf("Me() user = %#v", user)
	}
}

func TestAuthServiceLogoutRevokesSessionByRefreshToken(t *testing.T) {
	sessionRepo := &fakeSessionRepository{
		revokeByHashResult: models.UserSession{ID: "session-1", UserID: "user-1"},
	}
	eventRepo := &fakeAuthEventRepository{}

	service, err := NewAuthService(AuthServiceConfig{
		Repositories: fakeAuthRepositoryProvider{
			readOnlySet: repositories.AuthRepositorySet{Users: &fakeUserRepository{}, Sessions: sessionRepo, AuthEvents: eventRepo},
			txSet:       repositories.AuthRepositorySet{Users: &fakeUserRepository{}, Sessions: sessionRepo, AuthEvents: eventRepo},
		},
		Tokens:          fakeAccessTokenIssuer{token: "issued-access-token"},
		RefreshTokenTTL: 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	result, err := service.Logout(context.Background(), LogoutInput{
		UserID:       "user-1",
		RefreshToken: "raw-refresh-token",
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if !result.Revoked || result.SessionID != "session-1" {
		t.Fatalf("Logout() result = %#v", result)
	}
	if sessionRepo.revokedUserID != "user-1" || sessionRepo.revokedRefreshTokenHash == "" {
		t.Fatalf("revoke params = user %q hash %q", sessionRepo.revokedUserID, sessionRepo.revokedRefreshTokenHash)
	}
	if len(eventRepo.createdEvents) != 1 || eventRepo.createdEvents[0].EventType != models.AuthEventLogout {
		t.Fatalf("logout event = %#v", eventRepo.createdEvents)
	}
}

func TestAuthServiceLogoutIsIdempotentWhenSessionIsMissing(t *testing.T) {
	sessionRepo := &fakeSessionRepository{
		revokeByHashErr: repositories.ErrNotFound,
	}
	eventRepo := &fakeAuthEventRepository{}

	service, err := NewAuthService(AuthServiceConfig{
		Repositories: fakeAuthRepositoryProvider{
			readOnlySet: repositories.AuthRepositorySet{Users: &fakeUserRepository{}, Sessions: sessionRepo, AuthEvents: eventRepo},
			txSet:       repositories.AuthRepositorySet{Users: &fakeUserRepository{}, Sessions: sessionRepo, AuthEvents: eventRepo},
		},
		Tokens:          fakeAccessTokenIssuer{token: "issued-access-token"},
		RefreshTokenTTL: 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	result, err := service.Logout(context.Background(), LogoutInput{
		UserID:       "user-1",
		RefreshToken: "raw-refresh-token",
	})
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}
	if result.Revoked {
		t.Fatalf("Logout() result = %#v, want no-op", result)
	}
	if len(eventRepo.createdEvents) != 0 {
		t.Fatalf("logout event count = %d, want 0", len(eventRepo.createdEvents))
	}
}

func TestAuthServiceLogoutAllRecordsAuditEvent(t *testing.T) {
	sessionRepo := &fakeSessionRepository{
		revokeAllCount: 3,
	}
	eventRepo := &fakeAuthEventRepository{}

	service, err := NewAuthService(AuthServiceConfig{
		Repositories: fakeAuthRepositoryProvider{
			readOnlySet: repositories.AuthRepositorySet{Users: &fakeUserRepository{}, Sessions: sessionRepo, AuthEvents: eventRepo},
			txSet:       repositories.AuthRepositorySet{Users: &fakeUserRepository{}, Sessions: sessionRepo, AuthEvents: eventRepo},
		},
		Tokens:          fakeAccessTokenIssuer{token: "issued-access-token"},
		RefreshTokenTTL: 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	result, err := service.LogoutAll(context.Background(), LogoutAllInput{
		UserID:    "user-1",
		UserAgent: "Mozilla/5.0",
		IPAddress: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("LogoutAll() error = %v", err)
	}
	if result.RevokedSessions != 3 {
		t.Fatalf("LogoutAll() revoked sessions = %d, want 3", result.RevokedSessions)
	}
	if len(eventRepo.createdEvents) != 1 || eventRepo.createdEvents[0].EventType != models.AuthEventLogoutAll {
		t.Fatalf("logout-all event = %#v", eventRepo.createdEvents)
	}
}

type fakeAuthRepositoryProvider struct {
	readOnlySet repositories.AuthRepositorySet
	txSet       repositories.AuthRepositorySet
}

func (p fakeAuthRepositoryProvider) ReadOnly() repositories.AuthRepositorySet {
	return p.readOnlySet
}

func (p fakeAuthRepositoryProvider) WithinTx(_ context.Context, fn func(repositories.AuthRepositorySet) error) error {
	return fn(p.txSet)
}

type fakeUserRepository struct {
	usersByUsername map[string]models.User
	usersByID       map[string]models.User
	lastLoginUserID string
	lastLoginAt     time.Time
}

func (r *fakeUserRepository) FindByUsername(_ context.Context, username string) (models.User, error) {
	if user, ok := r.usersByUsername[username]; ok {
		return user, nil
	}

	return models.User{}, repositories.ErrNotFound
}

func (r *fakeUserRepository) FindByID(_ context.Context, userID string) (models.User, error) {
	if user, ok := r.usersByID[userID]; ok {
		return user, nil
	}

	for _, user := range r.usersByUsername {
		if user.ID == userID {
			return user, nil
		}
	}

	return models.User{}, repositories.ErrNotFound
}

func (r *fakeUserRepository) UpdateLastLoginAt(_ context.Context, userID string, lastLoginAt time.Time) error {
	r.lastLoginUserID = userID
	r.lastLoginAt = lastLoginAt
	return nil
}

type fakeSessionRepository struct {
	createdSessions         []repositories.CreateUserSessionParams
	revokeByHashResult      models.UserSession
	revokeByHashErr         error
	revokedUserID           string
	revokedRefreshTokenHash string
	revokeAllCount          int64
}

func (r *fakeSessionRepository) Create(_ context.Context, params repositories.CreateUserSessionParams) (models.UserSession, error) {
	r.createdSessions = append(r.createdSessions, params)
	return models.UserSession{
		ID:               "session-1",
		UserID:           params.UserID,
		RefreshTokenHash: params.RefreshTokenHash,
		TokenFamily:      params.TokenFamily,
		UserAgent:        params.UserAgent,
		IPAddress:        params.IPAddress,
		ExpiresAt:        params.ExpiresAt,
		CreatedAt:        params.CreatedAt,
		UpdatedAt:        params.CreatedAt,
	}, nil
}

func (r *fakeSessionRepository) RevokeByRefreshTokenHash(_ context.Context, userID, refreshTokenHash string, _ time.Time) (models.UserSession, error) {
	r.revokedUserID = userID
	r.revokedRefreshTokenHash = refreshTokenHash
	if r.revokeByHashErr != nil {
		return models.UserSession{}, r.revokeByHashErr
	}

	return r.revokeByHashResult, nil
}

func (r *fakeSessionRepository) RevokeAllByUserID(_ context.Context, _ string, _ time.Time) (int64, error) {
	return r.revokeAllCount, nil
}

type fakeAuthEventRepository struct {
	createdEvents []repositories.CreateAuthEventParams
}

func (r *fakeAuthEventRepository) Create(_ context.Context, params repositories.CreateAuthEventParams) (models.AuthEvent, error) {
	r.createdEvents = append(r.createdEvents, params)
	return models.AuthEvent{
		ID:        "event-1",
		UserID:    params.UserID,
		SessionID: params.SessionID,
		EventType: params.EventType,
		IPAddress: params.IPAddress,
		UserAgent: params.UserAgent,
		Metadata:  params.Metadata,
		CreatedAt: params.CreatedAt,
	}, nil
}

type fakeAccessTokenIssuer struct {
	token string
}

func (i fakeAccessTokenIssuer) IssueAccessToken(subject, role string) (string, utils.AccessTokenClaims, error) {
	return i.token, utils.AccessTokenClaims{
		Role:             role,
		RegisteredClaims: utils.AccessTokenClaims{}.RegisteredClaims,
	}, nil
}
