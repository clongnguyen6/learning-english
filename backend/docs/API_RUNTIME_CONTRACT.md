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

Current phase-0 readiness only checks the configured database handle seam. Later beads can extend readiness with more explicit dependency checks without changing the route contract.
