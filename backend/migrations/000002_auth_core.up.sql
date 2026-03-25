CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username varchar NOT NULL,
    password_hash varchar NOT NULL,
    display_name varchar,
    role varchar NOT NULL DEFAULT 'learner',
    status varchar NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    last_login_at timestamptz,
    CONSTRAINT users_username_not_blank CHECK (btrim(username) <> ''),
    CONSTRAINT users_password_hash_not_blank CHECK (btrim(password_hash) <> ''),
    CONSTRAINT users_role_valid CHECK (role IN ('learner', 'admin'))
);

CREATE UNIQUE INDEX users_username_key ON users (username);

CREATE TABLE user_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    refresh_token_hash varchar NOT NULL,
    token_family varchar NOT NULL,
    user_agent text,
    ip_address inet,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    last_used_at timestamptz,
    replaced_by_session_id uuid REFERENCES user_sessions (id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_sessions_refresh_token_hash_not_blank CHECK (btrim(refresh_token_hash) <> ''),
    CONSTRAINT user_sessions_token_family_not_blank CHECK (btrim(token_family) <> ''),
    CONSTRAINT user_sessions_expires_after_create CHECK (expires_at >= created_at),
    CONSTRAINT user_sessions_revoked_after_create CHECK (revoked_at IS NULL OR revoked_at >= created_at),
    CONSTRAINT user_sessions_last_used_after_create CHECK (last_used_at IS NULL OR last_used_at >= created_at),
    CONSTRAINT user_sessions_not_self_replaced CHECK (replaced_by_session_id IS NULL OR replaced_by_session_id <> id)
);

CREATE UNIQUE INDEX user_sessions_refresh_token_hash_key ON user_sessions (refresh_token_hash);
CREATE INDEX idx_user_sessions_user_active ON user_sessions (user_id, expires_at DESC) WHERE revoked_at IS NULL;
CREATE INDEX idx_user_sessions_token_family ON user_sessions (token_family);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions (expires_at);

CREATE TABLE auth_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES users (id) ON DELETE SET NULL,
    session_id uuid REFERENCES user_sessions (id) ON DELETE SET NULL,
    event_type varchar NOT NULL,
    ip_address inet,
    user_agent text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT auth_events_event_type_valid CHECK (
        event_type IN ('login_success', 'login_failed', 'refresh_reuse', 'logout', 'logout_all')
    )
);

CREATE INDEX idx_auth_events_user_created_at ON auth_events (user_id, created_at DESC);
CREATE INDEX idx_auth_events_session_created_at ON auth_events (session_id, created_at DESC);
CREATE INDEX idx_auth_events_event_type_created_at ON auth_events (event_type, created_at DESC);
