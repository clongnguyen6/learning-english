# API Runtime Contract

The canonical backend runtime contract now lives at [`backend/docs/API_RUNTIME_CONTRACT.md`](/Users/long/repo/personal/learning-english/backend/docs/API_RUNTIME_CONTRACT.md).

This top-level file remains as a pointer because project onboarding starts in the root docs before most contributors drill into backend-specific material.

Use the canonical backend doc for:
- response envelope and error envelope rules
- request ID propagation semantics
- route partitioning under `/api/v1`
- idempotency and optimistic concurrency expectations
- `GET /api/v1`, `GET /api/v1/healthz`, and `GET /api/v1/readyz`
