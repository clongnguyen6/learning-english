# Backend Scaffold

This directory is reserved for the Go API application described in `PLAN.md`.

Current scope:
- `cmd/api/` for the process entrypoint
- `internal/` for application packages and architecture boundaries
- `migrations/` for versioned SQL migrations
- `docs/` for backend-facing generated or maintained docs

Config contract:
- the backend will consume the root `.env.example` contract rather than inventing a second naming scheme
- required runtime vars should fail fast at startup once `learning-english-2a6.1.6` lands
- likely first-class backend config includes database connectivity, HTTP bind address, auth/token settings, CORS, app URLs, and storage settings

This bead only establishes the repository shape. Real backend shell work belongs to `learning-english-2a6.1.6`.
