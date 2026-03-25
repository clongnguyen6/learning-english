package models

import (
	"strings"
	"time"
)

const (
	UserStatusActive = "active"

	AuthEventLoginSuccess = "login_success"
	AuthEventLoginFailed  = "login_failed"
	AuthEventRefreshReuse = "refresh_reuse"
	AuthEventLogout       = "logout"
	AuthEventLogoutAll    = "logout_all"
)

type User struct {
	ID           string
	Username     string
	PasswordHash string
	DisplayName  string
	Role         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}

type UserSummary struct {
	ID          string
	Username    string
	DisplayName string
	Role        string
}

type UserSession struct {
	ID                  string
	UserID              string
	RefreshTokenHash    string
	TokenFamily         string
	UserAgent           string
	IPAddress           string
	ExpiresAt           time.Time
	RevokedAt           *time.Time
	LastUsedAt          *time.Time
	ReplacedBySessionID *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type AuthEvent struct {
	ID        string
	UserID    *string
	SessionID *string
	EventType string
	IPAddress string
	UserAgent string
	Metadata  map[string]any
	CreatedAt time.Time
}

func (u User) Summary() UserSummary {
	return UserSummary{
		ID:          strings.TrimSpace(u.ID),
		Username:    strings.TrimSpace(u.Username),
		DisplayName: strings.TrimSpace(u.DisplayName),
		Role:        strings.TrimSpace(u.Role),
	}
}

func (u User) IsActive() bool {
	return NormalizeUserStatus(u.Status) == UserStatusActive
}

func NormalizeUserStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}
