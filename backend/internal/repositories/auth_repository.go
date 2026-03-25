package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"learning-english/backend/internal/database"
	"learning-english/backend/internal/models"
)

var ErrNotFound = errors.New("repository record not found")

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (models.User, error)
	FindByID(ctx context.Context, userID string) (models.User, error)
	UpdateLastLoginAt(ctx context.Context, userID string, lastLoginAt time.Time) error
}

type SessionRepository interface {
	Create(ctx context.Context, params CreateUserSessionParams) (models.UserSession, error)
	RevokeByRefreshTokenHash(ctx context.Context, userID, refreshTokenHash string, revokedAt time.Time) (models.UserSession, error)
	RevokeAllByUserID(ctx context.Context, userID string, revokedAt time.Time) (int64, error)
}

type AuthEventRepository interface {
	Create(ctx context.Context, params CreateAuthEventParams) (models.AuthEvent, error)
}

type AuthRepositorySet struct {
	Users      UserRepository
	Sessions   SessionRepository
	AuthEvents AuthEventRepository
}

type AuthRepositoryProvider interface {
	ReadOnly() AuthRepositorySet
	WithinTx(ctx context.Context, fn func(AuthRepositorySet) error) error
}

type CreateUserSessionParams struct {
	UserID           string
	RefreshTokenHash string
	TokenFamily      string
	UserAgent        string
	IPAddress        string
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

type CreateAuthEventParams struct {
	UserID    *string
	SessionID *string
	EventType string
	IPAddress string
	UserAgent string
	Metadata  map[string]any
	CreatedAt time.Time
}

type SQLAuthRepositoryProvider struct {
	handle *database.Handle
}

type sqlUserRepository struct {
	exec database.DBTX
}

type sqlSessionRepository struct {
	exec database.DBTX
}

type sqlAuthEventRepository struct {
	exec database.DBTX
}

func NewSQLAuthRepositoryProvider(handle *database.Handle) *SQLAuthRepositoryProvider {
	return &SQLAuthRepositoryProvider{handle: handle}
}

func NewAuthRepositorySet(exec database.DBTX) AuthRepositorySet {
	return AuthRepositorySet{
		Users:      NewUserRepository(exec),
		Sessions:   NewSessionRepository(exec),
		AuthEvents: NewAuthEventRepository(exec),
	}
}

func NewUserRepository(exec database.DBTX) UserRepository {
	return &sqlUserRepository{exec: exec}
}

func NewSessionRepository(exec database.DBTX) SessionRepository {
	return &sqlSessionRepository{exec: exec}
}

func NewAuthEventRepository(exec database.DBTX) AuthEventRepository {
	return &sqlAuthEventRepository{exec: exec}
}

func (p *SQLAuthRepositoryProvider) ReadOnly() AuthRepositorySet {
	if p == nil || p.handle == nil {
		return AuthRepositorySet{}
	}

	return NewAuthRepositorySet(p.handle.DB())
}

func (p *SQLAuthRepositoryProvider) WithinTx(ctx context.Context, fn func(AuthRepositorySet) error) error {
	if p == nil || p.handle == nil {
		return errors.New("auth repository provider is not configured")
	}
	if fn == nil {
		return errors.New("auth repository callback is required")
	}

	return p.handle.WithTx(ctx, func(tx *sql.Tx) error {
		return fn(NewAuthRepositorySet(tx))
	})
}

func (r *sqlUserRepository) FindByUsername(ctx context.Context, username string) (models.User, error) {
	if r == nil || r.exec == nil {
		return models.User{}, errors.New("user repository is not configured")
	}

	username = strings.TrimSpace(username)
	row := r.exec.QueryRowContext(ctx, `
		SELECT
			id,
			username,
			password_hash,
			COALESCE(display_name, ''),
			role,
			COALESCE(status, ''),
			created_at,
			updated_at,
			last_login_at
		FROM users
		WHERE username = $1
		LIMIT 1
	`, username)

	user, err := scanUser(row)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *sqlUserRepository) FindByID(ctx context.Context, userID string) (models.User, error) {
	if r == nil || r.exec == nil {
		return models.User{}, errors.New("user repository is not configured")
	}

	userID = strings.TrimSpace(userID)
	row := r.exec.QueryRowContext(ctx, `
		SELECT
			id,
			username,
			password_hash,
			COALESCE(display_name, ''),
			role,
			COALESCE(status, ''),
			created_at,
			updated_at,
			last_login_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`, userID)

	user, err := scanUser(row)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *sqlUserRepository) UpdateLastLoginAt(ctx context.Context, userID string, lastLoginAt time.Time) error {
	if r == nil || r.exec == nil {
		return errors.New("user repository is not configured")
	}

	result, err := r.exec.ExecContext(ctx, `
		UPDATE users
		SET last_login_at = $2, updated_at = $2
		WHERE id = $1
	`, strings.TrimSpace(userID), lastLoginAt.UTC())
	if err != nil {
		return fmt.Errorf("update user last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read user last login update result: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *sqlSessionRepository) Create(ctx context.Context, params CreateUserSessionParams) (models.UserSession, error) {
	if r == nil || r.exec == nil {
		return models.UserSession{}, errors.New("session repository is not configured")
	}

	createdAt := params.CreatedAt.UTC()
	row := r.exec.QueryRowContext(ctx, `
		INSERT INTO user_sessions (
			user_id,
			refresh_token_hash,
			token_family,
			user_agent,
			ip_address,
			expires_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING
			id,
			user_id,
			refresh_token_hash,
			token_family,
			COALESCE(user_agent, ''),
			COALESCE(host(ip_address), ''),
			expires_at,
			revoked_at,
			last_used_at,
			replaced_by_session_id,
			created_at,
			updated_at
	`,
		strings.TrimSpace(params.UserID),
		strings.TrimSpace(params.RefreshTokenHash),
		strings.TrimSpace(params.TokenFamily),
		nullableTrimmedString(params.UserAgent),
		nullableTrimmedString(params.IPAddress),
		params.ExpiresAt.UTC(),
		createdAt,
	)

	session, err := scanSession(row)
	if err != nil {
		return models.UserSession{}, fmt.Errorf("create user session: %w", err)
	}

	return session, nil
}

func (r *sqlSessionRepository) RevokeByRefreshTokenHash(ctx context.Context, userID, refreshTokenHash string, revokedAt time.Time) (models.UserSession, error) {
	if r == nil || r.exec == nil {
		return models.UserSession{}, errors.New("session repository is not configured")
	}

	row := r.exec.QueryRowContext(ctx, `
		UPDATE user_sessions
		SET revoked_at = $3, updated_at = $3
		WHERE user_id = $1
		  AND refresh_token_hash = $2
		  AND revoked_at IS NULL
		RETURNING
			id,
			user_id,
			refresh_token_hash,
			token_family,
			COALESCE(user_agent, ''),
			COALESCE(host(ip_address), ''),
			expires_at,
			revoked_at,
			last_used_at,
			replaced_by_session_id,
			created_at,
			updated_at
	`,
		strings.TrimSpace(userID),
		strings.TrimSpace(refreshTokenHash),
		revokedAt.UTC(),
	)

	session, err := scanSession(row)
	if err != nil {
		return models.UserSession{}, err
	}

	return session, nil
}

func (r *sqlSessionRepository) RevokeAllByUserID(ctx context.Context, userID string, revokedAt time.Time) (int64, error) {
	if r == nil || r.exec == nil {
		return 0, errors.New("session repository is not configured")
	}

	result, err := r.exec.ExecContext(ctx, `
		UPDATE user_sessions
		SET revoked_at = $2, updated_at = $2
		WHERE user_id = $1
		  AND revoked_at IS NULL
	`, strings.TrimSpace(userID), revokedAt.UTC())
	if err != nil {
		return 0, fmt.Errorf("revoke user sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read revoked session count: %w", err)
	}

	return rowsAffected, nil
}

func (r *sqlAuthEventRepository) Create(ctx context.Context, params CreateAuthEventParams) (models.AuthEvent, error) {
	if r == nil || r.exec == nil {
		return models.AuthEvent{}, errors.New("auth event repository is not configured")
	}

	metadata := params.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return models.AuthEvent{}, fmt.Errorf("marshal auth event metadata: %w", err)
	}

	createdAt := params.CreatedAt.UTC()
	row := r.exec.QueryRowContext(ctx, `
		INSERT INTO auth_events (
			user_id,
			session_id,
			event_type,
			ip_address,
			user_agent,
			metadata,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)
		RETURNING
			id,
			user_id,
			session_id,
			event_type,
			COALESCE(host(ip_address), ''),
			COALESCE(user_agent, ''),
			metadata,
			created_at
	`,
		nullableStringPointer(params.UserID),
		nullableStringPointer(params.SessionID),
		strings.TrimSpace(params.EventType),
		nullableTrimmedString(params.IPAddress),
		nullableTrimmedString(params.UserAgent),
		string(metadataJSON),
		createdAt,
	)

	event, err := scanAuthEvent(row)
	if err != nil {
		return models.AuthEvent{}, fmt.Errorf("create auth event: %w", err)
	}

	return event, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(scanner rowScanner) (models.User, error) {
	var (
		user        models.User
		lastLoginAt sql.NullTime
	)

	err := scanner.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.DisplayName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrNotFound
		}

		return models.User{}, fmt.Errorf("scan user: %w", err)
	}

	if lastLoginAt.Valid {
		t := lastLoginAt.Time.UTC()
		user.LastLoginAt = &t
	}

	return user, nil
}

func scanSession(scanner rowScanner) (models.UserSession, error) {
	var (
		session             models.UserSession
		revokedAt           sql.NullTime
		lastUsedAt          sql.NullTime
		replacedBySessionID sql.NullString
	)

	err := scanner.Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.TokenFamily,
		&session.UserAgent,
		&session.IPAddress,
		&session.ExpiresAt,
		&revokedAt,
		&lastUsedAt,
		&replacedBySessionID,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.UserSession{}, ErrNotFound
		}

		return models.UserSession{}, fmt.Errorf("scan session: %w", err)
	}

	if revokedAt.Valid {
		t := revokedAt.Time.UTC()
		session.RevokedAt = &t
	}
	if lastUsedAt.Valid {
		t := lastUsedAt.Time.UTC()
		session.LastUsedAt = &t
	}
	if replacedBySessionID.Valid {
		value := strings.TrimSpace(replacedBySessionID.String)
		session.ReplacedBySessionID = &value
	}

	return session, nil
}

func scanAuthEvent(scanner rowScanner) (models.AuthEvent, error) {
	var (
		event       models.AuthEvent
		userID      sql.NullString
		sessionID   sql.NullString
		metadataRaw []byte
	)

	err := scanner.Scan(
		&event.ID,
		&userID,
		&sessionID,
		&event.EventType,
		&event.IPAddress,
		&event.UserAgent,
		&metadataRaw,
		&event.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.AuthEvent{}, ErrNotFound
		}

		return models.AuthEvent{}, fmt.Errorf("scan auth event: %w", err)
	}

	if userID.Valid {
		value := strings.TrimSpace(userID.String)
		event.UserID = &value
	}
	if sessionID.Valid {
		value := strings.TrimSpace(sessionID.String)
		event.SessionID = &value
	}

	event.Metadata = map[string]any{}
	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &event.Metadata); err != nil {
			return models.AuthEvent{}, fmt.Errorf("unmarshal auth event metadata: %w", err)
		}
	}

	return event, nil
}

func nullableTrimmedString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullableStringPointer(value *string) any {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return trimmed
}
