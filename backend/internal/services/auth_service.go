package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"learning-english/backend/internal/models"
	"learning-english/backend/internal/repositories"
	"learning-english/backend/internal/utils"
)

var (
	ErrUsernameBlank      = errors.New("username must not be blank")
	ErrUserIDBlank        = errors.New("user id must not be blank")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
)

type AccessTokenIssuer interface {
	IssueAccessToken(subject, role string) (string, utils.AccessTokenClaims, error)
}

type AuthServiceConfig struct {
	Repositories    repositories.AuthRepositoryProvider
	Tokens          AccessTokenIssuer
	RefreshTokenTTL time.Duration
	Now             func() time.Time
	Random          io.Reader
}

type AuthService struct {
	repositories    repositories.AuthRepositoryProvider
	tokens          AccessTokenIssuer
	refreshTokenTTL time.Duration
	now             func() time.Time
	random          io.Reader
}

type LoginInput struct {
	Username  string
	Password  string
	UserAgent string
	IPAddress string
}

type LoginResult struct {
	User              models.UserSummary
	AccessToken       string
	AccessTokenClaims utils.AccessTokenClaims
	RefreshToken      string
	RefreshExpiresAt  time.Time
	Session           models.UserSession
}

type LogoutInput struct {
	UserID       string
	RefreshToken string
	UserAgent    string
	IPAddress    string
}

type LogoutResult struct {
	Revoked   bool
	SessionID string
}

type LogoutAllInput struct {
	UserID    string
	UserAgent string
	IPAddress string
}

type LogoutAllResult struct {
	RevokedSessions int64
}

func NewAuthService(cfg AuthServiceConfig) (*AuthService, error) {
	if cfg.Repositories == nil {
		return nil, errors.New("auth repositories are required")
	}
	if cfg.Tokens == nil {
		return nil, errors.New("access token issuer is required")
	}
	if cfg.RefreshTokenTTL <= 0 {
		return nil, errors.New("refresh token ttl must be greater than zero")
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Random == nil {
		cfg.Random = rand.Reader
	}

	return &AuthService{
		repositories:    cfg.Repositories,
		tokens:          cfg.Tokens,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		now:             cfg.Now,
		random:          cfg.Random,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	if s == nil {
		return LoginResult{}, errors.New("auth service is not configured")
	}

	username := strings.TrimSpace(input.Username)
	if username == "" {
		return LoginResult{}, ErrUsernameBlank
	}
	if strings.TrimSpace(input.Password) == "" {
		return LoginResult{}, utils.ErrPasswordBlank
	}

	now := s.now().UTC()
	var result LoginResult

	err := s.repositories.WithinTx(ctx, func(set repositories.AuthRepositorySet) error {
		if err := validateAuthRepositorySet(set); err != nil {
			return err
		}

		user, err := set.Users.FindByUsername(ctx, username)
		if err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				if _, auditErr := set.AuthEvents.Create(ctx, repositories.CreateAuthEventParams{
					EventType: models.AuthEventLoginFailed,
					IPAddress: input.IPAddress,
					UserAgent: input.UserAgent,
					Metadata: map[string]any{
						"username": username,
						"reason":   "invalid_credentials",
					},
					CreatedAt: now,
				}); auditErr != nil {
					return auditErr
				}

				return ErrInvalidCredentials
			}

			return err
		}
		if !user.IsActive() {
			if _, auditErr := set.AuthEvents.Create(ctx, repositories.CreateAuthEventParams{
				UserID:    stringPointer(user.ID),
				EventType: models.AuthEventLoginFailed,
				IPAddress: input.IPAddress,
				UserAgent: input.UserAgent,
				Metadata: map[string]any{
					"username": username,
					"reason":   "inactive_user",
				},
				CreatedAt: now,
			}); auditErr != nil {
				return auditErr
			}

			return ErrInvalidCredentials
		}
		if err := utils.VerifyPassword(user.PasswordHash, input.Password); err != nil {
			if _, auditErr := set.AuthEvents.Create(ctx, repositories.CreateAuthEventParams{
				UserID:    stringPointer(user.ID),
				EventType: models.AuthEventLoginFailed,
				IPAddress: input.IPAddress,
				UserAgent: input.UserAgent,
				Metadata: map[string]any{
					"username": username,
					"reason":   "invalid_credentials",
				},
				CreatedAt: now,
			}); auditErr != nil {
				return auditErr
			}

			return ErrInvalidCredentials
		}

		accessToken, accessTokenClaims, err := s.tokens.IssueAccessToken(user.ID, user.Role)
		if err != nil {
			return fmt.Errorf("issue access token: %w", err)
		}

		refreshToken, err := s.generateOpaqueToken()
		if err != nil {
			return err
		}
		tokenFamily, err := s.generateOpaqueToken()
		if err != nil {
			return err
		}

		session, err := set.Sessions.Create(ctx, repositories.CreateUserSessionParams{
			UserID:           user.ID,
			RefreshTokenHash: hashOpaqueToken(refreshToken),
			TokenFamily:      tokenFamily,
			UserAgent:        input.UserAgent,
			IPAddress:        input.IPAddress,
			ExpiresAt:        now.Add(s.refreshTokenTTL),
			CreatedAt:        now,
		})
		if err != nil {
			return err
		}

		if err := set.Users.UpdateLastLoginAt(ctx, user.ID, now); err != nil {
			return err
		}

		if _, err := set.AuthEvents.Create(ctx, repositories.CreateAuthEventParams{
			UserID:    stringPointer(user.ID),
			SessionID: stringPointer(session.ID),
			EventType: models.AuthEventLoginSuccess,
			IPAddress: input.IPAddress,
			UserAgent: input.UserAgent,
			Metadata: map[string]any{
				"login_method": "password",
			},
			CreatedAt: now,
		}); err != nil {
			return err
		}

		user.LastLoginAt = &now
		result = LoginResult{
			User:              user.Summary(),
			AccessToken:       accessToken,
			AccessTokenClaims: accessTokenClaims,
			RefreshToken:      refreshToken,
			RefreshExpiresAt:  session.ExpiresAt,
			Session:           session,
		}

		return nil
	})
	if err != nil {
		return LoginResult{}, err
	}

	return result, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (models.UserSummary, error) {
	if s == nil {
		return models.UserSummary{}, errors.New("auth service is not configured")
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return models.UserSummary{}, ErrUserIDBlank
	}

	set := s.repositories.ReadOnly()
	if err := validateAuthRepositorySet(set); err != nil {
		return models.UserSummary{}, err
	}

	user, err := set.Users.FindByID(ctx, userID)
	if err != nil {
		return models.UserSummary{}, err
	}
	if !user.IsActive() {
		return models.UserSummary{}, ErrUserInactive
	}

	return user.Summary(), nil
}

func (s *AuthService) Logout(ctx context.Context, input LogoutInput) (LogoutResult, error) {
	if s == nil {
		return LogoutResult{}, errors.New("auth service is not configured")
	}

	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return LogoutResult{}, ErrUserIDBlank
	}

	refreshToken := strings.TrimSpace(input.RefreshToken)
	if refreshToken == "" {
		return LogoutResult{}, nil
	}

	now := s.now().UTC()
	var result LogoutResult

	err := s.repositories.WithinTx(ctx, func(set repositories.AuthRepositorySet) error {
		if err := validateAuthRepositorySet(set); err != nil {
			return err
		}

		session, err := set.Sessions.RevokeByRefreshTokenHash(ctx, userID, hashOpaqueToken(refreshToken), now)
		if err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return nil
			}

			return err
		}

		if _, err := set.AuthEvents.Create(ctx, repositories.CreateAuthEventParams{
			UserID:    stringPointer(userID),
			SessionID: stringPointer(session.ID),
			EventType: models.AuthEventLogout,
			IPAddress: input.IPAddress,
			UserAgent: input.UserAgent,
			Metadata: map[string]any{
				"scope": "current_device",
			},
			CreatedAt: now,
		}); err != nil {
			return err
		}

		result = LogoutResult{
			Revoked:   true,
			SessionID: session.ID,
		}

		return nil
	})
	if err != nil {
		return LogoutResult{}, err
	}

	return result, nil
}

