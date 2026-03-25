package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"learning-english/backend/internal/dto"
	"learning-english/backend/internal/models"
	"learning-english/backend/internal/repositories"
	"learning-english/backend/internal/services"
	"learning-english/backend/internal/utils"
)

type AuthHTTPService interface {
	Login(ctx context.Context, input services.LoginInput) (services.LoginResult, error)
	Refresh(ctx context.Context, input services.RefreshInput) (services.RefreshResult, error)
	Me(ctx context.Context, userID string) (models.UserSummary, error)
	Logout(ctx context.Context, input services.LogoutInput) (services.LogoutResult, error)
	LogoutAll(ctx context.Context, input services.LogoutAllInput) (services.LogoutAllResult, error)
}

type AuthHandlerConfig struct {
	Service       AuthHTTPService
	RefreshCookie utils.RefreshCookieConfig
}

type AuthHandler struct {
	service       AuthHTTPService
	refreshCookie utils.RefreshCookieConfig
}

func NewAuthHandler(cfg AuthHandlerConfig) (*AuthHandler, error) {
	if cfg.Service == nil {
		return nil, errors.New("auth service is required")
	}
	if _, err := utils.ExpireRefreshCookie(cfg.RefreshCookie); err != nil {
		return nil, err
	}

	return &AuthHandler{
		service:       cfg.Service,
		refreshCookie: cfg.RefreshCookie,
	}, nil
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	preventAuthCaching(w)

	var request dto.AuthLoginRequest
	if err := decodeJSONBody(r, &request); err != nil {
		writeValidationError(w, r, "invalid request body", nil)
		return
	}

	details := map[string]string{}
	if strings.TrimSpace(request.Username) == "" {
		details["username"] = "required"
	}
	if strings.TrimSpace(request.Password) == "" {
		details["password"] = "required"
	}
	if len(details) > 0 {
		writeValidationError(w, r, "invalid request body", details)
		return
	}

	result, err := h.service.Login(r.Context(), services.LoginInput{
		Username:  request.Username,
		Password:  request.Password,
		UserAgent: r.UserAgent(),
		IPAddress: clientIPAddress(r),
	})
	if err != nil {
		writeLoginError(w, r, err)
		return
	}

	if err := h.setRefreshCookie(w, result.RefreshToken, result.RefreshExpiresAt); err != nil {
		writeInternalServerError(w, r)
		return
	}

	WriteSuccess(w, http.StatusOK, dto.AuthSessionResponse{
		AccessToken:      result.AccessToken,
		TokenType:        "Bearer",
		ExpiresAt:        accessTokenExpiresAt(result.AccessTokenClaims),
		RefreshExpiresAt: result.RefreshExpiresAt.UTC(),
		User:             authUserDTO(result.User),
	}, requestMeta(r))
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	preventAuthCaching(w)

	result, err := h.service.Refresh(r.Context(), services.RefreshInput{
		RefreshToken: h.refreshTokenFromRequest(r),
		UserAgent:    r.UserAgent(),
		IPAddress:    clientIPAddress(r),
	})
	if err != nil {
		writeRefreshError(w, r, err)
		return
	}

	if err := h.setRefreshCookie(w, result.RefreshToken, result.RefreshExpiresAt); err != nil {
		writeInternalServerError(w, r)
		return
	}

	WriteSuccess(w, http.StatusOK, dto.AuthSessionResponse{
		AccessToken:      result.AccessToken,
		TokenType:        "Bearer",
		ExpiresAt:        accessTokenExpiresAt(result.AccessTokenClaims),
		RefreshExpiresAt: result.RefreshExpiresAt.UTC(),
		User:             authUserDTO(result.User),
	}, requestMeta(r))
}

func (h *AuthHandler) MeForUser(w http.ResponseWriter, r *http.Request, userID string) {
	preventAuthCaching(w)

	userID = strings.TrimSpace(userID)
	if userID == "" {
		writeUnauthorized(w, r, "authentication is required", nil)
		return
	}

	summary, err := h.service.Me(r.Context(), userID)
	if err != nil {
		writeAuthenticatedUserError(w, r, err)
		return
	}

	WriteSuccess(w, http.StatusOK, dto.AuthMeResponse{
		User: authUserDTO(summary),
	}, requestMeta(r))
}

