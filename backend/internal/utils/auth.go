package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"learning-english/backend/internal/policies"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordBlank        = errors.New("password must not be blank")
	ErrPasswordHashBlank    = errors.New("password hash must not be blank")
	ErrTokenBlank           = errors.New("token must not be blank")
	ErrTokenSubjectBlank    = errors.New("token subject must not be blank")
	ErrTokenRoleBlank       = errors.New("token role must not be blank")
	ErrTokenRoleInvalid     = errors.New("token role is invalid")
	ErrTokenExpired         = errors.New("token is expired")
	ErrTokenIssuerInvalid   = errors.New("token issuer is invalid")
	ErrTokenAudienceInvalid = errors.New("token audience is invalid")
)

type TokenManagerConfig struct {
	Issuer      string
	Audience    string
	HS256Secret string
	TTL         time.Duration
}

type TokenManager struct {
	issuer   string
	audience string
	secret   []byte
	ttl      time.Duration
	now      func() time.Time
}

type AccessTokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type RefreshCookieConfig struct {
	Name     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
}

func HashPassword(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", ErrPasswordBlank
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hashed), nil
}

func VerifyPassword(passwordHash, password string) error {
	if strings.TrimSpace(passwordHash) == "" {
		return ErrPasswordHashBlank
	}

	if strings.TrimSpace(password) == "" {
		return ErrPasswordBlank
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return fmt.Errorf("verify password: %w", err)
	}

	return nil
}

func NewTokenManager(cfg TokenManagerConfig) (*TokenManager, error) {
	if strings.TrimSpace(cfg.Issuer) == "" {
		return nil, fmt.Errorf("issuer is required")
	}

	if strings.TrimSpace(cfg.Audience) == "" {
		return nil, fmt.Errorf("audience is required")
	}

	if strings.TrimSpace(cfg.HS256Secret) == "" {
		return nil, fmt.Errorf("hs256 secret is required")
	}

	if cfg.TTL <= 0 {
		return nil, fmt.Errorf("ttl must be greater than zero")
	}

	return &TokenManager{
		issuer:   strings.TrimSpace(cfg.Issuer),
		audience: strings.TrimSpace(cfg.Audience),
		secret:   []byte(strings.TrimSpace(cfg.HS256Secret)),
		ttl:      cfg.TTL,
		now:      time.Now,
	}, nil
}

func (m *TokenManager) IssueAccessToken(subject, role string) (string, AccessTokenClaims, error) {
	if m == nil {
		return "", AccessTokenClaims{}, fmt.Errorf("token manager is not configured")
	}

	subject = strings.TrimSpace(subject)
	if subject == "" {
		return "", AccessTokenClaims{}, ErrTokenSubjectBlank
	}

	role = strings.TrimSpace(role)
	if role == "" {
		return "", AccessTokenClaims{}, ErrTokenRoleBlank
	}
	role = policies.NormalizeRole(role)
	if !policies.IsSupportedRole(role) {
		return "", AccessTokenClaims{}, ErrTokenRoleInvalid
	}

	now := m.now().UTC()
	claims := AccessTokenClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   subject,
			Audience:  jwt.ClaimStrings{m.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", AccessTokenClaims{}, fmt.Errorf("sign access token: %w", err)
	}

	return signed, claims, nil
}

func (m *TokenManager) ParseAccessToken(tokenString string) (AccessTokenClaims, error) {
	if m == nil {
		return AccessTokenClaims{}, fmt.Errorf("token manager is not configured")
	}

	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return AccessTokenClaims{}, ErrTokenBlank
	}

	claims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(parsed *jwt.Token) (any, error) {
		if parsed.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method %s", parsed.Method.Alg())
		}

		return m.secret, nil
	},
		jwt.WithTimeFunc(m.now),
		jwt.WithExpirationRequired(),
		jwt.WithAudience(m.audience),
		jwt.WithIssuer(m.issuer),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return AccessTokenClaims{}, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenInvalidAudience) {
			return AccessTokenClaims{}, ErrTokenAudienceInvalid
		}
		if errors.Is(err, jwt.ErrTokenInvalidIssuer) {
			return AccessTokenClaims{}, ErrTokenIssuerInvalid
		}

		return AccessTokenClaims{}, fmt.Errorf("parse access token: %w", err)
	}

	if !token.Valid {
		return AccessTokenClaims{}, fmt.Errorf("parse access token: token is invalid")
	}

	claims.Subject = strings.TrimSpace(claims.Subject)
	if claims.Subject == "" {
		return AccessTokenClaims{}, ErrTokenSubjectBlank
	}

	claims.Role = policies.NormalizeRole(claims.Role)
	if claims.Role == "" {
		return AccessTokenClaims{}, ErrTokenRoleBlank
	}
	if !policies.IsSupportedRole(claims.Role) {
		return AccessTokenClaims{}, ErrTokenRoleInvalid
	}

	return *claims, nil
}

func BuildRefreshCookie(cfg RefreshCookieConfig, value string, expiresAt time.Time) (*http.Cookie, error) {
	if strings.TrimSpace(cfg.Name) == "" {
		return nil, fmt.Errorf("refresh cookie name is required")
	}

	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("refresh cookie value is required")
	}

	if expiresAt.IsZero() {
		return nil, fmt.Errorf("refresh cookie expiry is required")
	}

	return &http.Cookie{
		Name:     cfg.Name,
		Value:    value,
		Path:     "/",
		Domain:   strings.TrimSpace(cfg.Domain),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
		Expires:  expiresAt.UTC(),
		MaxAge:   maxAgeUntil(expiresAt.UTC(), time.Now().UTC()),
	}, nil
}

func ExpireRefreshCookie(cfg RefreshCookieConfig) (*http.Cookie, error) {
	if strings.TrimSpace(cfg.Name) == "" {
		return nil, fmt.Errorf("refresh cookie name is required")
	}

	return &http.Cookie{
		Name:     cfg.Name,
		Value:    "",
		Path:     "/",
		Domain:   strings.TrimSpace(cfg.Domain),
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
		Expires:  time.Unix(0, 0).UTC(),
		MaxAge:   -1,
	}, nil
}

func containsAudience(audiences []string, expected string) bool {
	for _, audience := range audiences {
		if strings.TrimSpace(audience) == expected {
			return true
		}
	}

	return false
}

func maxAgeUntil(expiresAt, now time.Time) int {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	seconds := int(expiresAt.Sub(now).Seconds())

	if seconds < 1 {
		return 1
	}

	return seconds
}