func (s *AuthService) LogoutAll(ctx context.Context, input LogoutAllInput) (LogoutAllResult, error) {
	if s == nil {
		return LogoutAllResult{}, errors.New("auth service is not configured")
	}

	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return LogoutAllResult{}, ErrUserIDBlank
	}

	now := s.now().UTC()
	var result LogoutAllResult

	err := s.repositories.WithinTx(ctx, func(set repositories.AuthRepositorySet) error {
		if err := validateAuthRepositorySet(set); err != nil {
			return err
		}

		revokedSessions, err := set.Sessions.RevokeAllByUserID(ctx, userID, now)
		if err != nil {
			return err
		}

		if _, err := set.AuthEvents.Create(ctx, repositories.CreateAuthEventParams{
			UserID:    stringPointer(userID),
			EventType: models.AuthEventLogoutAll,
			IPAddress: input.IPAddress,
			UserAgent: input.UserAgent,
			Metadata: map[string]any{
				"scope":            "all_devices",
				"revoked_sessions": revokedSessions,
			},
			CreatedAt: now,
		}); err != nil {
			return err
		}

		result = LogoutAllResult{RevokedSessions: revokedSessions}
		return nil
	})
	if err != nil {
		return LogoutAllResult{}, err
	}

	return result, nil
}

func (s *AuthService) generateOpaqueToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := io.ReadFull(s.random, buffer); err != nil {
		return "", fmt.Errorf("generate opaque token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func hashOpaqueToken(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func validateAuthRepositorySet(set repositories.AuthRepositorySet) error {
	switch {
	case set.Users == nil:
		return errors.New("user repository is not configured")
	case set.Sessions == nil:
		return errors.New("session repository is not configured")
	case set.AuthEvents == nil:
		return errors.New("auth event repository is not configured")
	default:
		return nil
	}
}

func stringPointer(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return &value
}
