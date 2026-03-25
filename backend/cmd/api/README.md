# API Entrypoint

`cmd/api/` is the home for the backend process bootstrap and server startup code.

Current phase-0 scope:
- `main.go` loads validated runtime config and starts the minimal HTTP server
- graceful shutdown is wired so later DB, jobs, and background services can hook into the same process lifecycle
- the entrypoint now constructs the first logger, database, and router seams for the backend shell

Future beads extend this entrypoint with:
- database wiring
- auth and learner/admin route groups
- real database connectivity, jobs, and worker bootstrapping

Current runtime foundation now includes:
- request-id propagation
- recovery, logging, and CORS middleware
- `/api/v1`, `/api/v1/healthz`, and `/api/v1/readyz`
- structured JSON success/error envelopes documented in `backend/docs/API_RUNTIME_CONTRACT.md`
