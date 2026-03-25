# API Runtime Contract

This document locks the shared backend runtime contract before feature handlers multiply.

It is intentionally narrow. The goal is to make the cross-cutting rules explicit now so auth, learner APIs, admin APIs, OpenAPI generation, and frontend integration all build on the same foundation.

## Scope

- stable API base path: `/api/v1`
- learner endpoints live under `/api/v1/...`
- admin endpoints live under `/api/v1/admin/...`
- current phase-0 runtime foundation includes:
  - request ID propagation
  - structured JSON success/error envelopes
  - panic recovery
  - request logging
  - CORS handling
  - `GET /api/v1`
  - `GET /api/v1/healthz`
  - `GET /api/v1/readyz`
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/refresh`
  - `GET /api/v1/auth/me`
  - `POST /api/v1/auth/logout`
  - `POST /api/v1/auth/logout-all`

Phase-0 startup validation only hard-fails on configuration used by the current shell and readiness path. Future auth/storage beads should tighten secrets validation when those subsystems become active at runtime.

## Response Envelope

Success responses use:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": {
    "request_id": "0d9f..."
  }
}
```

Error responses use:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "invalid request body",
    "details": {
      "field": "required"
    }
  },
  "meta": {
    "request_id": "0d9f..."
  }
}
```

Contract rules:
- every API response returns JSON
- `meta.request_id` is included whenever a request reaches the HTTP runtime
- later feature handlers should reuse this envelope instead of inventing one-off shapes

## Error Codes

Current foundation/runtime codes:
- `NOT_FOUND`
- `METHOD_NOT_ALLOWED`
- `INTERNAL_SERVER_ERROR`
- `SERVICE_UNAVAILABLE`

Reserved application-level codes from `PLAN.md` that later beads should reuse:
- `UNAUTHORIZED`
- `FORBIDDEN`
- `VALIDATION_ERROR`
- `CONFLICT`
- `PRECONDITION_FAILED`
- `IDEMPOTENCY_CONFLICT`
- `RATE_LIMITED`
- `IMPORT_PARSE_ERROR`
- `IMPORT_VALIDATION_ERROR`
- `SESSION_REVOKED`

## Request IDs

- inbound request ID header name comes from `REQUEST_ID_HEADER`
- if the client sends a request ID, the backend reuses it
- if the client does not send one, the backend generates one
- the same request ID is echoed back in the response header and response body metadata
- logs must include the request ID so API failures can be correlated quickly

## Idempotency and Concurrency Expectations

These are foundation rules even before the feature handlers exist:

- retry-sensitive writes must accept `Idempotency-Key`
- mutable resources should expose `version` or `ETag`
- write endpoints with concurrency risk should use precondition checks such as `If-Match` or explicit `base_version`
- learner content reads for published projections should later expose cache validators such as `ETag` or `Last-Modified`

This bead does not implement those feature-specific write paths yet. It locks the runtime expectation so later handlers do not improvise incompatible behavior.

## Health Endpoints

`GET /api/v1/healthz`
- process-level liveness
- returns `200 OK` when the HTTP runtime is up

`GET /api/v1/readyz`
- dependency readiness
- returns `200 OK` when configured runtime dependencies are ready
- returns `503 Service Unavailable` with a structured error envelope when a required dependency is not ready

Current phase-0 readiness now pings the configured database connection. `200 OK` means the backend runtime can reach its database dependency; `503 Service Unavailable` means the runtime is up but the database is not reachable yet.

## Auth Endpoints

Current auth transport rules:
- access tokens are returned in the JSON response body and must be sent back as `Authorization: Bearer <token>`
- refresh tokens are never returned in the JSON response body
- refresh tokens are issued only via the configured HttpOnly cookie from `REFRESH_COOKIE_NAME`
- the refresh cookie always uses `HttpOnly`, `Path=/`, and the configured `Domain`, `Secure`, and `SameSite` settings
- auth responses should be treated as non-cacheable; the current handlers set `Cache-Control: no-store`

### `POST /api/v1/auth/login`

Request body:

```json
{
  "username": "long",
  "password": "secret123"
}
```

Success response:

```json
{
  "success": true,
  "data": {
    "access_token": "jwt",
    "token_type": "Bearer",
    "expires_at": "2026-03-26T05:15:00Z",
    "refresh_expires_at": "2026-04-25T04:00:00Z",
    "user": {
      "id": "uuid",
      "username": "long",
      "display_name": "Long",
      "role": "learner"
    }
  },
  "error": null,
  "meta": {
    "request_id": "0d9f..."
  }
}
```

Behavior:
- sets the refresh cookie on success
- returns `400 VALIDATION_ERROR` when `username` or `password` is blank
- returns `401 UNAUTHORIZED` for invalid username/password combinations

### `POST /api/v1/auth/refresh`

Behavior:
- reads the refresh token from the configured cookie
- rotates the refresh token on success and returns a new access token payload with the same shape as login
- returns `401 SESSION_REVOKED` when the refresh cookie is missing, expired, revoked, or flagged as reused

### `GET /api/v1/auth/me`

Behavior:
- requires a valid bearer access token
- returns the authenticated user summary in:

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "username": "long",
      "display_name": "Long",
      "role": "learner"
    }
  },
  "error": null,
  "meta": {
    "request_id": "0d9f..."
  }
}
```

### `POST /api/v1/auth/logout`

Behavior:
- requires a valid bearer access token for the current user
- revokes the current refresh session when the refresh cookie is present
- clears the refresh cookie even when the current device session is already absent
- returns:

```json
{
  "success": true,
  "data": {
    "revoked": true
  },
  "error": null,
  "meta": {
    "request_id": "0d9f..."
  }
}
```

### `POST /api/v1/auth/logout-all`

Behavior:
- requires a valid bearer access token for the current user
- revokes all sessions for that user
- clears the refresh cookie for the current browser
- returns:

```json
{
  "success": true,
  "data": {
    "revoked_sessions": 3
  },
  "error": null,
  "meta": {
    "request_id": "0d9f..."
  }
}
```

### Auth Error Mapping

- `400 VALIDATION_ERROR`: malformed JSON or missing required login fields
- `401 UNAUTHORIZED`: invalid credentials or missing/invalid bearer token on protected auth routes
- `401 SESSION_REVOKED`: missing, expired, revoked, or reuse-detected refresh session
- `500 INTERNAL_SERVER_ERROR`: unexpected repository/runtime failure while servicing auth requests