func (h *AuthHandler) LogoutForUser(w http.ResponseWriter, r *http.Request, userID string) {
	preventAuthCaching(w)

	userID = strings.TrimSpace(userID)
	if userID == "" {
		writeUnauthorized(w, r, "authentication is required", nil)
		return
	}

	result, err := h.service.Logout(r.Context(), services.LogoutInput{
		UserID:       userID,
		RefreshToken: h.refreshTokenFromRequest(r),
		UserAgent:    r.UserAgent(),
		IPAddress:    clientIPAddress(r),
	})
	if err != nil {
		writeAuthenticatedUserError(w, r, err)
		return
	}

	if err := h.clearRefreshCookie(w); err != nil {
		writeInternalServerError(w, r)
		return
	}

	WriteSuccess(w, http.StatusOK, dto.AuthLogoutResponse{
		Revoked: result.Revoked,
	}, requestMeta(r))
}

func (h *AuthHandler) LogoutAllForUser(w http.ResponseWriter, r *http.Request, userID string) {
	preventAuthCaching(w)

	userID = strings.TrimSpace(userID)
	if userID == "" {
		writeUnauthorized(w, r, "authentication is required", nil)
		return
	}

	result, err := h.service.LogoutAll(r.Context(), services.LogoutAllInput{
		UserID:    userID,
		UserAgent: r.UserAgent(),
		IPAddress: clientIPAddress(r),
	})
	if err != nil {
		writeAuthenticatedUserError(w, r, err)
		return
	}

	if err := h.clearRefreshCookie(w); err != nil {
		writeInternalServerError(w, r)
		return
	}

	WriteSuccess(w, http.StatusOK, dto.AuthLogoutAllResponse{
		RevokedSessions: result.RevokedSessions,
	}, requestMeta(r))
}

func (h *AuthHandler) refreshTokenFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(h.refreshCookie.Name)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(cookie.Value)
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, value string, expiresAt time.Time) error {
	cookie, err := utils.BuildRefreshCookie(h.refreshCookie, value, expiresAt)
	if err != nil {
		return fmt.Errorf("build refresh cookie: %w", err)
	}

	http.SetCookie(w, cookie)
	return nil
}

func (h *AuthHandler) clearRefreshCookie(w http.ResponseWriter) error {
	cookie, err := utils.ExpireRefreshCookie(h.refreshCookie)
	if err != nil {
		return fmt.Errorf("expire refresh cookie: %w", err)
	}

	http.SetCookie(w, cookie)
	return nil
}

func decodeJSONBody(r *http.Request, destination any) error {
	if r.Body == nil {
		return io.EOF
	}

	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain a single JSON object")
		}
		return err
	}

	return nil
}

func writeLoginError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidCredentials), errors.Is(err, services.ErrUserInactive):
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid username or password", nil, requestMeta(r))
	default:
		writeInternalServerError(w, r)
	}
}

func writeRefreshError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, services.ErrSessionRevoked):
		WriteError(w, http.StatusUnauthorized, "SESSION_REVOKED", "refresh session is revoked or expired", nil, requestMeta(r))
	default:
		writeInternalServerError(w, r)
	}
}

func writeAuthenticatedUserError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, services.ErrUserInactive), errors.Is(err, services.ErrUserIDBlank), errors.Is(err, repositories.ErrNotFound):
		writeUnauthorized(w, r, "authentication is required", nil)
	default:
		writeInternalServerError(w, r)
	}
}

func writeValidationError(w http.ResponseWriter, r *http.Request, message string, details any) {
	WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", message, details, requestMeta(r))
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request, message string, details any) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", message, details, requestMeta(r))
}

func writeInternalServerError(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "unexpected server error", nil, requestMeta(r))
}

func preventAuthCaching(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
}

func accessTokenExpiresAt(claims utils.AccessTokenClaims) time.Time {
	if claims.ExpiresAt == nil {
		return time.Time{}
	}

	return claims.ExpiresAt.Time.UTC()
}

func authUserDTO(user models.UserSummary) dto.AuthUser {
	return dto.AuthUser{
		ID:          strings.TrimSpace(user.ID),
		Username:    strings.TrimSpace(user.Username),
		DisplayName: strings.TrimSpace(user.DisplayName),
		Role:        strings.TrimSpace(user.Role),
	}
}

func clientIPAddress(r *http.Request) string {
	forwardedFor := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
	if forwardedFor != "" {
		return forwardedFor
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return strings.TrimSpace(host)
	}

	return strings.TrimSpace(r.RemoteAddr)
}
