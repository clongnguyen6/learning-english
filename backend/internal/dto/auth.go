package dto

import "time"

type AuthLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type AuthSessionResponse struct {
	AccessToken      string    `json:"access_token"`
	TokenType        string    `json:"token_type"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at,omitempty"`
	User             AuthUser  `json:"user"`
}

type AuthMeResponse struct {
	User AuthUser `json:"user"`
}

type AuthLogoutResponse struct {
	Revoked bool `json:"revoked"`
}

type AuthLogoutAllResponse struct {
	RevokedSessions int64 `json:"revoked_sessions"`
}
